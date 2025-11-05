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

func (m *mockECSClient) RunInstances(request *ecs.RunInstancesRequest) (*ecs.RunInstancesResponse, error) {
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

func (m *mockECSClient) DeleteInstance(request *ecs.DeleteInstanceRequest) (*ecs.DeleteInstanceResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DescribeDisks(request *ecs.DescribeDisksRequest) (*ecs.DescribeDisksResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DeleteDisk(request *ecs.DeleteDiskRequest) (*ecs.DeleteDiskResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DescribeNetworkInterfaces(request *ecs.DescribeNetworkInterfacesRequest) (*ecs.DescribeNetworkInterfacesResponse, error) {
	return nil, nil
}

func (m *mockECSClient) DeleteNetworkInterface(request *ecs.DeleteNetworkInterfaceRequest) (*ecs.DeleteNetworkInterfaceResponse, error) {
	return nil, nil
}
