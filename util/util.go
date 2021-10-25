package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

//ReqBody request body
type ReqBody struct {
	Project string `json:"project"`
	Module  string `json:"module"`
}

// 配置文件
type FileConfig struct {
	File string
}

// 解析config
func (c *FileConfig) ConfigTree() *toml.Tree {
	var ret *toml.Tree
	tree, err := toml.LoadFile(c.File)
	if err != nil {
		log.Println(err)
		return ret
	}
	return tree
}

// handle requets body
func HandleBody(resp http.ResponseWriter, req *http.Request, ch chan int, c *FileConfig) {
	defer func() { <-ch }()
	ch <- 1
	var jsbody ReqBody
	taskid := time.Now().UnixNano() / 1e6
	b, _ := ioutil.ReadAll(req.Body)
	err := json.Unmarshal(b, &jsbody)
	if err != nil {
		resp.Write([]byte(fmt.Sprintf("request body error: %s", err.Error())))
		return
	}
	logrus.Infof("taskid->%d  remoteIP->%s  method->%s  requestURL->%s%s  project->%s  module->%s ", taskid, req.RemoteAddr, req.Method, req.Host, req.RequestURI, jsbody.Project, jsbody.Module)
	defer req.Body.Close()
	rsyncServer(resp, jsbody.Project, jsbody.Module, taskid, c)
}

// rsync game server
func rsyncServer(resp http.ResponseWriter, project, module string, taskid int64, c *FileConfig) {
	// 项目名字和模块名都不能为空
	if project == "" || module == "" {
		resp.Write([]byte(fmt.Sprintf("taskid:%d request data error", taskid)))
		return
	} else {
		tomlTree := c.ConfigTree()
		if !tomlTree.Has(module) {
			resp.Write([]byte(fmt.Sprintf("taskid:%d module name error", taskid)))
			logrus.Errorf(fmt.Sprintf("taskid: %d module name: %s  error", taskid, module))
			return
		}
		if tomlTree.Get(fmt.Sprintf("%s.game", module)).(string) != project {
			resp.Write([]byte(fmt.Sprintf("taskid:%d project name error!please confirm project name: %s  and  module name: %s ", taskid, project, module)))
			logrus.Errorf(fmt.Sprintf("taskid:%d project name error!", taskid))
			return
		}
		// rsync 语句,不能和命令包在一起
		// --avvzpc --delete --password-file=/etc/xxx.scrt user@host::module path
		args := []string{
			"-avzpc",
			"--delete",
			"--contimeout=30",
			fmt.Sprintf("--password-file=%s", tomlTree.Get(fmt.Sprintf("%s.secrets", module)).(string)),
			fmt.Sprintf("%s@%s::%s", tomlTree.Get(fmt.Sprintf("%s.user", module)).(string), tomlTree.Get(fmt.Sprintf("%s.host", module)).(string), module),
			fmt.Sprintf("%s", tomlTree.Get(fmt.Sprintf("%s.path", module)).(string)),
		}
		logrus.Infof("taskid:%d rsync options: %s", taskid, args)
		// 执行rsync
		cmd := exec.Command("rsync", args...)
		var stdErr bytes.Buffer
		cmd.Stderr = &stdErr
		err := cmd.Run()
		if err != nil {
			resp.Write([]byte(fmt.Sprintf("taskid: %d rsync server error", taskid)))
			logrus.Errorf("taskid: %d eror:%s", taskid, stdErr.String())
			return
		}
		chrownStr := fmt.Sprintf("chown -R %s:%s %s", tomlTree.Get(fmt.Sprintf("%s.user", module)).(string), tomlTree.Get(fmt.Sprintf("%s.user", module)).(string), tomlTree.Get(fmt.Sprintf("%s.path", module)).(string))
		cmdChrown := exec.Command("sh", "-c", chrownStr)
		err = cmdChrown.Run()
		if err != nil {
			resp.Write([]byte(fmt.Sprintf("taskid: %d rsync server error.error: %s", taskid, err.Error())))
			logrus.Errorf("taskid: %d eror:%s", taskid, err.Error())
			return
		}
		resp.Write([]byte(fmt.Sprintf("taskid:%d pull game server ok", taskid)))
		logrus.Infof("taskid: %d pull server ok", taskid)
	}
}
