package net2

import (
	"encoding/binary"
	"net"
	"strings"
)

func Ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func Int2ip(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}

func GetLocalIp() (ip net.IP, err error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if v := ipnet.IP.To4(); v != nil {
				ip = v
				return
			}
		}
	}

	return
}

func GetInterfaceIp(name string) (ip net.IP, err error) {
	inter, err := net.InterfaceByName(name)

	if err != nil {
		return
	}

	addrs, err := inter.Addrs()

	if err != nil {
		return
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if v := ipnet.IP.To4(); v != nil {
				ip = v
				return
			}
		}
	}

	return
}

func InIpList(ip string, ips []string) bool {
	items := strings.Split(ip, ".")

	if len(items) != 4 {
		return false
	}

	ipC := strings.Join(items[:3], ".") + ".*"
	ipB := strings.Join(items[:2], ".") + ".*.*"

	for _, v := range ips {
		if v == ip || v == ipC || v == ipB {
			return true
		}
	}

	return false
}
