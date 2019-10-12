// package gorpc defines the Service and provides some wrappers to quickly start gorpc service.
package gorpc

import (
	"fmt"
	"github.com/hitzhangjie/go-rpc/config"
	"github.com/hitzhangjie/go-rpc/server"
	"os"
	"path/filepath"
	"strings"
)

// Service represents a service running somewhere, maybe deployed in multi-hosts or in Cloud.
//
// Service vs Server, these terms, as I see, they're two different views of running service.
// - Server, it's a running process or instance.
// - Service, it's deployed in public environment and provides service via `naming mechanism`.
//
// In go-rpc, you can start a `Server` via a `server.NewTcpServer()` or `server.NewUdpServer()`,
// If you want to register this service to remote naming service, you can use:
//
// 	method1:
//		```go
// 		gorpc.NewService()
//		```
//		`gorpc.NewService()` read configuration file service.ini and start `ServerModules` defined in `service.ini`.
//
// 	method2:
//		```go
// 		service := gorpc.NewService(name)
// 		service.RegisterServer(&server)
//		```
//
// method3:
//		```go
//		tcpSvr := NewTcpServer(...)
//		udpSvr := NewUdpServer(...)
//		service := gorpc.NewService(name)
//		service.RegisterModule(tcpSvr)
//		```
type Service struct {
	name    string
	version string
	server  *server.Server
}

// NewService create a new service
func NewService(name string) *Service {
	s := &Service{
		name:    name,
		version: "0.0.1",
		server:  nil,
	}
	return s
}

// Version set service version, each service can have serveral versions' api.
func (s *Service) Version(v string) *Service {
	s.version = v
	return s
}

func (s *Service) Handle(service interface{}) {
	// fixme service应该生成桩代码，里面定义好各个rpc名字与对应handler的映射关系
	// 类似于完成goneat中AddExec的操作！
	// 考虑不同业务协议的问题，可能有些业务协议使用的是int类型的cmd来区分接口，因此在协议之上还要抽象一个层，通过req体到rpc名字的映射，
	// func RpcName(req interface{}) string
}

func (s *Service) Start() error {

	if s.server == nil {
		return errServerNotInit
	}

	// start
	if err := s.server.Start(); err != nil {
		return err
	}
	// register
	if err := s.RegisterNaming(); err != nil {
		return err
	}
	// de-register
	<-s.server.Closed()
	if err := s.DeRegisterNaming(); err != nil {
		return err
	}

	return nil
}

func (s *Service) RegisterServer(svr *server.Server) {
	panic("implement me")
}

func (s *Service) RegisterSModule(mod *server.ServerModule) {
	panic("implement me")
}

func (s *Service) RegisterNaming() error {
	panic("implement me")
}

func (s *Service) DeRegisterNaming() error {
	panic("implement me")
}

// ListenAndServe quickly initialize Server and ServerModules and serve
func ListenAndServe(opts ...server.Option) {

	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fp := filepath.Join(d, "service.ini")
	cfg, err := config.LoadIniConfig(fp)
	if err != nil {
		panic(err)
	}

	svr, err := server.NewServer(opts...)
	if err != nil {
		panic(err)
	}

	for _, section := range cfg.Sections() {
		// enable support for protocols
		ok := strings.HasSuffix(section, "-service")
		if !ok {
			continue
		}
		codec := strings.TrimSuffix(section, "-service")

		// initialize tcp ServerModule
		tcpport := cfg.Int(section, "tcp.port", 0)
		if tcpport > 0 {
			mod, err := server.NewTcpServer("tcp4", fmt.Sprintf(":%d", tcpport), codec)
			if err != nil {
				panic(err)
			}
			mod.Register(svr)
		}

		// initialize udp ServerModule
		udpport := cfg.Int(section, "udp.port", 0)
		if udpport > 0 {
			mod, err := server.NewTcpServer("udp4", fmt.Sprintf(":%d", udpport), codec)
			if err != nil {
				panic(err)
			}
			mod.Register(svr)
		}
	}

	svr.Start()
}
