// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package errors contains the error codes returned by Alicloud APIs
package errors

import (
	"errors"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
)

// GetMCMErrorCodeForCreateMachine takes the error returned from the EC2API during the CreateMachine call and returns the corresponding MCM error code.
func GetMCMErrorCodeForCreateMachine(err error) codes.Code {
	var aliErr *tea.SDKError
	ok := errors.As(err, &aliErr)
	if ok {
		switch *aliErr.Code {
		case QuotaExceededDiskCapacity, QuotaExceededElasticQuota, OperationDeniedCloudSSDNotSupported, OperationDeniedNoStock, OperationDeniedZoneNotAllowed, OperationDeniedZoneSystemCategoryNotMatch, ZoneNotOnSale, ZoneNotOpen, InvalidVpcZoneNotSupported, InvalidInstanceTypeZoneNotSupported, InvalidZoneIDNotSupportShareEncryptedImage, ResourceNotAvailable:
			return codes.ResourceExhausted
		default:
			return codes.Internal
		}
	}
	return codes.Internal
}
