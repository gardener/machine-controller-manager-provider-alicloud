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

// Package alicloud contains the alicloud provider specific implementations to manage machines
package alicloud

import (
	"github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
)

// MachinePlugin implements the driver.Driver
// It also implements the PluginSPI interface
type MachinePlugin struct {
	SPI spi.PluginSPI
}

// NewAlicloudPlugin returns a new Alicloud machine plugin.
func NewAlicloudPlugin(pluginSPI spi.PluginSPI) driver.Driver {
	return &MachinePlugin{
		SPI: pluginSPI,
	}
}
