package util

import (
	"log"

	"github.com/pelletier/go-toml"
)

//client post body strucct
type SrcServer struct {
	Taskid  int64
	Project string //项目名称
	Module  string //版本明，需要和rsync模块名字对应
	Mu      *rsync.Mutex
}

// 解析config
func ConfigTree() *toml.Tree {
	var ret *toml.Tree
	tree, err := toml.LoadFile("config/server.toml")
	if err != nil {
		log.Println(err)
		return ret
	}
	return tree
}

// 获取唯一taskid
