package initcmd

const frameworkModule = "github.com/cyjaysong/renhe"

type tplData struct {
	Module          string
	FrameworkModule string
}

const mainGoTpl = `package main

import (
	_ "{{.Module}}/internal/logic"

	"{{.Module}}/internal/cmd"
)

func main() {
	cmd.Main.Run()
}
`

const cmdGoTpl = `package cmd

import (
	"{{.FrameworkModule}}/frame/r"
	"{{.FrameworkModule}}/net/rhttp"
	"{{.FrameworkModule}}/os/rctx"
	"github.com/labstack/echo/v5/middleware"
)

var Main = &struct {
	Run func()
}{
	Run: func() {
		// Trace 由配置 trace.enable 自动启停；退出时统一收尾
		defer r.Close()

		httpSrv := r.HttpSrv()
		httpSrv.Use(
			middleware.Recover(),
			rhttp.WriteBizResJSON(),
		)

		api := httpSrv.Group("/api")
		rhttp.EchoRegisterCtrlPointers(api)

		if err := httpSrv.Run(); err != nil {
			r.Log().Fatal(rctx.GetInitCtx(), "http server run failed", "err", err)
		}
	},
}
`

const configYamlTpl = `httpSrv:
  address: ":8000"
  # gracefulTimeout: "10s"    # SIGTERM 后等待进行中请求的时间（默认 10s）

database:
  default:
    # 连接池配置
    maxOpenConns: 25
    maxIdleConns: 5
    connMaxLifetime: "1h"
    connMaxIdleTime: "30m"
    # 数据库连接 DSN（启动时会 Ping）
    dsn: "postgres://postgres:password@127.0.0.1:5432/dbname?sslmode=disable"
    # slowQueryThreshold: "5s"  # 慢查询阈值；负数关闭慢查询标记
    # healthCheckInterval: "5s"
    # pingTimeout: "3s"
    # 从库 DSN 列表（可选，支持多个从库读写分离）
    # slave:
    #   - "postgres://postgres:password@127.0.0.1:5433/dbname?sslmode=disable"

# redis:
#   default:
#     address: ["127.0.0.1:6379"]
#     # mode: auto              # standalone | cluster | auto（默认 auto）
#     db: 0

logger:
  level: "info"            # debug/info/warn/error
  format: "text"           # text/json
  output: "console"        # console/file/both
  stack: false             # Error/Fatal/Panic 是否追加完整调用栈

# trace:
#   enable: true
#   exporter: stdout           # stdout | otlp | otlphttp | none
#   # otlp:
#   #   endpoint: "localhost:4318"
#   #   insecure: true
#   serviceName: "my-service"
#   sampler: 1.0
`

const hackConfigYamlTpl = `rh:
  gen:
    dao:
      # 数据库连接 DSN，格式: postgres://用户名:密码@地址:端口/数据库名?sslmode=disable
      - link: "postgres://postgres:password@127.0.0.1:5432/dbname?sslmode=disable"
        # 数据库 schema，默认 public
        schema: "public"
        # 指定要生成的表名，多个用逗号分隔，留空表示全部表
        tables: ""
        # 排除不需要生成的表名，多个用逗号分隔
        tablesEx: ""
        # 生成代码的基础路径
        path: "./internal"
        # table 元数据输出子路径（相对于 path）
        tablePath: "model/table"
        # do 代码输出子路径（相对于 path）
        doPath: "model/do"
        # entity 代码输出子路径（相对于 path）
        entityPath: "model/ent"
        # JSON 标签命名风格: Snake / CamelLower / Camel
        jsonCase: "Snake"
        # 生成 entity 时排除的字段，格式: "表名.字段名" 或 "*.字段名"，逗号分隔
        # entityFieldEx: "user.password, *.deleted_at"
        # 类型映射: 按数据库字段类型全局映射到 Go 类型
        # typeMapping:
        #   numeric:
        #     type: "decimal.Decimal"
        #     import: "github.com/shopspring/decimal"
        #     pkgAs: "decimalx"
        # 字段映射: 按 表名.字段名 精确映射（优先级高于 typeMapping）
        # fieldMapping:
        #   user.other:
        #     type: "userex.Profile"
        #     import: "example/internal/model/userex"
        #     pkgAs: "userex"
    service:
      # logic 源码目录，扫描其子包提取公开方法生成接口
      srcPath: "internal/logic"
      # service 接口输出目录
      dstPath: "internal/service"
`

const goModTpl = `module {{.Module}}

go 1.25.0

require (
	github.com/cyjaysong/renhe v0.0.0
)
`

const gitignoreTpl = `.idea/
.vscode/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
vendor/
.DS_Store
`
