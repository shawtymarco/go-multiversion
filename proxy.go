package multiversion

import "github.com/shawtymarco/go-multiversion/proxy"

// ProxyAdapter is a connection-local adapter for a standalone desktop proxy.
type ProxyAdapter = proxy.Adapter

func NewProxyAdapter(clientVersion string) (*ProxyAdapter, error) {
	return proxy.NewAdapter(clientVersion)
}
