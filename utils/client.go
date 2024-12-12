package utils

import (
	"crypto/tls"
	"net/http"
	"time"
)

// DefaultHTTPTransport is the default HTTP transport.
var DefaultHTTPTransport = &http.Transport{
	// Disable HTTP/2
	// Reason: https://www.bentasker.co.uk/posts/blog/software-development/golang-net-http-net-http-2-does-not-reliably-close-failed-connections-allowing-attempted-reuse.html
	TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
}

// DefaultHTTPClient is the default HTTP client.
var DefaultHTTPClient = &http.Client{
	Timeout:   time.Second * 30,
	Transport: DefaultHTTPTransport,
}
