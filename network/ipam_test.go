package network

import (
	"net"
	"testing"
)

func TestAllocate(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("192.168.0.1/24")
	ip, err := ipAllocator.Allocate(ipNet)
	t.Logf("alloc ip: %v, %v", ip, err)
}

func TestRelease(t *testing.T) {
	ip, ipNet, _ := net.ParseCIDR("192.168.0.1/24")
	if err := ipAllocator.Release(ipNet, &ip); err != nil {
		t.Errorf("Release error %v", err)
	}
}
