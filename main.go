package main

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dustingo/gameServerPublish/util"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

func main() {
	// 获取配置文件信息
	tomlTree := util.ConfigTree()
	if tomlTree == nil {
		logrus.Errorln("配置失败")
		return
	}
	// 获取服务器端口信息，日志目录信息，由于是interface类型，需要.(Type)转换才能使用
	port := tomlTree.Get("global.port").(string)
	logdir := tomlTree.Get("global.logdir").(string) + "publish.log"
	// 创建log轮转规则
	rotateWriter, _ := rotatelogs.New(
		logdir+".%Y%m%d",                                              //真实的日志名字格式
		rotatelogs.WithLinkName(logdir),                               //为当前的日志文件创建链接
		rotatelogs.WithMaxAge(time.Duration(604800)*time.Second),      // 保留7天的日志
		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second), //每24h轮转一次
	)
	// 同时输出到终端和日志文件
	multiWriter := io.MultiWriter(os.Stdout, rotateWriter)
	//logrus.SetReportCaller(true)
	logrus.SetOutput(multiWriter)
	/*
		Http Server and
	*/
	server := &http.Server{Addr: port, Handler: http.DefaultServeMux}
	http.HandleFunc("/pullserver", func(resp http.ResponseWriter, req *http.Request) {
		//不支持POST以外的方法
		if req.Method != "POST" {
			resp.Write([]byte("Not supported method except POST"))
			return
		}
		util.HandleBody(resp, req)
	})
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
