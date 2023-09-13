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

// Package errors contains the error codes returned by Alicloud APIs
package errors

import (
	alierr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
)

// CreateMachineErrorToMCMErrorCode takes the error returned from the EC2API during the CreateMachine call and returns the corresponding MCM error code.
func CreateMachineErrorToMCMErrorCode(err error) codes.Code {
	aliErr, ok := err.(*alierr.ServerError)
	if ok {
		switch aliErr.ErrorCode() {
		case QuotaExceededDiskCapacity, QuotaExceededElasticQuota, OperationDeniedCloudSSDNotSupported, OperationDeniedNoStock, OperationDeniedZoneNotAllowed, OperationDeniedZoneSystemCategoryNotMatch, ZoneNotOnSale, ZoneNotOpen, InvalidVpcZoneNotSupported, InvalidInstanceTypeZoneNotSupported, InvalidZoneIDNotSupportShareEncryptedImage, ResourceNotAvailable:
			return codes.ResourceExhausted
		default:
			return codes.Internal
		}
	}
	return codes.Internal
}
