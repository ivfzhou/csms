/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package util

import (
	"context"
	"errors"
	"flag"
	"net"
	"strconv"
	"strings"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/log"
)

var (
	// ErrLocalIPv4NotFound 未在本地系统找到 IPv4 地址。
	ErrLocalIPv4NotFound = errors.New("no valid ipv4 address found")
	// LocalIPv4 本地内网 IPv4 地址。
	LocalIPv4 []string
	// LocalInternetIPv4 本地外网 IPv4 地址。
	LocalInternetIPv4 []string
	// LocalIP 本地 IPv4 地址。
	LocalIP string
)

// AddIPCommandFlag 添加 IP 命令行参数。
func AddIPCommandFlag() {
	flag.StringVar(&LocalIP, consts.CommandFlagLocalIP, "", "specify local ipv4")
}

// InitializeLocalIP 解析 IP 地址。发生错误则退出程序。
func InitializeLocalIP(ctx context.Context) {
	if IPv4ToNumber(LocalIP) <= 0 {
		if len(LocalIP) > 0 {
			log.Warn(ctx, "invalid local ipv4 address", LocalIP)
		}
		ips, err := GetLocalIPv4()
		if err != nil {
			log.Fatal(ctx, consts.ExitCodeLocalIPNotFound, "failed to get local ip", err)
		}
		for _, v := range ips {
			if IsIntranet(v) {
				LocalIPv4 = append(LocalIPv4, v)
			} else {
				LocalInternetIPv4 = append(LocalInternetIPv4, v)
			}
		}
		log.Info(ctx, "local intranet ips", LocalIPv4)
		log.Info(ctx, "local internet ips", LocalInternetIPv4)
		if len(LocalInternetIPv4) > 0 {
			LocalIP = LocalInternetIPv4[0]
		} else if len(LocalIPv4) > 0 {
			LocalIP = LocalIPv4[0]
		}
		log.Info(ctx, "actual using local ip", LocalIP)
	}
}

// IPv4ToNumber ipv4 字符串转数字。若是非 ip 地址则返回 0。
func IPv4ToNumber(ip string) uint32 {
	result := uint32(0)
	arr := strings.Split(ip, ".")
	if len(arr) == 4 {
		num0, err := strconv.ParseUint(arr[0], 10, 32)
		if err != nil || num0 > 255 {
			return 0
		}
		num1, err := strconv.ParseUint(arr[1], 10, 32)
		if err != nil || num1 > 255 {
			return 0
		}
		num2, err := strconv.ParseUint(arr[2], 10, 32)
		if err != nil || num2 > 255 {
			return 0
		}
		num3, err := strconv.ParseUint(arr[3], 10, 32)
		if err != nil || num3 > 255 {
			return 0
		}
		result = uint32(num3)
		result |= uint32(num2) << 8
		result |= uint32(num1) << 16
		result |= uint32(num0) << 24
	}

	return result
}

// GetLocalIPv4 获取本地的 IPv4 地址。
func GetLocalIPv4() ([]string, error) {
	var ips []string
	// 获取所有网络接口。
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// 排除回环接口和未启用的接口。
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取接口地址。
		addrs, err2 := iface.Addrs()
		if err2 != nil {
			return nil, err2
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			// 排除 IPv6 地址和回环地址。
			if ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return nil, ErrLocalIPv4NotFound
	}

	return ips, nil
}

// IsIntranet 判断是否是内网 IP。
func IsIntranet(ipv4 string) bool {
	num := IPv4ToNumber(ipv4)
	if num>>16 == (192<<8 | 168) {
		return true
	}
	if num>>20 == (172<<4 | 16>>4) {
		return true
	}
	if num>>24 == 10 {
		return true
	}
	if num>>24 == 127 {
		return true
	}
	return false
}
