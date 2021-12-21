package observer

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
	"github.com/marsmay/golib/logger"
	"github.com/marsmay/golib/time2"
	"github.com/marsmay/golib/zookeeper"
)

const (
	SchemaHttp = "http"
	SchemaGrpc = "grpc"
)

type Service struct {
	Schema string `toml:"schema" json:"schema"`
	Name   string `toml:"name" json:"name"`
}

func (s *Service) GetName() string {
	return strings.Join([]string{s.Schema, s.Name}, ":")
}

func (s *Service) GetPath(basePath string) string {
	return strings.Join([]string{basePath, s.Schema, s.Name}, "/")
}

func (s *Service) String() string {
	return fmt.Sprintf("%+v", *s)
}

type Server struct {
	*Service
	Id      string `json:"id"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Ssl     bool   `json:"ssl"`
	RegTime int64  `json:"reg_time"`
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

func (s *Server) GetPath(basePath string) string {
	return strings.Join([]string{basePath, s.Schema, s.Name, s.Id}, "/")
}

func (s *Server) String() string {
	return fmt.Sprintf("%+v", *s)
}

func NewServer(schema, name, host string, port int, ssl bool) *Server {
	return &Server{
		Service: &Service{Schema: schema, Name: name},
		Id:      uuid.New().String(),
		Host:    host,
		Port:    port,
		Ssl:     ssl,
		RegTime: time2.NowMS(),
	}
}

type Config struct {
	ServicePath   string     `toml:"service_path" json:"service_path"`
	WatchServices []*Service `toml:"watch_services" json:"watch_services"`
}

type Observer struct {
	c            *Config
	zkClient     *zookeeper.Client
	logger       *logger.Logger
	locker       sync.RWMutex
	watchServers map[string][]*Server
	regServers   map[string]*Server
}

func (o *Observer) Register(server *Server) {
	path := server.GetPath(o.c.ServicePath)

	o.locker.Lock()
	o.regServers[path] = server
	o.locker.Unlock()

	o.zkClient.WatchNode(path, zk.EventNodeDeleted, func(event zk.Event) {
		o.locker.RLock()
		s, ok := o.regServers[path]
		o.locker.RUnlock()

		if !ok {
			return
		}

		value, _ := json.Marshal(s)
		err := o.zkClient.Create(path, value, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))

		if err != nil {
			o.logger.Errorf("register service failed | path: %s | data: %s ｜ error: %s", path, value, err)
			return
		}

		o.logger.Debugf("register service | path: %s | data: %s", path, value)
	})
}

func (o *Observer) Destroy() {
	o.locker.Lock()
	servers := o.regServers
	o.regServers = make(map[string]*Server, 4)
	o.locker.Unlock()

	for path := range servers {
		_ = o.zkClient.Delete(path)
	}
}

func (o *Observer) GetServer(schema, name string) (server *Server) {
	o.locker.RLock()
	defer o.locker.RUnlock()

	service := &Service{schema, name}
	servers := o.watchServers[service.GetName()]

	if len(servers) > 0 {
		server = servers[rand.Intn(len(servers))]
	}

	return
}

func (o *Observer) renew(event zk.Event) {
	items := strings.Split(event.Path, "/")

	if len(items) < 3 {
		o.logger.Errorf("invalid path | path: %s", event.Path)
		return
	}

	service := &Service{items[len(items)-2], items[len(items)-1]}
	datas, err := o.zkClient.GetNodes(event.Path)

	if err != nil {
		o.logger.Errorf("get child nodes failed | path: %s | error: %s", event.Path, err)
		return
	}

	servers := make([]*Server, 0, 16)

	for _, data := range datas {
		server := &Server{}

		if err := json.Unmarshal(data, server); err != nil {
			o.logger.Warningf("decode node data failed | data: %s | error: %s", data, err)
			continue
		}

		servers = append(servers, server)
	}

	o.locker.Lock()
	o.watchServers[service.GetName()] = servers
	o.locker.Unlock()

	o.logger.Debugf("renew services | services: %+v", o.watchServers)
}

func New(c *Config, zkClient *zookeeper.Client, logger *logger.Logger) (client *Observer, err error) {
	client = &Observer{
		c:            c,
		zkClient:     zkClient,
		logger:       logger,
		watchServers: make(map[string][]*Server, 16),
		regServers:   make(map[string]*Server, 4),
	}

	for _, service := range client.c.WatchServices {
		client.renew(zk.Event{
			Type: zk.EventNodeChildrenChanged,
			Path: service.GetPath(c.ServicePath),
		})
	}

	for _, service := range client.c.WatchServices {
		zkClient.WatchChildren(service.GetPath(c.ServicePath), zookeeper.EventTypeAll, client.renew)
	}

	return
}
