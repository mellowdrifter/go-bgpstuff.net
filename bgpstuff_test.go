package bgpstuff_test

import (
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

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			c := bgpstuff.NewBGPClient()
			got, exists, err := c.GetRoute(tc.ip)
			if tc.wantExists && !exists {
				t.Errorf("Prefix should exist, but exist returned false")
			}
			if !tc.wantExists && exists {
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
	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			c := bgpstuff.NewBGPClient()
			got, exists, err := c.GetOrigin(tc.ip)
			if tc.wantExists && !exists {
				t.Errorf("Origin should exist, but exist returned false")
			}
			if !tc.wantExists && exists {
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
	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			c := bgpstuff.NewBGPClient()
			got, _, exists, err := c.GetASPath(tc.ip)
			if tc.wantExists && !exists {
				t.Errorf("Origin should exist, but exist returned false")
			}
			if !tc.wantExists && exists {
				t.Errorf("Origin should not exist, but exist returned true")
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
	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			c := bgpstuff.NewBGPClient()
			got, exists, err := c.GetROA(tc.ip)
			if tc.wantExists && !exists {
				t.Errorf("Origin should exist, but exist returned false")
			}
			if !tc.wantExists && exists {
				t.Errorf("Origin should not exist, but exist returned true")
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
