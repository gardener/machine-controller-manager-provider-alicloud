// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package errors contains the error codes returned by Alicloud APIs
package errors

// constants for alicloud `RunInstances()` response error code which map to MCM `ResourceExhausted` code
const (
	// QuotaExceededDiskCapacity : The used capacity of disk type has exceeded the quota in the zone
	QuotaExceededDiskCapacity = "QuotaExceed.DiskCapacity"
	// QuotaExceededElasticQuota : The number of vCPUs assigned to the ECS instances has exceeded the quota in the zone. Please sign-on to Alibaba Cloud Console, and submit a quota increase application.
	QuotaExceededElasticQuota = "QuotaExceed.ElasticQuota"
	// OperationDeniedZoneSystemCategoryNotMatch : The specified Zone or cluster does not offer the specified disk category or the specified zone and cluster do not match.
	OperationDeniedZoneSystemCategoryNotMatch = "OperationDenied.ZoneSystemCategoryNotMatch"
	// OperationDeniedCloudSSDNotSupported : The specified available zone does not offer the cloud_ssd disk, use cloud_essd instead.
	OperationDeniedCloudSSDNotSupported = "OperationDenied.CloudSSDNotSupported"
	// OperationDeniedZoneNotAllowed : The creation of Instance to the specified Zone is not allowed.
	OperationDeniedZoneNotAllowed = "OperationDenied.ZoneNotAllowed"
	// OperationDeniedNoStock : The requested resource is sold out in the specified zone; try other types of resources or other regions and zones.
	OperationDeniedNoStock = "OperationDenied.NoStock"
	// ZoneNotOnSale : The resource in the specified zone is no longer available for sale. Please try other regions and zones.
	ZoneNotOnSale = "Zone.NotOnSale"
	// ZoneNotOpen : The specified zone is not granted to you to buy resources yet.
	ZoneNotOpen = "Zone.NotOpen"
	// InvalidVpcZoneNotSupported : The specified operation is not allowed in the zone to which your VPC belongs, please try in other zones.
	InvalidVpcZoneNotSupported = "InvalidVpcZone.NotSupported"
	// InvalidZoneIDNotSupportShareEncryptedImage : Creating instances by shared encrypted images is not supported in this zone.
	InvalidZoneIDNotSupportShareEncryptedImage = "InvalidZoneId.NotSupportShareEncryptedImage"
	// InvalidInstanceTypeZoneNotSupported : The specified zone does not support this instancetype.
	InvalidInstanceTypeZoneNotSupported = "InvalidInstanceType.ZoneNotSupported"
	// ResourceNotAvailable : Resource you requested is not available in this region or zone.
	ResourceNotAvailable = "ResourceNotAvailable"
)
