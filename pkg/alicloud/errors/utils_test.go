// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"github.com/alibabacloud-go/tea/tea"
	"testing"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	. "github.com/onsi/gomega"
)

type input struct {
	inputAliErrorCode string
	expectedCode      codes.Code
}

type responseContent struct {
	Code string
}

func TestCreateMachineErrorToMCMErrorCode(t *testing.T) {
	table := []input{
		{inputAliErrorCode: QuotaExceededDiskCapacity, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: QuotaExceededElasticQuota, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: OperationDeniedCloudSSDNotSupported, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: OperationDeniedNoStock, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: OperationDeniedZoneNotAllowed, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: OperationDeniedZoneSystemCategoryNotMatch, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: ZoneNotOnSale, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: ZoneNotOpen, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: InvalidVpcZoneNotSupported, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: InvalidResourceTypeNotSupported, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: InvalidInstanceTypeZoneNotSupported, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: InvalidZoneIDNotSupportShareEncryptedImage, expectedCode: codes.ResourceExhausted},
		{inputAliErrorCode: ResourceNotAvailable, expectedCode: codes.ResourceExhausted},
		// InvalidImageId can't be resolved by trying another zone, so not treated as ResourceExhausted
		{inputAliErrorCode: "InvalidImageId.NotFound", expectedCode: codes.Internal},
	}
	g := NewWithT(t)
	for _, entry := range table {
		g.Expect(GetMCMErrorCodeForCreateMachine(tea.NewSDKError(map[string]any{
			"statusCode": 403,
			"code":       entry.inputAliErrorCode,
			"message":    "some error happened on the server side",
		}))).To(Equal(entry.expectedCode))
	}
}
