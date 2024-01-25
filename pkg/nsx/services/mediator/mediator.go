package mediator

import (
	"context"
	"fmt"

	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/nsx-operator/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/nsx-operator/pkg/logger"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/common"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/node"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/subnet"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/subnetport"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/vpc"
	"github.com/vmware-tanzu/nsx-operator/pkg/util"
)

var log = logger.Log

// ServiceMediator We use mediator pattern to wrap all the services,
// embed all the services in ServiceMediator, so that we can mediate all the methods of all the services
// transparently to the caller, for example, in other packages, we can use ServiceMediator.GetVPCsByNamespace directly.
// In startCRDController function, we register the CRDService to the ServiceMediator, since only one controller writes to
// its own store and other controllers read from the store, so we don't need lock here.
type ServiceMediator struct {
	*vpc.VPCService
	*subnet.SubnetService
	*subnetport.SubnetPortService
	*node.NodeService
}

// This method is used for subnet service since vpc network config contains default subnet size
// and default subnet access mode.
func (m *ServiceMediator) GetVPCNetworkConfigByNamespace(ns string) *common.VPCNetworkConfigInfo {
	return m.VPCService.GetVPCNetworkConfigByNamespace(ns)
}

// GetAvailableSubnet returns available Subnet under SubnetSet, and creates Subnet if necessary.
func (serviceMediator *ServiceMediator) GetAvailableSubnet(subnetSet *v1alpha1.SubnetSet) (string, error) {
	subnetList := serviceMediator.SubnetStore.GetByIndex(common.TagScopeSubnetSetCRUID, string(subnetSet.GetUID()))
	for _, nsxSubnet := range subnetList {
		portNums := len(serviceMediator.GetPortsOfSubnet(*nsxSubnet.Id))
		totalIP := int(*nsxSubnet.Ipv4SubnetSize)
		if len(nsxSubnet.IpAddresses) > 0 {
			// totalIP will be overrided if IpAddresses are specified.
			totalIP, _ = util.CalculateIPFromCIDRs(nsxSubnet.IpAddresses)
		}
		if portNums < totalIP-3 {
			return *nsxSubnet.Path, nil
		}
	}
	namespace := &corev1.Namespace{}
	namespacedName := types.NamespacedName{
		Name: subnetSet.Namespace,
	}
	if err := serviceMediator.SubnetService.Client.Get(context.Background(), namespacedName, namespace); err != nil {
		return "", err
	}
	tags := serviceMediator.SubnetService.GenerateSubnetNSTags(subnetSet, string(namespace.UID))
	for k, v := range namespace.Labels {
		tags = append(tags, model.Tag{Scope: common.String(k), Tag: common.String(v)})
	}
	log.Info("the existing subnets are not available, creating new subnet", "subnetList", subnetList, "subnetSet.Name", subnetSet.Name, "subnetSet.Namespace", subnetSet.Namespace)
	vpcInfo, err := serviceMediator.GetNamespaceVPCInfo(subnetSet.Namespace)
	if err != nil {
		return "", err
	}
	return serviceMediator.CreateOrUpdateSubnet(subnetSet, *vpcInfo, tags)
}

func (serviceMediator *ServiceMediator) GetNamespaceVPCInfo(ns string) (*common.VPCResourceInfo, error) {
	vpcList := serviceMediator.GetVPCsByNamespace(ns)
	if len(vpcList) == 0 {
		return nil, fmt.Errorf("no vpc found for ns %s", ns)
	}
	vpcInfo, err := common.ParseVPCResourcePath(*vpcList[0].Path)
	if err != nil {
		err := fmt.Errorf("failed to parse NSX VPC path for VPC %s: %s", *vpcList[0].Id, err)
		return nil, err
	}
	return &vpcInfo, nil
}

func (serviceMediator *ServiceMediator) GetPortsOfSubnet(nsxSubnetID string) (ports []*model.VpcSubnetPort) {
	subnetPortList := serviceMediator.SubnetPortStore.GetByIndex(common.IndexKeySubnetID, nsxSubnetID)
	return subnetPortList
}

func (serviceMediator *ServiceMediator) GetNodeByName(nodeName string) (*model.HostTransportNode, error) {
	nodes := serviceMediator.NodeStore.GetByIndex(common.IndexKeyNodeName, nodeName)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("node %s not found", nodeName)
	}
	if len(nodes) > 1 {
		var nodeIDs []string
		for _, node := range nodes {
			nodeIDs = append(nodeIDs, *node.Id)
		}
		return nil, fmt.Errorf("multiple node IDs found for node %s: %v", nodeName, nodeIDs)
	}
	return nodes[0], nil
}
