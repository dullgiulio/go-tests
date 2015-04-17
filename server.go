package sima

import (
	"net"
	"net/http"
	"net/rpc"
	"os"
	"reflect"
	"sync"
)

type SimaNilT struct{}

var SimaNil SimaNilT = SimaNilT{}

type rpcServer struct {
	mux    *sync.Mutex
	server *rpc.Server
	objs   *Objects
	conf   *config
}

func newRpcServer() *rpcServer {
	r := &rpcServer{
		mux:    &sync.Mutex{},
		server: rpc.DefaultServer,
		objs:   NewObjects(),
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
	r.objs.Add(element.Name())
	r.server.Register(obj)
}

func (r *rpcServer) run() error {
	r.server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	l, e := net.Listen("tcp", r.conf.host)
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

func (s *SimaRpc) List(unused int, objs *Objects) error {
	defaultServer.mux.Lock()
	defer defaultServer.mux.Unlock()

	for i := range defaultServer.objs.Names {
		objs.Add(defaultServer.objs.Names[i])
	}
	return nil
}

func (s *SimaRpc) Exit(status int, unused *SimaNilT) error {
	os.Exit(status)
	return nil
}
