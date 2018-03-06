package net

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	networkSpliter = "@"
)

func ParseNetwork(str string) (network, addr string, err error) {
	if idx := strings.Index(str, networkSpliter); idx == -1 {
		err = fmt.Errorf("addr: \"%s\" error, must be network@tcp:port or network@unixsocket", str)
		return
	} else {
		network = str[:idx]
		addr = str[idx+1:]
		return
	}
}

// LocalIPs LocalIPs
func LocalIPs() ([]string, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(ifaces))
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return ips, err
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			if ip.IsLoopback() {
				continue
			}

			if ipv4 := ip.To4(); ipv4 != nil {
				ips = append(ips, ip.String())
			}

		}
	}

	return ips, nil
}

// IPStringToInt IPStringToInt
func IPStringToInt(ipstring string) int {
	ipSegs := strings.Split(ipstring, ".")
	var ipInt int
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

// IPIntToString IPIntToString
func IPIntToString(ipInt int) string {
	ipSegs := make([]string, 4)
	var len = len(ipSegs)
	buffer := bytes.NewBufferString("")
	for i := 0; i < len; i++ {
		tempInt := ipInt & 0xFF
		ipSegs[len-i-1] = strconv.Itoa(tempInt)
		ipInt = ipInt >> 8
	}
	for i := 0; i < len; i++ {
		buffer.WriteString(ipSegs[i])
		if i < len-1 {
			buffer.WriteString(".")
		}
	}
	return buffer.String()
}
