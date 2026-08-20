package application

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func validatePublicFeedURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return "", errors.New("RSS URL must be a public HTTPS address")
	}
	if err = validatePublicHost(context.Background(), parsed.Hostname()); err != nil {
		return "", err
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func validatePublicHost(ctx context.Context, host string) error {
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(addresses) == 0 {
		return errors.New("RSS host could not be resolved")
	}
	for _, address := range addresses {
		if !address.IP.IsGlobalUnicast() || address.IP.IsPrivate() || address.IP.IsLoopback() || address.IP.IsLinkLocalUnicast() {
			return errors.New("RSS URL must not resolve to a private or local address")
		}
	}
	return nil
}

func newPublicFeedClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			for _, resolved := range addresses {
				if resolved.IP.IsGlobalUnicast() && !resolved.IP.IsPrivate() && !resolved.IP.IsLoopback() && !resolved.IP.IsLinkLocalUnicast() {
					return dialer.DialContext(ctx, network, net.JoinHostPort(resolved.IP.String(), port))
				}
			}
			return nil, fmt.Errorf("feed host %q did not resolve to a public address", host)
		},
	}
	return &http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many feed redirects")
			}
			if request.URL.Scheme != "https" {
				return errors.New("feed redirected to a non-HTTPS address")
			}
			return validatePublicHost(request.Context(), request.URL.Hostname())
		},
	}
}
