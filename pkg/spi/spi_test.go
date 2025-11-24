// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package spi

import (
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
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
		Expect(*request.SystemDisk.Category).To(Equal("cloud_efficiency"))
		Expect(*request.SystemDisk.Size).To(Equal("50"))
		Expect(request.DataDisk).To(BeNil())
		Expect(request.Tag).To(ConsistOf(
			&ecs.RunInstancesRequestTag{
				Key:   tea.String("kubernetes.io/cluster/shoot--mcm"),
				Value: tea.String("1"),
			},
			&ecs.RunInstancesRequestTag{
				Key:   tea.String("kubernetes.io/role/worker/shoot--mcm"),
				Value: tea.String("1"),
			},
		))
	})

	It("should generate request of describing instance by machine Name", func() {
		request, err := pluginSPI.NewDescribeInstancesRequest(machineName, "", nil)
		Expect(err).To(BeNil())
		Expect(*request.InstanceName).To(Equal("plugin-test-machine"))
		Expect(request.InstanceIds).To(BeNil())
		Expect(request.Tag).To(BeNil())
	})

	It("should generate request of describing instance by provider ID", func() {
		request, err := pluginSPI.NewDescribeInstancesRequest("", instanceID, nil)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(BeNil())
		Expect(*request.InstanceIds).To(Equal("[\"i-u66kfxzhu3q9vm3l4a\"]"))
		Expect(request.Tag).To(BeNil())
	})

	It("should generate request of describing instance by tags", func() {
		request, err := pluginSPI.NewDescribeInstancesRequest("", "", providerSpec.Tags)
		Expect(err).To(BeNil())
		Expect(request.InstanceName).To(BeNil())
		Expect(request.InstanceIds).To(BeNil())
		Expect(request.Tag).To(ConsistOf(
			&ecs.DescribeInstancesRequestTag{
				Key:   tea.String("kubernetes.io/cluster/shoot--mcm"),
				Value: tea.String("1"),
			},
			&ecs.DescribeInstancesRequestTag{
				Key:   tea.String("kubernetes.io/role/worker/shoot--mcm"),
				Value: tea.String("1"),
			},
		))
	})

	It("should generate request of deleting instance", func() {
		request, err := pluginSPI.NewDeleteInstanceRequest(instanceID, true)
		Expect(err).To(BeNil())
		Expect(*request.InstanceId).To(Equal("i-u66kfxzhu3q9vm3l4a"))
		Expect(*request.Force).To(Equal(true))
	})

	It("should generate instance data disks", func() {
		dataDisks := pluginSPI.NewInstanceDataDisks(alicloudDataDisks, machineName)
		Expect(dataDisks).NotTo(BeEmpty())
		Expect(dataDisks).To(ConsistOf(
			&ecs.RunInstancesRequestDataDisk{
				Category:           tea.String("DiskEphemeralSSD"),
				Encrypted:          tea.String("true"),
				DiskName:           tea.String("plugin-test-machine-disk-1-data-disk"),
				Size:               tea.Int32(int32(50)),
				DeleteWithInstance: nil,
				Description:        tea.String(""),
			},
			&ecs.RunInstancesRequestDataDisk{
				Category:           tea.String(""),
				Encrypted:          tea.String("true"),
				DiskName:           tea.String("plugin-test-machine-disk-2-data-disk"),
				Size:               tea.Int32(int32(100)),
				DeleteWithInstance: tea.Bool(false),
				Description:        tea.String(""),
			},
			&ecs.RunInstancesRequestDataDisk{
				Category:           tea.String(""),
				Encrypted:          tea.String("false"),
				DiskName:           tea.String("plugin-test-machine-disk-3-data-disk"),
				Size:               tea.Int32(20),
				DeleteWithInstance: tea.Bool(true),
				Description:        tea.String(""),
			},
		))
	})
})
