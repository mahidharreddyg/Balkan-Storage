package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/auth"
	"github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/db"
	"github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/handlers"
	"github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/middleware"
)

func main() {
	_ = godotenv.Load(".env")

	if err := db.InitDB(); err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}
	defer db.CloseDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	maxSizeMB, _ := strconv.ParseInt(os.Getenv("MAX_FILE_SIZE_MB"), 10, 64)
	if maxSizeMB <= 0 {
		maxSizeMB = 10
	}
	maxRequestSize := maxSizeMB * 1024 * 1024
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestSize)
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	h := handlers.NewHandler(os.Getenv("STORAGE_PATH"))

	r.POST("/signup", h.SignupHandler)
	r.POST("/login", h.LoginHandler)
	r.GET("/ws/stats", h.StatsWS)

	r.GET("/verify-token", func(c *gin.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
        return
    }

    userID, username, err := auth.ParseJWT(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "user_id":  userID,
        "username": username,
        "status":   "valid",
    })
})

	authGroup := r.Group("/")
	authGroup.Use(handlers.AuthMiddleware, middleware.RateLimiter())
	{
		authGroup.POST("/upload", h.UploadHandler)
		authGroup.POST("/multi-upload", h.MultiUploadHandler)

		authGroup.GET("/files", h.ListFilesHandler)
		authGroup.GET("/files/search", h.SearchFilesHandler)
		authGroup.GET("/files/:id/download", h.DownloadHandler)
		authGroup.GET("/files/:id/preview", h.PreviewFileHandler)
		authGroup.PATCH("/files/:id/move", h.MoveFileHandler)
		authGroup.PATCH("/files/move-bulk", h.BulkMoveFilesHandler)

		authGroup.PATCH("/files/:id/tags", h.UpdateTagsHandler)
		authGroup.GET("/files/:id/tags", h.GetTagsHandler)

		authGroup.POST("/files/:id/add-editor", h.AddEditorHandler)
		authGroup.DELETE("/files/:id/remove-editor", h.RemoveEditorHandler)

		authGroup.GET("/files/:id/versions", h.ListFileVersionsHandler)
		authGroup.POST("/files/:id/restore-version/:version", h.RestoreFileVersionHandler)

		authGroup.PATCH("/files/:id/trash", h.TrashFileHandler)
		authGroup.PATCH("/files/:id/restore", h.RestoreFileHandler)
		authGroup.DELETE("/trash/:id", h.PermanentlyDeleteFileHandler)

		authGroup.POST("/folders", h.CreateFolderHandler)
		authGroup.GET("/folders", h.ListFoldersHandler)
		authGroup.PATCH("/folders/:id", h.RenameFolderHandler)
		authGroup.GET("/folders/:id/files", h.ListFolderFilesHandler)
		authGroup.GET("/folders/tree", h.GetFolderTreeHandler)
		authGroup.PATCH("/folders/:id/move", h.MoveFolderHandler)
		authGroup.PATCH("/folders/:id/trash", h.TrashFolderHandler)
		authGroup.PATCH("/folders/:id/restore", h.RestoreFolderHandler)
		authGroup.DELETE("/trash/folders/:id", h.PermanentlyDeleteFolderHandler)

		authGroup.GET("/trash", h.ListTrashHandler)
		authGroup.DELETE("/trash/empty", h.EmptyTrashHandler)

		authGroup.GET("/stats", h.StatsHandler)

		authGroup.GET("/admin/files", h.AdminListFiles)
		authGroup.GET("/admin/stats", h.AdminStats)

		authGroup.GET("/audit-logs", h.GetAuditLogsHandler)

		authGroup.POST("/share/:id", h.CreateShareHandler)
	}

	r.GET("/s/:token", h.AccessShareHandler)
	r.GET("/s/:token/download", h.DownloadShareHandler)
	r.GET("/s/:token/preview", h.PreviewShareHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s", port)
	r.Run(":" + port)
}