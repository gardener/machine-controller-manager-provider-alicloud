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

// Package alicloud contains the alicloud provider specific implementations to manage machines
package alicloud

import (
	"encoding/base64"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

// PluginSPI provides an interface to deal with cloud provider session
// You can optionally enhance this interface to add interface methods here
// You can use it to mock cloud provider calls
type PluginSPI interface {
	NewECSClient(secret *corev1.Secret, region string) (*ecs.Client, error)
	NewRunInstancesRequest(providerSpec *api.ProviderSpec, machineName string, userData []byte) (*ecs.RunInstancesRequest, error)
	NewDescribeInstancesRequest(machineName, providerID string, tags map[string]string) (*ecs.DescribeInstancesRequest, error)
	NewDeleteInstanceRequest(instanceID string, force bool) (*ecs.DeleteInstanceRequest, error)
	NewInstanceDataDisks(disks []api.AlicloudDataDisk, machineName string) []ecs.RunInstancesDataDisk
}

// MachinePlugin implements the driver.Driver
// It also implements the PluginSPI interface
type MachinePlugin struct {
	SPI PluginSPI
}

// PluginSPIImpl is the real implementation of SPI interface that makes the calls to the provider SDK.
type PluginSPIImpl struct{}

// NewECSClient returns a new instance of the ECS client.
func (pluginSPI *PluginSPIImpl) NewECSClient(secret *corev1.Secret, region string) (*ecs.Client, error) {
	accessKeyID := strings.TrimSpace(string(secret.Data[AlicloudAccessKeyID]))
	accessKeySecret := strings.TrimSpace(string(secret.Data[AlicloudAccessKeySecret]))
	ecsClient, err := ecs.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ecsClient, err
}

// NewRunInstancesRequest returns a new request of run instance.
func (pluginSPI *PluginSPIImpl) NewRunInstancesRequest(providerSpec *api.ProviderSpec, machineName string, userData []byte) (*ecs.RunInstancesRequest, error) {
	request := ecs.CreateRunInstancesRequest()

	request.ImageId = providerSpec.ImageID
	request.InstanceType = providerSpec.InstanceType
	request.RegionId = providerSpec.Region
	request.ZoneId = providerSpec.ZoneID
	request.SecurityGroupId = providerSpec.SecurityGroupID
	request.VSwitchId = providerSpec.VSwitchID
	request.PrivateIpAddress = providerSpec.PrivateIPAddress
	request.InstanceChargeType = providerSpec.InstanceChargeType
	request.InternetChargeType = providerSpec.InternetChargeType
	request.SpotStrategy = providerSpec.SpotStrategy
	request.IoOptimized = providerSpec.IoOptimized
	request.KeyPairName = providerSpec.KeyPairName

	if providerSpec.InternetMaxBandwidthIn != nil {
		request.InternetMaxBandwidthIn = requests.NewInteger(int(*providerSpec.InternetMaxBandwidthIn))
	}

	if providerSpec.InternetMaxBandwidthOut != nil {
		request.InternetMaxBandwidthOut = requests.NewInteger(int(*providerSpec.InternetMaxBandwidthOut))
	}

	if providerSpec.DataDisks != nil && len(providerSpec.DataDisks) > 0 {
		dataDisks := pluginSPI.NewInstanceDataDisks(providerSpec.DataDisks, machineName)
		request.DataDisk = &dataDisks
	}

	if providerSpec.SystemDisk != nil {
		request.SystemDiskCategory = providerSpec.SystemDisk.Category
		request.SystemDiskSize = fmt.Sprintf("%d", providerSpec.SystemDisk.Size)
	}

	tags, err := toInstanceTags(providerSpec.Tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	request.Tag = &tags
	request.InstanceName = machineName
	request.ClientToken = utils.GetUUIDV4()
	request.UserData = base64.StdEncoding.EncodeToString(userData)

	return request, nil
}

// NewDescribeInstancesRequest returns a new request of describe instance.
func (pluginSPI *PluginSPIImpl) NewDescribeInstancesRequest(machineName, providerID string, tags map[string]string) (*ecs.DescribeInstancesRequest, error) {
	request := ecs.CreateDescribeInstancesRequest()

	if providerID != "" {
		instanceID := decodeProviderID(providerID)
		request.InstanceIds = "[\"" + instanceID + "\"]"
	} else if machineName != "" {
		request.InstanceName = machineName
	} else {
		searchFilters := make(map[string]string)
		for k, v := range tags {
			if strings.Contains(k, "kubernetes.io/cluster/") || strings.Contains(k, "kubernetes.io/role/") {
				searchFilters[k] = v
			}
		}

		if len(searchFilters) < 2 {
			return nil, fmt.Errorf("Can't find VMs with none of machineID/Tag[kubernetes.io/cluster/*]/Tag[kubernetes.io/role/*]")
		}

		var tags []ecs.DescribeInstancesTag
		for k, v := range searchFilters {
			tags = append(tags, ecs.DescribeInstancesTag{
				Key:   k,
				Value: v,
			})
		}
		request.Tag = &tags
	}

	return request, nil
}

// NewDeleteInstanceRequest returns a new request of delete instance.
func (pluginSPI *PluginSPIImpl) NewDeleteInstanceRequest(instanceID string, force bool) (*ecs.DeleteInstanceRequest, error) {
	request := ecs.CreateDeleteInstanceRequest()

	request.InstanceId = instanceID
	request.Force = requests.NewBoolean(force)

	return request, nil
}

// NewInstanceDataDisks  instance.
func (pluginSPI *PluginSPIImpl) NewInstanceDataDisks(disks []api.AlicloudDataDisk, machineName string) []ecs.RunInstancesDataDisk {
	var instanceDataDisks []ecs.RunInstancesDataDisk

	for _, disk := range disks {
		instanceDataDisk := ecs.RunInstancesDataDisk{
			Category:    disk.Category,
			Encrypted:   strconv.FormatBool(disk.Encrypted),
			DiskName:    fmt.Sprintf("%s-%s-data-disk", machineName, disk.Name),
			Description: disk.Description,
			Size:        fmt.Sprintf("%d", disk.Size),
		}

		if disk.DeleteWithInstance != nil {
			instanceDataDisk.DeleteWithInstance = strconv.FormatBool(*disk.DeleteWithInstance)
		} else {
			instanceDataDisk.DeleteWithInstance = strconv.FormatBool(true)
		}

		if disk.Category == "DiskEphemeralSSD" {
			instanceDataDisk.DeleteWithInstance = ""
		}

		instanceDataDisks = append(instanceDataDisks, instanceDataDisk)
	}

	return instanceDataDisks
}

// NewAlicloudPlugin returns a new Alicloud machine plugin.
func NewAlicloudPlugin(pluginSPI PluginSPI) driver.Driver {
	return &MachinePlugin{
		SPI: pluginSPI,
	}
}
