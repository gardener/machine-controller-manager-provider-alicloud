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

// Package alicloud contains the cloud provider specific implementations to manage machines
package alicloud

import (
	"context"
	"fmt"

	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/klog"
)

// NOTE
//
// The basic working of the controller will work with just implementing the CreateMachine() & DeleteMachine() methods.
// You can first implement these two methods and check the working of the controller.
// Leaving the other methods to NOT_IMPLEMENTED error status.
// Once this works you can implement the rest of the methods.
//
// Also make sure each method return appropriate errors mentioned in `https://github.com/gardener/machine-controller-manager/blob/master/docs/development/machine_error_codes.md`

// CreateMachine handles a machine creation request
// REQUIRED METHOD
//
// REQUEST PARAMETERS (driver.CreateMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be created
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.CreateMachineResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                                ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                                Eg: gce://project-name/region/vm-ProviderID
// NodeName              string                   Returns the name of the node-object that the VM register's with Kubernetes.
//                                                This could be different from req.MachineName as well
// LastKnownState        string                   (Optional) Last known state of VM during the current operation.
//                                                Could be helpful to continue operations in future requests.
//
// OPTIONAL IMPLEMENTATION LOGIC
// It is optionally expected by the safety controller to use an identification mechanisms to map the VM Created by a providerSpec.
// These could be done using tag(s)/resource-groups etc.
// This logic is used by safety controller to delete orphan VMs which are not backed by any machine CRD
//
func (plugin *MachinePlugin) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine creation request has been processed for %q", req.Machine.Name)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderAlicloud {
		return nil, fmt.Errorf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, err := plugin.SPI.NewECSClient(req.Secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request, err := plugin.SPI.NewRunInstancesRequest(providerSpec, req.Machine.Name, req.Secret.Data[spi.AlicloudUserData])
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response, err := client.RunInstances(request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instanceID := response.InstanceIdSets.InstanceIdSet[0]

	klog.V(2).Infof("ECS instance %q created for machine %q", instanceID, req.Machine.Name)

	return &driver.CreateMachineResponse{
		ProviderID:     encodeProviderID(providerSpec.Region, instanceID),
		NodeName:       instanceIDToName(instanceID),
		LastKnownState: fmt.Sprintf("ECS instance %s created for machine %s", instanceID, req.Machine.Name),
	}, nil
}

// DeleteMachine handles a machine deletion request
//
// REQUEST PARAMETERS (driver.DeleteMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be deleted
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.DeleteMachineResponse)
// LastKnownState        bytes(blob)              (Optional) Last known state of VM during the current operation.
//                                                Could be helpful to continue operations in future requests.
//
func (plugin *MachinePlugin) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderAlicloud {
		return nil, fmt.Errorf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, err := plugin.SPI.NewECSClient(req.Secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instanceID := decodeProviderID(req.Machine.Spec.ProviderID)
	describeInstanceRequest, err := plugin.SPI.NewDescribeInstancesRequest("", instanceID, providerSpec.Tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response, err := client.DescribeInstances(describeInstanceRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instances := response.Instances.Instance
	if len(instances) == 0 {
		// No running instance exists with the given machineID
		errMessage := fmt.Sprintf("ECS instance not found backing this machine object with Provider ID: %v", req.Machine.Spec.ProviderID)
		klog.V(2).Infof(errMessage)

		return nil, status.Error(codes.NotFound, errMessage)
	}

	if instances[0].Status != "Running" && instances[0].Status != "Stopped" {
		return nil, status.Error(codes.Unavailable, "ECS instance not in running/stopped state")
	}

	deleteInstanceRequest, err := plugin.SPI.NewDeleteInstanceRequest(instanceID, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = client.DeleteInstance(deleteInstanceRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &driver.DeleteMachineResponse{
		LastKnownState: fmt.Sprintf("ECS instance %s deleted for machine %s", instanceID, req.Machine.Name),
	}, nil
}

// GetMachineStatus handles a machine get status request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.GetMachineStatusRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM status needs to be returned
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.GetMachineStatueResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                                ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                                Eg: gce://project-name/region/vm-ProviderID
// NodeName             string                    Returns the name of the node-object that the VM register's with Kubernetes.
//                                                This could be different from req.MachineName as well
//
// The request should return a NOT_FOUND (5) status error code if the machine is not existing
func (plugin *MachinePlugin) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderAlicloud {
		return nil, fmt.Errorf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
	}

	klog.V(2).Infof("Machine name found with %q", req.Machine.Name)
	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, err := plugin.SPI.NewECSClient(req.Secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request, err := plugin.SPI.NewDescribeInstancesRequest(req.Machine.Name, "", providerSpec.Tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instances := response.Instances.Instance
	if len(instances) == 0 {
		// No running instance exists with the given machineID
		klog.V(2).Infof("No matching instances found with %q", req.Machine.Name)
		errMessage := fmt.Sprintf("VM instance not found backing this machine object %v", req.Machine.Name)
		return nil, status.Error(codes.NotFound, errMessage)
	} else if len(instances) > 1 {
		instanceIDs := []string{}
		for _, instance := range instances {
			instanceIDs = append(instanceIDs, instance.InstanceId)
		}

		errMessage := fmt.Sprintf("multiple VM instances found backing this machine object. IDs for all backing VMs - %v ", instanceIDs)
		return nil, status.Error(codes.OutOfRange, errMessage)
	}

	klog.V(3).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)
	return &driver.GetMachineStatusResponse{
		NodeName:   instanceIDToName(instances[0].InstanceId),
		ProviderID: encodeProviderID(providerSpec.Region, instances[0].InstanceId),
	}, nil
}

// ListMachines lists all the machines possibilly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.ListMachinesRequest)
// MachineClass          *v1alpha1.MachineClass   MachineClass based on which VMs created have to be listed
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//                                           for all machine's who where possibilly created by this ProviderSpec
//
func (plugin *MachinePlugin) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been recieved for %q", req.MachineClass.Name)
	defer klog.V(2).Infof("List machines request has been recieved for %q", req.MachineClass.Name)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderAlicloud {
		return nil, fmt.Errorf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, err := plugin.SPI.NewECSClient(req.Secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request, err := plugin.SPI.NewDescribeInstancesRequest("", "", providerSpec.Tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instances := response.Instances.Instance
	listOfMachines := make(map[string]string)
	for _, instance := range instances {
		machineName := instance.InstanceName
		listOfMachines[encodeProviderID(providerSpec.Region, instance.InstanceId)] = machineName
	}

	return &driver.ListMachinesResponse{
		MachineList: listOfMachines,
	}, nil
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// REQUEST PARAMETERS (driver.GetVolumeIDsRequest)
// PVSpecList            []*corev1.PersistentVolumeSpec       PVSpecsList is a list PV specs for whom volume-IDs are required.
//
// RESPONSE PARAMETERS (driver.GetVolumeIDsResponse)
// VolumeIDs             []string                             VolumeIDs is a repeated list of VolumeIDs.
//
func (plugin *MachinePlugin) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been recieved for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	var volumeIDs []string
	for i := range req.PVSpecs {
		pvSpec := req.PVSpecs[i]
		if pvSpec.FlexVolume != nil && pvSpec.FlexVolume.Options != nil {
			if volumeID, ok := pvSpec.FlexVolume.Options["volumeId"]; ok {
				volumeIDs = append(volumeIDs, volumeID)
			}
		} else if pvSpec.CSI != nil && pvSpec.CSI.Driver == spi.AlicloudDriverName && pvSpec.CSI.VolumeHandle != "" {
			volumeIDs = append(volumeIDs, pvSpec.CSI.VolumeHandle)
		}
	}

	klog.V(2).Infof("GetVolumeIDs machines request has been processed successfully (%d/%d).", len(volumeIDs), len(req.PVSpecs))
	klog.V(4).Infof("GetVolumeIDs volumneIDs: %v", volumeIDs)

	return &driver.GetVolumeIDsResponse{
		VolumeIDs: volumeIDs,
	}, nil
}

// GenerateMachineClassForMigration helps in migration of one kind of machineClass CR to another kind.
// For instance an machineClass custom resource of `AWSMachineClass` to `MachineClass`.
// Implement this functionality only if something like this is desired in your setup.
// If you don't require this functionality leave is as is. (return Unimplemented)
//
// The following are the tasks typically expected out of this method
// 1. Validate if the incoming classSpec is valid one for migration (e.g. has the right kind).
// 2. Migrate/Copy over all the fields/spec from req.ProviderSpecificMachineClass to req.MachineClass
// For an example refer
//		https://github.com/prashanth26/machine-controller-manager-provider-gcp/blob/migration/pkg/gcp/machine_controller.go#L222-L233
//
// REQUEST PARAMETERS (driver.GenerateMachineClassForMigration)
// ProviderSpecificMachineClass    interface{}                             ProviderSpecificMachineClass is provider specfic machine class object (E.g. AWSMachineClass). Typecasting is required here.
// MachineClass 				   *v1alpha1.MachineClass                  MachineClass is the machine class object that is to be filled up by this method.
// ClassSpec                       *v1alpha1.ClassSpec                     Somemore classSpec details useful while migration.
//
// RESPONSE PARAMETERS (driver.GenerateMachineClassForMigration)
// NONE
//
func (plugin *MachinePlugin) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been recieved for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderAlicloud {
		return nil, fmt.Errorf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
	}

	alicloudMachineClass := req.ProviderSpecificMachineClass.(*v1alpha1.AlicloudMachineClass)
	if req.ClassSpec.Kind != AlicloudMachineClassKind {
		return nil, status.Error(codes.Internal, "Migration cannot be done for this machineClass kind")
	}

	return &driver.GenerateMachineClassForMigrationResponse{}, migrateMachineClass(alicloudMachineClass, req.MachineClass)
}
