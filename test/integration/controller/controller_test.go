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
/**
	Overview
		- Tests the provider specific Machine Controller
	Prerequisites
		- secret yaml file for the hyperscaler/provider passed as input
		- control cluster and target clusters kube-config passed as input (optional)
	BeforeSuite
		- Check and create control cluster and target clusters if required
		- Check and create crds ( machineclass, machines, machinesets and machinedeployment ) if required
		  using file available in kubernetes/crds directory of machine-controller-manager repo
		- Start the Machine Controller manager ( as goroutine )
		- apply secret resource for accesing the cloud provider service in the control cluster
		- Create machineclass resource from file available in kubernetes directory of provider specific repo in control cluster
	AfterSuite
		- Delete the control and target clusters // As of now we are reusing the cluster so this is not required

    Test: sameRegion Scheduling Strategy Test
        1) Create machine in same region/zone as target cluster and attach it to the cluster
           Expected Output
			 - should successfully attach the machine to the target cluster (new node added)
		2) Delete machine
			Expected Output
			 - should successfully delete the machine from the target cluster (less one node)
 **/

package controller_test

import (
	"github.com/gardener/machine-controller-manager-provider-alicloud/test/integration/provider"
	"github.com/gardener/machine-controller-manager/pkg/test/integration/common"
	. "github.com/onsi/ginkgo"
)

// the timeout is changed to accommodate for time taken by node-critical components to get ready. PR - https://github.com/gardener/machine-controller-manager/pull/778
var commons = common.NewIntegrationTestFramework(&provider.ResourcesTrackerImpl{}, 600)

var _ = BeforeSuite(commons.SetupBeforeSuite)

var _ = AfterSuite(commons.Cleanup)

var _ = Describe("Machine controllers test", func() {
	commons.BeforeEachCheck()
	commons.ControllerTests()
})
