module security-group

go 1.13

require (
	github.com/antihax/optional v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	paas.unicom.cn/dcs-sdk v0.0.0
	sigs.k8s.io/controller-runtime v0.5.0
)

replace paas.unicom.cn/dcs-sdk => ../dcs-sdk
