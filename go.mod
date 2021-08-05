module github.com/kubeflow/common

go 1.16

require (
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/go-openapi/spec v0.20.3
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/code-generator v0.21.3
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
	sigs.k8s.io/controller-runtime v0.9.5
	volcano.sh/apis v1.2.0-k8s1.19.6
)

replace (
	k8s.io/api => k8s.io/api v0.19.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.9
	k8s.io/client-go => k8s.io/client-go v0.19.9
	k8s.io/code-generator => k8s.io/code-generator v0.19.9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.2
	github.com/prometheus/client_golang v1.11.0 => github.com/prometheus/client_golang v1.7.1
)
