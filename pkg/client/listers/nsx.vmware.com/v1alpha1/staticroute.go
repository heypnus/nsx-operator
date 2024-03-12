/* Copyright © 2023 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: Apache-2.0 */

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/vmware-tanzu/nsx-operator/pkg/apis/nsx.vmware.com/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// StaticRouteLister helps list StaticRoutes.
// All objects returned here must be treated as read-only.
type StaticRouteLister interface {
	// List lists all StaticRoutes in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.StaticRoute, err error)
	// StaticRoutes returns an object that can list and get StaticRoutes.
	StaticRoutes(namespace string) StaticRouteNamespaceLister
	StaticRouteListerExpansion
}

// staticRouteLister implements the StaticRouteLister interface.
type staticRouteLister struct {
	indexer cache.Indexer
}

// NewStaticRouteLister returns a new StaticRouteLister.
func NewStaticRouteLister(indexer cache.Indexer) StaticRouteLister {
	return &staticRouteLister{indexer: indexer}
}

// List lists all StaticRoutes in the indexer.
func (s *staticRouteLister) List(selector labels.Selector) (ret []*v1alpha1.StaticRoute, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.StaticRoute))
	})
	return ret, err
}

// StaticRoutes returns an object that can list and get StaticRoutes.
func (s *staticRouteLister) StaticRoutes(namespace string) StaticRouteNamespaceLister {
	return staticRouteNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// StaticRouteNamespaceLister helps list and get StaticRoutes.
// All objects returned here must be treated as read-only.
type StaticRouteNamespaceLister interface {
	// List lists all StaticRoutes in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.StaticRoute, err error)
	// Get retrieves the StaticRoute from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.StaticRoute, error)
	StaticRouteNamespaceListerExpansion
}

// staticRouteNamespaceLister implements the StaticRouteNamespaceLister
// interface.
type staticRouteNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all StaticRoutes in the indexer for a given namespace.
func (s staticRouteNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.StaticRoute, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.StaticRoute))
	})
	return ret, err
}

// Get retrieves the StaticRoute from the indexer for a given namespace and name.
func (s staticRouteNamespaceLister) Get(name string) (*v1alpha1.StaticRoute, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("staticroute"), name)
	}
	return obj.(*v1alpha1.StaticRoute), nil
}
