// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package alicloud

import (
	"encoding/json"
	"fmt"
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-alicloud/pkg/alicloud/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
)

const (
	// ProviderAlicloud string const to identify Alicloud provider
	ProviderAlicloud = "Alicloud"
)

func decodeProviderSpec(machineClass *v1alpha1.MachineClass) (*api.ProviderSpec, error) {
	var providerSpec *api.ProviderSpec
	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}

func encodeProviderID(region, instanceID string) string {
	return fmt.Sprintf("%s.%s", region, instanceID)
}

func decodeProviderID(providerID string) string {
	splitProviderID := strings.Split(providerID, ".")
	return splitProviderID[len(splitProviderID)-1]
}

// Host name in Alicloud has relationship with Instance ID
// i-uf69zddmom11ci7est12 => izuf69zddmom11ci7est12z
func instanceIDToName(instanceID string) string {
	return strings.Replace(instanceID, "-", "z", 1) + "z"
}
