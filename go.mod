module github.com/fatedier/frp

go 1.23
require (
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/coreos/go-oidc/v3 v3.6.0
	github.com/fatedier/beego v0.0.0-20171024143340-6c6a4f5bd5eb
	github.com/fatedier/golib v0.1.1-0.20230725122706-dcbaee8eef40
	github.com/fatedier/kcp-go v2.0.4-0.20190803094908-fe8645b0a904+incompatible
	github.com/go-playground/validator/v10 v10.14.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/yamux v0.1.1
	github.com/onsi/ginkgo/v2 v2.21.0
	github.com/onsi/gomega v1.35.1
	github.com/pion/stun v0.6.1
	github.com/pires/go-proxyproto v0.7.0
	github.com/prometheus/client_golang v1.16.0
	github.com/quic-go/quic-go v0.37.4
	github.com/rodaine/table v1.1.0
	github.com/samber/lo v1.38.1
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/net v0.30.0
	golang.org/x/oauth2 v0.10.0
	golang.org/x/sync v0.8.0
	golang.org/x/time v0.7.0
	gopkg.in/ini.v1 v1.67.0
	k8s.io/apimachinery v0.32.0
	k8s.io/client-go v0.27.4
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-jose/go-jose/v3 v3.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.6 // indirect
	github.com/klauspost/reedsolomon v1.9.15 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/quic-go/qtls-go1-20 v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

// TODO(fatedier): Temporary use the modified version, update to the official version after merging into the official repository.
replace github.com/hashicorp/yamux => github.com/fatedier/yamux v0.0.0-20230628132301-7aca4898904d
