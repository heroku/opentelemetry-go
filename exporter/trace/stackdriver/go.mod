module go.opentelemetry.io/exporter/trace/stackdriver

go 1.12

replace go.opentelemetry.io => ../../..

require (
	cloud.google.com/go/trace v1.4.0
	github.com/golang/protobuf v1.5.2
	go.opentelemetry.io v0.0.0-20191021171549-9b5f5dd13acd
	golang.org/x/oauth2 v0.4.0
	google.golang.org/api v0.103.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
	google.golang.org/grpc v1.53.0
)
