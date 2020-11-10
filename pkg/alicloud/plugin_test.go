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
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
)

var _ = Describe("Machine Plugin", func(){
	var (
		providerSpec = &api.ProviderSpec{
			ImageID: "m-uf6jf6utod2nfs9x21iwse",
			InstanceType: "ecs.g6.large",
			Region: "cn-shanghai",
			ZoneID: "cn-shanghai-e",
			SecurityGroupID: "sg-uf69t4txlz6r18ybzxbx",
			VSwitchID: "vsw-uf6s1fjxxks65rk1tkrpm",
			InstanceChargeType: "PostPaid",
			InternetChargeType: "PayByTraffic",
			InternetMaxBandwidthIn: pointer.Int32Ptr(5),
			SpotStrategy: "NoSpot",
			KeyPairName: "shoot-ssh-publickey",
			Tags: map[string]string {
				"kubernetes.io/cluster/shoot--mcm": "1",
				"kubernetes.io/role/worker/shoot--mcm": "1",
			},
			SystemDisk: &api.AlicloudSystemDisk{
				Category: "cloud_efficiency",
				Size: int32(50),
			},
		}
		providerID = "cn-shanghai.i-u66kfxzhu3q9vm3l4a"
		machineName = "plugin-test-machine"
		userData = []byte("plugin-test-userdata")

		alicloudDataDisks = []api.AlicloudDataDisk{
			{
				Name: "disk-1",
				Category: "DiskEphemeralSSD",
				Encrypted: true,
				Size: int32(50),
			},
			{
				Name: "disk-2",
				DeleteWithInstance: pointer.BoolPtr(false),
				Encrypted: true,
				Size: int32(100),
			},
			{
				Name: "disk-3",
				Encrypted: false,
				Size: int32(20),
			},
		}

		plugin PluginSPIImpl
	)

	BeforeSuite(func(){
		plugin = PluginSPIImpl{}
	})

	It("should generate request of running instance", func(){
		request, err := plugin.NewRunInstancesRequest(providerSpec, machineName, userData)
		Expect(err).To(BeNil())
		Expect(request.SystemDiskCategory).To(Equal("cloud_efficiency"))
		Expect(request.SystemDiskSize).To(Equal("50"))
		Expect(request.DataDisk).To(BeNil())
		Expect(request.Tag).To(Equal(&[]ecs.RunInstancesTag{
			{
				Key: "kubernetes.io/cluster/shoot--mcm",
				Value: "1",
			},
			{
				Key: "kubernetes.io/role/worker/shoot--mcm",
				Value: "1",
			},
		}))
	})

	It("should generate request of describing instance by machine Name", func(){
		request, err := plugin.NewDescribeInstancesRequest(machineName, "", nil)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(Equal("plugin-test-machine"))
		Expect(request.InstanceIds).To(BeEmpty())
		Expect(request.Tag).To(BeNil())
	})

	It("should generate request of describing instance by provider ID", func(){
		request, err := plugin.NewDescribeInstancesRequest("", providerID, nil)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(BeEmpty())
		Expect(request.InstanceIds).To(Equal("[\"i-u66kfxzhu3q9vm3l4a\"]"))
		Expect(request.Tag).To(BeNil())
	})

	It("should generate request of describing instance by tags", func(){
		request, err := plugin.NewDescribeInstancesRequest("", "", providerSpec.Tags)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(BeEmpty())
		Expect(request.InstanceIds).To(BeEmpty())
		Expect(request.Tag).To(Equal(&[]ecs.DescribeInstancesTag{
			{
				Key: "kubernetes.io/cluster/shoot--mcm",
				Value: "1",
			},
			{
				Key: "kubernetes.io/role/worker/shoot--mcm",
				Value: "1",
			},
		}))
	})

	It("should generate request of deleting instance", func(){
		instanceID := decodeProviderID(providerID)
		request, err := plugin.NewDeleteInstanceRequest(instanceID, true)
		Expect(err).To(BeNil())
		Expect(request.InstanceId).To(Equal("i-u66kfxzhu3q9vm3l4a"))
		Expect(request.Force.GetValue()).To(Equal(true))
	})

	It("should generate instance data disks", func(){
		dataDisks := plugin.NewInstanceDataDisks(alicloudDataDisks, machineName)
		Expect(dataDisks).NotTo(BeEmpty())
		Expect(dataDisks).To(Equal([]ecs.RunInstancesDataDisk{
			{
				Category:    "DiskEphemeralSSD",
				Encrypted:   "true",
				DiskName:    "plugin-test-machine-disk-1-data-disk",
				Size:        "50",
				DeleteWithInstance: "",
			},
			{
				Encrypted:   "true",
				DiskName:    "plugin-test-machine-disk-2-data-disk",
				Size:        "100",
				DeleteWithInstance: "false",
			},
			{
				Encrypted:   "false",
				DiskName:    "plugin-test-machine-disk-3-data-disk",
				Size:        "20",
				DeleteWithInstance: "true",
			},
		}))
	})
})