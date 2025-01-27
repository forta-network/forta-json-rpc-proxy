package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/forta-network/forta-json-rpc-proxy/utils"
	"github.com/sirupsen/logrus"
)

var wrappedMethods = map[string]interface{}{
	"eth_sendRawTransaction": struct{}{},
	"eth_call":               struct{}{},
	"eth_estimateGas":        struct{}{},
}

var proxiedMethods = map[string]interface{}{
	"net_version":               struct{}{},
	"eth_chainId":               struct{}{},
	"eth_getBalance":            struct{}{},
	"eth_getTransactionCount":   struct{}{},
	"eth_getBlockByNumber":      struct{}{},
	"eth_getBlockByHash":        struct{}{},
	"eth_blockNumber":           struct{}{},
	"eth_getCode":               struct{}{},
	"eth_gasPrice":              struct{}{},
	"eth_getTransactionReceipt": struct{}{},
	"eth_feeHistory":            struct{}{},
	"eth_maxPriorityFeePerGas":  struct{}{},
}

// Proxy intercepts and forwards JSON-RPC requests.
type Proxy struct {
	rpcServer        *rpc.Server
	reverseProxy     *httputil.ReverseProxy
	apiKeyConfigured bool
	authHeaderVal    string
}

// NewProxy creates a new proxy which can handle HTTP requests with the help of a registered
// JSON-RPC service.
func NewProxy(service *wrapperService, target, apiKey string) *Proxy {
	rpcServer := rpc.NewServer()
	err := rpcServer.RegisterName("eth", service)
	if err != nil {
		logrus.WithError(err).Panic("failed to register rpc service to eth namespace")
	}
	targetURL, err := url.Parse(target)
	if err != nil {
		logrus.WithError(err).Panic("failed to parse target url for reverse proxy")
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.Transport = utils.DefaultHTTPTransport
	reverseProxy.Director = func(r *http.Request) {
		r.Host = targetURL.Host
		r.URL = targetURL
		r.Header.Del("Authorization") // strip proxy auth header
	}
	return &Proxy{
		rpcServer:        rpcServer,
		reverseProxy:     reverseProxy,
		apiKeyConfigured: len(apiKey) > 0,
		authHeaderVal:    fmt.Sprintf("Bearer %s", apiKey),
	}
}

// ServeHTTP implements http.Handler.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(b)) // replace request body
	var reqBody struct {
		Method string `json:"method"`
	}
	err := json.Unmarshal(b, &reqBody)
	if err != nil || reqBody.Method == "" {
		w.WriteHeader(200)
		w.Write(methodParseError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Handle wrapped methods by the handlers of the local service.
	if _, ok := wrappedMethods[reqBody.Method]; ok {
		logrus.WithField("method", reqBody.Method).Debug("received request for wrapped method")
		p.rpcServer.ServeHTTP(w, r)
		return
	}

	// Handle proxied methods by proxying to the target URL.
	if _, ok := proxiedMethods[reqBody.Method]; ok {
		logrus.WithField("method", reqBody.Method).Debug("received request for proxied method")
		p.reverseProxy.ServeHTTP(w, r)
		return
	}

	// Allow all proxied methods for requests with an API key (power user).
	// The wrapped methods are enabled for everyone and that's already handled
	// as part of the first case above.
	if p.apiKeyConfigured && r.Header.Get("Authorization") == p.authHeaderVal {
		logrus.WithField("method", reqBody.Method).Debug("received request for authorized method")
		p.reverseProxy.ServeHTTP(w, r)
		return
	}

	// Disallow other JSON-RPC methods.
	w.WriteHeader(200)
	w.Write(methodNotFoundError)
}
