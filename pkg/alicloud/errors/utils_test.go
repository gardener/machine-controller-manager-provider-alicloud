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
