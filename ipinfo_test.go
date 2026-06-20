package main

import "testing"

func TestIPIsInCIDR(t *testing.T) {
	tests := []struct {
		name    string
		ip      IP
		cidr    string
		want    bool
		wantErr bool
	}{
		{name: "empty cidr", ip: IP("203.0.113.10"), cidr: "", want: false},
		{name: "ipv4 in cidr", ip: IP("203.0.113.10"), cidr: "203.0.113.0/24", want: true},
		{name: "ipv4 out of cidr", ip: IP("203.0.114.10"), cidr: "203.0.113.0/24", want: false},
		{name: "ipv6 in cidr", ip: IP("2001:db8::1"), cidr: "2001:db8::/32", want: true},
		{name: "invalid ip", ip: IP("bad"), cidr: "203.0.113.0/24", wantErr: true},
		{name: "invalid cidr", ip: IP("203.0.113.10"), cidr: "bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ip.IsInCIDR(tt.cidr)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsInCIDR = %v, want %v", got, tt.want)
			}
		})
	}
}
