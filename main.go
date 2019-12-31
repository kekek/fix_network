package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/golang/glog"
	"wps.ktkt.com/monitor/fix_network/pkg/goodhosts"

	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
	"github.com/sparrc/go-ping"
	//"github.com/lextoumbourou/goodhosts"
	wpsHost "wps.ktkt.com/monitor/fix_network/internal/hosts"
	hostRename "wps.ktkt.com/monitor/fix_network/internal/hosts"
)

const (
	winHostFile = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	otherHost   = "/etc/hosts"
	testFile    = "./data/hosts"
)

var HostFile = testFile

var FixTarget string

func init() {
	if runtime.GOOS == "windows" {
		HostFile = winHostFile
	}

	flag.StringVar(&FixTarget, "url", "https://ktapi.zljgp.com/v1/user/permission", "需要修复的地址,例如 : https://ktapi.zljgp.com/v1/user/permission")

}
func main() {
	flag.Parse()

	if len(FixTarget) == 0 {
		glog.Info("url 参数不能为空")
		os.Exit(1)
	}

	originUrl, err := url.Parse(FixTarget)
	if err != nil {
		glog.Error("parse url failed.", err)
		return
	}

	// 当前连接的服务器ip地址
	host := originUrl.Hostname()
	currIp := pingV2(host).IPAddr().String()
	// userPublicIp 用户公网ip
	glog.Infof("host %s, currIp %s \n", host, currIp)

	IpLocation(currIp, "当前服务器网络信息")
	IpLocation("", "当前用户网络信息")

	//os.Exit(0)

	// dns 解析的服务器ip地址
	allIpList := lookUpDomain(host)
	if len(allIpList) == 0 {
		printResult(fmt.Sprintf("DNS 解析未解析 %s，请修复DNS", host))
		os.Exit(0)
	}

	glog.Infof("host %s, allIp %v \n", host, allIpList)

	// host file ip 信息
	hosts, err := wpsHost.NewHosts(HostFile)
	if err != nil {
		glog.Error("get host failed.")
	}
	//for _, line := range hosts.Lines {
	//	fmt.Println(line.Raw)
	//}

	err = hostRename.RenameHosts(HostFile)
	if err != nil {
		glog.Errorf("bak hostFile %s failed: %v", HostFile, err)
		return
	}
	
	err = hosts.RemoveAllHost(host)
	if err != nil {
		glog.Error("RemoveAllHost failed.", err)
	}

	hosts.Flush()

	for i := range allIpList {
		if ok := fixBind(hosts, allIpList[i], host); ok {
			printResult("修复完成，请打开软件查看是否正常")
			break
		}
	}
}

// ip , domain
// 操作绑定 host ， 测试是否通过， 通过则OK；不通过则删除。
func fixBind(hosts goodhosts.Hosts, ip, domain string) bool {
	glog.Infof("fixBind %s %s \n", ip, domain)
	has := hosts.Has(ip, domain)

	// 有则删除，没有则添加
	if has {
		err := hosts.Remove(ip, domain)
		if err != nil {
			glog.Error("fixBind Remove failed: ", err)
			return false
		}
	} else {
		err := hosts.Add(ip, domain)
		if err != nil {
			glog.Error("fixBind add failed: ", err)
			return false
		}
	}

	hosts.Flush()

	// test 是否能通
	ok := checkIsTimeOut(FixTarget)

	if !ok {
		if !has { // 没有通过，恢复原样, 原来没有则删除
			err := hosts.Remove(ip, domain)
			if err != nil {
				glog.Error("fixBind add recover failed: ", err)
				return false
			}
		}
	}

	return ok
}

func lookUpDomain(domainName string) []string {
	res := []string{}
	// 诊断 dns
	fmt.Printf("%s 开始查找dns %s \n", strings.Repeat("=", 20), domainName)
	ips, err := net.LookupIP(domainName)
	if err != nil {
		glog.Error("lookUpDomain: 获取ip错误", err)
		return res
	}

	for _, ip := range ips {
		res = append(res, ip.String())
	}

	return res
}

func pingV2(domainName string) (res *ping.Pinger) {
	glog.Infof("%s pingV2 : ping %s \n", strings.Repeat("=", 20),  domainName)
	pinger, err := ping.NewPinger(domainName)
	if err != nil {
		glog.Error(err)
		return
	}

	pinger.Count = 3

	// listen for ctrl-C signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			pinger.Stop()
		}
	}()

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	pinger.Run()

	return pinger
}

// 检查是否连通
func checkIsTimeOut(targetUrl string) bool {
	// 5 秒超时
	request := gorequest.New().Timeout(5 * time.Second)
	resp, body, err := request.Get(targetUrl).End()
	if err != nil {
		glog.Errorf("请求 %s 超时 \n", targetUrl)
		return false
	}

	glog.Info("checkIsTimeOut body: ", body)
	glog.Info("checkIsTimeOut resp", resp)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == 911 {
		return true
	}
	return false
}

func printResult(msg string) {
	fmt.Println("++++++++++诊断结果++++++++")
	fmt.Println(msg)
	fmt.Println("++++++++++END++++++++")
}

// 打印网络情况
func IpLocation(ip, msg string) {
	targetUrl := "https://www.ipip.net/ip.html"

	if len(ip) > 0 {
		targetUrl = fmt.Sprintf("https://www.ipip.net/ip/%s.html", ip)
	}

	request := gorequest.New().Timeout(5 * time.Second)
	resp, _, errs := request.Get(targetUrl).End()
	//resp, body, errs := request.Get(targetUrl).End()
	if errs != nil {
		glog.Errorf("请求 %s 超时 \n", targetUrl)
	}
	//glog.Info("checkIsTimeOut body: ", body)
	//glog.Info("checkIsTimeOut resp", resp)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		glog.Error("NewDocumentFromReader failed.", err)
		return
	}
	fmt.Printf("%s %s \n", strings.Repeat("=", 20), msg)
	doc.Find(".ipSearch").Find("table").First().Find("tr").Each(func(i int, e *goquery.Selection) {
		if i == 1 || i == 2 || i == 3 {
			tdList := e.ChildrenFiltered("td")
			title := strings.Trim(tdList.First().Text(), " \n\r\t")
			value := strings.Trim(tdList.Eq(1).Find("span").First().Text(), " \n\r\t")
			fmt.Printf("%s: %s\n", title, value)
		}
	})
}
