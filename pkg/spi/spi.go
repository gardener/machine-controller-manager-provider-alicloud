// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package spi

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"

	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
)

const (
	// AlicloudAccessKeyID is a constant for a key name that is part of the Alibaba cloud credentials.
	AlicloudAccessKeyID string = "alicloudAccessKeyID"
	// AlicloudAccessKeySecret is a constant for a key name that is part of the Alibaba cloud credentials.
	AlicloudAccessKeySecret string = "alicloudAccessKeySecret"
	// AlicloudAlternativeAccessKeyID is a constant for a key name of a secret containing the Alibaba cloud
	// credentials (access key id).
	AlicloudAlternativeAccessKeyID = "accessKeyID"
	// AlicloudAlternativeAccessKeySecret is a constant for a key name of a secret containing the Alibaba cloud
	// credentials (access key secret).
	AlicloudAlternativeAccessKeySecret = "accessKeySecret"
	// AlicloudUserData is a constant for user data
	AlicloudUserData string = "userData"
	// AlicloudDriverName is the name of the CSI driver for Alibaba Cloud
	AlicloudDriverName = "diskplugin.csi.alibabacloud.com"
)

// ECSClient provides an interface
type ECSClient interface {
	RunInstances(request *ecs.RunInstancesRequest) (*ecs.RunInstancesResponse, error)
	DescribeInstances(request *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error)
	DeleteInstance(request *ecs.DeleteInstanceRequest) (*ecs.DeleteInstanceResponse, error)
	DescribeDisks(request *ecs.DescribeDisksRequest) (*ecs.DescribeDisksResponse, error)
	DeleteDisk(request *ecs.DeleteDiskRequest) (*ecs.DeleteDiskResponse, error)
	DescribeNetworkInterfaces(request *ecs.DescribeNetworkInterfacesRequest) (*ecs.DescribeNetworkInterfacesResponse, error)
	DeleteNetworkInterface(request *ecs.DeleteNetworkInterfaceRequest) (*ecs.DeleteNetworkInterfaceResponse, error)
}

// PluginSPI provides an interface to deal with cloud provider session
// You can optionally enhance this interface to add interface methods here
// You can use it to mock cloud provider calls
type PluginSPI interface {
	NewECSClient(secret *corev1.Secret, region string) (ECSClient, error)
	NewRunInstancesRequest(providerSpec *api.ProviderSpec, machineName string, userData []byte) (*ecs.RunInstancesRequest, error)
	NewDescribeInstancesRequest(machineName, instanceID string, tags map[string]string) (*ecs.DescribeInstancesRequest, error)
	NewDeleteInstanceRequest(instanceID string, force bool) (*ecs.DeleteInstanceRequest, error)
	NewInstanceDataDisks(disks []api.AlicloudDataDisk, machineName string) []ecs.RunInstancesDataDisk
	NewRunInstanceTags(tags map[string]string) ([]ecs.RunInstancesTag, error)
}

// PluginSPIImpl is the real implementation of SPI interface that makes the calls to the provider SDK.
type PluginSPIImpl struct{}

// NewECSClient returns a new instance of the ECS client.
func (pluginSPI *PluginSPIImpl) NewECSClient(secret *corev1.Secret, region string) (ECSClient, error) {
	accessKeyID := extractCredentialsFromData(secret.Data, AlicloudAccessKeyID, AlicloudAlternativeAccessKeyID)
	accessKeySecret := extractCredentialsFromData(secret.Data, AlicloudAccessKeySecret, AlicloudAlternativeAccessKeySecret)
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

	if len(providerSpec.DataDisks) > 0 {
		dataDisks := pluginSPI.NewInstanceDataDisks(providerSpec.DataDisks, machineName)
		request.DataDisk = &dataDisks
	}

	if providerSpec.SystemDisk != nil {
		request.SystemDiskCategory = providerSpec.SystemDisk.Category
		request.SystemDiskSize = fmt.Sprintf("%d", providerSpec.SystemDisk.Size)
	}

	tags, err := pluginSPI.NewRunInstanceTags(providerSpec.Tags)
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
func (pluginSPI *PluginSPIImpl) NewDescribeInstancesRequest(machineName, instanceID string, tags map[string]string) (*ecs.DescribeInstancesRequest, error) {
	request := ecs.CreateDescribeInstancesRequest()

	if instanceID != "" {
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

// NewInstanceDataDisks returns instances data disks.
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

// NewRunInstanceTags returns tags of Running Instances.
func (pluginSPI *PluginSPIImpl) NewRunInstanceTags(tags map[string]string) ([]ecs.RunInstancesTag, error) {
	runInstancesTags := make([]ecs.RunInstancesTag, 0, 2)
	hasCluster, hasRole := false, false

	for k, v := range tags {
		if strings.Contains(k, "kubernetes.io/cluster/") {
			hasCluster = true
		} else if strings.Contains(k, "kubernetes.io/role/") {
			hasRole = true
		}
		runInstancesTags = append(runInstancesTags, ecs.RunInstancesTag{Key: k, Value: v})
	}

	if !hasCluster || !hasRole {
		err := fmt.Errorf("Tags should at least contains 2 keys, which are prefixed with kubernetes.io/cluster and kubernetes.io/role")
		return nil, err
	}

	return runInstancesTags, nil
}

// extractCredentialsFromData extracts and trims a value from the given data map. The first key that exists is being
// returned, otherwise, the next key is tried, etc. If no key exists then an empty string is returned.
func extractCredentialsFromData(data map[string][]byte, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			return strings.TrimSpace(string(val))
		}
	}
	return ""
}
