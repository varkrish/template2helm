module github.com/redhat-cop/template2helm

go 1.13

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

require (
	github.com/csfreak/dc2deploy v1.0.3
	github.com/essentialkaos/go-simpleyaml/v2 v2.1.3 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/openshift/api v0.0.0-20221013123534-96eec44e1979
	github.com/openshift/library-go v0.0.0-20221111030555-73ed40c0a938
	github.com/spf13/cobra v1.6.1
	helm.sh/helm/v3 v3.6.1
	honnef.co/go/tools v0.0.1-2020.1.4
	k8s.io/api v0.25.4
	k8s.io/apimachinery v0.25.4
	k8s.io/client-go v0.25.0 // indirect
	k8s.io/utils v0.0.0-20221108210102-8e77b1f39fe2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/yaml v1.3.0
)
