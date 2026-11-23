package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sublink/models"
	"sublink/node"

	"gorm.io/gorm"
)

// RenderRelayConfig builds a managed fragment containing explicitly enabled TCP and UDP routes.
func RenderRelayConfig(db *gorm.DB, serverID uint) (string, error) {
	var routes []models.RelayRoute
	if err := db.Where("relay_server_id = ? AND enabled = ?", serverID, true).Order("id asc").Find(&routes).Error; err != nil {
		return "", err
	}
	var output strings.Builder
	for _, route := range routes {
		if route.Protocol != "tcp" && route.Protocol != "udp" {
			return "", fmt.Errorf("route %d uses unsupported protocol %q", route.ID, route.Protocol)
		}
		var landing models.Node
		if err := db.First(&landing, "id = ?", route.LandingNodeID).Error; err != nil {
			return "", fmt.Errorf("route %d landing node is unavailable: %w", route.ID, err)
		}
		var host string
		var port int
		var err error
		if route.Protocol == "udp" {
			host, port, err = resolveUDPNodeTarget(landing.Link)
		} else {
			host, port, err = resolveTCPNodeTarget(landing.Link)
		}
		if err != nil {
			return "", fmt.Errorf("route %d: %w", route.ID, err)
		}
		addresses, err := resolvePublicRelayTargetIPs(host)
		if err != nil {
			return "", fmt.Errorf("route %d target rejected: %w", route.ID, err)
		}
		if route.Protocol == "udp" {
			fmt.Fprintf(&output, "# SUBLINKX-UDP-ROUTE\t%s\t%d\t%s\t%d\n", route.ListenAddress, route.ListenPort, addresses[0].String(), port)
			continue
		}
		listenAddress := net.JoinHostPort(route.ListenAddress, strconv.Itoa(int(route.ListenPort)))
		fmt.Fprintf(&output, "frontend sublinkx_relay_in_%d\n    bind %s\n    mode tcp\n    default_backend sublinkx_relay_out_%d\n\n", route.ID, listenAddress, route.ID)
		fmt.Fprintf(&output, "backend sublinkx_relay_out_%d\n    mode tcp\n", route.ID)
		for index, address := range addresses {
			targetAddress := net.JoinHostPort(address.String(), strconv.Itoa(port))
			fmt.Fprintf(&output, "    server landing_%d_%d %s check inter 5s fall 3 rise 2\n", route.ID, index+1, targetAddress)
		}
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func ValidateRelayNodeLink(link string) error {
	host, _, err := resolveTCPNodeTarget(link)
	if err != nil {
		return err
	}
	_, err = resolvePublicRelayTargetIPs(host)
	return err
}

func ValidateRelayNodeLinkForProtocol(link, protocol string) error {
	var host string
	if protocol == "udp" {
		var err error
		host, _, err = resolveUDPNodeTarget(link)
		if err != nil {
			return err
		}
	} else if protocol == "tcp" {
		var err error
		host, _, err = resolveTCPNodeTarget(link)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("relay protocol %q is unsupported", protocol)
	}
	_, err := resolvePublicRelayTargetIPs(host)
	return err
}

func resolveUDPNodeTarget(link string) (host string, port int, err error) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		return "", 0, errors.New("node link is not a valid URI")
	}
	switch strings.ToLower(u.Scheme) {
	case "hy", "hysteria", "hy2", "hysteria2", "tuic":
	default:
		return "", 0, fmt.Errorf("node protocol %q does not use a supported UDP transport", u.Scheme)
	}
	host = u.Hostname()
	port, err = strconv.Atoi(u.Port())
	if err != nil || host == "" || port < 1 || port > 65535 {
		return "", 0, errors.New("UDP node link must contain one valid host and port")
	}
	return host, port, nil
}

func resolveTCPNodeTarget(link string) (host string, port int, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			host, port, err = "", 0, errors.New("node link parser rejected malformed data")
		}
	}()
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		return "", 0, errors.New("node link is not a valid URI")
	}
	scheme := strings.ToLower(u.Scheme)
	transport := "tcp"
	switch scheme {
	case "vless", "trojan", "http", "https":
		host, port, transport = u.Hostname(), 0, strings.ToLower(u.Query().Get("type"))
		port, err = strconv.Atoi(u.Port())
	case "vmess":
		encoded := strings.TrimPrefix(link, "vmess://")
		decoded := node.Base64Decode(encoded)
		var payload struct {
			Add  string          `json:"add"`
			Net  string          `json:"net"`
			Port json.RawMessage `json:"port"`
		}
		if json.Unmarshal([]byte(decoded), &payload) != nil {
			return "", 0, errors.New("invalid VMess link")
		}
		host, transport = payload.Add, strings.ToLower(payload.Net)
		var portString string
		if json.Unmarshal(payload.Port, &portString) == nil {
			port, err = strconv.Atoi(portString)
		} else {
			var portNumber json.Number
			if json.Unmarshal(payload.Port, &portNumber) != nil {
				err = errors.New("invalid VMess port")
			} else {
				port, err = strconv.Atoi(portNumber.String())
			}
		}
	case "ss":
		decoded, decodeErr := node.DecodeSSURL(link)
		if decodeErr != nil {
			return "", 0, errors.New("invalid Shadowsocks link")
		}
		host, port = decoded.Server, decoded.Port
	case "ssr":
		decoded, decodeErr := node.DecodeSSRURL(link)
		if decodeErr != nil {
			return "", 0, errors.New("invalid ShadowsocksR link")
		}
		host, port = decoded.Server, decoded.Port
	case "hy", "hysteria", "hy2", "hysteria2", "tuic":
		return "", 0, fmt.Errorf("node protocol %q uses UDP/QUIC and cannot be relayed over TCP", scheme)
	default:
		return "", 0, fmt.Errorf("node protocol %q is not supported for relay", scheme)
	}
	if transport == "" {
		transport = "tcp"
	}
	switch transport {
	case "tcp", "ws", "grpc", "http", "h2", "httpupgrade", "xhttp", "splithttp":
	default:
		return "", 0, fmt.Errorf("node transport %q is not TCP", transport)
	}
	if err != nil || port < 1 || port > 65535 || host == "" {
		return "", 0, errors.New("node link must contain a valid host and port")
	}
	return host, port, nil
}

func validateRelayTarget(host string) error {
	if ip := net.ParseIP(host); ip != nil {
		if forbiddenRelayIP(ip) {
			return errors.New("target resolves to a non-public address")
		}
		return nil
	}
	if len(host) > 253 || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return errors.New("target host is not a valid hostname")
	}
	for _, label := range strings.Split(host, ".") {
		if !relayTargetLabel.MatchString(label) {
			return errors.New("target host is not a valid hostname")
		}
	}
	return nil
}

func resolvePublicRelayTargetIPs(host string) ([]net.IP, error) {
	if err := validateRelayTarget(host); err != nil {
		return nil, err
	}
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil || len(addresses) == 0 {
		return nil, errors.New("target host cannot be resolved")
	}
	unique := make([]net.IP, 0, len(addresses))
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		if forbiddenRelayIP(address) {
			return nil, errors.New("target host resolves to a non-public address")
		}
		key := address.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, address)
	}
	return unique, nil
}

var relayTargetLabel = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

var relayBlockedNetworks = func() []netip.Prefix {
	blocks := []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16",
		"172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16",
		"198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
		"::/128", "::1/128", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64", "2001::/23", "2002::/16",
		"fc00::/7", "fe80::/10", "fec0::/10", "ff00::/8",
	}
	prefixes := make([]netip.Prefix, 0, len(blocks))
	for _, block := range blocks {
		prefixes = append(prefixes, netip.MustParsePrefix(block))
	}
	return prefixes
}()

func forbiddenRelayIP(ip net.IP) bool {
	addr, err := netip.ParseAddr(ip.String())
	if err != nil {
		return true
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsMulticast() || addr.IsUnspecified() {
		return true
	}
	for _, prefix := range relayBlockedNetworks {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
