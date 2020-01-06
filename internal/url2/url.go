package url2

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"

	"github.com/kekek/fix_network/internal/logging"
	"github.com/sparrc/go-ping"
)

type SelfUrl struct {
	LawUrl string
	Host   string
	ipList []string
}

func New(addr string) (info *SelfUrl, err error) {
	info = &SelfUrl{
		LawUrl: addr,
		Host:   "",
	}

	addrInfo, err := url.Parse(addr)
	if err != nil {
		logging.Printf("parse url failed : %v \n", err)
		return
	} else if addrInfo == nil {
		return
	}

	info.Host = addrInfo.Hostname()
	return
}

// check dns
func (u *SelfUrl) AllIp() []string {

	if len(u.ipList) > 0 {
		return u.ipList
	}

	res := []string{}
	// 诊断 dns
	ips, err := net.LookupIP(u.Host)
	if err != nil {
		fmt.Println("lookUp ip failed.", err)
		return res
	}
	for _, ip := range ips {
		res = append(res, ip.String())
	}

	u.ipList = res

	return res
}

func (u *SelfUrl) AllIpV2() []string {

	if len(u.ipList) > 0 {
		return u.ipList
	}

	res := []string{}
	cmd := exec.Command("nslookup", u.Host)
	output, err := cmd.Output()
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Println(string(output))

	u.ipList = res

	return res
}

func (u *SelfUrl) AllIpTest() {
	// 解析ip地址
	ns, err := net.LookupHost(u.Host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s", err.Error())
		return
	}

	for _, n := range ns {
		fmt.Println(os.Stdout, n)
	}
}

func (u *SelfUrl) CurrIP() string {
	pingFunc := newPingV2(u.Host, true)
	if pingFunc == nil {
		return ""
	}

	pingFunc.Run()
	return pingFunc.IPAddr().String()
}

func (u *SelfUrl) Ping() string {
	pingFunc := newPingV2(u.Host, true)
	if pingFunc == nil {
		return ""
	}

	pingFunc.Run()
	return pingFunc.IPAddr().String()
}

func newPingV2(domainName string, debug ...bool) (res *ping.Pinger) {

	isLog := false
	if len(debug) > 0 {
		isLog = debug[0]
	}

	logging.Printf("%s pingV2 : ping %s \n", strings.Repeat("=", 20), domainName)
	pinger, err := ping.NewPinger(domainName)
	if err != nil {
		logging.Printf("pingV2 ping.NewPinger(%s) failed: %v", domainName, err)
		return nil
	}

	pinger.Count = 3
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	// listen for ctrl-C signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			pinger.Stop()
		}
	}()

	pinger.OnRecv = func(pkt *ping.Packet) {
		if isLog {
			logging.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
		}
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		if isLog {
			logging.Printf("--- %s ping statistics ---\n", stats.Addr)
			logging.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
				stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
			logging.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
				stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		}
	}

	logging.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	return pinger
}
