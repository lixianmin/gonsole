package tools

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/********************************************************************
created:    2020-06-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var hostIP string

func init() {
	hostIP = fetchIPAddress()
}

func GetHostIP() string {
	return hostIP
}

func fetchIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		_, _ = os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				var ip = ipnet.IP.String()
				return ip
			}
		}
	}

	return ""
}

func GetGPID(port int) string {
	var ip = GetHostIP()
	var now = time.Now().UnixNano() / 1000
	var pid = os.Getpid()
	var ret = fmt.Sprintf("%s/%d/%d/%d", ip, port, now, pid)
	return ret
}

func GetRawIP(r *http.Request) string {
	var rawIP = strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	if rawIP != "" {
		return rawIP
	}

	var ip = strings.Split(r.RemoteAddr, ":")[0]
	return ip
}

func GetIPNum(ipAddress string) (int64, error) {
	var ipList = strings.Split(ipAddress, ".")
	if len(ipList) != 4 {
		return 0, fmt.Errorf("invalid ip address: %q", ipAddress)
	}

	var result int64 = 0
	for _, ip := range ipList {
		var ipNum, err = strconv.ParseInt(ip, 10, 0)
		if err != nil {
			return 0, err
		}

		if ipNum < 0 || ipNum > 255 {
			return 0, errors.New("Out of range exception, address=" + ipAddress)
		}

		result = result*256 + ipNum
	}

	return result, nil
}
