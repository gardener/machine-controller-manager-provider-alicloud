// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

//go:generate mockgen -package=client -destination=mocks.go github.com/gardener/machine-controller-manager-provider-alicloud/pkg/spi PluginSPI

package spi
