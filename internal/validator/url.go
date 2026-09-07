package validator

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const MaxURLLength = 2048

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}

	if len(rawURL) > MaxURLLength {
		return fmt.Errorf("url exceeds maximum length of %d characters", MaxURLLength)
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are supported")
	}

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("URL must contain a host")
	}

	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("URL must contain a valid host")
	}

	// SSRF Protection: Check IP literals directly
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("URL host is not permitted: private, loopback, or link-local address")
		}
	} else {
		// Check blocked hostname suffixes
		lowerHost := strings.ToLower(host)
		if isBlockedHostname(lowerHost) {
			return fmt.Errorf("URL host is not permitted: loopback or internal host")
		}

		// Check DNS resolution for SSRF rebinding / local aliases (with fast timeout)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		var r net.Resolver
		ips, err := r.LookupIP(ctx, "ip", host)
		if err == nil {
			for _, ip := range ips {
				if isBlockedIP(ip) {
					return fmt.Errorf("URL host is not permitted: resolves to private, loopback, or link-local address")
				}
			}
		}
	}

	return nil
}

func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() {
		return true
	}
	return false
}

func isBlockedHostname(host string) bool {
	if host == "localhost" ||
		strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") ||
		strings.HasSuffix(host, ".lan") ||
		strings.HasSuffix(host, ".home") {
		return true
	}
	return false
}
