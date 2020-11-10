/*
Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved.

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

package mock

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	corev1 "k8s.io/api/core/v1"
)

// PluginSPIMock is the mock implementation of PluginSPI
type PluginSPIMock struct{}

// NewECSClient returns a mock instance of the ECS client.
func (pluginSPI *PluginSPIMock) NewECSClient(secret *corev1.Secret, region string) (*ecs.Client, error) {
	return nil, nil
}

