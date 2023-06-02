package tracing

type TracingConf interface {
	GetEndpoint() string
	GetCustomName() string
}
