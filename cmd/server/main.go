package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/auth"
	"github.com/imkerbos/ACME-Console/internal/config"
	"github.com/imkerbos/ACME-Console/internal/crypto"
	"github.com/imkerbos/ACME-Console/internal/handler"
	"github.com/imkerbos/ACME-Console/internal/logger"
	"github.com/imkerbos/ACME-Console/internal/model"
	"github.com/imkerbos/ACME-Console/internal/router"
	"github.com/imkerbos/ACME-Console/internal/scheduler"
	"github.com/imkerbos/ACME-Console/internal/service"
)

func main() {
	configPath := flag.String("config", "configs/config.dev.yaml", "path to config file")
	staticDir := flag.String("static", "", "path to static files directory (optional)")
	dev := flag.Bool("dev", true, "development mode")
	flag.Parse()

	// Set gin mode early
	if *dev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize logger
	if err := logger.Init(*dev); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", logger.Err(err))
	}

	logger.Info("Configuration loaded",
		logger.String("config", *configPath),
		logger.Int("port", cfg.Server.Port),
	)

	// Initialize database
	db, err := model.InitDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", logger.Err(err))
	}
	logger.Info("Database connected",
		logger.String("host", cfg.Database.Host),
		logger.String("database", cfg.Database.DBName),
	)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.ExpireHours)*time.Hour,
	)

	// Initialize setting service (reads from database)
	settingSvc := service.NewSettingService(db)

	// Initialize workspace service
	workspaceSvc := service.NewWorkspaceService(db)

	// Initialize notification service
	notificationSvc := service.NewNotificationService(db)

	// Initialize certificate service
	var certSvc *service.CertificateService

	// Check if encryption key is configured (required for real ACME)
	if cfg.Encryption.MasterKey != "" {
		encryptor, err := crypto.NewEncryptor(cfg.Encryption.MasterKey)
		if err != nil {
			logger.Fatal("Failed to initialize encryptor", logger.Err(err))
		}

		// Use database settings for ACME config
		legoSvc := service.NewLegoServiceWithSettings(db, settingSvc, encryptor)
		certSvc = service.NewCertificateServiceWithLego(db, legoSvc)
		logger.Info("ACME service initialized (production environment)")
	} else {
		// Fall back to mock service
		acmeSvc := service.NewAcmeShService()
		certSvc = service.NewCertificateService(db, acmeSvc)
		logger.Info("Using mock ACME service (no encryption key configured)")
	}

	// Initialize handlers
	handlers := &router.Handlers{
		Auth:         handler.NewAuthHandler(db, jwtManager),
		Certificate:  handler.NewCertificateHandler(certSvc),
		Challenge:    handler.NewChallengeHandler(certSvc),
		User:         handler.NewUserHandler(db),
		Setting:      handler.NewSettingHandler(settingSvc),
		Workspace:    handler.NewWorkspaceHandler(workspaceSvc),
		Notification: handler.NewNotificationHandler(notificationSvc),
	}

	// Setup static file serving
	var staticFS fs.FS
	if *staticDir != "" {
		staticFS = os.DirFS(*staticDir)
		logger.Info("Serving static files", logger.String("path", *staticDir))
	}

	// Setup router
	r := router.Setup(handlers, jwtManager, staticFS)

	// Start notification scheduler
	notifScheduler := scheduler.NewScheduler(notificationSvc)
	notifScheduler.Start()
	defer notifScheduler.Stop()

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Starting server", logger.String("address", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", logger.Err(err))
	}
}
