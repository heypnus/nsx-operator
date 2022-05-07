/* Copyright © 2021 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: Apache-2.0 */

package main

import (
	"context"
	"flag"
	"os"
	"time"

	zapu "go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/nsx-operator/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/nsx-operator/pkg/config"
	"github.com/vmware-tanzu/nsx-operator/pkg/controllers"
	"github.com/vmware-tanzu/nsx-operator/pkg/metrics"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var probeAddr, metricsAddr string
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8384", "The address the probe endpoint binds to.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8093", "The address the metrics endpoint binds to.")
	config.AddFlags()
	flag.Parse()

	cf, err := config.NewNSXOperatorConfigFromFile()
	if err != nil {
		os.Exit(1)
	}

	zpLevel := zapu.InfoLevel
	if cf.Debug == true {
		zpLevel = zapu.DebugLevel
	}

	opts := zap.Options{
		Development: true,
		Level:       zapu.NewAtomicLevelAt(zpLevel),
	}
	opts.BindFlags(flag.CommandLine)
	logf.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("starting NSX Operator")
	if metrics.AreMetricsExposed(cf) {
		metrics.InitializePrometheusMetrics()
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		MetricsBindAddress:     metricsAddr,
		LeaderElectionID:       "nsx-operator",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	securityReconcile := &controllers.SecurityPolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	nsxClient := nsx.GetClient(cf)
	if nsxClient == nil {
		setupLog.Error(err, "unable to get nsx client")
		os.Exit(1)
	}
	if service, err := services.InitializeSecurityPolicy(nsxClient, cf); err != nil {
		setupLog.Error(err, "unable to init securitypolicy service", "controller", "SecurityPolicy")
		os.Exit(1)
	} else {
		securityReconcile.Service = service
	}

	if err = securityReconcile.Start(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecurityPolicy")
		os.Exit(1)
	}

	if metrics.AreMetricsExposed(cf) {
		go updateHealthMetricsPeriodically(nsxClient)
	}

	if err := mgr.AddHealthzCheck("healthz", nsxClient.NSXChecker.CheckNSXHealth); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("rendering manifests")
	client := mgr.GetClient()
	decoder := serializer.NewCodecFactory(clientgoscheme.Scheme).UniversalDecoder()
	prometheusService := &v1.Service{}
	svcYaml, err := os.Open("/manifests/prometheus/service.yaml")
	if err != nil {
		setupLog.Error(err, "unable to retrieve prometheus service yaml", prometheusService.GetName())
		os.Exit(1)
	}
	data := yaml.NewYAMLOrJSONDecoder(svcYaml, 4096)
	if err := runtime.DecodeInto(decoder, data, prometheusService); err != nil {
		setupLog.Error(err, "unable to retrieve prometheus service", prometheusService.GetName())
	}
	if err := client.Update(context.TODO(), prometheusService); err != nil {
		setupLog.Error(err, "unable to apply prometheus service", prometheusService.GetName())
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// Function for fetching nsx health status and feeding it to the prometheus metric.
func getHealthStatus(nsxClient *nsx.Client) error {
	status := 1
	if err := nsxClient.NSXChecker.CheckNSXHealth(nil); err != nil {
		status = 0
	}
	// Record the new health status in metric.
	metrics.NSXOperatorHealthStats.Set(float64(status))
	return nil
}

// Periodically fetches health info.
func updateHealthMetricsPeriodically(nsxClient *nsx.Client) {
	for {
		if err := getHealthStatus(nsxClient); err != nil {
			setupLog.Error(err, "Failed to fetch health info")
		}
		select {
		case <-time.After(metrics.ScrapeTimeout):
		}
	}
}
