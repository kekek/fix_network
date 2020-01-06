package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/kekek/fix_network/internal/logging"
	"github.com/kekek/fix_network/internal/url2"
	"github.com/kekek/fix_network/pkg/goodhosts"
	"github.com/kekek/fix_network/pkg/util"

	//"github.com/lextoumbourou/goodhosts"
	wpsHost "github.com/kekek/fix_network/internal/hosts"
)

var Version string
var Date string

var showVer = flag.Bool("version", false, "show version")

const (
	winHostFile = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	otherHost   = "/etc/hosts"
	testFile    = "./data/hosts"
)

var HostFile = testFile

var domainList = []string{
	"https://clt.zljgp.com",
	"https://www.zljgp.com/logicians",
	"https://ktapi.zljgp.com",
	"https://mapi.zljgp.com/user/v1/mobile-signin-refresh-token",
	"https://message.zljgp.com",
	"https://mystock.zljgp.com",
	//"https://test.daklfjdlasjfldasjflas.com",
	//"zlj.docs.zljgp.com",
	//"metric.zljgp.com",
	//"ws.zljgp.com",
}

func init() {
	if runtime.GOOS == "windows" {
		HostFile = winHostFile
	}
}

func main() {

	flag.Parse()

	if *showVer {
		logging.Println("version : ", Version, "build date : ", Date)
		os.Exit(0)
	}

	logging.Init("./info.log")

	// 用户当前网络
	util.IpLocation("", "当前用户网络状态")

	for _, v := range domainList {

		printStart(fmt.Sprintf("检查：%s ", v))
		info, err := url2.New(v)
		if err != nil {
			logging.Printf("主机不存在，请修复dns或检查域名[%s]是否正确", v)
			printEnd("")
			continue
		}

		currIp := info.CurrIP()
		logging.Printf("hostName : %s,  currIp : %s \n", info.Host, currIp)

		util.IpLocation(currIp, fmt.Sprintf("服务器主机[%s(%s)]网络：", info.Host, currIp))

		if ok := util.CheckConnect(v); ok {

			logging.Printf("[%s] 网络连通正常 \n", v)
		} else {
			err := check(info)
			if err != nil {
				printResult(fmt.Sprintf("修复 %s 失败: %v", v, err))
			} else {
				printResult(fmt.Sprintf("修复 %s 成功.", v))
			}
		}

		printEnd("")
		logging.Println()
	}

	logging.Println("所有检查完成")
	logging.Println("所有检查正常完成后，尝试后仍然无法访问，请尝试命令修复lsp。以管理员身份运行：netsh winsock reset")

	InitSignal()
}

func check(info *url2.SelfUrl) error {
	// 备份host
	//err := wpsHost.RenameHosts(HostFile)
	err := wpsHost.BackupHosts(HostFile)
	if err != nil {
		return fmt.Errorf("bak hostFile %s failed: %v", HostFile, err)
	}

	// host file ip 信息
	hosts, err := wpsHost.NewHosts(HostFile)
	if err != nil {
		return fmt.Errorf("get host failed: %v", err)
	}

	err = hosts.RemoveAllHost(info.Host)
	if err != nil {
		return fmt.Errorf("RemoveAllHost failed: %v", err)
	}
	hosts.Flush()

	// 移除host 绑定后， 再次测试是否
	if ok := util.CheckConnect(info.LawUrl); ok {
		logging.Println("移除绑定后，连接成功")
		return nil
	}

	//info.AllIpV2()
	allIpList := info.AllIp()
	if len(allIpList) == 0 {
		return fmt.Errorf("DNS 解析未解析 %s，请尝试修复DNS", info.Host)
	}

	logging.Printf("host：%s, allIp：%v \n", info.Host, allIpList)

	for i := range allIpList {
		if ok := fixBind(hosts, allIpList[i], info); !ok {
			return errors.New("无法修复联系管理员")
		}
	}
	return nil
}

// ip , domain
// 操作绑定 host ， 测试是否通过， 通过则OK；不通过则删除。
func fixBind(hosts goodhosts.Hosts, ip string, info *url2.SelfUrl) bool {
	logging.Printf("fixBind %s %s \n", ip, info.Host)
	has := hosts.Has(ip, info.Host)

	// 有则删除，没有则添加
	if has {
		err := hosts.Remove(ip, info.Host)
		if err != nil {
			logging.Println("fixBind Remove failed: ", err)
			return false
		}
	} else {
		err := hosts.Add(ip, info.Host)
		if err != nil {
			logging.Println("fixBind add failed: ", err)
			return false
		}
	}

	hosts.Flush()

	wpsHost.RestartNetWork()

	// test 是否能通
	ok := util.CheckConnect(info.LawUrl)

	if !ok {
		if !has { // 没有通过，恢复原样, 原来没有则删除
			err := hosts.Remove(ip, info.Host)
			if err != nil {
				logging.Println("fixBind add recover failed: ", err)
				return false
			}
			hosts.Flush()
		}
	}

	return ok
}

func printResult(msg string) {
	logging.Println(msg)
}

func printStart(title string) {
	logging.Printf("\n %s BEGIN: %s %s \n", strings.Repeat("+", 20), title, strings.Repeat("+", 20))
}

func printEnd(title string) {
	logging.Printf("\n %s END %s %s \n", strings.Repeat("+", 20), title, strings.Repeat("+", 20))
}

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		default:
			return
		}
	}
}
