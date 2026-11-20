package gotenberg

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
)

type stubResolver struct {
	lookup func(host string) ([]netip.Addr, error)
}

func (s stubResolver) LookupNetIP(_ context.Context, _, host string) ([]netip.Addr, error) {
	return s.lookup(host)
}

func withStubResolver(t *testing.T, lookup func(host string) ([]netip.Addr, error)) {
	t.Helper()
	previous := outboundResolver
	outboundResolver = stubResolver{lookup: lookup}
	t.Cleanup(func() {
		outboundResolver = previous
	})
}

func mustAddrs(t *testing.T, values ...string) []netip.Addr {
	t.Helper()
	addrs := make([]netip.Addr, 0, len(values))
	for _, value := range values {
		addr, err := netip.ParseAddr(value)
		if err != nil {
			t.Fatalf("parse IP address %q: %v", value, err)
		}
		addrs = append(addrs, addr)
	}
	return addrs
}

func TestIsPublicIP(t *testing.T) {
	for _, tc := range []struct {
		addr   string
		public bool
	}{
		{addr: "1.1.1.1", public: true},
		{addr: "2606:4700:4700::1111", public: true},
		{addr: "127.0.0.1", public: false},
		{addr: "::1", public: false},
		{addr: "::ffff:127.0.0.1", public: false},
		{addr: "::ffff:10.0.0.1", public: false},
		{addr: "10.0.0.1", public: false},
		{addr: "169.254.169.254", public: false},
		{addr: "100.100.100.200", public: false},
		{addr: "198.18.0.1", public: false},
		{addr: "192.0.2.1", public: false},
		{addr: "255.255.255.255", public: false},
		{addr: "fd12:3456:789a::1", public: false},
		{addr: "0.0.0.0", public: false},
		{addr: "0.250.250.254", public: false},
		{addr: "64:ff9b::a00:1", public: false},
		{addr: "2001:db8::1", public: false},
		{addr: "fe80::1%lo", public: false},
		{addr: "::%lo", public: false},
		{addr: "ff02::1", public: false},
	} {
		t.Run(tc.addr, func(t *testing.T) {
			addr, err := netip.ParseAddr(tc.addr)
			if err != nil {
				t.Fatalf("parse IP address %q: %v", tc.addr, err)
			}
			if got := IsPublicIP(addr); got != tc.public {
				t.Fatalf("IsPublicIP(%q) = %t, want %t", tc.addr, got, tc.public)
			}
		})
	}
}

func TestDecideOutbound(t *testing.T) {
	tests := []struct {
		name       string
		rawURL     string
		allowList  *regexp2.Regexp
		denyList   *regexp2.Regexp
		lookup     func(host string) ([]netip.Addr, error)
		opts       []DecideOption
		wantErr    error
		wantBypass bool
		wantPinned []string
	}{
		{
			name:    "rejects IPv4-mapped loopback when enabled",
			rawURL:  "http://[::ffff:127.0.0.1]:8080/",
			opts:    []DecideOption{WithDenyPrivateIPs(true)},
			wantErr: ErrFiltered,
		},
		{
			name:    "rejects private DNS results when enabled",
			rawURL:  "https://internal.example/",
			lookup:  func(string) ([]netip.Addr, error) { return mustAddrs(t, "127.0.0.1"), nil },
			opts:    []DecideOption{WithDenyPrivateIPs(true)},
			wantErr: ErrFiltered,
		},
		{
			name:       "returns resolved addresses for connection pinning",
			rawURL:     "https://example.test/",
			lookup:     func(string) ([]netip.Addr, error) { return mustAddrs(t, "1.1.1.1"), nil },
			wantPinned: []string{"1.1.1.1"},
		},
		{
			name:       "normalizes the scheme before allow-list matching",
			rawURL:     "HTTP://example.test/",
			allowList:  regexp2.MustCompile(`^http://example\.test/`, 0),
			wantBypass: true,
		},
		{
			name:       "allow-list explicitly bypasses IP restrictions",
			rawURL:     "http://internal.example/",
			allowList:  regexp2.MustCompile(`^http://internal\.example/`, 0),
			opts:       []DecideOption{WithDenyPrivateIPs(true)},
			wantBypass: true,
		},
		{
			name:      "deny-list takes precedence over allow-list",
			rawURL:    "http://blocked.example/",
			allowList: regexp2.MustCompile(`^http://`, 0),
			denyList:  regexp2.MustCompile(`blocked\.example`, 0),
			wantErr:   ErrFiltered,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.lookup != nil {
				withStubResolver(t, tc.lookup)
			} else {
				withStubResolver(t, func(host string) ([]netip.Addr, error) {
					t.Fatalf("unexpected DNS lookup for %q", host)
					return nil, nil
				})
			}

			decision, err := DecideOutbound(
				context.Background(),
				tc.rawURL,
				tc.allowList,
				tc.denyList,
				time.Now().Add(5*time.Second),
				tc.opts...,
			)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("DecideOutbound() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("DecideOutbound() error = %v", err)
			}
			if decision.Bypass != tc.wantBypass {
				t.Fatalf("Bypass = %t, want %t", decision.Bypass, tc.wantBypass)
			}

			gotPinned := make([]string, 0, len(decision.Pinned))
			for _, addr := range decision.Pinned {
				gotPinned = append(gotPinned, addr.String())
			}
			if len(gotPinned) != len(tc.wantPinned) {
				t.Fatalf("Pinned = %v, want %v", gotPinned, tc.wantPinned)
			}
			for i := range gotPinned {
				if gotPinned[i] != tc.wantPinned[i] {
					t.Fatalf("Pinned = %v, want %v", gotPinned, tc.wantPinned)
				}
			}
		})
	}
}
