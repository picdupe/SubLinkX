package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ParseAllowedOrigins(config string) (map[string]struct{}, error) {
	origins := make(map[string]struct{})
	for _, value := range strings.Split(config, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		origin, err := normalizeOrigin(value)
		if err != nil {
			return nil, fmt.Errorf("invalid origin %q: %w", value, err)
		}
		origins[origin] = struct{}{}
	}
	return origins, nil
}

func CORS(allowedOrigins map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		normalized, err := normalizeOrigin(origin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "origin is not allowed"})
			return
		}
		if _, ok := allowedOrigins[normalized]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "origin is not allowed"})
			return
		}
		c.Header("Access-Control-Allow-Origin", normalized)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Origin, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")
		c.Header("Access-Control-Max-Age", "600")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func normalizeOrigin(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || u.Opaque != "" || u.User != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("expected an http(s) origin")
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("origin must not contain a path, query, or fragment")
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if strings.Contains(host, "%") {
		return "", fmt.Errorf("scoped IP addresses are not valid origins")
	}
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	port := u.Port()
	if port != "" {
		portNumber, parseErr := strconv.Atoi(port)
		if parseErr != nil || portNumber < 1 || portNumber > 65535 {
			return "", fmt.Errorf("invalid port")
		}
		if (u.Scheme == "http" && portNumber == 80) || (u.Scheme == "https" && portNumber == 443) {
			port = ""
		}
	}
	if port != "" {
		host = net.JoinHostPort(host, port)
	} else if net.ParseIP(host) != nil && strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return strings.ToLower(u.Scheme) + "://" + host, nil
}
