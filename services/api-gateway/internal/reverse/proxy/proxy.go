package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/logger"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/response"
)

type Gateway struct {
	routes map[string]*httputil.ReverseProxy
	log    *logger.Logger
}

func NewGateway(routes map[string]*url.URL, log *logger.Logger) *Gateway {
	proxies := make(map[string]*httputil.ReverseProxy, len(routes))
	for host, target := range routes {
		proxies[strings.ToLower(host)] = newProxy(target, log)
	}
	return &Gateway{routes: proxies, log: log}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := normalizeHost(r.Host)
	proxy := g.routes[host]
	if proxy == nil {
		response.Error(w, http.StatusBadGateway, "unknown host")
		return
	}
	proxy.ServeHTTP(w, r)
}

func newProxy(target *url.URL, log *logger.Logger) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalHost := r.Host
		originalScheme := forwardedProto(r)
		clientIP := clientIP(r)
		originalDirector(r)
		r.Host = target.Host
		r.Header.Set("X-Forwarded-Host", originalHost)
		r.Header.Set("X-Forwarded-Proto", originalScheme)
		r.Header.Set("X-Real-IP", clientIP)
		appendForwardedFor(r, clientIP)
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("proxy error", map[string]any{
			"host":   r.Host,
			"path":   r.URL.Path,
			"target": target.String(),
			"error":  err.Error(),
		})
		response.Error(w, http.StatusBadGateway, "upstream unavailable")
	}
	return proxy
}

func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func forwardedProto(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func appendForwardedFor(r *http.Request, ip string) {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		r.Header.Set("X-Forwarded-For", prior+", "+ip)
		return
	}
	r.Header.Set("X-Forwarded-For", ip)
}
