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
	"shalom/internal/service"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	ctx := context.Background()
	pgPool, err := pgxpool.New(ctx, cfg.Database.PostgresURL)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()
	log.Println("Connected to PostgreSQL")

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.MongoURL))
	if err != nil {
		log.Fatalf("Unable to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)
	mongoDB := mongoClient.Database(cfg.Database.MongoDB)
	log.Println("Connected to MongoDB")

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, rdb)

	// Initialize services
	authService := service.NewAuthService(pgPool, rdb, authMiddleware)
	meetingService := service.NewMeetingService(pgPool, cfg)
	chatService := service.NewChatService(mongoDB)
	bandwidthOptimizer := service.NewBandwidthOptimizer(rdb)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	meetingHandler := handler.NewMeetingHandler(meetingService)
	chatHandler := handler.NewChatHandler(chatService)
	bandwidthHandler := handler.NewBandwidthHandler(bandwidthOptimizer)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(rdb, 100, time.Minute))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "shalom"})
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
