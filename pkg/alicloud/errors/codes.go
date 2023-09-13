/*
Copyright (c) 2023 SAP SE or an SAP affiliate company. All rights reserved.
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
package errors

// constants for alicloud `RunInstances()` response error code which map to MCM `ResourceExhausted` code
const(
// QuotaExceededDiskCapacity: The used capacity of disk type has exceeded the quota in the zone
QuotaExceededDiskCapacity="QuotaExceed.DiskCapacity"
// QuotaExceededElasticQuota: The number of vCPUs assigned to the ECS instances has exceeded the quota in the zone. Please sign-on to Alibaba Cloud Console, and submit a quota increase application.
QuotaExceededElasticQuota="QuotaExceed.ElasticQuota"
// OperationDeniedZoneSystemCategoryNotMatch: The specified Zone or cluster does not offer the specified disk category or the specified zone and cluster do not match.
OperationDeniedZoneSystemCategoryNotMatch="OperationDenied.ZoneSystemCategoryNotMatch"
// OperationDeniedCloudSSDNotSupported: The specified available zone does not offer the cloud_ssd disk, use cloud_essd instead.
OperationDeniedCloudSSDNotSupported="OperationDenied.CloudSSDNotSupported"
// OperationDeniedZoneNotAllowed: The creation of Instance to the specified Zone is not allowed.
OperationDeniedZoneNotAllowed="OperationDenied.ZoneNotAllowed"
// OperationDeniedNoStock: The requested resource is sold out in the specified zone; try other types of resources or other regions and zones.
OperationDeniedNoStock="OperationDenied.NoStock"
// ZoneNotOnSale: The resource in the specified zone is no longer available for sale. Please try other regions and zones.
ZoneNotOnSale="Zone.NotOnSale"
// ZoneNotOpen: The specified zone is not granted to you to buy resources yet.
ZoneNotOpen="Zone.NotOpen"
// InvalidVpcZone.NotSupported: The specified operation is not allowed in the zone to which your VPC belongs, please try in other zones.
InvalidVpcZoneNotSupported="InvalidVpcZone.NotSupported"
// InvalidZoneIdNotSupportShareEncryptedImage: Creating instances by shared encrypted images is not supported in this zone.
InvalidZoneIdNotSupportShareEncryptedImage="InvalidZoneId.NotSupportShareEncryptedImage"
// InvalidInstanceType.ZoneNotSupported: The specified zone does not support this instancetype.
InvalidInstanceTypeZoneNotSupported="InvalidInstanceType.ZoneNotSupported"
// ResourceNotAvailable: Resource you requested is not available in this region or zone.
ResourceNotAvailable="ResourceNotAvailable"
)