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
	h := header("sima")

	if r.conf.discover {
		h.output("objects", strings.Join(r.objs, ", "))
		os.Exit(1)
	}

	r.server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	l, e := net.Listen(r.conf.proto, r.conf.addr)
	if e != nil {
		h.output("error", e.Error())
		return e
	}

	h.output("ready", "started http server")

	if err := http.Serve(l, nil); err != nil {
		h.output("error", e.Error())
	}
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
