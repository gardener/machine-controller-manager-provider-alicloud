/*
Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mock

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

// ClientImplMock is the mock implement of ECSClient interface
type ClientImplMock struct{}

func (mockClient *ClientImplMock) RunInstances(request *ecs.RunInstancesRequest) (*ecs.RunInstancesResponse, error) {
	return &ecs.RunInstancesResponse{
		InstanceIdSets: ecs.InstanceIdSets{
			InstanceIdSet: []string{"i-mockinstanceid"},
		},
	}, nil
}

func (mockClient *ClientImplMock) DescribeInstances(request *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error) {
	var response ecs.DescribeInstancesResponse
	if len(request.InstanceIds) == 0 {
		response = ecs.DescribeInstancesResponse{
			Instances: ecs.Instances{
				Instance: []ecs.Instance{},
			},
		}
	} else {
		response = ecs.DescribeInstancesResponse{
			Instances: ecs.Instances{
				Instance: []ecs.Instance{
					{
						InstanceId: "i-mockinstanceid",
						Status:     "Running",
					},
				},
			},
		}
	}

	return &response, nil
}

func (mockClient *ClientImplMock) DeleteInstance(request *ecs.DeleteInstanceRequest) (*ecs.DeleteInstanceResponse, error) {
	return &ecs.DeleteInstanceResponse{}, nil
}
