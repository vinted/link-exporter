package reflector

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/vinted/link-exporter/internal/config"
)

func Start() {

	var wg sync.WaitGroup
	for _, iface := range config.Config.Interfaces {
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
			slog.Error(fmt.Sprintf("Interface %s don't have an link-local address\n", iface))
		}

		wg.Add(1)
		go startServer(ipv6Addr.String(), iface, &wg)
	}
	wg.Wait()
}

func startServer(ipv6Addr string, iface string, wg *sync.WaitGroup) {
	defer wg.Done()
	udpAddr, err := net.ResolveUDPAddr("udp6", "["+ipv6Addr+"%"+iface+"]:"+config.Config.MonitorPort)

	if err != nil {
		slog.Error(fmt.Sprintf("%s \n", err))
		os.Exit(1)
	}

	l, err := net.ListenUDP("udp6", udpAddr)
	if err != nil {
		slog.Error(fmt.Sprintf("Error listening:", err.Error()))
	}
	slog.Info(fmt.Sprintf("Started udp reflector on %s:%s", ipv6Addr, config.Config.MonitorPort))
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
