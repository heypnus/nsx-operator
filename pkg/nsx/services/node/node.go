package node

import (
	"fmt"
	"sync"

	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"k8s.io/client-go/tools/cache"

	"github.com/vmware-tanzu/nsx-operator/pkg/logger"
	servicecommon "github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/common"
)

var (
	log              = logger.Log
	ResourceTypeNode = servicecommon.ResourceTypeNode
	MarkedForDelete  = true
)

type NodeService struct {
	servicecommon.Service
	NodeStore *NodeStore
}

func InitializeNode(service servicecommon.Service) (*NodeService, error) {
	wg := sync.WaitGroup{}
	wgDone := make(chan bool)
	fatalErrors := make(chan error)

	wg.Add(1)

	nodeService := &NodeService{Service: service}

	nodeService.NodeStore = &NodeStore{
		ResourceStore: servicecommon.ResourceStore{
			Indexer: cache.NewIndexer(
				keyFunc,
				cache.Indexers{
					servicecommon.IndexKeyNodeName: nodeIndexByNodeName,
				},
			),
			BindingType: model.HostTransportNodeBindingType(),
		},
	}
	// TODO: confirm whether we can remove the following intialization because node doesn't have the cluster tag so it's a dry run
	go nodeService.InitializeResourceStore(&wg, fatalErrors, ResourceTypeNode, nil, nodeService.NodeStore)

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-fatalErrors:
		close(fatalErrors)
		return nodeService, err
	}

	return nodeService, nil

}

func (service *NodeService) SyncNodeStore(nodeName string) error {
	nodes := service.NodeStore.GetByIndex(servicecommon.IndexKeyNodeName, nodeName)
	if len(nodes) > 1 {
		return fmt.Errorf("multiple nodes found for node name %s", nodeName)
	}
	// TODO: confirm whether we need to resync the node info frmo NSX
	if len(nodes) == 1 {
		log.Info("node alreay cached", "node.Fqdn", nodes[0].NodeDeploymentInfo.Fqdn, "node.Id", *nodes[0].Id)
		// updatedNode, err := service.NSXClient.HostTransPortNodesClient.Get("default", "default", nodes[0].Id)
		// if err != nil {
		// 	return fmt.Errorf("failed to get HostTransPortNode for node %s: %s", nodeName, err)
		// }
		// node.NodeStore.Operate(updatedNode)
	}
	nodeResults, err := service.NSXClient.HostTransPortNodesClient.List("default", "default", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		log.Error(err, "failed to list HostTransportNodes")
	}
	for _, node := range nodeResults.Results {
		if *node.NodeDeploymentInfo.Fqdn == nodeName {
			service.NodeStore.Operate(&node)
			break
		}
	}
	return nil
}
