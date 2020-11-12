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
	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
	corev1 "k8s.io/api/core/v1"
)

// PluginSPIMock is the mock implementation of PluginSPI
type PluginSPIMock struct{}

// NewECSClient returns a mock instance of the ECS client.
func (pluginSPIMock *PluginSPIMock) NewECSClient(secret *corev1.Secret, region string) (spi.ECSClient, error) {
	return &ClientImplMock{}, nil
}

// NewRunInstancesRequest returns a mock request of running instances.
func (pluginSPIMock *PluginSPIMock) NewRunInstancesRequest(providerSpec *api.ProviderSpec, machineName string, userData []byte) (*ecs.RunInstancesRequest, error) {
	pluginSPI := spi.PluginSPIImpl{}
	return pluginSPI.NewRunInstancesRequest(providerSpec, machineName, userData)
}

// NewRunInstancesRequest returns a mock request of describing instances.
func (pluginSPIMock *PluginSPIMock) NewDescribeInstancesRequest(machineName, providerID string, tags map[string]string) (*ecs.DescribeInstancesRequest, error) {
	pluginSPI := spi.PluginSPIImpl{}
	return pluginSPI.NewDescribeInstancesRequest(machineName, providerID, tags)
}

// NewDeleteInstanceRequest returns a mock request of deleting instances.
func (pluginSPIMock *PluginSPIMock) NewDeleteInstanceRequest(instanceID string, force bool) (*ecs.DeleteInstanceRequest, error) {
	pluginSPI := spi.PluginSPIImpl{}
	return pluginSPI.NewDeleteInstanceRequest(instanceID, force)
}

// NewInstanceDataDisks return a mock data disks
func (pluginSPIMock *PluginSPIMock) NewInstanceDataDisks(disks []api.AlicloudDataDisk, machineName string) []ecs.RunInstancesDataDisk {
	pluginSPI := spi.PluginSPIImpl{}
	return pluginSPI.NewInstanceDataDisks(disks, machineName)
}

// NewInstanceDataDisks return a mock tags
func (pluginSPIMock *PluginSPIMock) NewRunInstanceTags(tags map[string]string) ([]ecs.RunInstancesTag, error) {
	pluginSPI := spi.PluginSPIImpl{}
	return pluginSPI.NewRunInstanceTags(tags)
}
