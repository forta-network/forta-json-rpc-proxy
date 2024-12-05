package clients

import (
	"crypto/tls"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
	// Disable HTTP/2
	// Reason: https://www.bentasker.co.uk/posts/blog/software-development/golang-net-http-net-http-2-does-not-reliably-close-failed-connections-allowing-attempted-reuse.html
	Transport: &http.Transport{
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
	},
}
