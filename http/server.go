package http

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/marsmay/golib/logger"
)

type TlsConfig struct {
	Enable   bool   `toml:"enable" json:"enable"`
	CertPath string `toml:"cert_path" json:"cert_path"`
	KeyPath  string `toml:"key_path" json:"key_path"`
}

type Config struct {
	Host    string     `toml:"host" json:"host"`
	Port    int        `toml:"port" json:"port"`
	Charset string     `toml:"charset" json:"charset"`
	Gzip    bool       `toml:"gzip" json:"gzip"`
	PProf   bool       `toml:"pprof" json:"pprof"`
	Tls     *TlsConfig `toml:"tls" json:"tls"`
}

func (c *Config) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

func DefaultConfig() *Config {
	return &Config{
		Port:    80,
		Charset: "UTF-8",
		Tls: &TlsConfig{
			Enable: false,
		},
	}
}

type Router interface {
	RegHttpHandler(app *gin.Engine)
	GetIdentifier(ctx *gin.Context) string
}

type Server struct {
	sync.Mutex
	config   *Config
	router   Router
	server   *http.Server
	logger   *logger.Logger
	ctx      context.Context
	canceler func()
}

func (s *Server) Running() bool {
	select {
	case <-s.ctx.Done():
		return false
	default:
		return true
	}
}

// recovery panic (500)
func (s *Server) Recovery(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			if ctx.IsAborted() {
				return
			}

			var stacktrace string

			for i := 1; ; i++ {
				_, f, l, got := runtime.Caller(i)

				if !got {
					break
				}

				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}

			request := fmt.Sprintf("%s | %s %s | %s | %s", ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL.RequestURI(), ctx.Request.UserAgent(), s.router.GetIdentifier(ctx))
			s.logger.Error(fmt.Sprintf("recovered panic:\nRequest: %s\nTrace: %s\n%s", request, err, stacktrace))

			ctx.Status(http.StatusInternalServerError)
			ctx.Abort()
		}
	}()

	ctx.Next()
}

// record access log
func (s *Server) AccessLog(ctx *gin.Context) {
	start := time.Now()
	ctx.Next()

	idf := s.router.GetIdentifier(ctx)
	statusCode, useTime, clientIp := ctx.Writer.Status(), time.Since(start), ctx.ClientIP()
	uri, method, userAgent := ctx.Request.URL.RequestURI(), ctx.Request.Method, ctx.Request.UserAgent()
	s.logger.Infof("request: %d | %4v | %s | %s %s | %s | %s", statusCode, useTime, clientIp, method, uri, userAgent, idf)
}

func CrossDomain(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Next()
}

func UnGzip(ctx *gin.Context) {
	ctx.Request.Header.Del("Accept-Encoding")
	ctx.Next()
}

func (s *Server) Start() {
	go func() {
		var err error

		if s.config.Tls.Enable {
			err = s.server.ListenAndServeTLS(s.config.Tls.CertPath, s.config.Tls.KeyPath)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed && s.Running() {
			s.logger.Errorf("can't serve at <%s> | error: %s", s.config.GetAddr(), err)
		}
	}()
}

func (s *Server) Stop() {
	s.canceler()

	ctx, canceler := context.WithTimeout(context.Background(), 3*time.Second)
	defer canceler()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorf("server shutdown error: %s", err)
	}
}

func NewServer(c *Config, r Router, l *logger.Logger) *Server {
	server := &Server{config: c, router: r, logger: l}
	server.ctx, server.canceler = context.WithCancel(context.Background())

	// set gin
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = l
	gin.DefaultErrorWriter = l

	// create gin instance
	router := gin.New()
	router.Use(server.Recovery)
	router.Use(server.AccessLog)

	// enable gzip
	if c.Gzip {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	// enable pprof
	if c.PProf {
		pprof.Register(router)
	}

	// set route
	server.router.RegHttpHandler(router)

	// set server
	server.server = &http.Server{
		Addr:    c.GetAddr(),
		Handler: router,
	}

	return server
}
