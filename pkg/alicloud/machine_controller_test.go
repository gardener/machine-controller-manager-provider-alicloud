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

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	mockclient "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/mock/client"
	mockspi "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/mock/spi"
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
)

var _ = Describe("Machine Controller", func() {
	var (
		machineName      = "mock-machine-name"
		nodeName         = "izmockinstanceidz"
		machineClassName = "mock-machine-class-name"
		providerID       = "cn-shanghai.i-mockinstanceid"
		instanceID       = "i-mockinstanceid"

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
		providerSpecRaw, _ = json.Marshal(providerSpec)

		machine = &v1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name: machineName,
			},
			Spec: v1alpha1.MachineSpec{
				ProviderID: providerID,
			},
		}
		machineClass = &v1alpha1.MachineClass{
			Provider: ProviderAlicloud,
			ObjectMeta: metav1.ObjectMeta{
				Name: machineClassName,
			},
			ProviderSpec: runtime.RawExtension{
				Raw: providerSpecRaw,
			},
		}

		providerSecret = &corev1.Secret{}

		runInstancesRequest = &ecs.RunInstancesRequest{}
		runInstanceResponse = &ecs.RunInstancesResponse{
			InstanceIdSets: ecs.InstanceIdSets{
				InstanceIdSet: []string{
					instanceID,
				},
			},
		}

		deleteInstanceRequest = &ecs.DeleteInstanceRequest{
			InstanceId: instanceID,
			Force:      requests.NewBoolean(true),
		}
		deleteInstanceResponse = &ecs.DeleteInstanceResponse{}

		describeInstanceRequest = &ecs.DescribeInstancesRequest{
			InstanceIds: "[\"" + instanceID + "\"]",
		}
		describeInstanceResponse = &ecs.DescribeInstancesResponse{
			Instances: ecs.Instances{
				Instance: []ecs.Instance{
					{
						Status:       "Running",
						InstanceId:   instanceID,
						InstanceName: machineName,
					},
				},
			},
		}

		ctx               = context.Background()
		ctrl              *gomock.Controller
		mockPluginSPI     *mockspi.MockPluginSPI
		mockECSClient     *mockclient.MockECSClient
		mockMachinePlugin driver.Driver
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPluginSPI = mockspi.NewMockPluginSPI(ctrl)
		mockECSClient = mockclient.NewMockECSClient(ctrl)
		mockMachinePlugin = NewAlicloudPlugin(mockPluginSPI)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should create machine successfully", func() {
		var (
			createMachineRequest = driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       providerSecret,
			}
			createMachineResponse = &driver.CreateMachineResponse{
				ProviderID:     providerID,
				NodeName:       nodeName,
				LastKnownState: "ECS instance i-mockinstanceid created for machine mock-machine-name",
			}
		)

		gomock.InOrder(
			mockPluginSPI.EXPECT().NewECSClient(createMachineRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
			mockPluginSPI.EXPECT().NewRunInstancesRequest(providerSpec, createMachineRequest.Machine.Name, createMachineRequest.Secret.Data[spi.AlicloudUserData]).Return(runInstancesRequest, nil),
			mockECSClient.EXPECT().RunInstances(runInstancesRequest).Return(runInstanceResponse, nil),
		)

		response, err := mockMachinePlugin.CreateMachine(ctx, &createMachineRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(createMachineResponse))
	})

	It("should delete machine successfully", func() {
		var (
			deleteMachineRequest = &driver.DeleteMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       providerSecret,
			}
			deleteMachineResponse = &driver.DeleteMachineResponse{
				LastKnownState: "ECS instance i-mockinstanceid deleted for machine mock-machine-name",
			}
		)

		gomock.InOrder(
			mockPluginSPI.EXPECT().NewECSClient(deleteMachineRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
			mockPluginSPI.EXPECT().NewDescribeInstancesRequest("", instanceID, providerSpec.Tags).Return(describeInstanceRequest, nil),
			mockECSClient.EXPECT().DescribeInstances(describeInstanceRequest).Return(describeInstanceResponse, nil),
			mockPluginSPI.EXPECT().NewDeleteInstanceRequest(instanceID, true).Return(deleteInstanceRequest, nil),
			mockECSClient.EXPECT().DeleteInstance(deleteInstanceRequest).Return(deleteInstanceResponse, nil),
		)

		response, err := mockMachinePlugin.DeleteMachine(ctx, deleteMachineRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(deleteMachineResponse))
	})

	It("should get machine status successfully", func() {
		var (
			getMachineStatusRequest = &driver.GetMachineStatusRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       providerSecret,
			}
			getMahineStatusResponse = &driver.GetMachineStatusResponse{
				ProviderID: providerID,
				NodeName:   nodeName,
			}
		)

		gomock.InOrder(
			mockPluginSPI.EXPECT().NewECSClient(getMachineStatusRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
			mockPluginSPI.EXPECT().NewDescribeInstancesRequest(getMachineStatusRequest.Machine.Name, "", providerSpec.Tags).Return(describeInstanceRequest, nil),
			mockECSClient.EXPECT().DescribeInstances(describeInstanceRequest).Return(describeInstanceResponse, nil),
		)

		response, err := mockMachinePlugin.GetMachineStatus(ctx, getMachineStatusRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(getMahineStatusResponse))
	})

	It("should list machines successfully", func() {
		var (
			listMachinesRequest = &driver.ListMachinesRequest{
				MachineClass: machineClass,
				Secret:       providerSecret,
			}
			listMachinesResponse = &driver.ListMachinesResponse{
				MachineList: map[string]string{
					providerID: machineName,
				},
			}
		)

		gomock.InOrder(
			mockPluginSPI.EXPECT().NewECSClient(listMachinesRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
			mockPluginSPI.EXPECT().NewDescribeInstancesRequest("", "", providerSpec.Tags).Return(describeInstanceRequest, nil),
			mockECSClient.EXPECT().DescribeInstances(describeInstanceRequest).Return(describeInstanceResponse, nil),
		)

		response, err := mockMachinePlugin.ListMachines(ctx, listMachinesRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(listMachinesResponse))
	})

	It("should get volume IDs successfully", func() {
		var (
			getVolumeIDsRequest = &driver.GetVolumeIDsRequest{
				PVSpecs: []*corev1.PersistentVolumeSpec{
					{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							FlexVolume: &corev1.FlexPersistentVolumeSource{
								Driver: "alicloud/disk",
								FSType: "ext4",
								Options: map[string]string{
									"volumeId": "d-mockflexvolumeid",
								},
							},
						},
					},
					{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver:       spi.AlicloudDriverName,
								FSType:       "ext4",
								VolumeHandle: "d-mockcsivolumeid",
							},
						},
					},
				},
			}
		)

		response, err := mockMachinePlugin.GetVolumeIDs(ctx, getVolumeIDsRequest)
		Expect(err).To(BeNil())
		Expect(response.VolumeIDs).To(ConsistOf("d-mockflexvolumeid", "d-mockcsivolumeid"))
	})
})
