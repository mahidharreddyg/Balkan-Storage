package handlers

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "mime/multipart"
    "regexp"

    "github.com/gin-gonic/gin"
    "github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/db"
)

func requireAdmin(c *gin.Context) bool {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "unauthenticated"})
		return false
	}
	var role string
	err := db.Pool.QueryRow(c, "SELECT role FROM users WHERE id=$1", userID).Scan(&role)
	if err != nil || role != "admin" {
		c.JSON(403, gin.H{"error": "forbidden"})
		return false
	}
	return true
}
func validateFilename(name string) bool {
	if len(name) == 0 || len(name) > 255 {
		return false
	}
	return true
}

func validateTag(t string) bool {
	if len(t) == 0 || len(t) > 50 {
		return false
	}
	return true
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var usernameRegex = regexp.MustCompile(`^[A-Za-z0-9_.\-]{3,64}$`)

func validateEmail(e string) bool {
	return emailRegex.MatchString(e)
}
func validateUsername(u string) bool {
	return usernameRegex.MatchString(u)
}
func computeSHA256(file multipart.File) (string, int64, error) {
    h := sha256.New()
    size, err := io.Copy(h, file)
    if err != nil {
        return "", 0, err
    }
    return hex.EncodeToString(h.Sum(nil)), size, nil
}

func NewHandler(storagePath string) *Handler {
    return &Handler{StoragePath: storagePath}
}