// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
