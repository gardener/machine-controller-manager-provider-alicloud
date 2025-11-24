package provider

import (
	"context"
	"fmt"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
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

// getOrphanedInstances returns list of Orphan resources.
func getOrphanedInstances(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var instancesID []string
	var tags = []*ecs.DescribeInstancesRequestTag{{Key: &tagName, Value: &tagValue}}
	input := ecs.DescribeInstancesRequest{}
	input.Status = tea.String("running")
	input.Tag = tags

	var providerSpec *api.ProviderSpec
	err := json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Failed to unmarshal ProviderSpec: %v", err)
	}
	input.RegionId = tea.String(providerSpec.Region)

	result, err := sess.DescribeInstances(&input)
	if err != nil {
		return instancesID, err
	}
	for _, instance := range result.Body.Instances.Instance {
		instancesID = append(instancesID, *instance.InstanceId)
	}
	return instancesID, nil
}

// getOrphanedDisks returns list of Orphan disks.
func getOrphanedDisks(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var volumeID []string
	var tags = []*ecs.DescribeDisksRequestTag{{Key: &tagName, Value: &tagValue}}
	input := ecs.DescribeDisksRequest{}
	input.Status = tea.String("Available")
	input.Tag = tags

	var providerSpec *api.ProviderSpec
	err := json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Failed to unmarshal ProviderSpec: %v", err)
	}
	input.RegionId = tea.String(providerSpec.Region)

	result, err := sess.DescribeDisks(&input)
	if err != nil {
		return volumeID, err
	}
	for _, disk := range result.Body.Disks.Disk {
		volumeID = append(volumeID, *disk.DiskId)
	}
	return volumeID, nil
}

// getOrphanedNICs returns list of Orphan NICs
func getOrphanedNICs(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var nicIDs []string
	var tags = []*ecs.DescribeNetworkInterfacesRequestTag{{Key: &tagName, Value: &tagValue}}
	input := ecs.DescribeNetworkInterfacesRequest{}
	input.Tag = tags

	var providerSpec *api.ProviderSpec
	err := json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Failed to unmarshal ProviderSpec: %v", err)
	}
	input.RegionId = tea.String(providerSpec.Region)

	result, err := sess.DescribeNetworkInterfaces(&input)
	if err != nil {
		return nicIDs, err
	}
	for _, nic := range result.Body.NetworkInterfaceSets.NetworkInterfaceSet {
		nicIDs = append(nicIDs, *nic.NetworkInterfaceId)
	}
	return nicIDs, nil
}

func cleanOrphanResources(instanceIds []string, volumeIds []string, NICIds []string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) (delErrInstanceID []string, delErrVolumeIds []string, delErrNICs []string) {

	for _, instanceID := range instanceIds {
		if err := terminateInstance(instanceID, machineClass, secretData); err != nil {
			fmt.Printf("error in deleting instance : %v", err)
			delErrInstanceID = append(delErrInstanceID, instanceID)
		}
	}

	for _, volumeID := range volumeIds {
		if err := deleteVolume(volumeID, machineClass, secretData); err != nil {
			fmt.Printf("error in deleting volume : %v", err)
			delErrVolumeIds = append(delErrVolumeIds, volumeID)
		}
	}

	for _, nicID := range NICIds {
		if err := deleteNIC(nicID, machineClass, secretData); err != nil {
			fmt.Printf("error in deleting volume : %v", err)
			delErrNICs = append(delErrNICs, nicID)
		}
	}

	return
}

func deleteNIC(nicID string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := ecs.DeleteNetworkInterfaceRequest{}
	input.NetworkInterfaceId = tea.String(nicID)
	_, err := sess.DeleteNetworkInterface(&input)
	if err != nil {
		return err
	}
	return nil
}

func deleteVolume(diskID string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := ecs.DeleteDiskRequest{}
	input.DiskId = tea.String(diskID)
	_, err := sess.DeleteDisk(&input)
	if err != nil {
		return err
	}
	return nil
}

func terminateInstance(instanceID string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := ecs.DeleteInstanceRequest{}
	input.InstanceId = tea.String(instanceID)
	_, err := sess.DeleteInstance(&input)
	if err != nil {
		return err
	}
	return nil
}
