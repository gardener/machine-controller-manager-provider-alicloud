// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package provider contains the cloud provider specific implementations to manage machines
package alicloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alibabacloud-go/tea/tea"

	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
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
			Body: &ecs.RunInstancesResponseBody{
				InstanceIdSets: &ecs.RunInstancesResponseBodyInstanceIdSets{
					InstanceIdSet: []*string{
						tea.String(instanceID),
					},
				},
			},
		}

		deleteInstanceRequest = &ecs.DeleteInstanceRequest{
			InstanceId: tea.String(instanceID),
			Force:      tea.Bool(true),
		}
		deleteInstanceResponse = &ecs.DeleteInstanceResponse{}

		describeInstanceRequest = &ecs.DescribeInstancesRequest{
			InstanceIds: tea.String("[\"" + instanceID + "\"]"),
			RegionId:    tea.String(providerSpec.Region),
		}
		describeInstanceResponse = &ecs.DescribeInstancesResponse{
			Body: &ecs.DescribeInstancesResponseBody{
				TotalCount: tea.Int32(1),
				Instances: &ecs.DescribeInstancesResponseBodyInstances{
					Instance: []*ecs.DescribeInstancesResponseBodyInstancesInstance{
						{
							Status:       tea.String("Running"),
							InstanceId:   tea.String(instanceID),
							InstanceName: tea.String(machineName),
						},
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

		describeInstanceRequest = &ecs.DescribeInstancesRequest{
			InstanceIds: tea.String("[\"" + instanceID + "\"]"),
			RegionId:    tea.String(providerSpec.Region),
		}
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

	Describe("should delete machine successfully", func() {
		It("when machine.spec.providerID is set", func() {
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
				mockPluginSPI.EXPECT().NewDescribeInstancesRequest("", instanceID, providerSpec.Region, providerSpec.Tags).Return(describeInstanceRequest, nil),
				mockECSClient.EXPECT().DescribeInstances(describeInstanceRequest).Return(describeInstanceResponse, nil),
				mockPluginSPI.EXPECT().NewDeleteInstanceRequest(instanceID, true).Return(deleteInstanceRequest, nil),
				mockECSClient.EXPECT().DeleteInstance(deleteInstanceRequest).Return(deleteInstanceResponse, nil),
			)

			response, err := mockMachinePlugin.DeleteMachine(ctx, deleteMachineRequest)
			Expect(err).To(BeNil())
			Expect(response).To(Equal(deleteMachineResponse))
		})
		It("when machine.spec.providerID is not set", func() {
			var (
				deleteMachineRequest = &driver.DeleteMachineRequest{
					Machine:      machine,
					MachineClass: machineClass,
					Secret:       providerSecret,
				}
				deleteMachineResponse = &driver.DeleteMachineResponse{
					LastKnownState: "ECS instance(s) [i-mockinstanceid] deleted for machine mock-machine-name",
				}
			)

			gomock.InOrder(
				mockPluginSPI.EXPECT().NewECSClient(deleteMachineRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
				mockPluginSPI.EXPECT().NewDescribeInstancesRequest(deleteMachineRequest.Machine.Name, "", providerSpec.Region, providerSpec.Tags).Return(describeInstanceRequest, nil),
				mockECSClient.EXPECT().DescribeInstances(describeInstanceRequest).Return(describeInstanceResponse, nil),
				mockPluginSPI.EXPECT().NewDeleteInstanceRequest(instanceID, true).Return(deleteInstanceRequest, nil),
				mockECSClient.EXPECT().DeleteInstance(deleteInstanceRequest).Return(deleteInstanceResponse, nil),
			)

			deleteMachineRequest.Machine.Spec.ProviderID = ""
			response, err := mockMachinePlugin.DeleteMachine(ctx, deleteMachineRequest)
			Expect(err).To(BeNil())
			Expect(response).To(Equal(deleteMachineResponse))
			deleteMachineRequest.Machine.Spec.ProviderID = providerID //Need to add this value back as other tests are dependent on this
		})

		It("when machine.spec.providerID is not set and multiple instances exist across pages", func() {
			var (
				deleteMachineRequest = &driver.DeleteMachineRequest{
					Machine:      machine,
					MachineClass: machineClass,
					Secret:       providerSecret,
				}
				pageSize = 2
				// Expect all instances to be deleted
				deleteMachineResponse = &driver.DeleteMachineResponse{
					LastKnownState: fmt.Sprintf("ECS instance(s) %v deleted for machine %s",
						func() []string {
							ids := make([]string, 0, pageSize+1)
							for i := 0; i < pageSize; i++ {
								ids = append(ids, fmt.Sprintf("i-page1-%d", i))
							}
							ids = append(ids, "i-page2-0")
							return ids
						}(),
						machineName),
				}
			)

			page1Instances := make([]*ecs.DescribeInstancesResponseBodyInstancesInstance, pageSize)
			for i := 0; i < pageSize; i++ {
				id := fmt.Sprintf("i-page1-%d", i)
				name := fmt.Sprintf("machine-page1-%d", i)
				page1Instances[i] = &ecs.DescribeInstancesResponseBodyInstancesInstance{
					InstanceId:   tea.String(id),
					InstanceName: tea.String(name),
				}
			}
			page2Instances := []*ecs.DescribeInstancesResponseBodyInstancesInstance{
				{
					InstanceId:   tea.String("i-page2-0"),
					InstanceName: tea.String("machine-page2-0"),
				},
			}

			describeInstanceResponsePage1 := &ecs.DescribeInstancesResponse{
				Body: &ecs.DescribeInstancesResponseBody{
					NextToken: tea.String("token-page-2"),
					Instances: &ecs.DescribeInstancesResponseBodyInstances{
						Instance: page1Instances,
					},
				},
			}
			describeInstanceResponsePage2 := &ecs.DescribeInstancesResponse{
				Body: &ecs.DescribeInstancesResponseBody{
					NextToken: tea.String(""),
					Instances: &ecs.DescribeInstancesResponseBodyInstances{
						Instance: page2Instances,
					},
				},
			}

			gomock.InOrder(
				mockPluginSPI.EXPECT().NewECSClient(deleteMachineRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
				mockPluginSPI.EXPECT().NewDescribeInstancesRequest(deleteMachineRequest.Machine.Name, "", providerSpec.Region, providerSpec.Tags).Return(describeInstanceRequest, nil),

				mockECSClient.EXPECT().DescribeInstances(gomock.AssignableToTypeOf(describeInstanceRequest)).DoAndReturn(func(req *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error) {
					if req.NextToken != nil && *req.NextToken != "" {
						return nil, fmt.Errorf("expected empty NextToken, got %s", *req.NextToken)
					}
					return describeInstanceResponsePage1, nil
				}),
				mockECSClient.EXPECT().DescribeInstances(gomock.AssignableToTypeOf(describeInstanceRequest)).DoAndReturn(func(req *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error) {
					if req.NextToken == nil || *req.NextToken != "token-page-2" {
						return nil, fmt.Errorf("expected NextToken token-page-2, got %v", req.NextToken)
					}
					return describeInstanceResponsePage2, nil
				}),
			)

			for _, inst := range page1Instances {
				req := &ecs.DeleteInstanceRequest{InstanceId: inst.InstanceId, Force: tea.Bool(true)}
				mockPluginSPI.EXPECT().NewDeleteInstanceRequest(*inst.InstanceId, true).Return(req, nil)
				mockECSClient.EXPECT().DeleteInstance(req).Return(&ecs.DeleteInstanceResponse{}, nil)
			}
			for _, inst := range page2Instances {
				req := &ecs.DeleteInstanceRequest{InstanceId: inst.InstanceId, Force: tea.Bool(true)}
				mockPluginSPI.EXPECT().NewDeleteInstanceRequest(*inst.InstanceId, true).Return(req, nil)
				mockECSClient.EXPECT().DeleteInstance(req).Return(&ecs.DeleteInstanceResponse{}, nil)
			}

			deleteMachineRequest.Machine.Spec.ProviderID = ""
			response, err := mockMachinePlugin.DeleteMachine(ctx, deleteMachineRequest)
			Expect(err).To(BeNil())
			Expect(response).To(Equal(deleteMachineResponse))
			deleteMachineRequest.Machine.Spec.ProviderID = providerID
		})
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
			mockPluginSPI.EXPECT().NewDescribeInstancesRequest(getMachineStatusRequest.Machine.Name, "", providerSpec.Region, providerSpec.Tags).Return(describeInstanceRequest, nil),
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
			mockPluginSPI.EXPECT().NewDescribeInstancesRequest("", "", providerSpec.Region, providerSpec.Tags).Return(describeInstanceRequest, nil),
			mockECSClient.EXPECT().DescribeInstances(describeInstanceRequest).Return(describeInstanceResponse, nil),
		)

		response, err := mockMachinePlugin.ListMachines(ctx, listMachinesRequest)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(listMachinesResponse))
	})

	It("should list machines successfully across multiple pages", func() {
		var (
			listMachinesRequest = &driver.ListMachinesRequest{
				MachineClass: machineClass,
				Secret:       providerSecret,
			}
			pageSize       = 2
			page1Instances = make([]*ecs.DescribeInstancesResponseBodyInstancesInstance, pageSize)
			page2Instances = []*ecs.DescribeInstancesResponseBodyInstancesInstance{
				{
					InstanceId:   tea.String("i-page2-0"),
					InstanceName: tea.String("machine-page2-0"),
				},
			}
		)

		for i := range pageSize {
			id := fmt.Sprintf("i-page1-%d", i)
			name := fmt.Sprintf("machine-page1-%d", i)
			page1Instances[i] = &ecs.DescribeInstancesResponseBodyInstancesInstance{
				InstanceId:   tea.String(id),
				InstanceName: tea.String(name),
			}
		}

		var (
			describeInstanceResponsePage1 = &ecs.DescribeInstancesResponse{
				Body: &ecs.DescribeInstancesResponseBody{
					NextToken: tea.String("token-page-2"),
					Instances: &ecs.DescribeInstancesResponseBodyInstances{
						Instance: page1Instances,
					},
				},
			}
			describeInstanceResponsePage2 = &ecs.DescribeInstancesResponse{
				Body: &ecs.DescribeInstancesResponseBody{
					NextToken: tea.String(""),
					Instances: &ecs.DescribeInstancesResponseBodyInstances{
						Instance: page2Instances,
					},
				},
			}
		)

		expectedMachineList := make(map[string]string)
		for _, inst := range page1Instances {
			expectedMachineList[encodeProviderID(providerSpec.Region, *inst.InstanceId)] = *inst.InstanceName
		}
		for _, inst := range page2Instances {
			expectedMachineList[encodeProviderID(providerSpec.Region, *inst.InstanceId)] = *inst.InstanceName
		}
		listMachinesResponse := &driver.ListMachinesResponse{
			MachineList: expectedMachineList,
		}

		gomock.InOrder(
			mockPluginSPI.EXPECT().NewECSClient(listMachinesRequest.Secret, providerSpec.Region).Return(mockECSClient, nil),
			mockPluginSPI.EXPECT().NewDescribeInstancesRequest("", "", providerSpec.Region, providerSpec.Tags).Return(describeInstanceRequest, nil),

			mockECSClient.EXPECT().DescribeInstances(gomock.AssignableToTypeOf(describeInstanceRequest)).DoAndReturn(func(req *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error) {
				if req.NextToken != nil && *req.NextToken != "" {
					return nil, fmt.Errorf("expected empty NextToken, got %s", *req.NextToken)
				}
				return describeInstanceResponsePage1, nil
			}),

			mockECSClient.EXPECT().DescribeInstances(gomock.AssignableToTypeOf(describeInstanceRequest)).DoAndReturn(func(req *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error) {
				if req.NextToken == nil || *req.NextToken != "token-page-2" {
					return nil, fmt.Errorf("expected NextToken token-page-2, got %v", req.NextToken)
				}
				return describeInstanceResponsePage2, nil
			}),
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
