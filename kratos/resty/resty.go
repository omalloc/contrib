package resty

import (
	"crypto/tls"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

/**
 * go-resty/v2 支持 otel 链路跟踪的版本封装
 */

// TracerTransport 自定义 http-transport
type TracerTransport struct {
	*http.Transport

	debug       bool
	tracer      trace.Tracer
	propagation propagation.TextMapPropagator
}

// ClientOption 自定义属性
type ClientOption struct {
	propagators propagation.TextMapPropagator
	tracer      trace.Tracer
	timeout     time.Duration
	transport   *http.Transport
	debug       bool
}

type Option func(*ClientOption)

func WithTracer(tracer trace.Tracer) Option {
	return func(o *ClientOption) {
		o.tracer = tracer
	}
}
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return func(o *ClientOption) {
		o.propagators = propagators
	}
}
func WithTimeout(timeout time.Duration) Option {
	return func(o *ClientOption) {
		o.timeout = timeout
	}
}
func WithTransport(transport *http.Transport) Option {
	return func(o *ClientOption) {
		o.transport = transport
	}
}
func WithDebug(debug bool) Option {
	return func(o *ClientOption) {
		o.debug = debug
	}
}

func New(opts ...Option) *resty.Client {
	o := &ClientOption{
		propagators: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
		tracer: otel.GetTracerProvider().
			Tracer("go-resty", trace.WithInstrumentationVersion("0.1.0")),
		timeout: 30 * time.Second,
		debug:   false,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.transport == nil {
		o.transport = createTransport(o.timeout)
	}

	c := resty.New()
	c.Debug = o.debug
	c.SetTransport(&TracerTransport{
		Transport:   o.transport,
		propagation: o.propagators,
		tracer:      o.tracer,
		debug:       o.debug, // debug 状态同时附加到 client 中，用来控制是否记录 body 到链路跟踪
	})

	return c
}

func createTransport(timeout time.Duration) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:     false,
		DisableCompression:    false,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       1000,
	}
}

func (r *TracerTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	attrs := make([]attribute.KeyValue, 0)
	ctx, span := r.tracer.Start(req.Context(),
		fmt.Sprintf("%s %s%s", req.Method, req.Host, req.URL.Path),
		trace.WithSpanKind(trace.SpanKindClient))
	r.propagation.Inject(ctx, propagation.HeaderCarrier(req.Header))
	defer span.End()

	attrs = append(attrs, peerAttr(req.RemoteAddr)...)
	attrs = append(attrs, semconv.HTTPClientAttributesFromHTTPRequest(req)...)
	attrs = append(attrs, semconv.HTTPTargetKey.String(req.URL.Path))

	reqBody, err := req.GetBody()
	if err == nil && reqBody != nil {
		data, _ := io.ReadAll(reqBody)
		if len(data) > 0 {
			attrs = append(attrs, attribute.String("http.request.body", string(data)))
		}
	}

	req = req.WithContext(ctx)
	r.propagation.Inject(ctx, propagation.HeaderCarrier(req.Header))
	resp, err = r.Transport.RoundTrip(req)
	if err != nil {
		span.RecordError(err, trace.WithTimestamp(time.Now()))
		return
	}

	// 如果有错误，记录一下错误状态，4xx 5xx
	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, resp.Status)
	}

	if resp.Body != nil {
		respDump, _ := httputil.DumpResponse(resp, r.debug)
		if len(respDump) > 0 {
			attrs = append(attrs, attribute.String("http.response.data", string(respDump)))
		}
	}

	attrs = append(attrs, semconv.HTTPStatusCodeKey.Int(resp.StatusCode))
	span.SetAttributes(attrs...)

	return resp, nil
}

func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}
