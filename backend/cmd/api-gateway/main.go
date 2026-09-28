package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"shalom/internal/config"
	"shalom/internal/handler"
	"shalom/internal/middleware"
	"shalom/internal/repository"
	"shalom/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// ============================================================
	// DATABASE CONNECTIONS
	// ============================================================

	// PostgreSQL with connection pooling
	pgConfig, err := pgxpool.ParseConfig(cfg.Database.PostgresURL)
	if err != nil {
		log.Fatalf("Unable to parse PostgreSQL config: %v", err)
	}
	pgConfig.MaxConns = 25
	pgConfig.MinConns = 5
	pgConfig.MaxConnLifetime = 30 * time.Minute
	pgConfig.MaxConnIdleTime = 10 * time.Minute
	pgConfig.HealthCheckPeriod = 5 * time.Minute

	pgPool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	// Verify PostgreSQL connection
	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL (pool: 5-25 connections)")

	// MongoDB
	mongoOpts := options.Client().ApplyURI(cfg.Database.MongoURL)
	mongoOpts.SetMaxPoolSize(50)
	mongoOpts.SetMinPoolSize(5)
	mongoClient, err := mongo.Connect(ctx, mongoOpts)
	if err != nil {
		log.Fatalf("Unable to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Unable to ping MongoDB: %v", err)
	}
	mongoDB := mongoClient.Database(cfg.Database.MongoDB)
	log.Println("Connected to MongoDB")

	// Redis with connection pooling
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     20,
		MinIdleConns: 5,
		PoolTimeout:  30 * time.Second,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// ============================================================
	// REPOSITORIES
	// ============================================================

	repos := repository.NewRepositories(pgPool)

	// ============================================================
	// SERVICES
	// ============================================================

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, rdb)
	authService := service.NewAuthService(rdb, authMiddleware, repos.User)
	meetingService := service.NewMeetingService(cfg, repos.Meeting)
	chatService := service.NewChatService(mongoDB)
	bandwidthOptimizer := service.NewBandwidthOptimizer(rdb)

	// ============================================================
	// HANDLERS
	// ============================================================

	authHandler := handler.NewAuthHandler(authService)
	meetingHandler := handler.NewMeetingHandler(meetingService)
	chatHandler := handler.NewChatHandler(chatService)
	bandwidthHandler := handler.NewBandwidthHandler(bandwidthOptimizer)

	// ============================================================
	// ROUTER
	// ============================================================

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(rdb, 100, time.Minute))

	// Health endpoints
	r.GET("/health", func(c *gin.Context) {
		dbOK := pgPool.Ping(c.Request.Context()) == nil
		redisOK := rdb.Ping(c.Request.Context()).Err() == nil
		status := "ok"
		code := http.StatusOK
		if !dbOK || !redisOK {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{
			"status":     status,
			"service":    "shalom",
			"postgres":   dbOK,
			"redis":      redisOK,
			"uptime":     time.Since(startTime).String(),
		})
	})

	r.GET("/health/ready", func(c *gin.Context) {
		if err := pgPool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "postgres not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Public routes
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// Protected routes
	auth := r.Group("/api")
	auth.Use(authMiddleware.AuthRequired())
	{
		// Auth
		auth.POST("/auth/logout", authHandler.Logout)
		auth.GET("/auth/profile", authHandler.GetProfile)

		// Meetings
		auth.POST("/meetings", meetingHandler.CreateMeeting)
		auth.GET("/meetings/:id", meetingHandler.GetMeeting)
		auth.POST("/meetings/:id/start", meetingHandler.StartMeeting)
		auth.POST("/meetings/join", meetingHandler.JoinMeeting)
		auth.POST("/meetings/:id/end", meetingHandler.EndMeeting)
		auth.GET("/meetings/:id/participants", meetingHandler.GetParticipants)
		auth.PATCH("/meetings/:id/participant", meetingHandler.UpdateParticipant)
		auth.DELETE("/meetings/:id/participants/:userId", meetingHandler.KickParticipant)

		// Chat
		auth.POST("/chat/messages/text", chatHandler.SendTextMessage)
		auth.POST("/chat/messages/file", chatHandler.SendFileMessage)
		auth.POST("/chat/messages/voice", chatHandler.SendVoiceMessage)
		auth.POST("/chat/messages/poll", chatHandler.SendPoll)
		auth.POST("/chat/messages/poll/vote", chatHandler.VotePoll)
		auth.POST("/chat/messages/reaction", chatHandler.AddReaction)
		auth.DELETE("/chat/messages/reaction", chatHandler.RemoveReaction)
		auth.PATCH("/chat/messages/edit", chatHandler.EditMessage)
		auth.DELETE("/chat/messages", chatHandler.DeleteMessage)
		auth.GET("/chat/rooms/:roomId/messages", chatHandler.GetMessages)
		auth.POST("/chat/rooms", chatHandler.CreateChatRoom)
		auth.GET("/chat/rooms", chatHandler.GetChatRooms)

		// Bandwidth
		auth.GET("/bandwidth/profile", bandwidthHandler.GetProfile)
		auth.POST("/bandwidth/switch", bandwidthHandler.SwitchMode)
		auth.POST("/bandwidth/report", bandwidthHandler.ReportNetwork)
		auth.GET("/bandwidth/presets", bandwidthHandler.GetPresets)
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Shalom server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

var startTime = time.Now()
