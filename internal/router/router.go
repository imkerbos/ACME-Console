package router

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/auth"
	"github.com/imkerbos/ACME-Console/internal/handler"
	"github.com/imkerbos/ACME-Console/internal/middleware"
	"github.com/imkerbos/ACME-Console/internal/response"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	Certificate  *handler.CertificateHandler
	Challenge    *handler.ChallengeHandler
	User         *handler.UserHandler
	Setting      *handler.SettingHandler
	Workspace    *handler.WorkspaceHandler
	Notification *handler.NotificationHandler
}

func Setup(handlers *Handlers, jwtManager *auth.JWTManager, staticFS fs.FS) *gin.Engine {
	r := gin.New()

	// Global middleware - order matters
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health check (no auth required)
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", handlers.Auth.Login)
		}

		// Public settings (no auth required)
		v1.GET("/settings/site", handlers.Setting.GetSite)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtManager))
		{
			// Auth routes (auth required)
			protected.GET("/auth/me", handlers.Auth.GetCurrentUser)
			protected.POST("/auth/change-password", handlers.Auth.ChangePassword)
			protected.PUT("/auth/profile", handlers.Auth.UpdateProfile)

			// Workspace endpoints
			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", handlers.Workspace.List)
				workspaces.POST("", handlers.Workspace.Create)
				workspaces.GET("/:id", handlers.Workspace.Get)
				workspaces.PUT("/:id", handlers.Workspace.Update)
				workspaces.DELETE("/:id", handlers.Workspace.Delete)
				workspaces.GET("/:id/members", handlers.Workspace.ListMembers)
				workspaces.POST("/:id/members", handlers.Workspace.AddMember)
				workspaces.PUT("/:id/members/:userId", handlers.Workspace.UpdateMember)
				workspaces.DELETE("/:id/members/:userId", handlers.Workspace.RemoveMember)
			}

			// Certificate endpoints
			certs := protected.Group("/certificates")
			{
				certs.POST("", handlers.Certificate.Create)
				certs.GET("", handlers.Certificate.List)
				certs.GET("/:id", handlers.Certificate.Get)
				certs.DELETE("/:id", handlers.Certificate.Delete)
				certs.POST("/:id/pre-verify", handlers.Certificate.PreVerify)
				certs.POST("/:id/verify", handlers.Certificate.Verify)
				certs.GET("/:id/download", handlers.Certificate.Download)
				certs.GET("/:id/notification-logs", handlers.Notification.ListLogs)

				// Challenge endpoints (nested under certificates)
				certs.GET("/:id/challenges", handlers.Challenge.List)
				certs.GET("/:id/challenges/export", handlers.Challenge.Export)
			}

			// Notification endpoints
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", handlers.Notification.List)
				notifications.POST("", handlers.Notification.Create)
				notifications.GET("/:id", handlers.Notification.Get)
				notifications.PUT("/:id", handlers.Notification.Update)
				notifications.DELETE("/:id", handlers.Notification.Delete)
				notifications.POST("/:id/test", handlers.Notification.Test)
			}

			// Admin routes (admin role required)
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminAuth())
			{
				// User management
				users := admin.Group("/users")
				{
					users.GET("", handlers.User.List)
					users.POST("", handlers.User.Create)
					users.GET("/:id", handlers.User.Get)
					users.PUT("/:id", handlers.User.Update)
					users.DELETE("/:id", handlers.User.Delete)
					users.POST("/:id/reset-password", handlers.User.ResetPassword)
				}

				// Settings management
				settings := admin.Group("/settings")
				{
					settings.GET("", handlers.Setting.List)
					settings.PUT("", handlers.Setting.Update)
					settings.GET("/acme", handlers.Setting.GetACME)
					settings.PUT("/acme", handlers.Setting.UpdateACME)
					settings.PUT("/site", handlers.Setting.UpdateSite)
				}
			}
		}
	}

	// Serve static files if provided
	if staticFS != nil {
		r.Use(staticFileHandler(staticFS))
	}

	return r
}

// staticFileHandler serves static files and falls back to index.html for SPA routing
func staticFileHandler(staticFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))

	return func(c *gin.Context) {
		// Skip API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}

		// Skip health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Try to serve the file
		path := c.Request.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		if _, err := fs.Stat(staticFS, path[1:]); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// Fall back to index.html for SPA routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
