/* Copyright © 2023 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: Apache-2.0 */

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	nsxvmwarecomv1alpha1 "github.com/vmware-tanzu/nsx-operator/pkg/apis/nsx.vmware.com/v1alpha1"
	versioned "github.com/vmware-tanzu/nsx-operator/pkg/client/clientset/versioned"
	internalinterfaces "github.com/vmware-tanzu/nsx-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/vmware-tanzu/nsx-operator/pkg/client/listers/nsx.vmware.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SecurityPolicyInformer provides access to a shared informer and lister for
// SecurityPolicies.
type SecurityPolicyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.SecurityPolicyLister
}

type securityPolicyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewSecurityPolicyInformer constructs a new informer for SecurityPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSecurityPolicyInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSecurityPolicyInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredSecurityPolicyInformer constructs a new informer for SecurityPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSecurityPolicyInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NsxV1alpha1().SecurityPolicies(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NsxV1alpha1().SecurityPolicies(namespace).Watch(context.TODO(), options)
			},
		},
		&nsxvmwarecomv1alpha1.SecurityPolicy{},
		resyncPeriod,
		indexers,
	)
}

func (f *securityPolicyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSecurityPolicyInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *securityPolicyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&nsxvmwarecomv1alpha1.SecurityPolicy{}, f.defaultInformer)
}

func (f *securityPolicyInformer) Lister() v1alpha1.SecurityPolicyLister {
	return v1alpha1.NewSecurityPolicyLister(f.Informer().GetIndexer())
}
