module github.com/keikoproj/iam-manager

go 1.12

require (
	github.com/aws/aws-sdk-go v1.25.38
	github.com/go-logr/logr v0.1.0
	github.com/golang/mock v1.4.0
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/pborman/uuid v0.0.0-20170612153648-e790cca94e6c
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/sys v0.0.0-20190422165155-953cdadca894 // indirect
	golang.org/x/tools v0.0.0-20200221224223-e1da425f72fd // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.3.0
	sigs.k8s.io/controller-runtime v0.2.2
)
