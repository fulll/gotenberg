package gotenberg

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

// ErrNonPublicIP indicates that an outbound URL targets a non-public IP address.
var ErrNonPublicIP = errors.New("non-public IP")

// ErrPublicIP indicates that an outbound URL targets a public IP address.
var ErrPublicIP = errors.New("public IP")

type netipResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

var outboundResolver netipResolver = net.DefaultResolver

var outboundDialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var nonPublicIPv4Prefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
}

var nonPublicIPv6Prefixes = []netip.Prefix{
	netip.MustParsePrefix("::/96"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("100:0:0:1::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("3fff::/20"),
	netip.MustParsePrefix("5f00::/16"),
	netip.MustParsePrefix("fec0::/10"),
}

// IsPublicIP reports whether addr is publicly routable. Special-purpose
// ranges are rejected, and IPv6 zones and IPv4-mapped addresses are removed
// before classification.
func IsPublicIP(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}

	addr = addr.WithZone("").Unmap()
	if addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified() ||
		addr.IsInterfaceLocalMulticast() {
		return false
	}

	prefixes := nonPublicIPv6Prefixes
	if addr.Is4() {
		prefixes = nonPublicIPv4Prefixes
	}
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return false
		}
	}

	return true
}

// OutboundDecision contains the resolved addresses used for an outbound request.
type OutboundDecision struct {
	// Bypass is true when an explicit allow-list match permits a direct connection.
	Bypass bool
	// Pinned contains the host addresses resolved during URL validation.
	Pinned []netip.Addr
}

type outboundDecisionKey struct{}

type decideConfig struct {
	denyPrivateIPs bool
	denyPublicIPs  bool
}

// DecideOption customizes IP-class checks for an outbound URL.
type DecideOption func(*decideConfig)

// WithDenyPrivateIPs rejects URLs whose host resolves to a non-public IP address.
func WithDenyPrivateIPs(deny bool) DecideOption {
	return func(c *decideConfig) {
		c.denyPrivateIPs = deny
	}
}

// WithDenyPublicIPs rejects URLs whose host resolves to a public IP address.
func WithDenyPublicIPs(deny bool) DecideOption {
	return func(c *decideConfig) {
		c.denyPublicIPs = deny
	}
}

func resolveHost(ctx context.Context, host string, denyPrivate, denyPublic bool) ([]netip.Addr, error) {
	if host == "" {
		return nil, errors.New("empty host")
	}

	check := func(addr netip.Addr) error {
		public := IsPublicIP(addr)
		if denyPublic && public {
			return fmt.Errorf("%q: %w", addr, ErrPublicIP)
		}
		if denyPrivate && !public {
			return fmt.Errorf("%q: %w", addr, ErrNonPublicIP)
		}
		return nil
	}

	if addr, err := netip.ParseAddr(host); err == nil {
		if err := check(addr); err != nil {
			return nil, err
		}
		return []netip.Addr{addr}, nil
	}

	addrs, err := outboundResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("resolve %q: no addresses returned", host)
	}
	for _, addr := range addrs {
		if err := check(addr); err != nil {
			return nil, fmt.Errorf("%q resolves to rejected address: %w", host, err)
		}
	}

	return addrs, nil
}

func matchOutboundPattern(pattern *regexp2.Regexp, value string, deadline time.Time, options regexp2.RegexOptions) (bool, error) {
	if pattern == nil || pattern.String() == "" {
		return false, nil
	}

	compiled := regexp2.MustCompile(pattern.String(), options)
	compiled.MatchTimeout = time.Until(deadline)
	matched, err := compiled.MatchString(value)
	if err != nil {
		if time.Now().After(deadline) {
			return false, context.DeadlineExceeded
		}
		return false, fmt.Errorf("'%s' cannot handle '%s': %w", compiled.String(), value, err)
	}

	return matched, nil
}

// DecideOutbound filters a URL and resolves its host for connection pinning.
// IP-class restrictions are opt-in; regex allow and deny lists are always applied.
func DecideOutbound(ctx context.Context, rawURL string, allowList, denyList *regexp2.Regexp, deadline time.Time, opts ...DecideOption) (OutboundDecision, error) {
	cfg := decideConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return OutboundDecision{}, fmt.Errorf("parse URL %q: %w", rawURL, ErrFiltered)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	normalized := parsed.String()

	allowMatched, err := matchOutboundPattern(allowList, normalized, deadline, 0)
	if err != nil {
		return OutboundDecision{}, err
	}
	if allowList != nil && allowList.String() != "" && !allowMatched {
		return OutboundDecision{}, fmt.Errorf("'%s' does not match the expression from the allowed list: %w", normalized, ErrFiltered)
	}

	denyMatched, err := matchOutboundPattern(denyList, normalized, deadline, regexp2.IgnoreCase)
	if err != nil {
		return OutboundDecision{}, err
	}
	if denyMatched {
		return OutboundDecision{}, fmt.Errorf("'%s' matches the expression from the denied list: %w", normalized, ErrFiltered)
	}

	if allowMatched {
		return OutboundDecision{Bypass: true}, nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return OutboundDecision{}, nil
	}

	host := parsed.Hostname()
	if host == "" {
		return OutboundDecision{}, fmt.Errorf("URL %q has no host: %w", rawURL, ErrFiltered)
	}

	addrs, err := resolveHost(ctx, host, cfg.denyPrivateIPs, cfg.denyPublicIPs)
	if err != nil {
		switch {
		case errors.Is(err, ErrNonPublicIP), errors.Is(err, ErrPublicIP):
			return OutboundDecision{}, fmt.Errorf("'%s' targets a rejected address: %w", normalized, ErrFiltered)
		default:
			return OutboundDecision{}, fmt.Errorf("validate '%s' host: %w", normalized, err)
		}
	}

	return OutboundDecision{Pinned: addrs}, nil
}

// FilterOutboundURL validates a URL against regex lists and optional IP-class restrictions.
func FilterOutboundURL(ctx context.Context, rawURL string, allowList, denyList *regexp2.Regexp, deadline time.Time, opts ...DecideOption) error {
	_, err := DecideOutbound(ctx, rawURL, allowList, denyList, deadline, opts...)
	return err
}

type outboundRoundTripper struct {
	base      http.RoundTripper
	proxy     func(*http.Request) (*url.URL, error)
	allowList *regexp2.Regexp
	denyList  *regexp2.Regexp
	opts      []DecideOption
}

func (rt *outboundRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	deadline, ok := req.Context().Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}

	decision, err := DecideOutbound(req.Context(), req.URL.String(), rt.allowList, rt.denyList, deadline, rt.opts...)
	if err != nil {
		return nil, err
	}

	if rt.proxy != nil {
		proxyURL, err := rt.proxy(req)
		if err != nil {
			return nil, err
		}
		if proxyURL != nil {
			decision = OutboundDecision{Bypass: true}
		}
	}

	ctx := context.WithValue(req.Context(), outboundDecisionKey{}, decision)
	return rt.base.RoundTrip(req.WithContext(ctx))
}

// NewOutboundHTTPClient creates a client that validates outbound URLs and pins
// direct connections to the IP addresses resolved during validation.
func NewOutboundHTTPClient(timeout time.Duration, allowList, denyList *regexp2.Regexp, opts ...DecideOption) *http.Client {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.DialContext = secureDialContext
	return &http.Client{
		Timeout: timeout,
		Transport: &outboundRoundTripper{
			base:      base,
			proxy:     base.Proxy,
			allowList: allowList,
			denyList:  denyList,
			opts:      opts,
		},
	}
}

func secureDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("split host:port %q: %w", addr, err)
	}

	if decision, ok := ctx.Value(outboundDecisionKey{}).(OutboundDecision); ok {
		if decision.Bypass {
			return outboundDialer.DialContext(ctx, network, addr)
		}
		if len(decision.Pinned) > 0 {
			return DialPinned(ctx, network, decision.Pinned, port)
		}
	}

	addrs, err := resolveHost(ctx, host, false, false)
	if err != nil {
		return nil, err
	}
	return DialPinned(ctx, network, addrs, port)
}

// DialPinned connects to one of the supplied addresses without resolving the hostname again.
func DialPinned(ctx context.Context, network string, addrs []netip.Addr, port string) (net.Conn, error) {
	var lastErr error
	for _, addr := range addrs {
		conn, err := outboundDialer.DialContext(ctx, network, net.JoinHostPort(addr.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		return nil, errors.New("no addresses to dial")
	}
	return nil, lastErr
}
