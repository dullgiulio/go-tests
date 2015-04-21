package sima

import (
	"net"
	"net/http"
	"net/rpc"
	"os"
	"reflect"
	"strings"
	"sync"
)

type SimaNilT struct{}

var SimaNil SimaNilT = SimaNilT{}

type rpcServer struct {
	mux    *sync.Mutex
	server *rpc.Server
	objs   []string
	conf   *config
}

func newRpcServer() *rpcServer {
	r := &rpcServer{
		mux:    &sync.Mutex{},
		server: rpc.DefaultServer,
		objs:   make([]string, 0),
		conf:   makeConfig(), // conf remains fixed after this point
	}
	r.register(&SimaRpc{})

	return r
}

var defaultServer = newRpcServer()

func (r *rpcServer) register(obj interface{}) {
	r.mux.Lock()
	defer r.mux.Unlock()

	element := reflect.TypeOf(obj).Elem()
	r.objs = append(r.objs, element.Name())
	r.server.Register(obj)
}

func (r *rpcServer) run() error {
	if r.conf.discover {
		hw := newHeaderWriter(os.Stdout)
		hw.put("objects", strings.Join(r.objs, ", "))
		hw.end()
		os.Exit(1)
	}

	r.server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	l, e := net.Listen(r.conf.proto, r.conf.addr)
	if e != nil {
		return e
	}
	http.Serve(l, nil)
	return nil
}

func Register(obj interface{}) {
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
