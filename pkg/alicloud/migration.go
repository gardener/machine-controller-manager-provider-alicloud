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

package alicloud

import (
	"encoding/json"

	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ProviderAlicloud string const to identify Alicloud provider
	ProviderAlicloud = "Alicloud"
)

func migrateMachineClass(alicloudMachineClass *v1alpha1.AlicloudMachineClass, machineClass *v1alpha1.MachineClass) error {

	providerSpec := &api.ProviderSpec{
		APIVersion:              api.V1alpha1,
		ImageID:                 alicloudMachineClass.Spec.ImageID,
		InstanceType:            alicloudMachineClass.Spec.InstanceType,
		Region:                  alicloudMachineClass.Spec.Region,
		ZoneID:                  alicloudMachineClass.Spec.ZoneID,
		SecurityGroupID:         alicloudMachineClass.Spec.SecurityGroupID,
		VSwitchID:               alicloudMachineClass.Spec.VSwitchID,
		PrivateIPAddress:        alicloudMachineClass.Spec.PrivateIPAddress,
		InstanceChargeType:      alicloudMachineClass.Spec.InstanceChargeType,
		InternetChargeType:      alicloudMachineClass.Spec.InternetChargeType,
		InternetMaxBandwidthIn:  alicloudMachineClass.Spec.InternetMaxBandwidthIn,
		InternetMaxBandwidthOut: alicloudMachineClass.Spec.InternetMaxBandwidthOut,
		SpotStrategy:            alicloudMachineClass.Spec.SpotStrategy,
		IoOptimized:             alicloudMachineClass.Spec.IoOptimized,
		Tags:                    alicloudMachineClass.Spec.Tags,
		KeyPairName:             alicloudMachineClass.Spec.KeyPairName,
		SystemDisk: &api.AlicloudSystemDisk{
			Category: alicloudMachineClass.Spec.SystemDisk.Category,
			Size:     alicloudMachineClass.Spec.SystemDisk.Size,
		},
		DataDisks: []api.AlicloudDataDisk{},
	}

	for _, dataDisk := range alicloudMachineClass.Spec.DataDisks {
		providerSpec.DataDisks = append(providerSpec.DataDisks, api.AlicloudDataDisk{
			Name:               dataDisk.Name,
			Category:           dataDisk.Category,
			Description:        dataDisk.Description,
			Encrypted:          dataDisk.Encrypted,
			DeleteWithInstance: dataDisk.DeleteWithInstance,
			Size:               dataDisk.Size,
		})
	}

	providerSpecRaw, err := json.Marshal(providerSpec)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	machineClass.Name = alicloudMachineClass.Name
	machineClass.Labels = alicloudMachineClass.Labels
	machineClass.Annotations = alicloudMachineClass.Annotations
	machineClass.Finalizers = alicloudMachineClass.Finalizers
	machineClass.ProviderSpec = runtime.RawExtension{
		Raw: providerSpecRaw,
	}
	machineClass.SecretRef = alicloudMachineClass.Spec.SecretRef
	machineClass.CredentialsSecretRef = alicloudMachineClass.Spec.CredentialsSecretRef
	machineClass.Provider = ProviderAlicloud

	return nil
}
