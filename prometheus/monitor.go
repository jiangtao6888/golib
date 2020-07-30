package prometheus

import (
	stdCtx "context"
	"errors"
	"fmt"
	"github.com/Zivn/golib/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	TypeCounter = iota
	TypeGauge
	TypeHistogram
	TypeSummary
)

var (
	// DefaultBuckets prometheus buckets in seconds.
	DefaultBuckets = []float64{0.1, 0.3, 0.5, 1.0, 3.0, 5.0}
)

type VectorConfig struct {
	Name             string   `toml:"name" json:"name"`
	Desc             string   `toml:"desc" json:"desc"`
	Type             int      `toml:"type" json:"type"`
	Labels           []string `toml:"labels" json:"labels"`
	IgnoreConstLabel bool     `toml:"ignore_const_label" json:"ignore_const_label"`
}

type Config struct {
	ConstLabels map[string]string `toml:"const_labels" json:"const_labels"`
	Vectors     []*VectorConfig   `toml:"vectors" json:"vectors"`
}

type Vector struct {
	config *VectorConfig
	vec    prometheus.Collector
	logger *logger.Logger
}

func (v *Vector) Trigger(value float64, labels ...string) {
	if len(labels) != len(v.config.Labels) {
		v.logger.Errorf("invalid vector labels | name: %s | labels: %+v", v.config.Name, labels)
		return
	}

	switch v.config.Type {
	case TypeGauge:
		v.vec.(*prometheus.GaugeVec).WithLabelValues(labels...).Set(value)
	case TypeHistogram:
		v.vec.(*prometheus.HistogramVec).WithLabelValues(labels...).Observe(value)
	case TypeSummary:
		v.vec.(*prometheus.SummaryVec).WithLabelValues(labels...).Observe(value)
	case TypeCounter:
		v.vec.(*prometheus.CounterVec).WithLabelValues(labels...).Inc()
	}
}

// NOTE: vector type must is Histogram or Summary
func (v *Vector) HttpInterceptor(ctx context.Context) {
	start := time.Now()
	ctx.Next()

	r := ctx.Request()
	statusCode := strconv.Itoa(ctx.GetStatusCode())
	duration := float64(time.Since(start).Nanoseconds()) / 1000000000
	labels := []string{statusCode, r.Method, r.URL.Path}

	v.Trigger(duration, labels...)
}

func getClietIP(ctx stdCtx.Context) (ip string, err error) {
	pr, ok := peer.FromContext(ctx)

	if !ok {
		err = fmt.Errorf("invoke FromContext() failed")
		return
	}

	if pr.Addr == net.Addr(nil) {
		err = fmt.Errorf("peer.Addr is nil")
		return
	}

	ip = strings.Split(pr.Addr.String(), ":")[0]
	return
}

// NOTE: vector type must is Histogram or Summary
func (v *Vector) GrpcServerUnaryInterceptor(ctx stdCtx.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	clientIp := "unknown"

	if ip, e := getClietIP(ctx); e == nil {
		clientIp = ip
	}

	resp, err = handler(ctx, req)
	duration := float64(time.Since(start).Nanoseconds()) / 1000000000
	labels := []string{info.FullMethod, clientIp}

	v.Trigger(duration, labels...)
	return
}

// NOTE: vector type must is Histogram or Summary
func (v *Vector) GrpcServerStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	start := time.Now()
	clientIp := "unknown"

	if ip, e := getClietIP(stream.Context()); e == nil {
		clientIp = ip
	}

	err = handler(srv, stream)
	duration := float64(time.Since(start).Nanoseconds()) / 1000000000
	labels := []string{info.FullMethod, clientIp}

	v.Trigger(duration, labels...)
	return
}

// NOTE: vector type must is Histogram or Summary
func (v *Vector) GrpcClientUnaryInterceptor(ctx stdCtx.Context, method string, req, resp interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, options ...grpc.CallOption) (err error) {
	start := time.Now()
	err = invoker(ctx, method, req, resp, conn, options...)
	duration := float64(time.Since(start).Nanoseconds()) / 1000000000
	labels := []string{method, conn.Target()}

	v.Trigger(duration, labels...)
	return
}

// NOTE: vector type must is Histogram or Summary
func (v *Vector) GrpcClientStreamInterceptor(ctx stdCtx.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, options ...grpc.CallOption) (stream grpc.ClientStream, err error) {
	start := time.Now()
	stream, err = streamer(ctx, desc, conn, method, options...)
	duration := float64(time.Since(start).Nanoseconds()) / 1000000000
	labels := []string{method, conn.Target()}

	v.Trigger(duration, labels...)
	return
}

func (v *Vector) Config() *VectorConfig {
	return v.config
}

type Monitor struct {
	config  *Config
	vectors map[string]*Vector
	logger  *logger.Logger
}

func (m *Monitor) Register(config *VectorConfig) (err error) {
	constLabels := m.config.ConstLabels

	if config.IgnoreConstLabel {
		constLabels = nil
	}

	var vec prometheus.Collector

	switch config.Type {
	case TypeHistogram:
		vec = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
				Buckets:     DefaultBuckets,
			},
			config.Labels,
		)
	case TypeGauge:
		vec = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
			},
			config.Labels,
		)
	case TypeSummary:
		vec = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
			},
			config.Labels,
		)
	case TypeCounter:
		vec = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
			},
			config.Labels,
		)
	default:
		err = errors.New("invalid monitor type")
		return
	}

	m.vectors[config.Name] = &Vector{config: config, vec: vec, logger: m.logger}
	prometheus.MustRegister(vec)
	return
}

func (m *Monitor) Trigger(name string, value float64, labels ...string) {
	vector, ok := m.vectors[name]

	if !ok {
		m.logger.Errorf("unknown monitor vector | name: %s", name)
		return
	}

	vector.Trigger(value, labels...)
}

func (m *Monitor) Vector(name string) (vector *Vector) {
	vec, ok := m.vectors[name]

	if !ok {
		m.logger.Errorf("unknown monitor vector | name: %s", name)
		return
	}

	vector = vec
	return
}

func (m *Monitor) Metrics() context.Handler {
	return iris.FromStd(promhttp.Handler())
}

func New(config *Config, logger *logger.Logger) (monitor *Monitor, err error) {
	monitor = &Monitor{
		config:  config,
		vectors: make(map[string]*Vector, len(config.Vectors)),
		logger:  logger,
	}

	for _, c := range config.Vectors {
		if err = monitor.Register(c); err != nil {
			break
		}
	}

	return
}
