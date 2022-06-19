package provider

import (
	"context"
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

// getOrphanesInstances returns list of Orphan resources that couldn't be deleted
func getOrphanedInstances(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var instancesID []string
	var tags = &[]ecs.DescribeInstancesTag{{Key: tagName, Value: tagValue}}
	input := ecs.CreateDescribeInstancesRequest()
	input.InstanceName = "instance-state-name"
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

// getOrphanesDisks returns list of Orphan disks
func getOrphanedDisks(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	var volumeID []string
	var tags = &[]ecs.DescribeDisksTag{{Key: tagName, Value: tagValue}}
	input := ecs.CreateDescribeDisksRequest()
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
	var NICIDs []string
	var tags = &[]ecs.DescribeNetworkInterfacesTag{{Key: tagName, Value: tagValue}}
	input := ecs.CreateDescribeNetworkInterfacesRequest()
	input.Tag = tags

	result, err := sess.DescribeNetworkInterfaces(input)
	if err != nil {
		return NICIDs, err
	}
	for _, nic := range result.NetworkInterfaceSets.NetworkInterfaceSet {
		NICIDs = append(NICIDs, nic.NetworkInterfaceId)
	}
	return NICIDs, nil
}

func cleanOrphanResources(instanceIds []string, volumeIds []string, NICIds []string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) (delErrInstanceId []string, delErrVolumeIds []string, delErrNICs []string) {

	for _, instanceId := range instanceIds {
		if err := terminateInstance(instanceId, machineClass, secretData); err != nil {
			delErrInstanceId = append(delErrInstanceId, instanceId)
		}
	}

	for _, volumeId := range volumeIds {
		if err := deleteVolume(volumeId, machineClass, secretData); err != nil {
			delErrVolumeIds = append(delErrVolumeIds, volumeId)
		}
	}

	for _, nicId := range NICIds {
		if err := deleteNIC(nicId, machineClass, secretData); err != nil {
			delErrNICs = append(delErrNICs, nicId)
		}
	}

	return
}

func deleteNIC(nicId string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := &ecs.DeleteNetworkInterfaceRequest{NetworkInterfaceId: nicId}
	_, err := sess.DeleteNetworkInterface(input)
	if err != nil {
		return err
	}
	return nil
}

func deleteVolume(diskId string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := &ecs.DeleteDiskRequest{DiskId: diskId}
	_, err := sess.DeleteDisk(input)
	if err != nil {
		return err
	}
	return nil
}

func terminateInstance(instanceId string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	sess := newSession(machineClass, &v1.Secret{Data: secretData})
	input := &ecs.DeleteInstanceRequest{InstanceId: instanceId}
	_, err := sess.DeleteInstance(input)
	if err != nil {
		return err
	}
	return nil
}
