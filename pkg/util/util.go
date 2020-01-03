package util

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
	"wps.ktkt.com/monitor/fix_network/internal/logging"
)

func CheckConnect(url string) bool {
	logging.Printf("%s 网络连接状态CheckConnect[%s] \n", strings.Repeat("=", 20), url)

	// 5 秒超时
	request := gorequest.New().Timeout(5 * time.Second)
	resp, _, errs := request.Get(url).End()
	//resp, body, err := request.Get(targetUrl).End()
	if errs != nil {
		logging.Println("测试连接失败", errs)
		return false
	}

	logging.Println("resp.StatusCode : ", resp.StatusCode)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == 911 {
		return true
	}
	return false
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
		logging.Printf("请求 %s 超时 \n", targetUrl)
	}
	//glog.Info("checkIsTimeOut body: ", body)
	//glog.Info("checkIsTimeOut resp", resp)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logging.Printf("NewDocumentFromReader failed.", err)
		return
	}
	logging.Printf("%s %s \n", strings.Repeat("=", 20), msg)
	doc.Find(".ipSearch").Find("table").First().Find("tr").Each(func(i int, e *goquery.Selection) {
		if i == 1 || i == 2 || i == 3 {
			tdList := e.ChildrenFiltered("td")
			title := strings.Trim(tdList.First().Text(), " \n\r\t")
			value := strings.Trim(tdList.Eq(1).Find("span").First().Text(), " \n\r\t")
			logging.Printf("%s: %s\n", title, value)
		}
	})
}

func RestartNetWork() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("ipconfig", "/flushdns")
		if err := cmd.Run(); err != nil {
			logging.Println(err.Error())
		} else {
			logging.Println("exec : ipconfig /flushdns 新 DNS 解析")
		}

	} else {
		// todo
		//if exec.Command("sudo", "cp", "hosts", "/etc/").Run() == nil {
		//	log.Println("hosts 替换成功")
		//}
		//if exec.Command("sudo", "/etc/init.d/networking", "restart").Run() == nil {
		//	log.Println("新 DNS 解析生效")
		//}
	}
}

// 备份文件
func RenameHosts(hostPath string) error {
	newPath := fmt.Sprintf("%s.%s.bak", hostPath, time.Now().Format("20060102150405"))
	logging.Println(strings.Repeat("=", 5), time.Now(), "开始备份文件", hostPath, newPath)

	if runtime.GOOS == "windows" {
		err := exec.Command("copy", hostPath, newPath).Run()
		if err != nil {
			logging.Println("备份失败", err.Error())
			return err
		}

	} else {
		// 备份文件
		err := exec.Command("cp", hostPath, newPath).Run()
		if err != nil {
			logging.Println("备份失败", err)
			return err
		}
	}
	logging.Println("备份成功")

	return nil
}