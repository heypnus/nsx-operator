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

// NSXServiceAccountInformer provides access to a shared informer and lister for
// NSXServiceAccounts.
type NSXServiceAccountInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NSXServiceAccountLister
}

type nSXServiceAccountInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNSXServiceAccountInformer constructs a new informer for NSXServiceAccount type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNSXServiceAccountInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNSXServiceAccountInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNSXServiceAccountInformer constructs a new informer for NSXServiceAccount type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNSXServiceAccountInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NsxV1alpha1().NSXServiceAccounts(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NsxV1alpha1().NSXServiceAccounts(namespace).Watch(context.TODO(), options)
			},
		},
		&nsxvmwarecomv1alpha1.NSXServiceAccount{},
		resyncPeriod,
		indexers,
	)
}

func (f *nSXServiceAccountInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNSXServiceAccountInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nSXServiceAccountInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&nsxvmwarecomv1alpha1.NSXServiceAccount{}, f.defaultInformer)
}

func (f *nSXServiceAccountInformer) Lister() v1alpha1.NSXServiceAccountLister {
	return v1alpha1.NewNSXServiceAccountLister(f.Informer().GetIndexer())
}
