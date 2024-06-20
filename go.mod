module github.com/pentops/j5

go 1.22.3

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.34.2-20240508200655-46a4cf4ba109.2
	github.com/aws/aws-sdk-go-v2/config v1.27.18
	github.com/aws/aws-sdk-go-v2/service/ecr v1.28.5
	github.com/aws/aws-sdk-go-v2/service/s3 v1.55.1
	github.com/bufbuild/protocompile v0.14.0
	github.com/bufbuild/protoyaml-go v0.1.9
	github.com/docker/docker v26.1.4+incompatible
	github.com/go-git/go-git/v5 v5.12.0
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/iancoleman/strcase v0.2.0
	github.com/jhump/protoreflect v1.16.0
	github.com/pentops/flowtest v0.0.0-20240623020118-02bf89f30b87
	github.com/pentops/log.go v0.0.0-20240612184720-79634792a21e
	github.com/pentops/prototools v0.0.0-20240614222903-e9c266e2e96b
	github.com/pentops/runner v0.0.0-20240525192419-d655233635e9
	github.com/ryanuber/go-glob v1.0.0
	github.com/shopspring/decimal v1.4.0
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.17.1
	go.opentelemetry.io/otel/trace v1.27.0
	golang.org/x/mod v0.18.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240610135401-a8a62080eff3
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
)

require (
	buf.build/gen/go/bufbuild/buf/grpc/go v1.4.0-20240517182809-0b7e724f50eb.1 // indirect
	buf.build/gen/go/bufbuild/buf/protocolbuffers/go v1.34.2-20240517182809-0b7e724f50eb.2 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.27.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.12 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/bufbuild/protovalidate-go v0.6.2 // indirect
	github.com/cloudflare/circl v1.3.9 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/cel-go v0.20.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.27.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240610135401-a8a62080eff3 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.1 // indirect
)
