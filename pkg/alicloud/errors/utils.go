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
		case QuotaExceededDiskCapacity, QuotaExceededElasticQuota, OperationDeniedCloudSSDNotSupported, OperationDeniedNoStock, OperationDeniedZoneNotAllowed, OperationDeniedZoneSystemCategoryNotMatch, ZoneNotOnSale, ZoneNotOpen, InvalidVpcZoneNotSupported, InvalidInstanceTypeZoneNotSupported, InvalidZoneIdNotSupportShareEncryptedImage, ResourceNotAvailable:
			return codes.ResourceExhausted
		default:
			return codes.Internal
		}
	}
	return codes.Internal
}
