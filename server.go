package sima

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"reflect"
	"strings"
)

type SimaNilT struct{}

type rpcServer struct {
	server *rpc.Server
	objs   []string
	conf   *config
}

func newRpcServer() *rpcServer {
	r := &rpcServer{
		server: rpc.DefaultServer,
		objs:   make([]string, 0),
		conf:   makeConfig(), // conf remains fixed after this point
	}
	r.register(&SimaRpc{})
	return r
}

var defaultServer = newRpcServer()

func (r *rpcServer) register(obj interface{}) {
	element := reflect.TypeOf(obj).Elem()
	r.objs = append(r.objs, element.Name())
	r.server.Register(obj)
}

type connection interface {
	addr() string
	retries() int
}

type tcp int

func (t *tcp) addr() string {
	if *t < 1024 {
		// Only use unprivileged ports
		*t = 1023
	}

	*t = *t + 1
	return fmt.Sprintf("%d", *t)
}

func (t *tcp) retries() int {
	return 500
}

type unix struct{}

func (u *unix) addr() string {
	// TODO: Add a directory
	return randstr(8)
}

func (u *unix) retries() int {
	return 4
}

func (r *rpcServer) run() error {
	var conn connection
	var err error
	var listener net.Listener

	h := header(r.conf.prefix)
	h.output("objects", strings.Join(r.objs, ", "))

	switch r.conf.proto {
	case "tcp":
		conn = new(tcp)
	default:
		r.conf.proto = "unix"
		conn = new(unix)
	}

	for i := 0; i < conn.retries(); i++ {
		r.conf.addr = conn.addr()
		r.server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
		listener, err = net.Listen(r.conf.proto, r.conf.addr)
		if err == nil {
			break
		}
	}

	if err != nil {
		h.output("fatal",
			NewErrConnectionFailed(fmt.Sprintf("Could not connect in %d attemps, using %s protocol", conn.retries(), r.conf.proto)).Error())
		return err
	}

	h.output("ready", fmt.Sprintf("proto=%s addr=%s", r.conf.proto, r.conf.addr))
	if err := http.Serve(listener, nil); err != nil {
		h.output("fatal", fmt.Sprintf("err-http-serve: %s", err.Error()))
		return err
	}
	return nil
}

func Register(obj interface{}) {
	// TODO: panic() if run() has been called
	defaultServer.register(obj)
}

func Run() error {
	return defaultServer.run()
}

type SimaRpc struct{}

func NewSimaRpc() *SimaRpc {
	return &SimaRpc{}
}

func (s *SimaRpc) Exit(status int, unused *SimaNilT) error {
	os.Exit(status)
	return nil
}
