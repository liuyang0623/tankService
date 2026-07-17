// Package main is the entry point for the Go Service API server.
//
// @title           Go Service API
// @version         1.0
// @description     WeChat mini-program backend service
// @host            localhost:3000
// @BasePath        /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
package main

import (
	"log"

	_ "go-service/docs"
	"go-service/internal/auth"
	"go-service/internal/diary"
	"go-service/internal/notebook"
	"go-service/internal/follow"
	"go-service/internal/interactions"
	"go-service/internal/message"
	"go-service/internal/notification"
	"go-service/internal/posts"
	"go-service/internal/upload"
	"go-service/internal/users"
	"go-service/pkg/config"
	"go-service/pkg/database"
	"go-service/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// setupRouter creates and configures the Gin engine with CORS middleware
// and registers the /api/v1 (or cfg.APIPrefix) route group.
// Extracted for testability — main() calls this and then runs the server.
func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Configure CORS
	corsConfig := cors.Config{
		AllowOrigins: []string{cfg.CORSOrigin},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},
		AllowCredentials: true,
	}

	// When CORSOrigin is "*", use AllowAllOrigins instead (wildcard + credentials is invalid)
	if cfg.CORSOrigin == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowOrigins = nil
		corsConfig.AllowCredentials = false
	}

	r.Use(cors.New(corsConfig))

	// Global OPTIONS handler so CORS preflight requests are served for any path.
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})

	// Message service and hub (declared before route blocks so WS endpoint can reference it)
	msgSvc := message.NewMessageService(db)
	msgHub := message.NewHub()
	var msgHandler *message.MessageHandler

	// Register the API route group.
	v1 := r.Group("/" + cfg.APIPrefix)
	{
		authSvc := auth.NewAuthService(db, cfg.JWTSecret, cfg.WechatAppID, cfg.WechatSecret)
		authHandler := auth.NewAuthHandler(authSvc)
		v1.POST("/auth/wechat/login", authHandler.WechatLogin)

		// User routes
		userSvc := users.NewUserService(db)
		userHandler := users.NewUserHandler(userSvc)

		// Notification service (declared before follow so follow can trigger notifications)
		notificationSvc := notification.NewNotificationService(db)
		notificationHandler := notification.NewNotificationHandler(notificationSvc)

		// Follow service (also provides follow stats to the user handler)
		followSvc := follow.NewFollowService(db)
		followSvc.SetNotifier(notificationSvc)
		followHandler := follow.NewFollowHandler(followSvc)
		userHandler.SetFollowStats(followSvc)

		// Protected user routes (require JWT)
		authorized := v1.Group("")
		authorized.Use(middleware.JWTMiddleware(cfg.JWTSecret))
		authorized.GET("/users/profile", userHandler.GetProfile)
		authorized.PATCH("/users/profile", userHandler.UpdateProfile)
		authorized.POST("/users/:id/follow", followHandler.ToggleFollow)

		// Optional-auth user routes (personalize isFollowing when logged in)
		optionalUser := v1.Group("")
		optionalUser.Use(middleware.OptionalJWTMiddleware(cfg.JWTSecret))
		optionalUser.GET("/users/:id", userHandler.GetUser)
		optionalUser.GET("/users/:id/followers", followHandler.ListFollowers)
		optionalUser.GET("/users/:id/following", followHandler.ListFollowing)

		// Post routes
		postSvc := posts.NewPostService(db)
		postHandler := posts.NewPostHandler(postSvc)

		// Interactions
		interactionSvc := interactions.NewInteractionService(db)
		interactionHandler := interactions.NewInteractionHandler(interactionSvc)

		// Public post routes (optional auth: personalize isLiked/isFavorited when logged in)
		publicPost := v1.Group("")
		publicPost.Use(middleware.OptionalJWTMiddleware(cfg.JWTSecret))
		publicPost.GET("/posts", postHandler.FindAll)
		publicPost.GET("/categories", postHandler.ListCategories)
		publicPost.GET("/posts/:id/comments", interactionHandler.GetPostComments)
		publicPost.GET("/posts/:id", postHandler.FindOne)
		publicPost.GET("/users/:id/posts", postHandler.FindUserPosts)

		// Protected post routes (require JWT)
		authorized.POST("/posts", postHandler.Create)
		authorized.GET("/posts/drafts", postHandler.FindDrafts)
		authorized.GET("/posts/my", postHandler.FindMyPosts)
		authorized.PATCH("/posts/:id", postHandler.Update)
		authorized.DELETE("/posts/:id", postHandler.Remove)
		authorized.POST("/posts/:id/publish", postHandler.Publish)

		// Interaction routes (require JWT)
		authorized.POST("/posts/:id/like", interactionHandler.LikePost)
		authorized.POST("/posts/:id/favorite", interactionHandler.FavoritePost)
		authorized.GET("/users/me/favorites", interactionHandler.GetUserFavorites)
		authorized.POST("/comments", interactionHandler.CreateComment)
		authorized.DELETE("/comments/:id", interactionHandler.DeleteComment)

		// Diary routes (require JWT — all private, only self)
			diarySvc := diary.NewDiaryService(db)
			diaryHandler := diary.NewDiaryHandler(diarySvc)
			authorized.POST("/diaries", diaryHandler.Create)
			authorized.GET("/diaries", diaryHandler.FindMine)
			authorized.GET("/diaries/:id", diaryHandler.FindOne)
			authorized.PATCH("/diaries/:id", diaryHandler.Update)
			authorized.DELETE("/diaries/:id", diaryHandler.Remove)

			// Notebook routes (require JWT — all private, only self)
			notebookSvc := notebook.NewNotebookService(db)
			notebookHandler := notebook.NewNotebookHandler(notebookSvc)
			authorized.POST("/notebooks", notebookHandler.Create)
			authorized.GET("/notebooks", notebookHandler.FindMine)
			authorized.PATCH("/notebooks/:id", notebookHandler.Update)
			authorized.DELETE("/notebooks/:id", notebookHandler.Remove)

			// Upload routes (require JWT)
		uploadSvc := upload.NewUpYunService(cfg.UpyunBucket, cfg.UpyunOperator, cfg.UpyunPassword, cfg.UpyunEndpoint, cfg.UpyunDomain)
		uploadHandler := upload.NewUploadHandler(uploadSvc)
		authorized.POST("/upload/image", uploadHandler.UploadImage)
		authorized.POST("/upload/file", uploadHandler.UploadFile)

		// Message routes (require JWT)
		msgHandler = message.NewMessageHandler(msgSvc, msgHub, cfg.JWTSecret)
		authorized.GET("/conversations", msgHandler.ListConversations)
		authorized.GET("/conversations/:id/messages", msgHandler.GetMessages)
		authorized.POST("/messages", msgHandler.SendMessage)
		authorized.POST("/conversations/:id/read", msgHandler.MarkRead)

		// Notification routes (require JWT)
		authorized.GET("/notifications", notificationHandler.List)
		authorized.POST("/notifications/read", notificationHandler.MarkRead)
		authorized.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	}

	// WebSocket endpoint (no JWT middleware; token validated inside handler)
	r.GET("/ws", msgHandler.HandleWS)

	// Swagger UI endpoint
	r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the WebSocket hub event loop
	go msgHub.Run()

	return r
}

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Connect to the database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 3. Set up the router (CORS + route groups)
	r := setupRouter(database.DB, cfg)

	// Auto Migrate Database
	if err := database.AutoMigrate(
		&users.User{},
		&posts.Post{},
		&posts.PostImage{},
		&posts.Topic{},
		&diary.Diary{},
		&diary.DiaryImage{},
		&notebook.Notebook{},
		&interactions.Comment{},
		&interactions.Like{},
		&interactions.Favorite{},
		&follow.Follow{},
		&message.Conversation{},
		&message.Message{},
		&notification.Notification{},
	); err != nil {
		log.Fatalf("failed to auto migrate database: %v", err)
	}

	// 4. Start the HTTP server
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}