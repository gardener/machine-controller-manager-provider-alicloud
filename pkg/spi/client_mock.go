package spi

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

// mockECSClient is a simple mock for testing DescribeAllInstances
type mockECSClient struct {
	responses []*ecs.DescribeInstancesResponse
	callCount int
	err       error
}

func (m *mockECSClient) RunInstances(_ *ecs.RunInstancesRequest) (*ecs.RunInstancesResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DescribeInstances(request *ecs.DescribeInstancesRequest) (*ecs.DescribeInstancesResponse, error) {
	m.callCount++

	if m.err != nil {
		return nil, m.err
	}

	// Get page number from request (defaults to 1 if not set)
	pageIndex := 0
	if pageNum, err := request.PageNumber.GetValue(); err == nil {
		pageIndex = pageNum - 1
	}

	if pageIndex >= 0 && pageIndex < len(m.responses) {
		return m.responses[pageIndex], nil
	}

	return &ecs.DescribeInstancesResponse{}, nil
}

func (m *mockECSClient) DeleteInstance(_ *ecs.DeleteInstanceRequest) (*ecs.DeleteInstanceResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DescribeDisks(_ *ecs.DescribeDisksRequest) (*ecs.DescribeDisksResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DeleteDisk(_ *ecs.DeleteDiskRequest) (*ecs.DeleteDiskResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DescribeNetworkInterfaces(_ *ecs.DescribeNetworkInterfacesRequest) (*ecs.DescribeNetworkInterfacesResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DeleteNetworkInterface(_ *ecs.DeleteNetworkInterfaceRequest) (*ecs.DeleteNetworkInterfaceResponse, error) {
	return nil, nil
}
