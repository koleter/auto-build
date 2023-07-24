# auto-build
golang 自动编译系统

## Install
```shell
TODO
```

## Usage
```shell
#normal
./auto-build ./config.toml
#nohup
nohup ./auto-build ./config.toml > nohup.log 2>&1 &
```

## 配置文件说明
```toml
port = 8000 # 监听端口
log_path = "./log"
log_level = "DEBUG"
bare_path = "./bare"
record_path = "./buildlog" # 编译/程序运行 log
go_env_path = "./goenv" # go 环境安装目录
default_go_path = "./workspace/" # 针对 gomod 的 gopath 目录(缓存包)
dest_path = "./output/" # 输出文件目录
sql_file = "./dev.db" # sqlite 文件位置,会自动创建
web_path = "./dist/" # 前端目录,可以用下面的前端项目编译后的 dist 目录
```

## TODO
- [x] webhook接到请求后编译
- [x] 删除 goenv/task/project
- [x] ~~添加 golang/新建项目添加进度~~ 本地自动下载 go env 包
- [x] git后端问题
- [x] ~~url 优化,每个任务一个单独的 url,然后工程有个 latest url ,当选择工程的可以快捷下载最新的~~ 目前方案没问题
- [ ] 增加编译后 hook
- [ ] 增加编译前后脚本,可以再编译前或编译后执行脚本(主要是编译 c 库)
- [x] ~~自动编译选项删除,改成在 igit 配置后就自动编译~~ 还是需要
- [x] 项目不在单独目录下,而是每次 git clone -depth=1 
- [x] golang 包不用用户上传,自己下载

## 前端
[github.com/hash-rabbit/auto-build-web](https://github.com/hash-rabbit/auto-build-web)