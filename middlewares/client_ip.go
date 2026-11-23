package middlewares

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/gin-gonic/gin"
)

func ParseAllowedClientIPs(config string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for _, value := range strings.Split(config, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if addr, err := netip.ParseAddr(value); err == nil {
			addr = addr.Unmap()
			bits := 128
			if addr.Is4() {
				bits = 32
			}
			prefixes = append(prefixes, netip.PrefixFrom(addr, bits))
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("invalid IP address or CIDR %q", value)
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return nil, fmt.Errorf("invalid IPv4-mapped CIDR %q", value)
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func ClientIPAllowlist(allowed []netip.Prefix) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowed) == 0 {
			addr, err := netip.ParseAddr(strings.TrimSpace(c.ClientIP()))
			if err == nil && addr.Unmap().IsLoopback() {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "client IP is not allowed"})
			return
		}
		addr, err := netip.ParseAddr(strings.TrimSpace(c.ClientIP()))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "client IP is not allowed"})
			return
		}
		addr = addr.Unmap()
		for _, prefix := range allowed {
			if prefix.Contains(addr) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "client IP is not allowed"})
	}
}
