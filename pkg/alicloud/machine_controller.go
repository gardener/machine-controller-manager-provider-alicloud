// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package alicloud contains the cloud provider specific implementations to manage machines
package alicloud

import (
	"context"
	"fmt"

	maperror "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/errors"
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/klog/v2"
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
//
//	ProviderID typically matches with the node.Spec.ProviderID on the node object.
//	Eg: gce://project-name/region/vm-ProviderID
//
// NodeName              string                   Returns the name of the node-object that the VM register's with Kubernetes.
//
//	This could be different from req.MachineName as well
//
// LastKnownState        string                   (Optional) Last known state of VM during the current operation.
//
//	Could be helpful to continue operations in future requests.
//
// OPTIONAL IMPLEMENTATION LOGIC
// It is optionally expected by the safety controller to use an identification mechanisms to map the VM Created by a providerSpec.
// These could be done using tag(s)/resource-groups etc.
// This logic is used by safety controller to delete orphan VMs which are not backed by any machine CRD
func (plugin *MachinePlugin) CreateMachine(_ context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine creation request has been processed for %q", req.Machine.Name)

	// Check if provider in the MachineClass is the provider we support
	if req.MachineClass.Provider != ProviderAlicloud {
		err := fmt.Errorf("requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
		return nil, status.Error(maperror.GetMCMErrorCodeForCreateMachine(err), err.Error())
	}

	instanceID, err := GetInstanceIDFromRunInstancesResponse(response)
	if err != nil {
		errMessage := fmt.Sprintf("ECS instance creation failed for machine %s: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Internal, errMessage)
	}

	klog.V(2).Infof("ECS instance %q created for machine %q", *instanceID, req.Machine.Name)

	return &driver.CreateMachineResponse{
		ProviderID:     encodeProviderID(providerSpec.Region, *instanceID),
		NodeName:       instanceIDToName(*instanceID),
		LastKnownState: fmt.Sprintf("ECS instance %s created for machine %s", *instanceID, req.Machine.Name),
	}, nil
}

// InitializeMachine handles VM initialization for Alibaba Cloud VM's. Currently, un-implemented.
func (plugin *MachinePlugin) InitializeMachine(_ context.Context, _ *driver.InitializeMachineRequest) (*driver.InitializeMachineResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Alibaba Cloud Provider does not yet implement InitializeMachine")
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
//
//	Could be helpful to continue operations in future requests.
func (plugin *MachinePlugin) DeleteMachine(_ context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	// Check if provider in the MachineClass is the provider we support
	if req.MachineClass.Provider != ProviderAlicloud {
		err := fmt.Errorf("requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, err := plugin.SPI.NewECSClient(req.Secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	lastKnownState := ""

	if req.Machine.Spec.ProviderID != "" {
		instanceID := decodeProviderID(req.Machine.Spec.ProviderID)
		describeInstanceRequest, err := plugin.SPI.NewDescribeInstancesRequest("", instanceID, providerSpec.Region, providerSpec.Tags)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		instances, err := plugin.GetAllInstances(client, describeInstanceRequest)
		if err != nil {
			klog.Errorf("error while fetching instance details for instanceID %s: %v", instanceID, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		klog.V(3).Infof("Total %d instances found for instanceID %s", len(instances), instanceID)
		if len(instances) == 0 {
			// No running instance exists with the given machineID
			errMessage := fmt.Sprintf("ECS instance not found backing this machine object with Provider ID: %v", req.Machine.Spec.ProviderID)
			klog.Error(errMessage)

			return nil, status.Error(codes.NotFound, errMessage)
		}

		if *instances[0].Status != "Running" && *instances[0].Status != "Stopped" {
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
		lastKnownState = fmt.Sprintf("ECS instance %s deleted for machine %s", instanceID, req.Machine.Name)
	} else {
		klog.V(2).Infof("No provider ID set for machine %s. Checking if backing ECS instance is present.", req.Machine.Name)
		describeInstanceRequest, err := plugin.SPI.NewDescribeInstancesRequest(req.Machine.Name, "", providerSpec.Region, providerSpec.Tags)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		instances, err := plugin.GetAllInstances(client, describeInstanceRequest)
		if err != nil {
			klog.Errorf("error while fetching instance details for machine object %s: %v", req.Machine.Name, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		klog.V(3).Infof("Total %d instances found for machine %s", len(instances), req.Machine.Name)
		if len(instances) == 0 {
			// No running instance exists with the given machineName
			klog.V(2).Infof("No backing ECS instance found. Termination successful for machine object %q", req.Machine.Name)
			return &driver.DeleteMachineResponse{}, nil
		}

		var deletedInstances = make([]string, 0, len(instances))
		for _, instance := range instances {
			deleteInstanceRequest, err := plugin.SPI.NewDeleteInstanceRequest(*instance.InstanceId, true)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			_, err = client.DeleteInstance(deleteInstanceRequest)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			klog.V(3).Infof("ECS instance %s deleted for machine %s", *instance.InstanceId, *instance.InstanceName)
			deletedInstances = append(deletedInstances, *instance.InstanceId)
		}
		lastKnownState = fmt.Sprintf("ECS instance(s) %v deleted for machine %s", deletedInstances, req.Machine.Name)
	}

	return &driver.DeleteMachineResponse{
		LastKnownState: lastKnownState,
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
//
//	ProviderID typically matches with the node.Spec.ProviderID on the node object.
//	Eg: gce://project-name/region/vm-ProviderID
//
// NodeName             string                    Returns the name of the node-object that the VM register's with Kubernetes.
//
//	This could be different from req.MachineName as well
//
// The request should return a NOT_FOUND (5) status error code if the machine is not existing
func (plugin *MachinePlugin) GetMachineStatus(_ context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)

	// Check if provider in the MachineClass is the provider we support
	if req.MachineClass.Provider != ProviderAlicloud {
		err := fmt.Errorf("requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	request, err := plugin.SPI.NewDescribeInstancesRequest(req.Machine.Name, "", providerSpec.Region, providerSpec.Tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instances, err := plugin.GetAllInstances(client, request)
	if err != nil {
		klog.Errorf("error while fetching instance details for machine object %s: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.V(3).Infof("Total %d instances found for machine %s", len(instances), req.Machine.Name)
	if len(instances) == 0 {
		// No running instance exists with the given machineID
		klog.V(2).Infof("No matching instances found with %q", req.Machine.Name)
		errMessage := fmt.Sprintf("VM instance not found backing this machine object %v", req.Machine.Name)
		return nil, status.Error(codes.NotFound, errMessage)
	} else if len(instances) > 1 {
		var instanceIDs []string
		for _, instance := range instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}

		errMessage := fmt.Sprintf("multiple VM instances found backing this machine object. IDs for all backing VMs - %v ", instanceIDs)
		return nil, status.Error(codes.OutOfRange, errMessage)
	}

	klog.V(3).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)
	return &driver.GetMachineStatusResponse{
		NodeName:   instanceIDToName(*instances[0].InstanceId),
		ProviderID: encodeProviderID(providerSpec.Region, *instances[0].InstanceId),
	}, nil
}

// ListMachines lists all the machines possibly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc.
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.ListMachinesRequest)
// MachineClass          *v1alpha1.MachineClass   MachineClass based on which VMs created have to be listed
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//
//	for all machines which were possibly created by this ProviderSpec
func (plugin *MachinePlugin) ListMachines(_ context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been received for %q", req.MachineClass.Name)
	defer klog.V(2).Infof("List machines request has been received for %q", req.MachineClass.Name)

	// Check if provider in the MachineClass is the provider we support
	if req.MachineClass.Provider != ProviderAlicloud {
		err := fmt.Errorf("requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderAlicloud)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	client, err := plugin.SPI.NewECSClient(req.Secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request, err := plugin.SPI.NewDescribeInstancesRequest("", "", providerSpec.Region, providerSpec.Tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instances, err := plugin.GetAllInstances(client, request)
	if err != nil {
		klog.Errorf("error while fetching instance details for machines: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.V(3).Infof("Total %d instances found for listing machines for machine class %q", len(instances), req.MachineClass.Name)
	listOfMachines := make(map[string]string)
	for _, instance := range instances {
		machineName := *instance.InstanceName
		listOfMachines[encodeProviderID(providerSpec.Region, *instance.InstanceId)] = machineName
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
func (plugin *MachinePlugin) GetVolumeIDs(_ context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been received for %q", req.PVSpecs)
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
	klog.V(4).Infof("GetVolumeIDs: %v", volumeIDs)

	return &driver.GetVolumeIDsResponse{
		VolumeIDs: volumeIDs,
	}, nil
}
