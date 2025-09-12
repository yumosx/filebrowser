## 优点
1. 支持多种登录鉴权机制 jwt
2. 支持文件上传, 下载，删除,
3. 支持大文件上传
4. 支持文件搜索
5. 使用 Blot,轻量支持持久化 
6. 配置灵活

## docker 部署

1. 构建
```powershell
docker build -f Dockerfile.simple -t filebrowser:sc .
```

2. 运行
```powershell
# 把 C:\mywww 映射成 /srv，访问 http://localhost:8080
docker run -d --name fb -p 8080:80 -v C:\mywww:/srv filebrowser:sc
```

参数说明
- `-d`：后台运行
- `--name fb`：容器起个短名，方便后续 `docker stop fb` / `docker logs fb`
- `-p 8080:80`：主机 8080 端口映射容器 80 端口
- `-v C:\mywww:/srv`：把本地目录当 Web 根目录；想换盘直接改路径

3. 常用后续命令
```powershell
# 看日志
docker logs -f fb

# 停止/删除容器
docker stop fb
docker rm fb
```

4. 浏览器访问  
   打开 [http://localhost:8080](http://localhost:8080) 即可使用 Filebrowser。

## 本地直接 build

```go
go build .

./filebrowser
```

```json
{
  "port": 8080,
  "baseURL": "",
  "address": "",
  "log": "stdout",
  "database": "/database/filebrowser.db",
  "root": "/srv"
}
```

3. 命令行配置

```go
Usage:
  filebrowser [flags]
  filebrowser [command]

Available Commands:
  cmds        Command runner management utility
  completion  Generate the autocompletion script for the specified shell
  config      Configuration management utility
  hash        Hashes a password
  help        Help about any command
  rules       Rules management utility
  upgrade     Upgrades an old configuration
  users       Users management utility
  version     Print the version number

Flags:
  -a, --address string                     address to listen on (default "127.0.0.1")
  -b, --baseurl string                     base url
      --cache-dir string                   file cache directory (disabled if empty)
  -t, --cert string                        tls certificate
  -c, --config string                      config file path
  -d, --database string                    database path (default "./filebrowser.db")
      --disable-exec                       disables Command Runner feature (default true)
      --disable-preview-resize             disable resize of image previews
      --disable-thumbnails                 disable image thumbnails
      --disable-type-detection-by-header   disables type detection by reading file headers
  -h, --help                               help for filebrowser
      --img-processors int                 image processors count (default 4)
  -k, --key string                         tls key
  -l, --log string                         log output (default "stdout")
      --noauth                             use the noauth auther when using quick setup
      --password string                    hashed password for the first user when using quick config
  -p, --port string                        port to listen on (default "8080")
  -r, --root string                        root to prepend to relative paths (default ".")
      --socket string                      socket to listen to (cannot be used with address, port, cert nor key flags)
      --socket-perm uint32                 unix socket file permissions (default 438)
      --token-expiration-time string       user session timeout (default "2h")
      --username string                    username for the first user when using quick config (default "admin")

Use "filebrowser [command] --help" for more information about a command.
```


## 后端接口地址:
```go
// 文件/目录资源操作（浏览、创建、修改、删除、部分更新）
// 浏览与读取
api.PathPrefix("/resources").Handler(monkey(resourceGetHandler, "/api/resources")).Methods("GET")
// 删除文件/目录
api.PathPrefix("/resources").Handler(monkey(resourceDeleteHandler(fileCache), "/api/resources")).Methods("DELETE")
// 新建、复制、移动、上传等创建类操作
api.PathPrefix("/resources").Handler(monkey(resourcePostHandler(fileCache), "/api/resources")).Methods("POST")
// 全量修改（如重命名、属性覆写）
api.PathPrefix("/resources").Handler(monkey(resourcePutHandler, "/api/resources")).Methods("PUT")
// 局部修改（如批量元数据或权限）
api.PathPrefix("/resources").Handler(monkey(resourcePatchHandler(fileCache), "/api/resources")).Methods("PATCH")
```