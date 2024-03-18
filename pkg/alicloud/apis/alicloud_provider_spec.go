// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package api

const (
	// V1alpha1 is the constant for API version of machine controller manager
	V1alpha1 = "mcm.gardener.cloud/v1alpha1"
)

// ProviderSpec is the spec to be used while parsing the calls.
type ProviderSpec struct {
	APIVersion              string              `json:"apiVersion,omitempty"`
	ImageID                 string              `json:"imageID"`
	InstanceType            string              `json:"instanceType"`
	Region                  string              `json:"region"`
	ZoneID                  string              `json:"zoneID,omitempty"`
	SecurityGroupID         string              `json:"securityGroupID,omitempty"`
	VSwitchID               string              `json:"vSwitchID"`
	PrivateIPAddress        string              `json:"privateIPAddress,omitempty"`
	SystemDisk              *AlicloudSystemDisk `json:"systemDisk,omitempty"`
	DataDisks               []AlicloudDataDisk  `json:"dataDisks,omitempty"`
	InstanceChargeType      string              `json:"instanceChargeType,omitempty"`
	InternetChargeType      string              `json:"internetChargeType,omitempty"`
	InternetMaxBandwidthIn  *int                `json:"internetMaxBandwidthIn,omitempty"`
	InternetMaxBandwidthOut *int                `json:"internetMaxBandwidthOut,omitempty"`
	SpotStrategy            string              `json:"spotStrategy,omitempty"`
	IoOptimized             string              `json:"IoOptimized,omitempty"`
	Tags                    map[string]string   `json:"tags,omitempty"`
	KeyPairName             string              `json:"keyPairName"`
}

// AlicloudDataDisk describes DataDisk for Alicloud.
type AlicloudDataDisk struct {
	Name               string `json:"name,omitEmpty"`
	Category           string `json:"category,omitEmpty"`
	Description        string `json:"description,omitEmpty"`
	Encrypted          bool   `json:"encrypted,omitEmpty"`
	DeleteWithInstance *bool  `json:"deleteWithInstance,omitEmpty"`
	Size               int    `json:"size,omitEmpty"`
}

// AlicloudSystemDisk describes SystemDisk for Alicloud.
type AlicloudSystemDisk struct {
	Category string `json:"category"`
	Size     int    `json:"size"`
}
