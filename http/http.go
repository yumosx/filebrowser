package http

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

type modifyRequest struct {
	What  string   `json:"what"`  // Answer to: what data type?
	Which []string `json:"which"` // Answer to: which fields?
}

func NewHandler(
	imgSvc ImgService,
	fileCache FileCache,
	store *storage.Storage,
	server *settings.Server,
	assetsFs fs.FS,
) (http.Handler, error) {
	server.Clean()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
			next.ServeHTTP(w, r)
		})
	})
	index, static := getStaticHandlers(store, server, assetsFs)

	// NOTE: This fixes the issue where it would redirect if people did not put a
	// trailing slash in the end. I hate this decision since this allows some awful
	// URLs https://www.gorillatoolkit.org/pkg/mux#Router.SkipClean
	r = r.SkipClean(true)

	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	// 健康检查接口：用于容器/负载均衡存活探测
	r.HandleFunc("/health", healthHandler)
	// 静态资源：前端打包产物与公共资源
	r.PathPrefix("/static").Handler(static)
	// SPA 入口：未匹配路由均交给前端处理
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	// 认证相关
	tokenExpirationTime := server.GetTokenExpirationTime(DefaultTokenExpirationTime)
	// 登录：获取访问令牌
	api.Handle("/login", monkey(loginHandler(tokenExpirationTime), ""))
	// 注册：创建新用户（如启用）
	api.Handle("/signup", monkey(signupHandler, ""))
	// 续签：刷新访问令牌
	api.Handle("/renew", monkey(renewHandler(tokenExpirationTime), ""))

	// 用户管理
	users := api.PathPrefix("/users").Subrouter()
	// 列表用户
	users.Handle("", monkey(usersGetHandler, "")).Methods("GET")
	// 新增用户
	users.Handle("", monkey(userPostHandler, "")).Methods("POST")
	// 更新用户
	users.Handle("/{id:[0-9]+}", monkey(userPutHandler, "")).Methods("PUT")
	// 获取用户详情
	users.Handle("/{id:[0-9]+}", monkey(userGetHandler, "")).Methods("GET")
	// 删除用户
	users.Handle("/{id:[0-9]+}", monkey(userDeleteHandler, "")).Methods("DELETE")

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

	// 大文件断点续传（TUS 协议）
	// 创建上传会话
	api.PathPrefix("/tus").Handler(monkey(tusPostHandler(), "/api/tus")).Methods("POST")
	// 查询/校验会话
	api.PathPrefix("/tus").Handler(monkey(tusHeadHandler(), "/api/tus")).Methods("HEAD", "GET")
	// 续传分片
	api.PathPrefix("/tus").Handler(monkey(tusPatchHandler(), "/api/tus")).Methods("PATCH")
	// 取消上传
	api.PathPrefix("/tus").Handler(monkey(tusDeleteHandler(), "/api/tus")).Methods("DELETE")

	// 磁盘/配额使用情况
	api.PathPrefix("/usage").Handler(monkey(diskUsage, "/api/usage")).Methods("GET")

	// 外链与分享
	// 列表分享项
	api.Path("/shares").Handler(monkey(shareListHandler, "/api/shares")).Methods("GET")
	// 获取单个分享的公开信息
	api.PathPrefix("/share").Handler(monkey(shareGetsHandler, "/api/share")).Methods("GET")
	// 创建分享链接
	api.PathPrefix("/share").Handler(monkey(sharePostHandler, "/api/share")).Methods("POST")
	// 删除分享
	api.PathPrefix("/share").Handler(monkey(shareDeleteHandler, "/api/share")).Methods("DELETE")

	// 系统设置
	// 获取设置
	api.Handle("/settings", monkey(settingsGetHandler, "")).Methods("GET")
	// 更新设置
	api.Handle("/settings", monkey(settingsPutHandler, "")).Methods("PUT")

	// 原始文件下载/直链访问
	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	// 预览图/缩略图：size 为尺寸参数，path 为目标路径
	api.PathPrefix("/preview/{size}/{path:.*}").
		Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	// 终端命令执行（如启用）
	api.PathPrefix("/command").Handler(monkey(commandsHandler, "/api/command")).Methods("GET")
	// 全局搜索
	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")
	// 字幕获取（媒体文件配套）
	api.PathPrefix("/subtitle").Handler(monkey(subtitleHandler, "/api/subtitle")).Methods("GET")

	// 公共访问（无需登录）
	public := api.PathPrefix("/public").Subrouter()
	// 公共直链下载
	public.PathPrefix("/dl").Handler(monkey(publicDlHandler, "/api/public/dl/")).Methods("GET")
	// 通过分享链接访问
	public.PathPrefix("/share").Handler(monkey(publicShareHandler, "/api/public/share/")).Methods("GET")

	return stripPrefix(server.BaseURL, r), nil
}
