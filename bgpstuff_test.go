package bgpstuff_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mellowdrifter/go-bgpstuff.net"
)

func TestRoute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip         string
		want       string
		wantExists bool
		wantErr    bool
	}{
		{
			ip:         "1.1.1.1",
			want:       "1.1.1.0/24",
			wantExists: true,
		},
		{
			ip:      "10.1.1.1",
			wantErr: true,
		},
		{
			ip:         "2600::",
			want:       "2600::/48",
			wantExists: true,
		},
		{
			ip:         "19.1.1.1",
			wantExists: false,
		},
		{
			ip:      "🥺",
			wantErr: true,
		},
	}

	// TODO: Interface?
	c := bgpstuff.NewBGPClient(true)

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			got, err := c.GetRoute(tc.ip)
			if tc.wantExists && got == nil {
				t.Errorf("Prefix should exist, but exist returned false")
			}
			if !tc.wantExists && got != nil {
				t.Errorf("Prefix should not exist, but exist returned true")
			}
			if tc.wantErr && err == nil {
				t.Error("Expected error, but no error returned")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("No error expected, but got error: %v", err)
			}
			_, want, _ := net.ParseCIDR(tc.want)
			if !cmp.Equal(got, want) {
				t.Errorf("Got: %s, Want: %s", got, want)
			}
		})
	}
}

func TestOrigin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip         string
		want       int
		wantExists bool
		wantErr    bool
	}{
		{
			ip:         "1.1.1.1",
			want:       13335,
			wantExists: true,
		},
		{
			ip:      "10.1.1.1",
			wantErr: true,
		},
		{
			ip:         "2600::",
			want:       1239,
			wantExists: true,
		},
		{
			ip:         "19.1.1.1",
			wantExists: false,
		},
		{
			ip:      "🥺",
			wantErr: true,
		},
	}
	c := bgpstuff.NewBGPClient(true)
	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			got, err := c.GetOrigin(tc.ip)
			if tc.wantExists && got == 0 {
				t.Errorf("Origin should exist, but exist returned false")
			}
			if !tc.wantExists && got != 0 {
				t.Errorf("Origin should not exist, but exist returned true")
			}
			if tc.wantErr && err == nil {
				t.Error("Expected error, but no error returned")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("No error expected, but got error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Got: %d, Want: %d", got, tc.want)
			}
		})
	}
}

func TestASPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip         string
		wantSrc    int
		wantExists bool
		wantErr    bool
	}{
		{
			ip:         "1.1.1.1",
			wantSrc:    13335,
			wantExists: true,
		},
		{
			ip:      "10.1.1.1",
			wantErr: true,
		},
		{
			ip:         "2600::",
			wantSrc:    1239,
			wantExists: true,
		},
		{
			ip:         "19.1.1.1",
			wantExists: false,
		},
		{
			ip:      "🥺",
			wantErr: true,
		},
	}
	c := bgpstuff.NewBGPClient(true)
	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			got, _, err := c.GetASPath(tc.ip)
			if tc.wantExists && got == nil {
				t.Errorf("AS-path should exist, but exist returned false")
			}
			if !tc.wantExists && got != nil {
				t.Errorf("AS-path should not exist, but exist returned true")
			}
			if tc.wantErr && err == nil {
				t.Error("Expected error, but no error returned")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("No error expected, but got error: %v", err)
			}
			if tc.wantExists {
				if len(got) < 2 {
					t.Errorf("ASPath is only %d long: %v", len(got), got)
				}
				if got[len(got)-1] != tc.wantSrc {
					t.Errorf("Got: %d, Want: %d", got, tc.wantSrc)
				}
			}
		})
	}
}

func TestROA(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip         string
		want       string
		wantExists bool
		wantErr    bool
	}{
		{
			ip:         "1.1.1.1",
			want:       "VALID",
			wantExists: true,
		},
		{
			ip:      "10.1.1.1",
			wantErr: true,
		},
		{
			ip:         "2600::",
			want:       "UNKNOWN",
			wantExists: true,
		},
		{
			ip:         "19.1.1.1",
			wantExists: false,
		},
		{
			ip:      "🥺",
			wantErr: true,
		},
	}
	c := bgpstuff.NewBGPClient(true)
	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			got, err := c.GetROA(tc.ip)
			if tc.wantExists && got == "" {
				t.Errorf("ROA should exist, but exist returned false")
			}
			if !tc.wantExists && got != "" {
				t.Errorf("ROA should not exist, but exist returned true: %v", got)
			}
			if tc.wantErr && err == nil {
				t.Error("Expected error, but no error returned")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("No error expected, but got error: %v", err)
			}
			if tc.wantExists {
				if got != tc.want {
					t.Errorf("Got: %s, Want: %s", got, tc.want)
				}
			}
		})
	}
}

func TestASName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		asn        int
		want       string
		wantExists bool
		wantErr    bool
	}{
		{
			asn:        3356,
			want:       "LEVEL3",
			wantExists: true,
		},
		{
			asn:     0,
			wantErr: true,
		},
		{
			asn: 4199999999,
		},
	}
	c := bgpstuff.NewBGPClient(true)
	c.GetASNames()
	for _, tc := range tests {
		t.Run(fmt.Sprint(tc.asn), func(t *testing.T) {
			name, err := c.GetASName(tc.asn)
			if tc.wantExists && name == "" {
				t.Errorf("AS name should exist, but exist returned false")
			}
			if !tc.wantExists && name != "" {
				t.Errorf("AS name should not exist, but exist returned true")
			}
			if tc.wantErr && err == nil {
				t.Error("Expected error, but no error returned")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("No error expected, but got error: %v", err)
			}
			if tc.wantExists {
				if name != tc.want {
					t.Errorf("Got: %s, Want: %s", name, tc.want)
				}
			}
		})
	}
}

func TestASNames(t *testing.T) {
	c := bgpstuff.NewBGPClient(true)
	if err := c.GetASNames(); err != nil {
		t.Errorf("got error: %v", err)
	}
	if len(c.ASNames) < 100000 {
		t.Errorf("wanted at least 100k prefixes, but got %d", len(c.ASNames))
	}

	if c.ASNames[3356] != "LEVEL3" {
		t.Errorf("expected LEVEL3, got %s", c.ASNames[3356])
	}
}

func TestInvalids(t *testing.T) {
	c := bgpstuff.NewBGPClient(true)
	if err := c.GetInvalids(); err != nil {
		t.Errorf("got error: %v", err)
	}
	if len(c.Invalids) == 0 {
		t.Errorf("Should have some invalids, but seeing %d invalids", len(c.Invalids))
	}

	if len(c.Invalids[13335]) != 3 {
		t.Errorf("cloudflare advertises three invalid prefixes, but seeing %d: %v", len(c.Invalids[13335]), c.Invalids[13335])
	}
}

func TestInvalid(t *testing.T) {
	c := bgpstuff.NewBGPClient(true)
	_, err := c.GetInvalid(13335)
	if err == nil {
		t.Errorf("expected error, but got none")
	}

	c.GetInvalids()
	prefixes, err := c.GetInvalid(13335)
	if err != nil {
		t.Fatal(err)
	}
	if len(prefixes) == 0 {
		t.Fatalf("wanted some invalids, but returned %d", len(prefixes))
	}
	if len(prefixes) != 3 {
		t.Errorf("cloudflare advertises three invalid prefixes, but seeing %d: %v", len(c.Invalids[13335]), c.Invalids[13335])
	}
}

func TestSourced(t *testing.T) {
	c := bgpstuff.NewBGPClient(true)
	prefixes, v4, v6, err := c.GetSourced(15169)
	if err != nil {
		t.Fatal(err)
	}

	if (v4 == 0) || (v6 == 0) {
		t.Errorf("AS15169 should be advertising more than zero IPv4 and IPv6 addresses, but got IPv4: %d and IPv6: %d", v4, v6)
	}

	_, dns, _ := net.ParseCIDR("8.8.8.0/24")
	if !containsSubnet(dns, prefixes) {
		t.Error("Expected to see 8.8.8.0/24, but not found")
	}
}

func containsSubnet(prefix *net.IPNet, prefixes []*net.IPNet) bool {
	for _, v := range prefixes {
		if v.String() == prefix.String() {
			return true
		}
	}
	return false
}

func TestTotals(t *testing.T) {
	c := bgpstuff.NewBGPClient(true)
	ipv4, ipv6, err := c.GetTotals()
	if err != nil {
		t.Fatal(err)
	}

	if ipv4 < 800000 || ipv6 < 100000 {
		t.Errorf("Expecting a certain amount of prefixes, but got %d IPv4 and %d IPv6", ipv4, ipv6)
	}
}
