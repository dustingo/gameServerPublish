#### About
gameServerPublish 是一款用Go原生http实现的,用于远程研发游戏服务器端拉取的web api

#### Running
git clone https://github.com/dustingo/gameServerPublish.git  
go build or go run main.go
systemctl start gameServerPublish  
systemctl stop gameServerPublish  
关闭http server时，会等待业务处理结束再结束。但是会遵循systemd的超时时间限制.
#### Config
config/server.toml
- 为全局配置 \
  端口和日志目录  
  [global]

  port = ":8890"

  logdir = "/export/publish/"

- 项目配置 \
  [zzbq_dev] #双方约定好的远程rsync 模块名字

  game = "zzbq" #此模块属于的项目名称

  user = "mhxzx" #rsync时使用的用户

  host = "192.168.137.128" #对方服务器地址

  secrets = "/etc/rsyncd_zzbq.scrt" #用于rsync验证的密码文件

  path = "/home/mhxzx/zzbq_package" #拉取的服务器端本地存放位置,"zzbq_package"最好要和rsync需要拉取的目录名字相同

#### Client
method: POST  
Content-Type: application/json  
Body: {"project": "fshx","module":"fshx_dev" } #project为game名称,module 服务器端版本的rsync模块名称  
e.g:  
curl --request POST   --url http://localhost:8890/pullserver   --header 'content-type: application/json'    --data '{"project": "fshx","package":"fshx_dev" }'
#### to do
为提高API的安全性，将会采取验证Token的方式来触发业务
