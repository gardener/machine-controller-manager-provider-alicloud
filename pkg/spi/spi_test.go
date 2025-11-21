// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package spi

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"

	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
)

var pluginSPI PluginSPIImpl
var _ = BeforeSuite(func() {
	pluginSPI = PluginSPIImpl{}
})

var _ = Describe("Plugin SPI", func() {

	var (
		internetMaxBandwidthIn  = 5
		internetMaxBandwidthOut = 5
		providerSpec            = &api.ProviderSpec{
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
		instanceID  = "i-u66kfxzhu3q9vm3l4a"
		machineName = "plugin-test-machine"
		userData    = []byte("plugin-test-userdata")

		alicloudDataDisks = []api.AlicloudDataDisk{
			{
				Name:      "disk-1",
				Category:  "DiskEphemeralSSD",
				Encrypted: true,
				Size:      50,
			},
			{
				Name:               "disk-2",
				DeleteWithInstance: pointer.BoolPtr(false),
				Encrypted:          true,
				Size:               100,
			},
			{
				Name:      "disk-3",
				Encrypted: false,
				Size:      20,
			},
		}
	)

	It("should generate request of running instance", func() {
		request, err := pluginSPI.NewRunInstancesRequest(providerSpec, machineName, userData)
		Expect(err).To(BeNil())
		Expect(request.SystemDiskCategory).To(Equal("cloud_efficiency"))
		Expect(request.SystemDiskSize).To(Equal("50"))
		Expect(request.DataDisk).To(BeNil())
		Expect(*request.Tag).To(ConsistOf(
			ecs.RunInstancesTag{
				Key:   "kubernetes.io/cluster/shoot--mcm",
				Value: "1",
			}, ecs.RunInstancesTag{
				Key:   "kubernetes.io/role/worker/shoot--mcm",
				Value: "1",
			},
		))
	})

	It("should generate request of describing instance by machine Name", func() {
		request, err := pluginSPI.NewDescribeInstancesRequest(machineName, "", nil)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(Equal("plugin-test-machine"))
		Expect(request.InstanceIds).To(BeEmpty())
		Expect(request.Tag).To(BeNil())
	})

	It("should generate request of describing instance by provider ID", func() {
		request, err := pluginSPI.NewDescribeInstancesRequest("", instanceID, nil)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(BeEmpty())
		Expect(request.InstanceIds).To(Equal("[\"i-u66kfxzhu3q9vm3l4a\"]"))
		Expect(request.Tag).To(BeNil())
	})

	It("should generate request of describing instance by tags", func() {
		request, err := pluginSPI.NewDescribeInstancesRequest("", "", providerSpec.Tags)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(BeEmpty())
		Expect(request.InstanceIds).To(BeEmpty())
		Expect(*request.Tag).To(ConsistOf(
			ecs.DescribeInstancesTag{
				Key:   "kubernetes.io/cluster/shoot--mcm",
				Value: "1",
			}, ecs.DescribeInstancesTag{
				Key:   "kubernetes.io/role/worker/shoot--mcm",
				Value: "1",
			},
		))
	})

	It("should fetch all instances from a single page", func() {
		mockClient := &mockECSClient{
			responses: []*ecs.DescribeInstancesResponse{
				{
					TotalCount: 2,
					Instances: ecs.Instances{
						Instance: []ecs.Instance{
							{
								InstanceId:   "i-instance1",
								InstanceName: "machine-1",
								Status:       "Running",
							},
							{
								InstanceId:   "i-instance2",
								InstanceName: "machine-2",
								Status:       "Running",
							},
						},
					},
				},
			},
		}

		instances, err := pluginSPI.DescribeAllInstances(mockClient, "", "", providerSpec.Tags)
		Expect(err).To(BeNil())
		Expect(instances).To(HaveLen(2))
		Expect(instances[0].InstanceId).To(Equal("i-instance1"))
		Expect(instances[1].InstanceId).To(Equal("i-instance2"))
		Expect(mockClient.callCount).To(Equal(1))
	})

	It("should fetch all instances from multiple pages", func() {
		mockClient := &mockECSClient{
			responses: []*ecs.DescribeInstancesResponse{
				{
					TotalCount: 250,
					Instances: ecs.Instances{
						Instance: make([]ecs.Instance, 100),
					},
				},
				{
					TotalCount: 250,
					Instances: ecs.Instances{
						Instance: make([]ecs.Instance, 100),
					},
				},
				{
					TotalCount: 250,
					Instances: ecs.Instances{
						Instance: make([]ecs.Instance, 50),
					},
				},
			},
		}

		instances, err := pluginSPI.DescribeAllInstances(mockClient, "", "", providerSpec.Tags)
		Expect(err).To(BeNil())
		Expect(instances).To(HaveLen(250))
		Expect(mockClient.callCount).To(Equal(3))
	})

	It("should handle empty results", func() {
		mockClient := &mockECSClient{
			responses: []*ecs.DescribeInstancesResponse{
				{
					TotalCount: 0,
					Instances: ecs.Instances{
						Instance: []ecs.Instance{},
					},
				},
			},
		}

		instances, err := pluginSPI.DescribeAllInstances(mockClient, "non-existent-machine", "", nil)
		Expect(err).To(BeNil())
		Expect(instances).To(BeEmpty())
		Expect(mockClient.callCount).To(Equal(1))
	})

	It("should return error when API call fails", func() {
		mockClient := &mockECSClient{
			err: fmt.Errorf("API Error: failed to describe instances"),
		}

		instances, err := pluginSPI.DescribeAllInstances(mockClient, "", "", providerSpec.Tags)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("API Error"))
		Expect(instances).To(BeNil())
		Expect(mockClient.callCount).To(Equal(1))
	})

	It("should generate request of deleting instance", func() {
		request, err := pluginSPI.NewDeleteInstanceRequest(instanceID, true)
		Expect(err).To(BeNil())
		Expect(request.InstanceId).To(Equal("i-u66kfxzhu3q9vm3l4a"))
		Expect(request.Force.GetValue()).To(Equal(true))
	})

	It("should generate instance data disks", func() {
		dataDisks := pluginSPI.NewInstanceDataDisks(alicloudDataDisks, machineName)
		Expect(dataDisks).NotTo(BeEmpty())
		Expect(dataDisks).To(ConsistOf(
			ecs.RunInstancesDataDisk{
				Category:           "DiskEphemeralSSD",
				Encrypted:          "true",
				DiskName:           "plugin-test-machine-disk-1-data-disk",
				Size:               "50",
				DeleteWithInstance: "",
			}, ecs.RunInstancesDataDisk{
				Encrypted:          "true",
				DiskName:           "plugin-test-machine-disk-2-data-disk",
				Size:               "100",
				DeleteWithInstance: "false",
			}, ecs.RunInstancesDataDisk{
				Encrypted:          "false",
				DiskName:           "plugin-test-machine-disk-3-data-disk",
				Size:               "20",
				DeleteWithInstance: "true",
			},
		))
	})
})
