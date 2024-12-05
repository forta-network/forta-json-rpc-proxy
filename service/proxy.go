package service

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
)

// Proxy intercepts and forwards JSON-RPC requests.
type Proxy struct {
	rpcServer *rpc.Server
}

// NewProxy creates a new proxy which can handle HTTP requests with the help of a registered
// JSON-RPC service.
func NewProxy(service *Service) *Proxy {
	rpcServer := rpc.NewServer()
	err := rpcServer.RegisterName("eth", service)
	if err != nil {
		panic(err) // this should really panic
	}
	return &Proxy{rpcServer: rpcServer}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.rpcServer.ServeHTTP(w, r)
}
