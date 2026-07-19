# 部署手册：GitHub 推送 main 自动部署到腾讯云轻量服务器

本项目通过 **GitHub Actions** 实现「推送 `main` 分支 → 自动部署」。
部署方式为 **SSH 登录服务器本地构建**：Actions 通过 SSH 登录服务器，拉取最新代码后
在服务器上 `docker compose up -d --build` 重新构建并重启，无需镜像仓库。

架构：

```
GitHub push main
      │
      ▼
GitHub Actions (.github/workflows/deploy.yml)
      │  SSH
      ▼
腾讯云服务器  /opt/tankService
      │  docker compose up -d --build
      ▼
┌─────────────────────────────┐
│ Nginx (80/443, 公网)        │
│   └─反代→ app 容器 127.0.0.1:3000 │
│           └─连→ mysql 容器 127.0.0.1:3306 │
└─────────────────────────────┘
```

- `app`、`mysql` 容器均只绑定 `127.0.0.1`，不对外网暴露
- 外部仅通过 Nginx 经 `80/443` 访问
- MySQL 数据持久化在 Docker volume `mysql_data`

---

## 一、服务器端一次性配置

> 以下命令在腾讯云服务器上执行。假设登录用户为 `root`（若为 `ubuntu` 等，命令前按需加 `sudo`）。

### 1. 确认依赖

```bash
docker --version
docker compose version          # 需 v2（compose 插件形式）
nginx -v
# 安装 certbot（用于签发 Let's Encrypt HTTPS 证书）
apt update && apt install -y certbot python3-certbot-nginx
```

### 2. 克隆项目到固定目录

```bash
mkdir -p /opt && cd /opt
# 方式 A：HTTPS（公开仓库或后续用 token）
git clone https://github.com/liuyang0623/tankService.git
# 方式 B：SSH（需在服务器上配 GitHub deploy key）
# git clone git@github.com:liuyang0623/tankService.git
cd tankService
```

> 部署路径固定为 `/opt/tankService`，需与 GitHub Secret `PROJECT_PATH` 一致。

### 3. 创建生产 `.env`

```bash
cp .env.example .env
vi .env
```

必须填写的真实值：

| 变量 | 说明 |
|------|------|
| `DB_PASSWORD` | MySQL 密码（compose 会用它建库并连接） |
| `DB_NAME` / `DB_USER` | 库名 / 用户名（保持与 compose 默认一致即可） |
| `JWT_SECRET` | JWT 签名密钥，务必用强随机值 |
| `WECHAT_APPID` / `WECHAT_SECRET` | 微信小程序凭证 |
| `WECHAT_SUBSCRIBE_TPL_FOLLOW` | 订阅消息模板 ID |
| `UPYUN_*` | 又拍云存储配置 |
| `CORS_ORIGIN` | 前端域名，如 `https://tank.dayangge.site`；不要在生产用 `*` |

> `.env` 已被 `.gitignore` 排除，不会进仓库，也不会被自动部署覆盖。

### 4. 首次启动

```bash
docker compose up -d --build
docker compose ps          # 确认 app 与 mysql 均为 Up / healthy
curl -fsS http://127.0.0.1:3000/api/docs/index.html >/dev/null && echo OK
```

### 5. 配置 Nginx 反向代理

```bash
# server_name 已在模板中设为 tank.dayangge.site，直接软链到 nginx
ln -sf /opt/tankService/deploy/nginx/go-service.conf \
  /etc/nginx/conf.d/go-service.conf

nginx -t && systemctl reload nginx
```

### 6. 签发 HTTPS 证书

```bash
certbot --nginx -d tank.dayangge.site
# certbot 会自动改写上面的 conf：加 443/SSL、加 80→443 跳转，并配置自动续期
```

### 7. 腾讯云安全组

在轻量服务器控制台「防火墙」放行：

- `80`（HTTP，certbot 验证 + 跳转用）
- `443`（HTTPS）
- `22`（SSH，供 GitHub Actions 登录）

`3000`、`3306` **不要**对外放行（已绑 `127.0.0.1`）。

---

## 二、GitHub 端一次性配置

### 1. 生成专用部署密钥（在你本地电脑执行）

```bash
ssh-keygen -t ed25519 -C "github-deploy-tankService" -f ~/.ssh/tank_deploy -N ""
```

生成两个文件：`~/.ssh/tank_deploy`（私钥）、`~/.ssh/tank_deploy.pub`（公钥）。

### 2. 把公钥加到服务器

```bash
# 将公钥内容追加到服务器登录用户的 authorized_keys
cat ~/.ssh/tank_deploy.pub | ssh 用户名@服务器IP 'cat >> ~/.ssh/authorized_keys'
```

验证专用密钥能免密登录：

```bash
ssh -i ~/.ssh/tank_deploy 用户名@服务器IP 'echo login-ok'
```

### 3. 在 GitHub 仓库配置 Secrets

仓库页面 → **Settings → Secrets and variables → Actions → New repository secret**，添加：

| Secret 名 | 值 |
|-----------|-----|
| `SERVER_HOST` | 服务器公网 IP |
| `SERVER_USER` | 登录用户名（如 `root` / `ubuntu`） |
| `SERVER_PORT` | `22` |
| `SSH_PRIVATE_KEY` | `~/.ssh/tank_deploy` **私钥全文**（含 `-----BEGIN/END-----` 行） |
| `PROJECT_PATH` | `/opt/tankService` |

> 私钥全文获取：`cat ~/.ssh/tank_deploy`，整段复制粘贴。

---

## 三、日常使用

### 自动部署

推送到 `main` 即自动部署：

```bash
git push origin main
```

在 GitHub 仓库 **Actions** 标签页可看到部署进度与日志。

### 手动部署

Actions 页面 → 选 **Deploy to Tencent Cloud** → **Run workflow**（`workflow_dispatch`）。

### 回滚

- **推荐**：在 GitHub 上对问题提交做 `Revert` 再 push，触发一次干净的重新部署
- **应急**：SSH 登录服务器，`cd /opt/tankService && git reset --hard <好的commit> && docker compose up -d --build`

### 常用排查命令（服务器上）

```bash
docker compose ps                    # 容器状态
docker compose logs -f app           # 实时应用日志
docker compose logs --tail=100 app   # 最近 100 行
docker compose restart app           # 只重启应用
docker compose down && docker compose up -d --build   # 完全重建
```

---

## 四、部署链路说明

`.github/workflows/deploy.yml` 在服务器上依次执行：

1. `git fetch --all && git reset --hard origin/main` —— 强制对齐远端 main（丢弃服务器本地改动）
2. `docker compose up -d --build` —— 重新构建镜像并滚动重启
3. `docker image prune -f` —— 清理悬空旧镜像，防磁盘堆积
4. 健康检查 —— 轮询 `http://127.0.0.1:3000/api/docs/index.html`，最多等 60s；失败则打印应用日志并让 job 失败

> 健康检查用无鉴权的 Swagger 文档路由（`cmd/server/main.go` 注册的 `/api/docs`），
> 返回 200 即证明服务已正常启动。
