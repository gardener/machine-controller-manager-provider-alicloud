package alicloud

import (
	"fmt"

	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// GetInstanceIDFromRunInstancesResponse is a utility function to extract instance ID from RunInstancesResponse
func GetInstanceIDFromRunInstancesResponse(resp *ecs.RunInstancesResponse) (*string, error) {
	if resp == nil ||
		resp.Body == nil ||
		resp.Body.InstanceIdSets == nil ||
		len(resp.Body.InstanceIdSets.InstanceIdSet) == 0 {

		return nil, fmt.Errorf("instance ID missing")
	}

	for _, instanceID := range resp.Body.InstanceIdSets.InstanceIdSet {
		if instanceID != nil {
			return instanceID, nil
		}
	}

	return nil, fmt.Errorf("instance ID missing")
}

// GetInstancesFromDescribeInstancesResponse is a utility function to extract instances from DescribeInstancesResponse
func GetInstancesFromDescribeInstancesResponse(resp *ecs.DescribeInstancesResponse) ([]*ecs.DescribeInstancesResponseBodyInstancesInstance, error) {
	if resp == nil ||
		resp.Body == nil ||
		resp.Body.Instances == nil {

		return nil, fmt.Errorf("invalid response")
	}

	return resp.Body.Instances.Instance, nil
}

// GetAllInstances is a utility function to get all instances matching the DescribeInstancesRequest with pagination
func (plugin *MachinePlugin) GetAllInstances(client spi.ECSClient, request *ecs.DescribeInstancesRequest) ([]*ecs.DescribeInstancesResponseBodyInstancesInstance, error) {
	var instances []*ecs.DescribeInstancesResponseBodyInstancesInstance
	pageNumber := 0
	for {
		response, err := client.DescribeInstances(request)
		if err != nil {
			return nil, err
		}
		pageInstances, err := GetInstancesFromDescribeInstancesResponse(response)
		if err != nil {
			return nil, err
		}
		instances = append(instances, pageInstances...)
		pageNumber++
		klog.V(3).Infof("Fetched %d/%d instances (Page %d)", len(instances), ptr.Deref(response.Body.TotalCount, 0), pageNumber)

		if ptr.Deref(response.Body.NextToken, "") == "" {
			break
		}
		request.NextToken = response.Body.NextToken
	}
	return instances, nil
}
