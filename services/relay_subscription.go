package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"sublink/node"
)

// BuildRelaySubscriptionLink returns a copy of a node link pointed at a relay
// listener. The original node link is never modified.
func BuildRelaySubscriptionLink(link, relayHost string, relayPort uint16, routeName, protocol string) (string, error) {
	var host string
	var err error
	if protocol == "udp" {
		host, _, err = resolveUDPNodeTarget(link)
	} else if protocol == "tcp" {
		host, _, err = resolveTCPNodeTarget(link)
	} else {
		return "", fmt.Errorf("relay protocol %q is unsupported", protocol)
	}
	if err != nil {
		return "", err
	}
	if err := validateRelayTarget(host); err != nil {
		return "", err
	}
	port := strconv.Itoa(int(relayPort))
	scheme := strings.ToLower(strings.SplitN(link, "://", 2)[0])
	if protocol == "udp" {
		if scheme != "hy" && scheme != "hysteria" && scheme != "hy2" && scheme != "hysteria2" && scheme != "tuic" {
			return "", fmt.Errorf("node protocol %q cannot be derived for UDP relay", scheme)
		}
		u, err := url.Parse(link)
		if err != nil {
			return "", errors.New("invalid UDP node URI")
		}
		u.Host = net.JoinHostPort(relayHost, port)
		u.Fragment = relayDisplayName(u.Fragment, routeName)
		return u.String(), nil
	}
	switch scheme {
	case "vless", "trojan", "http", "https":
		u, err := url.Parse(link)
		if err != nil {
			return "", errors.New("invalid node URI")
		}
		name := u.Fragment
		u.Host = net.JoinHostPort(relayHost, port)
		u.Fragment = relayDisplayName(name, routeName)
		return u.String(), nil
	case "vmess":
		return relayVMessLink(link, relayHost, port, routeName)
	case "ss":
		return relaySSLink(link, relayHost, port, routeName)
	case "ssr":
		return relaySSRLink(link, relayHost, port, routeName)
	default:
		return "", fmt.Errorf("node protocol %q cannot be derived for a relay subscription", scheme)
	}
}

func relayDisplayName(original, route string) string {
	original = strings.TrimSpace(original)
	route = strings.TrimSpace(route)
	if original == "" {
		return route
	}
	if route == "" {
		return original
	}
	return original + " - " + route
}

func relayVMessLink(link, relayHost, relayPort, routeName string) (string, error) {
	payload := node.Base64Decode(strings.TrimPrefix(link, "vmess://"))
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &fields); err != nil {
		return "", errors.New("invalid VMess link")
	}
	var originalName string
	_ = json.Unmarshal(fields["ps"], &originalName)
	fields["add"], _ = json.Marshal(relayHost)
	fields["port"], _ = json.Marshal(relayPort)
	fields["ps"], _ = json.Marshal(relayDisplayName(originalName, routeName))
	updated, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return "vmess://" + node.Base64Encode(string(updated)), nil
}

func relaySSLink(link, relayHost, relayPort, routeName string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", errors.New("invalid Shadowsocks link")
	}
	originalName := u.Fragment
	userInfo := ""
	query := u.RawQuery
	if u.User != nil && u.Port() != "" {
		userInfo = u.User.String()
	} else {
		decoded := node.Base64Decode(u.Host)
		legacy, parseErr := url.Parse("ss://" + decoded)
		if parseErr != nil || legacy.User == nil || legacy.Port() == "" {
			return "", errors.New("unsupported Shadowsocks link format")
		}
		userInfo = legacy.User.String()
		if query == "" {
			query = legacy.RawQuery
		}
		if originalName == "" {
			originalName = legacy.Fragment
		}
	}
	decodedUserInfo := node.Base64Decode(userInfo)
	if strings.Contains(decodedUserInfo, ":") {
		userInfo = node.Base64Encode(decodedUserInfo)
	}
	var suffix string
	if query != "" {
		suffix = "?" + query
	}
	return "ss://" + userInfo + "@" + net.JoinHostPort(relayHost, relayPort) + suffix + "#" + url.QueryEscape(relayDisplayName(originalName, routeName)), nil
}

func relaySSRLink(link, relayHost, relayPort, routeName string) (string, error) {
	decoded := node.Base64Decode(strings.TrimPrefix(link, "ssr://"))
	parsed, err := node.DecodeSSRURL(link)
	if err != nil {
		return "", errors.New("invalid ShadowsocksR link")
	}
	addressEnd, err := ssrAddressEnd(decoded)
	if err != nil {
		return "", err
	}
	displayName := relayDisplayName(parsed.Qurey.Remarks, routeName)
	parts := strings.SplitN(decoded, "/?", 2)
	query := url.Values{}
	if len(parts) == 2 {
		query, err = url.ParseQuery(parts[1])
		if err != nil {
			return "", errors.New("invalid ShadowsocksR query")
		}
	}
	query.Set("remarks", node.Base64Encode(displayName))
	updated := net.JoinHostPort(relayHost, relayPort) + decoded[addressEnd:] + "/?" + query.Encode()
	return "ssr://" + node.Base64Encode(updated), nil
}

func ssrAddressEnd(payload string) (int, error) {
	firstColon := -1
	if strings.HasPrefix(payload, "[") {
		closingBracket := strings.IndexByte(payload, ']')
		if closingBracket < 0 || closingBracket+1 >= len(payload) || payload[closingBracket+1] != ':' {
			return 0, errors.New("invalid ShadowsocksR address")
		}
		firstColon = closingBracket + 1
	} else {
		firstColon = strings.IndexByte(payload, ':')
	}
	if firstColon < 0 {
		return 0, errors.New("invalid ShadowsocksR address")
	}
	secondOffset := strings.IndexByte(payload[firstColon+1:], ':')
	if secondOffset < 0 {
		return 0, errors.New("invalid ShadowsocksR port")
	}
	return firstColon + 1 + secondOffset, nil
}
