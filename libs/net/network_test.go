package net

import (
	"testing"
)

func TestLocalIPs(t *testing.T) {
	ips, err := LocalIPs()
	t.Logf("ip:%v, err:%v", ips, err)

	for _, v := range ips {
		ipInt := IPStringToInt(v)
		t.Logf("ip: %s, int: %d", v, ipInt)
		t.Logf("ip: %s, int: %d", IPIntToString(ipInt), ipInt)
	}
}
