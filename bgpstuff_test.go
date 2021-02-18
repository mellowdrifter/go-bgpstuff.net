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
		ip      string
		want    string
		wantErr bool
	}{
		{
			ip:   "1.1.1.1",
			want: "1.1.1.0/24",
		},
		{
			ip:   "8.8.8.1",
			want: "8.8.8.0/24",
		},
		{
			ip:   "11.1.1.1",
			want: "11.1.1.0/24",
		},
		{
			ip:      "🥺",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			c := &bgpstuff.Client{}
			got, err := c.GetRoute(tc.ip)
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
