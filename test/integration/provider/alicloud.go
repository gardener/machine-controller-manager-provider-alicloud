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
package provider

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"

	providerDriver "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
)

/**
	Orphaned Resources
	- VMs:
		Describe instances with specified tag name:<cluster-name>
		Report/Print out instances found
		Describe volumes attached to the instance (using instance id)
		Report/Print out volumes found
		Delete attached volumes found
		Terminate instances found
	- Disks:
		Describe volumes with tag status:available
		Report/Print out volumes found
		Delete identified volumes
**/

func newSession(machineClass *v1alpha1.MachineClass, secret *v1.Secret) spi.ECSClient {
	var (
		providerSpec *api.ProviderSpec
		sPI          spi.PluginSPIImpl
	)

	err := json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Error occured while performing unmarshal %s", err.Error())
	}
	sess, err := sPI.NewECSClient(secret, providerSpec.Region)
	if err != nil {
		log.Printf("Error occured while creating new session %s", err)
	}
	return sess
}

func getMachines(machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	var machines []string
	var sPI spi.PluginSPIImpl
	driverProvider := providerDriver.NewAlicloudPlugin(&sPI)
	machineList, err := driverProvider.ListMachines(context.TODO(), &driver.ListMachinesRequest{
		MachineClass: machineClass,
		Secret:       &v1.Secret{Data: secretData},
	})
	if err != nil {
		return nil, err
	} else if len(machineList.MachineList) != 0 {
		for _, machine := range machineList.MachineList {
			machines = append(machines, machine)
		}
	}
	return machines, nil
}

// getOrphanedInstances returns list of Orphan resources that couldn't be deleted.
func getOrphanedInstances(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var instancesID []string
	var tags = &[]ecs.DescribeInstancesTag{{Key: tagName, Value: tagValue}}
	input := ecs.CreateDescribeInstancesRequest()
	input.Status = "running"
	input.Tag = tags

	result, err := sess.DescribeInstances(input)
	if err != nil {
		return instancesID, err
	}
	for _, instance := range result.Instances.Instance {
		instancesID = append(instancesID, instance.InstanceId)
	}
	return instancesID, nil
}

// getOrphanedDisks returns list of Orphan disks.
func getOrphanedDisks(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var volumeID []string
	var tags = &[]ecs.DescribeDisksTag{{Key: tagName, Value: tagValue}}
	input := ecs.CreateDescribeDisksRequest()
	input.Status = "Available"
	input.Tag = tags
	result, err := sess.DescribeDisks(input)
	if err != nil {
		return volumeID, err
	}
	for _, disk := range result.Disks.Disk {
		volumeID = append(volumeID, disk.DiskId)
	}
	return volumeID, nil
}

// getOrphanedNICs returns list of Orphan NICs
func getOrphanedNICs(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var nicIDs []string
	var tags = &[]ecs.DescribeNetworkInterfacesTag{{Key: tagName, Value: tagValue}}
	input := ecs.CreateDescribeNetworkInterfacesRequest()
	input.Tag = tags

	result, err := sess.DescribeNetworkInterfaces(input)
	if err != nil {
		return nicIDs, err
	}
	for _, nic := range result.NetworkInterfaceSets.NetworkInterfaceSet {
		nicIDs = append(nicIDs, nic.NetworkInterfaceId)
	}
	return nicIDs, nil
}

func cleanOrphanResources(instanceIds []string, volumeIds []string, NICIds []string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) (delErrInstanceId []string, delErrVolumeIds []string, delErrNICs []string) {

	for _, instanceId := range instanceIds {
		if err := terminateInstance(instanceId, machineClass, secretData); err != nil {
			fmt.Printf("error in deleting instance : %v", err)
			delErrInstanceId = append(delErrInstanceId, instanceId)
		}
	}

	for _, volumeId := range volumeIds {
		if err := deleteVolume(volumeId, machineClass, secretData); err != nil {
			fmt.Printf("error in deleting volume : %v", err)
			delErrVolumeIds = append(delErrVolumeIds, volumeId)
		}
	}

	for _, nicId := range NICIds {
		if err := deleteNIC(nicId, machineClass, secretData); err != nil {
			fmt.Printf("error in deleting volume : %v", err)
			delErrNICs = append(delErrNICs, nicId)
		}
	}

	return
}

func deleteNIC(nicId string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := ecs.CreateDeleteNetworkInterfaceRequest()
	input.NetworkInterfaceId = nicId
	_, err := sess.DeleteNetworkInterface(input)
	if err != nil {
		return err
	}
	return nil
}

func deleteVolume(diskId string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := ecs.CreateDeleteDiskRequest()
	input.DiskId = diskId
	_, err := sess.DeleteDisk(input)
	if err != nil {
		return err
	}
	return nil
}

func terminateInstance(instanceId string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := ecs.CreateDeleteInstanceRequest()
	input.InstanceId = instanceId
	_, err := sess.DeleteInstance(input)
	if err != nil {
		return err
	}
	return nil
}
