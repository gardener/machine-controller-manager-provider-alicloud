package alicloud

import (
	"fmt"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
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
