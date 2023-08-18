/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: Apache-2.0 */

package node

import (
	"context"
	"os"

	"github.com/vmware-tanzu/nsx-operator/pkg/logger"
	"github.com/vmware-tanzu/nsx-operator/pkg/metrics"
	v1 "k8s.io/api/core/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/nsx-operator/pkg/controllers/common"
	servicecommon "github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/common"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/node"
)

var (
	log               = logger.Log
	MetricResTypeNode = common.MetricResTypeNode
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme  *apimachineryruntime.Scheme
	Service *node.NodeService
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	node := &v1.Node{}
	log.Info("reconciling node", "node", req.NamespacedName)

	metrics.CounterInc(r.Service.NSXConfig, metrics.ControllerSyncTotal, MetricResTypeNode)

	if err := r.Client.Get(ctx, req.NamespacedName, node); err != nil {
		log.Error(err, "unable to fetch node", "req", req.NamespacedName)
		return common.ResultNormal, client.IgnoreNotFound(err)
	}
	if err := r.Service.SyncNodeStore(node.Name); err != nil {
		log.Error(err, "failed to sync node store", "req", req.NamespacedName)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Node{}).
		Complete(r)
}

func StartNodeController(mgr ctrl.Manager, commonService servicecommon.Service) {
	nodePortReconciler := NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	if nodeService, err := node.InitializeNode(commonService); err != nil {
		log.Error(err, "failed to initialize node commonService", "controller", "Node")
		os.Exit(1)
	} else {
		nodePortReconciler.Service = nodeService
		common.ServiceMediator.NodeService = nodePortReconciler.Service
	}

	if err := nodePortReconciler.Start(mgr); err != nil {
		log.Error(err, "failed to create controller", "controller", "Node")
		os.Exit(1)
	}
}

func (r *NodeReconciler) Start(mgr ctrl.Manager) error {
	err := r.SetupWithManager(mgr)
	if err != nil {
		return err
	}
	return nil
}

func updateFail(r *NodeReconciler, c *context.Context, o *v1.Node, e *error) {
	metrics.CounterInc(r.Service.NSXConfig, metrics.ControllerUpdateFailTotal, MetricResTypeNode)
}

func deleteFail(r *NodeReconciler, c *context.Context, o *v1.Node, e *error) {
	metrics.CounterInc(r.Service.NSXConfig, metrics.ControllerDeleteFailTotal, MetricResTypeNode)
}

func updateSuccess(r *NodeReconciler, c *context.Context, o *v1.Node) {
	metrics.CounterInc(r.Service.NSXConfig, metrics.ControllerUpdateSuccessTotal, MetricResTypeNode)
}

func deleteSuccess(r *NodeReconciler, _ *context.Context, _ *v1.Node) {
	metrics.CounterInc(r.Service.NSXConfig, metrics.ControllerDeleteSuccessTotal, MetricResTypeNode)
}
