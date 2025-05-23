package reflector

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/vinted/link-exporter/internal/config"
)

func Start() {

	var wg sync.WaitGroup
	for _, iface := range config.Config.Interfaces {
		ipv6Addr, _ := getInterfaceIP(iface)
		wg.Add(1)
		go startServer(ipv6Addr.String(), iface, &wg)
	}
	wg.Wait()
}

func getInterfaceIP(iface string) (net.IP, error) {
	var ipv6Addr net.IP
	ief, err := net.InterfaceByName(iface)
	if err != nil {
		slog.Error(fmt.Sprintf("%s \n", err))
	}
	addrs, err := ief.Addrs()
	if err != nil {
		slog.Error(fmt.Sprintf("%s \n", err))
	}
	for _, addr := range addrs {
		if addr.(*net.IPNet).IP.IsLinkLocalUnicast() == false {
			continue
		}
		ipv6Addr = addr.(*net.IPNet).IP
		break
	}

	if ipv6Addr == nil {
		slog.Error(fmt.Sprintf("Interface %s don't have an link-local address", iface))
		return nil, fmt.Errorf("Interface has no link-local address")
	}
	return ipv6Addr, nil
}
func startServer(ipv6Addr string, iface string, wg *sync.WaitGroup) {
	hb, _ := time.ParseDuration(config.Config.HeartBeatInterval)
	var udpAddr *net.UDPAddr
	var err error
	defer wg.Done()
	for {
		udpAddr, err = net.ResolveUDPAddr("udp6", "["+ipv6Addr+"%"+iface+"]:"+config.Config.MonitorPort)
		if err != nil {
			slog.Debug(fmt.Sprintf("Interface is not available: %s", err))
			time.Sleep(hb)
			ip, err := getInterfaceIP(iface)
			if err == nil {
				ipv6Addr = ip.String()
				time.Sleep(5 * time.Second)
			}
		} else {
			break
		}
	}

	l, err := net.ListenUDP("udp6", udpAddr)
	if err != nil {
		slog.Error(fmt.Sprintf("Error listening:", err.Error()))
	}
	slog.Info(fmt.Sprintf("Started UDP reflector on %s:%s", ipv6Addr, config.Config.MonitorPort))
	defer l.Close()
	for {
		buf := make([]byte, 19)
		_, addr, err := l.ReadFromUDP(buf)
		if err != nil {
			slog.Error(fmt.Sprintf("Got error while reading from socket: ", err))
			return
		}
		go reflectData(l, addr, buf)
	}
}

func reflectData(l *net.UDPConn, addr *net.UDPAddr, buf []byte) {
	l.WriteToUDP(buf, addr)
}
