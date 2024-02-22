package utils

import (
	"log"
	"net"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// 获取主机内存使用
func GetMemoryInfo() (memTotal uint64, memUsedPercent float64) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	memTotal = memInfo.Total
	memUsedPercent = memInfo.UsedPercent

	return
}

// 获取ip
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// 获取主机cpou使用率
func GetCPUPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}
