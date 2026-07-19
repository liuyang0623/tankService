# 部署实战全记录：从 GitHub 到腾讯云自动部署

> 本文档记录 tankService 从零实现「推送 main 分支自动部署到腾讯云轻量服务器」的**完整实战过程**，
> 包含每一步操作和实际踩到的 6 个坑及解决方式。
>
> - 想要**纯净的操作手册**（无踩坑记录）→ 看 [`DEPLOY.md`](./DEPLOY.md)
> - 想了解**真实过程、避免重复踩坑** → 看本文档

---

## 目标架构

```
本地开发 → git push main
                │
                ▼
        GitHub Actions（.github/workflows/deploy.yml）
                │  SSH 登录
                ▼
        腾讯云轻量服务器 /opt/tankService
                │  git reset --hard + docker compose up -d --build
                ▼
┌───────────────────────────────────────────────┐
│ 宝塔 Nginx（80/443，公网，HTTPS）              │
│   └─反代→ go-service-app 容器 127.0.0.1:3000   │
│              └─连→ go-service-mysql 容器 3306  │
└───────────────────────────────────────────────┘
        域名：tank.dayangge.site
```

**关键设计**：app 和 mysql 容器都只绑定 `127.0.0.1`，不对外网暴露；外部仅通过 Nginx 经 80/443 访问。

---

## 环境信息

| 项 | 值 |
|----|-----|
| 服务器 | 腾讯云轻量，系统 **OpenCloudOS**（RHEL 系，用 `yum`/`dnf`，**不是** `apt`） |
| Web 服务器 | **宝塔面板**自带 Nginx（配置在 `/www/server/nginx/`，**不是** `/etc/nginx/`） |
| 已装组件 | Docker、Docker Compose v2、Nginx |
| 域名 | tank.dayangge.site（已解析到服务器公网 IP） |
| 仓库 | github.com/liuyang0623/tankService |
| 项目 | Go 1.25 + Gin + MySQL 8.0 + Swagger |

---

## 一、准备阶段：新增部署文件（本地）

项目原本缺少部署配置，新增了 4 个文件：

| 文件 | 作用 |
|------|------|
| `Dockerfile` | 多阶段构建：`golang:1.25-alpine` 编译 + `alpine` 运行，`CGO_ENABLED=0` 静态链接 |
| `.github/workflows/deploy.yml` | push main / 手动触发；SSH 登录服务器拉代码、构建、健康检查 |
| `deploy/nginx/go-service.conf` | Nginx 反代模板（反代到 127.0.0.1:3000，含 WebSocket 支持） |
| `docs/DEPLOY.md` | 纯净操作手册 |

> **健康检查探活路径**用的是 `/api/docs/index.html`——这是项目里唯一无需 JWT 鉴权的路由
> （Swagger 文档，`cmd/server/main.go` 注册在 `/api/docs/*any`），curl 返回 200 即证明服务就绪。

**关键点**：`docs/` 目录（`docs.go`、`swagger.json`、`swagger.yaml`）已提交到仓库，
且被 `main.go` 以 `_ "go-service/docs"` 导入，因此 Dockerfile 构建阶段**无需**安装 swag、无需 `swag init`，直接编译。

---

## 二、服务器端一次性配置

### 1. 安装依赖（OpenCloudOS 用 yum，不是 apt）

```bash
yum install -y epel-release
yum install -y certbot python3-certbot-nginx git
```

### 2. 克隆项目到固定目录

```bash
mkdir -p /opt && cd /opt
git clone https://github.com/liuyang0623/tankService.git
cd tankService
```

> 部署路径固定为 `/opt/tankService`，需与 GitHub Secret `PROJECT_PATH` 一致。

### 3. 创建生产 `.env`

```bash
cp .env.example .env
vi .env
```

必填真实值：`DB_PASSWORD`、`JWT_SECRET`、`WECHAT_APPID`、`WECHAT_SECRET`、
`WECHAT_SUBSCRIBE_TPL_FOLLOW`、`UPYUN_*`；`CORS_ORIGIN` 填 `https://tank.dayangge.site`。

> **`DB_USER` 不能填 `root`**（原因见「踩坑 5」），填 `go_service` 之类的普通用户名。
>
> `.env` 已被 `.gitignore` 排除，不会进仓库，也不会被自动部署覆盖。

### 4. 首次启动并验证

```bash
docker compose up -d --build
docker compose ps    # app 和 mysql 都应为 Up / healthy
curl -fsS http://127.0.0.1:3000/api/docs/index.html >/dev/null && echo OK
```

### 5. 配置反向代理 + HTTPS（宝塔面板）

因服务器用**宝塔面板**管理 Nginx，走面板而非命令行（原因见「踩坑 6」）：

1. 宝塔面板 → **网站 → 添加站点**，域名填 `tank.dayangge.site`，PHP 选「纯静态」
2. 站点 → **反向代理 → 添加**，目标 URL `http://127.0.0.1:3000`，发送域名 `$host`
3. 站点 → **SSL → Let's Encrypt**，勾选域名申请证书，成功后开启「强制 HTTPS」

### 6. 腾讯云安全组放行端口

轻量服务器控制台 → **防火墙**，放行 `80`、`443`、`22`。**不要**放行 `3000`、`3306`。

---

## 三、GitHub Actions 自动部署配置

### 1. 服务器上生成专用部署密钥

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/gh_deploy -N ""
cat ~/.ssh/gh_deploy.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
cat ~/.ssh/gh_deploy          # 复制私钥全文（含 BEGIN/END 行）
```

### 2. GitHub 仓库配 5 个 Secrets

仓库 → **Settings → Secrets and variables → Actions → New repository secret**：

| Secret 名 | 值 |
|-----------|-----|
| `SERVER_HOST` | 服务器公网 IP |
| `SERVER_USER` | `root` |
| `SERVER_PORT` | `22` |
| `SSH_PRIVATE_KEY` | 上一步复制的私钥全文 |
| `PROJECT_PATH` | `/opt/tankService` |

### 3. 触发并验证

```bash
git commit --allow-empty -m "test: 触发自动部署"
git push origin main
```

GitHub 仓库 **Actions** 页看任务，变绿即成功。之后**每次 push main 自动部署**。

**验收**：浏览器打开 `https://tank.dayangge.site/api/docs/index.html`，有 Swagger 文档且地址栏有 🔒。

---

## 四、踩坑记录与解决方式

实战中依次遇到 6 个问题，全部解决。按出现顺序：

### 坑 1：`apt: command not found`

- **现象**：`apt update` 报 `command not found`。
- **原因**：服务器系统是 **OpenCloudOS**（RHEL 系），包管理器是 `yum`/`dnf`，不是 Debian 系的 `apt`。
- **解决**：改用 `yum install -y epel-release && yum install -y certbot python3-certbot-nginx git`。
  certbot 在 EPEL 源里，需先装 `epel-release`。

### 坑 2：`git pull` 报 `Failure when receiving data from the peer`

- **现象**：服务器 `git pull` / `git fetch` 传输中途被掐断，但 `git clone` 能成功。
- **原因**：国内服务器访问 GitHub HTTPS 不稳定，长传输易被中断。
- **解决**：加大 Git 缓冲并关压缩，多重试；或直接删掉重新 `clone`（clone 更容易成功）：
  ```bash
  git config --global http.postBuffer 524288000
  git config --global core.compression 0
  ```

### 坑 3：`failed to read dockerfile: no such file or directory`

- **现象**：服务器 `docker compose up --build` 找不到 Dockerfile。
- **原因**：Dockerfile 是本地新建并 commit 的，但**本地领先远端 11 个提交，一直没 push 到 GitHub**，
  服务器自然拉不到。
- **解决**：先在本地 `git push origin main`，再到服务器 `git fetch --all && git reset --hard origin/main`。
  **教训**：自动部署的前提是代码已经在 GitHub 上——先 push，服务器才拉得到。

### 坑 4：`go mod download` 报 `i/o timeout`（proxy.golang.org）

- **现象**：Docker 构建到 `go mod download` 卡住，最终 `dial tcp ...443: i/o timeout`。
- **原因**：默认 Go 模块代理是 `proxy.golang.org`（Google 服务），国内服务器直连不通。
- **解决**：Dockerfile 构建阶段加国内代理，`go mod download` 前插入：
  ```dockerfile
  ENV GOPROXY=https://goproxy.cn,direct
  ```
  （`goproxy.cn` 是七牛维护的国内代理，`direct` 为兜底直连。已提交到仓库。）

### 坑 5：MySQL 容器 unhealthy，`MYSQL_USER="root"` 报错

- **现象**：`dependency failed to start: container go-service-mysql is unhealthy`。
  日志：`MYSQL_USER="root" ... cannot be used for the root user`。
- **原因**：`.env` 里 `DB_USER=root`，compose 会把它传给 MySQL 的 `MYSQL_USER`。
  MySQL 8 禁止用 `MYSQL_USER=root` 再建 root 用户（root 已内置），容器直接退出。
- **解决**：改 `.env` 里 `DB_USER=root` → `DB_USER=go_service`（任意非 root 名）；
  因坏配置已初始化过数据卷，需 `docker compose down -v` 清卷后重启：
  ```bash
  docker compose down -v      # 库内无业务数据时才可删卷
  docker compose up -d
  ```

### 坑 6：`ln` 软链 Nginx 配置失败，`/etc/nginx/conf.d/` 不存在

- **现象**：`ln -sf ... /etc/nginx/conf.d/go-service.conf` 报 `No such file or directory`，
  但 `nginx -t` 显示配置文件在 `/www/server/nginx/conf/nginx.conf`。
- **原因**：服务器 Nginx 是**宝塔面板**装的，配置目录是 `/www/server/nginx/`，
  没有标准的 `/etc/nginx/conf.d/`。手动塞文件还会和面板冲突。
- **解决**：改用**宝塔面板**建站 → 反向代理（目标 `http://127.0.0.1:3000`）→ SSL 申请 Let's Encrypt 证书 + 强制 HTTPS。
  面板反代默认支持 WebSocket，项目消息 ws 服务正常。

---

## 五、日常使用

配置完成后，日常开发只需：

```bash
# 本地改代码 → 提交 → 推送
git add .
git commit -m "feat: xxx"
git push origin main
```

推送后 GitHub Actions 自动完成：SSH 登录服务器 → 拉最新代码 → 重新构建镜像 → 重启容器 → 健康检查。
到仓库 Actions 页看进度，变绿即上线。

**注意**：部署脚本含 `git reset --hard origin/main`，会覆盖服务器上对**仓库内文件**的手动改动。
`.env` 被 gitignore 排除，不受影响；但不要在服务器上直接改仓库内的其他文件。
