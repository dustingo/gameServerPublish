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

/*
实际执行任务逻辑
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

*/
