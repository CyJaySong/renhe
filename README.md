# RenHe

RenHe 是一个轻量级 Go Web 应用框架，基于成熟的第三方库组合而成，提供开箱即用的配置管理、日志、数据库、缓存、HTTP 服务和代码生成能力。

## 特性

- **统一门面** — 通过 `r.Cfg()` / `r.Log()` / `r.DB()` / `r.Redis()` / `r.HttpSrv()` 快速访问所有组件
- **YAML 配置驱动** — 基于 Viper，按优先级搜索 `manifest/config > manifest > config > .` 目录
- **结构化日志** — 基于标准库 log/slog（rlog 门面），支持 text/JSON 格式、console/file/both 输出、文件轮转、调用栈打印
- **PostgreSQL + 读写分离** — 基于 Bun ORM，支持主从多节点、Round-Robin 负载均衡、连接池健康检查
- **Redis** — 基于 go-redis，自动推断单机/集群模式，使用 UniversalClient 统一接口
- **HTTP 服务** — 基于 Echo v5，支持结构体元数据驱动的控制器自动注册
- **参数校验** — 基于 go-playground/validator，内置 `v` / `v2` 双 tag 验证器
- **CLI 工具** — `rh` 命令行，提供项目初始化、DAO/Service 代码生成（支持 typeMapping/fieldMapping）
- **OpenTelemetry** — 日志自动关联 Trace/Span ID；HTTP/DB/Redis 可观测（HTTP 中间件可关）

## 项目结构

```
renhe/
├── cmd/rh/                 # CLI 工具（rh init / rh gen dao / rh gen service）
├── database/
│   ├── rdb/                # PostgreSQL 数据库封装（Bun ORM，主从分离）
│   └── redis/              # Redis 封装（go-redis，单机/集群自适应）
├── example/                # 示例项目
├── frame/r/                # 门面包，全局组件快捷入口
├── net/rhttp/              # HTTP 服务封装（Echo v5，控制器自动注册）
├── os/
│   ├── rcfg/               # 配置管理（Viper）
│   ├── rctx/               # 上下文与 initCtx
│   ├── rlog/               # 日志服务（log/slog）
│   └── rotrace/            # OpenTelemetry TracerProvider
└── util/rvalid/            # 参数校验（go-playground/validator）
```

## 快速开始

### 安装 CLI

```bash
go install github.com/cyjaysong/renhe/cmd/rh@latest
```

### 初始化项目

```bash
rh init myproject
cd myproject && go mod tidy
```

### 配置文件

在 `manifest/config/config.yaml` 中配置：

```yaml
httpSrv:
  address: ":8000"
  # gracefulTimeout: "10s"    # SIGTERM 后等待进行中请求（默认 10s）

database:
  default:
    dsn: "postgres://user:pass@127.0.0.1:5432/dbname?sslmode=disable"
    maxOpenConns: 25
    maxIdleConns: 5
    connMaxLifetime: "1h"
    connMaxIdleTime: "30m"
    # slowQueryThreshold: "5s"  # SQL 慢查询阈值；负数关闭
    # healthCheckInterval: "5s"
    # pingTimeout: "3s"

redis:
  default:
    address: ["127.0.0.1:6379"]
    # mode: auto              # standalone | cluster | auto（默认 auto）
    db: 0

logger:
  level: "info"              # debug/info/warn/error
  format: "text"             # text/json
  output: "console"          # console/file/both
  stack: false               # Error/Fatal/Panic 是否追加完整调用栈
  # file:
  #   path: "logs/app.log"
  #   maxSize: 128           # 单文件最大 MB
  #   rotateTime: "every_day"  # 兼容字段；当前按体积轮转
  #   backupNum: 20
  #   backupTime: 168        # 小时（168 = 7天）
  #   compress: false
  #   buffSize: 8192

trace:
  enable: true               # 自动 Init；false/缺省不启
  exporter: stdout           # stdout | otlp | otlphttp | none
  # otlp:
  #   endpoint: "localhost:4318"
  #   insecure: true
  serviceName: "my-service"
  sampler: 1.0               # 0.0~1.0
```

### 编写业务代码

```go
package main

import (
    "github.com/cyjaysong/renhe/frame/r"
    "github.com/cyjaysong/renhe/net/rhttp"
    "github.com/cyjaysong/renhe/os/rctx"
    "github.com/labstack/echo/v5/middleware"
)

func main() {
    // Trace 由 trace.enable 自动启停；退出统一收尾
    defer r.Close()

    httpSrv := r.HttpSrv()
    httpSrv.Use(
        middleware.Recover(),
        rhttp.WriteBizResJSON(),
    )

    api := httpSrv.Group("/api")
    rhttp.EchoRegisterCtrlPointers(api, new(YourController))

    if err := httpSrv.Run(); err != nil {
        r.Log().Fatal(rctx.GetInitCtx(), "http server run failed", "err", err)
    }
}
```

控制器方法签名约定：

```go
func (c *YourController) Hello(ctx *echo.Context, req *HelloReq) (*HelloRes, error)
```

- `*BizReq` 内嵌 `rhttp.HttpApiMeta`，用 tag 声明 `path` / `method` / `name`
- Ctrl **不允许**辅助公开方法：每个公开方法都必须是上述签名且带 `HttpApiMeta`，否则启动 Fatal
- 成功返回的 `*BizRes` 由 `WriteBizResJSON` 写出；也可自行 `ctx.JSON`

### 代码生成

```bash
# 生成 DAO（entity/do/table）
rh gen dao

# 生成 Service 接口
rh gen service
```

代码生成配置通过 `hack/config.yaml` 管理，支持 `typeMapping` 和 `fieldMapping` 自定义类型映射，也支持通过 `pkgAs` 为导入包指定别名。敏感字段建议用 `entityFieldEx` 排除（如 `user.password`）。

## 核心依赖

| 组件 | 库 |
|------|------|
| HTTP | [Echo v5](https://github.com/labstack/echo) |
| ORM | [Bun](https://github.com/uptrace/bun) + PostgreSQL |
| 缓存 | [go-redis v9](https://github.com/redis/go-redis) |
| 日志 | [log/slog](https://pkg.go.dev/log/slog)（经 rlog 门面） |
| 配置 | [Viper](https://github.com/spf13/viper) |
| 校验 | [go-playground/validator](https://github.com/go-playground/validator) |
| CLI | [Cobra](https://github.com/spf13/cobra) |

## License

MIT
