module github.com/gardener/machine-controller-manager-provider-alicloud

go 1.15

require (
	github.com/aliyun/alibaba-cloud-sdk-go v0.0.0-20180828111155-cad214d7d71f
	github.com/gardener/machine-controller-manager v0.36.0
	github.com/golang/mock v1.4.4
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	k8s.io/api v0.16.8
	k8s.io/apimachinery v0.16.8
	k8s.io/component-base v0.16.8
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.16.8 // v0.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8 // v0.16.8
	k8s.io/apiserver => k8s.io/apiserver v0.16.8 // v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8 // v0.16.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.8 // v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8 // v0.16.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)
