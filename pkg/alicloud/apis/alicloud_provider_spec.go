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
	Category  string `json:"category"`
	Size      int    `json:"size"`
	Encrypted bool   `json:"encrypted,omitEmpty"`
}
