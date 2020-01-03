package logging

import (
	"io"
	"log"
	"os"
)

var DefaultLog *log.Logger

func Init(path string) {
	fi, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	DefaultLog = log.New(io.MultiWriter(os.Stdout, fi), "",log.Ldate|log.Ltime|log.Lshortfile)
}

func Printf(format string, v ...interface{}) {
	DefaultLog.Printf(format, v...)
}

func Println(v ...interface{})  {
	DefaultLog.Println(v...)
}