package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dustingo/gameServerPublish/util"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

//限制同时同步100个服务器端
var job = make(chan int, 100)

//配置文件
var filePath = flag.String("config", "", "server config path")

func main() {
	flag.Parse()
	configFile := util.FileConfig{
		File: *filePath,
	}
	// 获取配置文件信息
	//tomlTree := util.ConfigTree()
	tomlTree := configFile.ConfigTree()
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
	// 日志中打印文件信息
	//logrus.SetReportCaller(true)
	logrus.SetOutput(multiWriter)
	//信号
	exit := make(chan os.Signal, 1)
	//退出通知
	done := make(chan bool, 1)
	signal.Notify(exit, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		sigs := <-exit
		logrus.Infoln("received signal: ", sigs)
		done <- true
	}()
	/*
		Http Server and handler
	*/
	server := &http.Server{Addr: port, Handler: http.DefaultServeMux}
	http.HandleFunc("/pullserver", func(resp http.ResponseWriter, req *http.Request) {
		//不支持POST以外的方法
		if req.Method != "POST" {
			resp.Write([]byte("Not support method except POST"))
			return
		}
		//  处理request信息，执行业务
		util.HandleBody(resp, req, job, &configFile)
	})
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
	<-done
	for {
		if len(job) == 0 {
			logrus.Infoln("http stopped")
			break
		} else {
			logrus.Printf("job running: %d,please wait...", len(job))
		}
		time.Sleep(2 * time.Second)
	}
	server.Close()

}
