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

// Package provider contains the cloud provider specific implementations to manage machines
package alicloud

import (
	"context"
	"encoding/json"

	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/mock"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Machine Controller", func() {
	var (
		internetMaxBandwidthIn  = 5
		internetMaxBandwidthOut = 5
		providerSpec            = &api.ProviderSpec{
			APIVersion:              api.V1alpha1,
			ImageID:                 "m-uf6jf6utod2nfs9x21iwse",
			InstanceType:            "ecs.g6.large",
			Region:                  "cn-shanghai",
			ZoneID:                  "cn-shanghai-e",
			SecurityGroupID:         "sg-uf69t4txlz6r18ybzxbx",
			VSwitchID:               "vsw-uf6s1fjxxks65rk1tkrpm",
			InstanceChargeType:      "PostPaid",
			InternetChargeType:      "PayByTraffic",
			InternetMaxBandwidthIn:  &internetMaxBandwidthIn,
			InternetMaxBandwidthOut: &internetMaxBandwidthOut,
			SpotStrategy:            "NoSpot",
			KeyPairName:             "shoot-ssh-publickey",
			Tags: map[string]string{
				"kubernetes.io/cluster/shoot--mcm":     "1",
				"kubernetes.io/role/worker/shoot--mcm": "1",
			},
			SystemDisk: &api.AlicloudSystemDisk{
				Category: "cloud_efficiency",
				Size:     50,
			},
		}
		alicloudMachineClassSpec = &v1alpha1.AlicloudMachineClassSpec{
			ImageID:                 "m-uf6jf6utod2nfs9x21iwse",
			InstanceType:            "ecs.g6.large",
			Region:                  "cn-shanghai",
			ZoneID:                  "cn-shanghai-e",
			SecurityGroupID:         "sg-uf69t4txlz6r18ybzxbx",
			VSwitchID:               "vsw-uf6s1fjxxks65rk1tkrpm",
			InstanceChargeType:      "PostPaid",
			InternetChargeType:      "PayByTraffic",
			InternetMaxBandwidthIn:  &internetMaxBandwidthIn,
			InternetMaxBandwidthOut: &internetMaxBandwidthOut,
			SpotStrategy:            "NoSpot",
			KeyPairName:             "shoot-ssh-publickey",
			Tags: map[string]string{
				"kubernetes.io/cluster/shoot--mcm":     "1",
				"kubernetes.io/role/worker/shoot--mcm": "1",
			},
			SystemDisk: &v1alpha1.AlicloudSystemDisk{
				Category: "cloud_efficiency",
				Size:     50,
			},
		}
		providerSpecRaw, _ = json.Marshal(providerSpec)

		machineName      = "mock-machine-name"
		machineClassName = "mock-machine-class-name"

		providerId = "cn-shanghai.i-mockinstanceid"

		ctx               = context.Background()
		MachinePluginMock = NewAlicloudPlugin(&mock.PluginSPIMock{})
	)

	It("should create machine successfully", func() {
		var (
			createMachineRequest = driver.CreateMachineRequest{
				Machine: &v1alpha1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: machineName,
					},
				},
				MachineClass: &v1alpha1.MachineClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: machineClassName,
					},
					ProviderSpec: runtime.RawExtension{
						Raw: providerSpecRaw,
					},
				},
				Secret: &corev1.Secret{},
			}
			createMachineResponse = &driver.CreateMachineResponse{
				ProviderID:     "cn-shanghai.i-mockinstanceid",
				NodeName:       "izmockinstanceidz",
				LastKnownState: "ECS instance i-mockinstanceid created for machine mock-machine-name",
			}
		)

		response, err := MachinePluginMock.CreateMachine(ctx, &createMachineRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(createMachineResponse))
	})

	It("should delete machine successfully", func() {
		var (
			deleteMachineRequest = driver.DeleteMachineRequest{
				Machine: &v1alpha1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: machineName,
					},
					Spec: v1alpha1.MachineSpec{
						ProviderID: providerId,
					},
				},
				MachineClass: &v1alpha1.MachineClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: machineClassName,
					},
					ProviderSpec: runtime.RawExtension{
						Raw: providerSpecRaw,
					},
				},
				Secret: &corev1.Secret{},
			}
			deleteMachineResponse = &driver.DeleteMachineResponse{
				LastKnownState: "ECS instance i-mockinstanceid deleted for machine mock-machine-name",
			}
		)

		response, err := MachinePluginMock.DeleteMachine(ctx, &deleteMachineRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(deleteMachineResponse))
	})

	It("should migrate old machine class to the new one", func() {
		var (
			migrateMachineClassRequest = &driver.GenerateMachineClassForMigrationRequest{
				ProviderSpecificMachineClass: &v1alpha1.AlicloudMachineClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "machine-class-migration",
					},
					Spec: *alicloudMachineClassSpec,
				},
				MachineClass: &v1alpha1.MachineClass{},
				ClassSpec: &v1alpha1.ClassSpec{
					Kind: AlicloudMachineClassKind,
				},
			}

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "machine-class-migration",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
				Provider: ProviderAlicloud,
			}

			migrateMachineClassResponse = &driver.GenerateMachineClassForMigrationResponse{}
		)
		response, err := MachinePluginMock.GenerateMachineClassForMigration(ctx, migrateMachineClassRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(migrateMachineClassResponse))
		Expect(migrateMachineClassRequest.MachineClass).To(Equal(machineClass))
	})
})
