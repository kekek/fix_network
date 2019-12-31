package hosts

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	//
	//"github.com/parnurzeal/gorequest"
	//yaml "gopkg.in/yaml.v1"
)

var (
	ErrorBindRelationExist = errors.New("ip 和 域名 已经绑定")
)

func Fix(ip, domainName, hostPath string) error {
	fmt.Println(strings.Repeat("=", 10), time.Now(), "开始修改hosts", ip, domainName)

	// 备份文件
	err := renameHosts(hostPath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	//var tmpBuf bytes.Buffer // 排重使用
	var MultiBindFlag bool  // 是否重复绑定

	cont, err := getHostsContent(hostPath)
	if err != nil {
		return err
	}

	for i := range cont{

		if strings.HasPrefix(cont[i], "#") || len(cont[i]) == 0 || string(cont[i]) == "#" {
			_, err = buf.WriteString(cont[i])
			if err != nil {
				fmt.Println("write failed.", err)
				return err
			}

			buf.WriteString("\r\n")
		} else {

			// 已有绑定，注释掉
			if strings.Index(cont[i], domainName) > 0 {
				buf.WriteString("# ")
				_, err = buf.WriteString(cont[i])
				if err != nil {
					fmt.Println("write failed.", err)
					return err
				}
			}
		}

		// ip 和域名已绑定
		if strings.Index(cont[i], domainName) > -1 && strings.Index(cont[i], ip) > -1 {
			MultiBindFlag = true
		}
	}

	// 重复绑定返回
	if !MultiBindFlag {
		// 重新绑定
		headLine := "\r\n\r\n# 自动修复域名 " + domainName + " \r\n# Last updated at " + time.Now().Format("2006-01-02 15:04:05") + "\r\n# Hosts Start \r\n\r\n"
		baseLine := "\r\n# Hosts End"

		buf.WriteString(headLine)
		newHostDomain := fmt.Sprintf("%s %s \r\n", ip, domainName)
		buf.WriteString(newHostDomain)
		buf.WriteString(baseLine)
	}

	err = ioutil.WriteFile(hostPath, buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}

	restartNetWork()
	return nil
}

func getHostsContent(hostPath string) ([]string, error) {
	res := []string{}
	// 读取文件，按行遍历, 替换 host
	fi, err := os.Open(hostPath)
	if err != nil {
		fmt.Printf("打开 %s 失败: %s \n", hostPath, err)
		return res, err
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		res = append(res, string(a))
	}

	return res, nil
}

func restartNetWork() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("ipconfig", "/flushdns")
		if err := cmd.Run(); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("新 DNS 解析生效")
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
func renameHosts(hostPath string) error {
	newPath := fmt.Sprintf("%s.%s.bak", hostPath, time.Now().Format("20060102150405"))
	fmt.Println(strings.Repeat("=", 5), time.Now(), "开始备份文件", hostPath, newPath)

	if runtime.GOOS == "windows" {
		err := exec.Command("copy", hostPath, newPath).Run()
		if err != nil {
			log.Println("备份失败", err.Error())
			return err
		}

	} else {
		// 备份文件
		err := exec.Command("cp", hostPath, newPath).Run()
		if err != nil {
			log.Println("备份失败", err)
			return err
		}
	}
	fmt.Println("备份成功")

	return nil
}


// ParseHosts takes in hosts file content and returns a map of parsed results.
// map[ip][]string{domain1,...}
func ParseHosts(hostsFileContent []byte) (map[string][]string, error) {
	hostsMap := map[string][]string{}
	for _, line := range strings.Split(strings.Trim(string(hostsFileContent), " \t\r\n"), "\n") {
		line = strings.Replace(strings.Trim(line, " \t"), "\t", " ", -1)
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		}
		pieces := strings.SplitN(line, " ", 2)
		if len(pieces) > 1 && len(pieces[0]) > 0 {
			if names := strings.Fields(pieces[1]); len(names) > 0 {
				if _, ok := hostsMap[pieces[0]]; ok {
					hostsMap[pieces[0]] = append(hostsMap[pieces[0]], names...)
				} else {
					hostsMap[pieces[0]] = names
				}
			}
		}
	}
	return hostsMap, nil
}
