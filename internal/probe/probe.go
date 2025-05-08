package probe

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/vinted/link-exporter/internal/config"
	"github.com/vinted/link-exporter/internal/ip_neighbor"
	"github.com/vinted/link-exporter/internal/prom"
)

func Start() {
	hb, _ := time.ParseDuration(config.Config.HeartBeatInterval)
	for {
		neighbors := ip_neighbor.ListNeighbors()
		for _, iface := range config.Config.Interfaces {
			for _, neigh := range neighbors {
				if neigh.InterfaceName == iface {
					heartBeat(iface, neigh.IP)
				}
			}
		}
		time.Sleep(hb)
	}
}

func heartBeat(ifname string, neigh_ip netip.Addr) {
	var wg sync.WaitGroup
	udpAddr := "[" + neigh_ip.String() + "%" + ifname + "]:" + config.Config.MonitorPort
	conn, err := net.Dial("udp6", udpAddr)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating UDP connection on port %s error: %v", ifname, err))
		return
	}
	wg.Add(1)
	go getReply(conn, ifname, &wg)
	wg.Wait()
	bs := make([]byte, 8)
	nsec := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(bs, uint64(nsec))
	conn.Write(bs)
	slog.Debug("heartBeat sent")
}

func getReply(conn net.Conn, ifname string, wg *sync.WaitGroup) {
	deadline, _ := time.ParseDuration(config.Config.HeartBeatTimeout)
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(deadline))
	reply := make([]byte, 8)
	wg.Done()
	_, err := conn.Read(reply)
	reply_nsec := time.Now().UnixNano()
	latency := 0.0
	error_count := prom.MetricsCtx[ifname].PacketLoss
	if err != nil {
		error_count += 1
		slog.Error(fmt.Sprintf("Error: UDP read on port %s error: %v", ifname, err))
	}
	if err == nil {
		reply_data := binary.LittleEndian.Uint64(reply)
		latency = float64(uint64(reply_nsec) - reply_data)
		latency = latency / 1000 / 1000
		slog.Debug(fmt.Sprintf("Received reply on %s. Latency: %f ms", ifname, latency))
	}
	prom.MetricsCtx[ifname] = prom.LinkMetrics{
		PacketLoss: error_count,
		Latency:    latency,
	}
}
