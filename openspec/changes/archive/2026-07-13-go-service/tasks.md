## 1. 项目骨架（project-bootstrap）

- [x] 1.1 初始化 Go 模块（go mod init），添加核心依赖（gin, gorm, mysql driver, jwt, swag, godotenv）
- [x] 1.2 创建目录结构：cmd/server/, internal/, pkg/config/, pkg/database/, pkg/middleware/, pkg/response/
- [x] 1.3 实现配置加载（pkg/config）：读取 .env，定义 Config 结构体
- [x] 1.4 实现数据库连接（pkg/database）：GORM + MySQL，AutoMigrate 所有数据模型
- [x] 1.5 实现统一响应格式（pkg/response）：Success/Error 帮助函数
- [x] 1.6 实现主程序入口（cmd/server/main.go）：加载配置、连接数据库、注册路由、启动服务
- [x] 1.7 集成 Swagger（swaggo/swag）：main.go 添加注释，配置 /api/docs 路由
- [x] 1.8 启动验证：`go run ./cmd/server/` 成功启动，访问 /api/docs 可见 Swagger UI

## 2. 认证模块（auth）

- [x] 2.1 定义 User 数据模型（internal/auth/model.go 或 internal/users/model.go）
- [x] 2.2 实现 JWT 中间件（pkg/middleware/jwt.go）：解析 Bearer token，注入 userID 到 context
- [x] 2.3 实现微信登录 service（internal/auth/service.go）：调用微信 jscode2session，查找/创建用户，签发 JWT
- [x] 2.4 实现认证 handler（internal/auth/handler.go）：POST /api/v1/auth/wechat/login，添加 Swagger 注释
- [x] 2.5 注册 auth 路由，验证：POST /api/v1/auth/wechat/login 返回 token

## 3. 用户模块（users）

- [x] 3.1 实现用户 service（internal/users/service.go）：GetProfile, UpdateProfile, FindOne
- [x] 3.2 实现用户 handler（internal/users/handler.go）：GET/PATCH /api/v1/users/profile，GET /api/v1/users/:id，添加 Swagger 注释
- [x] 3.3 注册 users 路由（profile 接口需 JWT 中间件），验证接口可正常返回

## 4. 文章模块（posts）

- [x] 4.1 定义 Post、PostImage、Topic、PostTopic 数据模型（internal/posts/model.go）
- [x] 4.2 实现文章 service（internal/posts/service.go）：Create, FindAll, FindOne, Update, Remove, Publish, FindDrafts, FindUserPosts
- [x] 4.3 实现文章 handler（internal/posts/handler.go）：所有文章 CRUD 接口，添加 Swagger 注释
- [x] 4.4 注册 posts 路由，验证：创建草稿、发布文章、获取列表、获取详情

## 5. 互动模块（interactions）

- [x] 5.1 定义 Like、Favorite、Comment 数据模型（internal/interactions/model.go）
- [x] 5.2 实现互动 service（internal/interactions/service.go）：LikePost, FavoritePost, GetUserFavorites, CreateComment, GetPostComments, DeleteComment
- [x] 5.3 实现互动 handler（internal/interactions/handler.go）：所有互动接口，添加 Swagger 注释
- [x] 5.4 注册 interactions 路由，验证：点赞/取消点赞、收藏、评论功能正常

## 6. 上传模块（upload）

- [x] 6.1 实现又拍云签名服务（internal/upload/service.go）：HMAC-SHA1 签名、生成文件路径、构造 Authorization 头
- [x] 6.2 实现上传 handler（internal/upload/handler.go）：POST /api/v1/upload/image 和 /api/v1/upload/file，文件类型和大小校验，添加 Swagger 注释
- [x] 6.3 注册 upload 路由（需 JWT 中间件），验证：上传图片返回 URL
