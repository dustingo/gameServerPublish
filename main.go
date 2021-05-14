package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/dustingo/gameServerPublish/util"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// post或者get传递的body channel
//var ch = make(chan map[string]string)

// body信息的字典
//var m = make(map[string]string, 2)

// 存储每个请求的http.ResponseWriter channel
//var ret = make(chan http.ResponseWriter)

// 存储每个请求的taskid（时间戳）
//var id = make(chan int64, 1)

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
	//定义接收到的请求的构体
	var src util.SrcServer
	/*
		Http Server and
	*/
	server := &http.Server{Addr: port, Handler: http.DefaultServeMux}
	http.HandleFunc("/pullserver", func(resp http.ResponseWriter, req *http.Request) {
		//使用毫秒时间戳作为唯一taskid
		//taskid := time.Now().UnixNano() / 1e6 //Format("20060102150405")
		//id <- taskid
		//不支持POST以外的方法
		if req.Method != "POST" {
			resp.Write([]byte("Not supported method except POST"))
			return
		}
		body, _ := ioutil.ReadAll(req.Body)

		srcBody := json.NewDecoder(req.Body)
		srcBody.Decode(&src)
		logrus.Infof("taskid->%d  remoteIP->%s  method->%s  requestURL->%s%s  project->%s  module->%s ", taskid, req.RemoteAddr, req.Method, req.Host, req.RequestURI, src.Project, src.Module)
		defer req.Body.Close()
		m["project"] = src.Project
		m["module"] = src.Module
		//ch <- m
		//ret <- resp
	})
	go Start()
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// 实际执行任务逻辑
func RsyncServer(m map[string]string, id chan int64) {
	defer delete(m, "project")
	defer delete(m, "module")
	project := m["project"]
	module := m["module"]
	taskid := <-id
	// 项目名字和模块名都不能为空
	if project == "" || module == "" {
		resp := <-ret
		resp.Write([]byte(fmt.Sprintf("taskid:%d request data error", taskid)))
		return
	} else {
		tomlTree := util.ConfigTree()
		if !tomlTree.Has(module) {
			resp := <-ret
			resp.Write([]byte(fmt.Sprintf("taskid:%d module name error", taskid)))
			logrus.Errorf(fmt.Sprintf("taskid: %d module name: %s  error", taskid, module))
			return
		}
		if tomlTree.Get(fmt.Sprintf("%s.game", module)).(string) != project {
			resp := <-ret
			resp.Write([]byte(fmt.Sprintf("taskid:%d project name error!please confirm project name: %s  and  module name: %s ", taskid, project, module)))
			logrus.Errorf(fmt.Sprintf("taskid:%d project name error!", taskid))
			return
		}
		// rsync 语句,不能和命令包在一起
		// --avvzpc --delete --password-file=/etc/xxx.scrt user@host::module path
		args := []string{
			"-avzpc",
			"--delete",
			fmt.Sprintf("--password-file=%s", tomlTree.Get(fmt.Sprintf("%s.secrets", module)).(string)),
			fmt.Sprintf("%s@%s::%s", tomlTree.Get(fmt.Sprintf("%s.user", module)).(string), tomlTree.Get(fmt.Sprintf("%s.host", module)).(string), module),
			fmt.Sprintf("%s", tomlTree.Get(fmt.Sprintf("%s.path", module)).(string)),
		}
		logrus.Infoln(args)
		//构造cmd struct
		cmd := exec.Command("rsync", args...)
		var stdErr bytes.Buffer
		cmd.Stderr = &stdErr
		err := cmd.Run()
		if err != nil {
			resp := <-ret
			resp.Write([]byte(fmt.Sprintf("taskid: %d rsync server error", taskid)))
			logrus.Errorf("taskid: %d eror:%s", taskid, stdErr.String())
			return
		}
		fmt.Println(project, module)
		resp := <-ret
		resp.Write([]byte(fmt.Sprintf("taskid:%d pull game server ok", taskid)))
	}
}

// 处理请求的goroutine
func (c ChanInfo) Start() {
	for {
		select {
		case info := c.Ch:
			go RsyncServer(info, id)
		}
	}
}
