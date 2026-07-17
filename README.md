# MSE Gateway Domain TLS Updater

一个基于 `Gin + React` 的 MSE 网关域名 TLS 管理应用，支持：

- 多账号配置与切换（每个账号可配置多个网关）。
- 按 `config.yaml` 配置批量更新指定网关下所有域名的 TLS 版本。
- 页面输入单个域名，校验是否属于当前网关后执行 TLS 替换。
- Dry-run 预览（仅计算，不实际更新）。
- 操作审计日志记录与查询。
- 查看当前网关域名列表及 TLS 信息。
- 用户注册/登录、角色权限管理（`root` / `admin` / `watcher`）。
- 忘记密码（注册邮箱找回、一次性重置链接、30 分钟有效、频率限制、安全审计、改密成功邮件通知）。

## 目录结构

```text
.
├── cmd/server/main.go            # 后端入口
├── internal/config               # 配置加载
├── internal/audit                # 审计日志
├── internal/mse                  # MSE SDK 调用封装
├── internal/service              # 业务逻辑（批量更新/单域名更新）
├── internal/httpapi              # Gin 路由与接口
├── frontend                      # React 前端
└── config.yaml                   # 应用配置
```

## 配置说明

编辑根目录 `config.yaml`：

- `server.port`：服务端口（默认 `8080`）。
- `aliyun.accept_language`：阿里云 API 返回语言（`zh` / `en`）。
- `update.protocol` / `update.must_https` / `update.http2` / `update.batch_tls_version`：TLS 更新默认参数。
- `audit.enabled` / `audit.file`：审计配置（当前审计数据落 MySQL，`audit.file` 保留兼容）。
- `auth.jwt_secret` / `auth.token_expiry`：登录 token 配置。
- `auth.db_*`：MySQL 连接配置。
- `auth.root_initial_*`：首次启动初始化根用户信息。
- `auth.smtp_*`：SMTP 配置，用于忘记密码与改密成功通知。

配置示例（建议使用仓库 `config-example.yaml`，并复制为本地 `config.yaml`）：

```yaml
# 服务监听配置
server:
  # HTTP 服务端口
  port: 8080

# 阿里云 MSE 账号与网关配置
aliyun:
  # OpenAPI 返回语言：zh / en
  accept_language: zh

# 批量/单次更新的公共默认参数
update:
  protocol: HTTPS              # 协议类型：HTTP / HTTPS
  must_https: false            # 是否强制 HTTPS
  http2: close                 # Http2 配置：open / close / globalConfig
  batch_tls_version: tlsv1.2   # 兜底 TLS 版本（当网关未配置 batch_tls_version 时生效）

# 审计日志配置
audit:
  enabled: true                # 是否开启审计日志
  file: "./logs/audit.log"     # 审计日志文件路径（JSON 行格式）

# 更新接口鉴权配置
auth:
  update_token: "xxxxxxx"
  jwt_secret: "your-jwt-secret-key-change-in-production"
  token_expiry: "24h"
  db_host: "127.0.0.1"
  db_port: 3306
  db_user: "root"
  db_password: "xxxxxx"
  db_name: "gateway_tls"
  root_initial_name: "root"
  root_initial_phone: "13800000000"
  root_initial_email: "root@example.com"
  root_initial_password: "Admin@AiLab.2026" #默认登录密码，登录后请修改
  smtp_host: "smtp.163.com"
  smtp_port: 465
  smtp_user: "example@163.com"
  smtp_pass: "smtp_authorization_code"
  smtp_from: "example@163.com"
```

> 安全建议：
> - `config.yaml` 请仅用于本地或受控环境，不要提交真实密钥到仓库。
> - `smtp_pass` 需使用邮箱 SMTP 授权码，而非邮箱登录密码。

## 启动后端

```bash
go mod tidy
go run ./cmd/server
```

默认端口：`8080`（可在 `config.yaml` 修改）。

## 启动前端（开发模式）

```bash
cd frontend
npm install
npm run dev
```

默认地址：`http://localhost:5173`，并通过 Vite 代理请求后端 `http://localhost:8080`。

## 构建并由 Gin 托管前端静态文件

```bash
cd frontend
npm install
npm run build
cd ..
go run ./cmd/server
```

构建后访问 `http://localhost:8080/` 即可打开页面。

## Docker 构建与运行

根目录已提供 `Dockerfile`（前端/后端多阶段构建，最终镜像基于 `nginx`，并反向代理到容器内 Go 服务）。

```bash
# 1) 构建镜像
docker build -t mse-gateway-domain:latest .

# 2) 准备配置文件（按需修改）
cp config-example.yaml config.yaml

# 3) 启动容器（示例：宿主机 8080 -> 容器 80）
docker run --rm -p 8080:80 \
  -v "$(pwd)/config.yaml:/app/config.yaml:ro" \
  --name mse-gateway-domain \
  mse-gateway-domain:latest
```

> 注意：容器需要能够访问你配置中的 MySQL 地址（`auth.db_host` / `auth.db_port`）。

## 后端 API

### 公开接口（无需登录）

- `GET /api/health`：健康检查。
- `GET /api/captcha`：获取 4 位数字验证码图片（5 分钟有效）。
- `GET /api/check-exists?field=username|phone|email&value=...`：检查注册字段是否存在。
- `POST /api/register`：用户注册（需传 `captcha_token` + `captcha_code`，验证码校验通过后才可注册）。
- `POST /api/login`：登录（支持用户名或手机号，需传 `captcha_token` + `captcha_code`，验证码校验通过后才可登录）。
- `POST /api/reset-password`：忘记密码（传入 `account`，系统向注册邮箱发送重置链接）。
- `POST /api/change-password`：使用重置 token 设置新密码。

### 登录后接口（需 `Authorization: Bearer <jwt>`）

- `GET /api/user-info`：当前登录用户信息。
- `PUT /api/user-profile`：更新当前用户资料。
- `GET /api/tls-versions`：获取可选 TLS 版本。
- `GET /api/accounts`：获取可选账号列表。
- `GET /api/gateways?account_id=...`：获取指定账号网关列表。
- `GET /api/domains?...`：分页获取网关域名。

### 管理员接口（`admin` / `root`）

- `POST /api/batch-update`：批量更新 TLS（支持 dry-run；`account_id=all` 时遍历全部账号网关）。
- `POST /api/update-domain`：更新单个域名 TLS。
- `GET /api/audit-logs`：查询审计日志。
- `POST /api/accounts` / `GET /api/accounts/all` / `DELETE /api/accounts/:id`：账号管理。
- `POST /api/gateways` / `GET /api/gateways/all` / `DELETE /api/gateways/:id`：网关管理。

### 超级管理员接口（`root`）

- `GET /api/users`：用户列表。
- `POST /api/users`：创建用户。
- `PUT /api/users/:id/role`：变更角色。
- `PUT /api/users/:id/reset-password`：重置用户密码。
- `DELETE /api/users/:id`：删除用户。

## 忘记密码流程说明

- 用户在登录页输入用户名或手机号后点击“忘记密码”。
- 系统按账号查找绑定邮箱并发送重置链接（链接直达 `/reset-password?token=...`）。
- token 有效期 30 分钟，且一次性使用。
- 同一账号 1 小时最多请求 5 次重置。
- 新密码必须与旧密码不同。
- 找回和改密会记录 IP、设备、时间用于安全审计。
- 密码修改成功后发送通知邮件（标题：`MSE网关管理系统密码重置成功通知`）。

## MSE API 参考

- `UpdateGatewayDomain` 文档：
  - https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-updategatewaydomain
