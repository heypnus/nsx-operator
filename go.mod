module github.com/vmware-tanzu/nsx-operator

go 1.19

replace (
	github.com/vmware-tanzu/nsx-operator/pkg/apis => ./pkg/apis
	github.com/vmware/go-vmware-nsxt => github.com/mengdie-song/go-vmware-nsxt v0.0.0-20220921033217-dbd234470e30 // inventory-keeper includes all commits from github.com/gran-vmv/go-vmware-nsxt v0.0.0-20211206034609-cd7ffaf2c996
	github.com/vmware/vsphere-automation-sdk-go/lib => github.com/TaoZou1/vsphere-automation-sdk-go/lib v0.5.4
	github.com/vmware/vsphere-automation-sdk-go/runtime => github.com/TaoZou1/vsphere-automation-sdk-go/runtime v0.5.4
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt => github.com/TaoZou1/vsphere-automation-sdk-go/services/nsxt v0.9.5
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp => github.com/TaoZou1/vsphere-automation-sdk-go/services/nsxt-mp v0.3.1-0.20221020082725-84e89979deb6
)

require (
	github.com/agiledragon/gomonkey v2.0.2+incompatible
	github.com/agiledragon/gomonkey/v2 v2.9.0
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/coreos/go-semver v0.3.0
	github.com/deckarep/golang-set v1.8.0
	github.com/go-logr/logr v1.2.4
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.2.0
	github.com/kevinburke/ssh_config v1.2.0
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/vmware-tanzu/nsx-operator/pkg/apis v0.0.0-00010101000000-000000000000
	github.com/vmware-tanzu/vm-operator/api v1.8.2
	github.com/vmware/govmomi v0.27.4
	github.com/vmware/vsphere-automation-sdk-go/lib v0.5.2
	github.com/vmware/vsphere-automation-sdk-go/runtime v0.5.2
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt v0.11.0
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp v0.3.0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.11.0
	golang.org/x/time v0.3.0
	gopkg.in/ini.v1 v1.66.4
	k8s.io/api v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/client-go v0.26.1
	k8s.io/code-generator v0.26.1
	sigs.k8s.io/controller-runtime v0.14.5
)

require (
	github.com/beevik/etree v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gibson042/canonicaljson-go v1.0.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	gitlab.eng.vmware.com/vapi-sdk/vsphere-automation-sdk-go/runtime v0.7.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.13.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/term v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.26.1 // indirect
	k8s.io/component-base v0.26.1 // indirect
	k8s.io/gengo v0.0.0-20220902162205-c0856e24416d // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
