package svc

import (
	"net"
	"net/http"
)

func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: t.dialTimeout,
		}).DialContext,
		ResponseHeaderTimeout: t.responseHeaderTimeout,
	}
	return transport.RoundTrip(req)
}
