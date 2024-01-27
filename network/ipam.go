package network

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/miniDocker/network/ipam/subnet.json"

type IPAM struct {
	// 文件地址
	SubnetAllocatorPath string
	// 网段-位图
	Subnets *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// load 加载网段地址分配信息
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer func(subnetConfigFile *os.File) {
		if err := subnetConfigFile.Close(); err != nil {
			log.Errorf("Close file %s error: %v", ipam.SubnetAllocatorPath, err)
		}
	}(subnetConfigFile)
	if err != nil {
		return err
	}

	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(subnetJson[:n], ipam.Subnets); err != nil {
		log.Errorf("Error parse allocation info, %v", err)
		return err
	}
	return nil
}

// dump 存储网段地址分配信息
func (ipam *IPAM) dump() error {
	// 创建文件夹
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ipamConfigFileDir, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 打开文件
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer func(subnetConfigFile *os.File) {
		if err := subnetConfigFile.Close(); err != nil {
			log.Errorf("Close file %s error: %v", ipam.SubnetAllocatorPath, err)
		}
	}(subnetConfigFile)
	if err != nil {
		return err
	}

	// json编码
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	// 存储json
	if _, err = subnetConfigFile.Write(ipamConfigJson); err != nil {
		return err
	}

	return nil
}

// Allocate 分配IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {

	// 初始化 储存网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}

	// 从文件中加载已经分配的网段信息
	if err = ipam.load(); err != nil {
		log.Errorf("Error load allocation info, %v", err)
	}
	// 避免 192.168.0.1/24
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// 返回子网对应的位数和总位数
	// 127.0.0.0/24 则返回 8 24
	one, size := subnet.Mask.Size()

	// 如果没有分配过这个网段
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// 初始化地址位图,  第一个地址和最后一个地址作废
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", (1<<uint8(size-one))-2)
	}

	// 查找一个可用的ip地址
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
			ipAlloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipAlloc)
			ip = subnet.IP
			// 根据偏移计算IP地址
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			// 第一个地址作废
			ip[3] += 1
			break
		}
	}

	if err := ipam.dump(); err != nil {
		return nil, err
	}
	return
}

// Release 释放一个IP地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	if err := ipam.load(); err != nil {
		log.Errorf("Error load allocation info, %v", err)
	}

	// 避免 192.168.0.1/24
	_, subnet, _ = net.ParseCIDR(subnet.String())
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
	ipAlloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipAlloc)

	if err := ipam.dump(); err != nil {
		return err
	}
	return nil
}
