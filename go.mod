module github.com/keikoproj/iam-manager

go 1.16

require (
	github.com/aws/aws-sdk-go v1.25.38
	github.com/go-logr/logr v0.1.0
	github.com/golang/mock v1.6.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.8.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/sys v0.0.0-20220406163625-3f8b81556e12 // indirect
	golang.org/x/tools v0.1.10 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
	rsc.io/quote/v3 v3.1.0 // indirect
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/controller-tools v0.2.5 // indirect
)
