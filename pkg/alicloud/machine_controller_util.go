/*
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.

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

package alicloud

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

const (
	// AlicloudAccessKeyID is a constant for a key name that is part of the Alibaba cloud credentials.
	AlicloudAccessKeyID string = "alicloudAccessKeyID"
	// AlicloudAccessKeySecret is a constant for a key name that is part of the Alibaba cloud credentials.
	AlicloudAccessKeySecret string = "alicloudAccessKeySecret"
	// AlicloudUserData is a constant for user data
	AlicloudUserData string = "userData"
)

func newECSClient(secret *corev1.Secret, region string) (*ecs.Client, error) {
	accessKeyID := strings.TrimSpace(string(secret.Data[AlicloudAccessKeyID]))
	accessKeySecret := strings.TrimSpace(string(secret.Data[AlicloudAccessKeySecret]))
	ecsClient, err := ecs.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ecsClient, err
}

func decodeProviderSpec(machineClass *v1alpha1.MachineClass) (*api.ProviderSpec, error) {
	var providerSpec *api.ProviderSpec
	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}

func generateDataDiskRequests(disks []api.AlicloudDataDisk, machineName string) []ecs.RunInstancesDataDisk {
	var dataDiskRequests []ecs.RunInstancesDataDisk
	for _, disk := range disks {
		dataDiskRequest := ecs.RunInstancesDataDisk{
			Category:    disk.Category,
			Encrypted:   strconv.FormatBool(disk.Encrypted),
			DiskName:    fmt.Sprintf("%s-%s-data-disk", machineName, disk.Name),
			Description: disk.Description,
			Size:        fmt.Sprintf("%d", disk.Size),
		}

		if disk.DeleteWithInstance != nil {
			dataDiskRequest.DeleteWithInstance = strconv.FormatBool(*disk.DeleteWithInstance)
		} else {
			dataDiskRequest.DeleteWithInstance = strconv.FormatBool(true)
		}

		if disk.Category == "DiskEphemeralSSD" {
			dataDiskRequest.DeleteWithInstance = ""
		}

		dataDiskRequests = append(dataDiskRequests, dataDiskRequest)
	}

	return dataDiskRequests
}

func toInstanceTags(tags map[string]string) ([]ecs.RunInstancesTag, error) {
	result := []ecs.RunInstancesTag{{}, {}}
	hasCluster := false
	hasRole := false

	for k, v := range tags {
		if strings.Contains(k, "kubernetes.io/cluster/") {
			hasCluster = true
			result[0].Key = k
			result[0].Value = v
		} else if strings.Contains(k, "kubernetes.io/role/") {
			hasRole = true
			result[1].Key = k
			result[1].Value = v
		} else {
			result = append(result, ecs.RunInstancesTag{Key: k, Value: v})
		}
	}

	if !hasCluster || !hasRole {
		err := fmt.Errorf("Tags should at least contains 2 keys, which are prefixed with kubernetes.io/cluster and kubernetes.io/role")
		return nil, err
	}

	return result, nil
}

func encodeProviderID(region, instanceID string) string {
	return fmt.Sprintf("%s.%s", region, instanceID)
}

func decodeProviderID(providerID string) string {
	splitProviderID := strings.Split(providerID, ".")
	return splitProviderID[len(splitProviderID)-1]
}

func describeInstances(providerID string, providerSpec *api.ProviderSpec, client *ecs.Client) ([]ecs.Instance, error) {
	request := ecs.CreateDescribeInstancesRequest()

	if providerID != "" {
		instanceID := decodeProviderID(providerID)
		request.InstanceIds = "[\"" + instanceID + "\"]"
	} else {
		searchFilters := make(map[string]string)
		for k, v := range providerSpec.Tags {
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

	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, err
	}

	return response.Instances.Instance, nil
}
