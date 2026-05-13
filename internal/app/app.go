package app

import (
	"DeepSight/internal/agent/research"
	"DeepSight/internal/agent/tools"
	"DeepSight/internal/config"
	"DeepSight/internal/database"
	"DeepSight/internal/handler"
	"DeepSight/internal/middleware"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"
	"DeepSight/internal/service"
	"DeepSight/internal/util/jwt"
	"DeepSight/internal/worker"
	"fmt"

	"github.com/gin-gonic/gin"
)

func Run() error {
	cfg, err := config.Load("configs/application.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := database.Initialize(&cfg.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	if err := database.InitializeRedis(&cfg.Redis); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}
	if err := database.InitializeRabbitMQ(&cfg.RabbitMQ); err != nil {
		return fmt.Errorf("failed to initialize rabbitmq: %w", err)
	}
	if err := database.InitializeRustFS(&cfg.RustFS); err != nil {
		return fmt.Errorf("failed to initialize rustfs: %w", err)
	}
	if err := service.InitializeLLM(&cfg.OpenAI); err != nil {
		return fmt.Errorf("failed to initialize llm: %w", err)
	}
	jwt.Initialize(&cfg.JWT)

	db := database.GetDB()
	if err := db.AutoMigrate(
		&model.User{},
		&model.KnowledgeBase{},
		&model.File{},
		&model.Chunk{},
		&model.Conversation{},
		&model.Message{},
		&model.AnalysisReport{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg.JWT.ExpireDuration())
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, userService)

	kbRepo := repository.NewKnowledgeBaseRepository(db)
	fileRepo := repository.NewFileRepository(db)
	kbService := service.NewKnowledgeBaseService(kbRepo, fileRepo)
	kbHandler := handler.NewKnowledgeBaseHandler(kbService)
	chunkRepo := repository.NewChunkRepository(db)
	fileService := service.NewFileService(fileRepo, chunkRepo, kbRepo)
	fileHandler := handler.NewFileHandler(fileService)

	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	chatService := service.NewChatService(convRepo, msgRepo, chunkRepo, fileRepo, kbRepo)
	chatHandler := handler.NewChatHandler(chatService)

	analysisRepo := repository.NewAnalysisRepository(db)
	analysisService := service.NewAnalysisService(
		analysisRepo,
		func(userID, kbID, convID uint) service.ResearchRunner {
			webSearch := tools.NewWebSearchTool(cfg.Tavily.APIKey)
			return research.NewCoordinator(
				service.LLM,
				analysisRepo,
				chunkRepo,
				fileRepo,
				kbRepo,
				msgRepo,
				webSearch,
				userID,
				kbID,
				convID,
			)
		},
	)
	analysisHandler := handler.NewAnalysisHandler(analysisService)

	// 启动文件处理 Worker
	fileWorker := worker.NewFileWorker(fileRepo, chunkRepo)
	if err := fileWorker.Start(); err != nil {
		return fmt.Errorf("failed to start file worker: %w", err)
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// CORS 中间件
	router.Use(middleware.CorsMiddleware())

	apiV1 := router.Group("/api/v1")
	apiV1.POST("/register", authHandler.Register)
	apiV1.POST("/auth/login", authHandler.Login)

	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(authService))
	protected.POST("/auth/logout", authHandler.Logout)
	userGroup := protected.Group("/users")
	userHandler.RegisterRoutes(userGroup)

	kbsGroup := protected.Group("/knowledge-bases")
	kbHandler.RegisterRoutes(kbsGroup)
	fileHandler.RegisterRoutes(kbsGroup)

	analysisGroup := protected.Group("/analysis")
	analysisHandler.RegisterRoutes(analysisGroup)

	chatGroup := protected.Group("/conversations")
	chatHandler.RegisterRoutes(chatGroup)

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
