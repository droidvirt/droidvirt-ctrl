module github.com/droidvirt/droidvirt-ctrl

go 1.14

require (
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/go-openapi/spec v0.19.4
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.18.2
	k8s.io/gengo v0.0.0-20200114144118-36b2048a9120
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	kubevirt.io/client-go v0.33.0
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/controller-tools v0.3.0
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2

replace github.com/docker/docker v0.0.0-00010101000000-000000000000 => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
