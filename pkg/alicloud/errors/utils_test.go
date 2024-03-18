// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"testing"

	alierr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
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
		{inputAliErrorCode: OperationDeniedNoStock, expectedCode: codes.ResourceExhausted},
		// InvalidImageId can't be resolved by trying another zone, so not treated as ResourceExhausted
		{inputAliErrorCode: "InvalidImageId.NotFound", expectedCode: codes.Internal},
	}
	g := NewWithT(t)
	for _, entry := range table {
		jsonResponse, err := json.Marshal(responseContent{entry.inputAliErrorCode})
		g.Expect(err).To(BeNil())
		inputError := alierr.NewServerError(403, string(jsonResponse), "some error happened on the server side")
		g.Expect(GetMCMErrorCodeForCreateMachine(inputError)).To(Equal(entry.expectedCode))
	}
}
