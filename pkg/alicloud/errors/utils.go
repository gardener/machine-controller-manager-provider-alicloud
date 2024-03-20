// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package errors contains the error codes returned by Alicloud APIs
package errors

import (
	alierr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
)

// GetMCMErrorCodeForCreateMachine takes the error returned from the EC2API during the CreateMachine call and returns the corresponding MCM error code.
func GetMCMErrorCodeForCreateMachine(err error) codes.Code {
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
