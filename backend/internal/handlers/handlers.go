package handlers

import (
  "crypto/rand"
  "encoding/hex"
  "fmt"
  "io"
  "mime/multipart"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"
  "strconv"
  "encoding/json"

  "github.com/gin-gonic/gin"
  "github.com/lib/pq"

  "github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/auth"
  "github.com/BalkanID-University/vit-2026-capstone-internship-hiring-task-mahidharreddyg/backend/internal/db"
)

type Handler struct {
  StoragePath string
}

// upload file
func (h *Handler) UploadHandler(c *gin.Context) {
    maxSizeMB, _ := strconv.ParseInt(os.Getenv("MAX_FILE_SIZE_MB"), 10, 64)
    if maxSizeMB > 0 {
        maxSize := maxSizeMB * 1024 * 1024
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
    }

    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "failed to read file"})
        return
    }
    defer file.Close()

    filename := filepath.Base(header.Filename)
    filename = strings.ReplaceAll(filename, "..", "")
    if !validateFilename(filename) {
        c.JSON(400, gin.H{"error": "invalid filename"})
        return
    }

    detected, valid := detectAndValidateMIME(file, header.Header.Get("Content-Type"))
    if !valid {
        c.JSON(400, gin.H{
            "error":    "MIME type mismatch",
            "declared": header.Header.Get("Content-Type"),
            "detected": detected,
            "status":   "rejected",
        })
        return
    }
    file.Seek(0, 0)

    hash, size, err := computeSHA256(file)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to compute hash"})
        return
    }
    file.Seek(0, 0)

    userID := c.GetInt64("user_id")
    if userID == 0 {
        c.JSON(401, gin.H{"error": "unauthenticated"})
        return
    }

    var quota int64
    err = db.Pool.QueryRow(c, "SELECT quota FROM users WHERE id=$1", userID).Scan(&quota)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to get quota"})
        return
    }
    var used int64
    err = db.Pool.QueryRow(c, "SELECT COALESCE(SUM(size),0) FROM files WHERE owner_id=$1 AND trashed=false", userID).Scan(&used)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to calculate usage"})
        return
    }
    if quota > 0 && used+size > quota {
        c.JSON(413, gin.H{"error": "quota exceeded"})
        return
    }

    blobPath := filepath.Join(h.StoragePath, hash)
    var blobID int64
    err = db.Pool.QueryRow(c, "SELECT id FROM blobs WHERE hash=$1", hash).Scan(&blobID)
    if err == nil {
        _, _ = db.Pool.Exec(c, "UPDATE blobs SET ref_count = ref_count + 1 WHERE id=$1", blobID)
    } else {
        out, err := os.Create(blobPath)
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to save file"})
            return
        }
        defer out.Close()
        _, _ = io.Copy(out, file)

        err = db.Pool.QueryRow(
            c,
            "INSERT INTO blobs (hash, size, path, ref_count, created_at) VALUES ($1,$2,$3,$4,$5) RETURNING id",
            hash, size, blobPath, 1, time.Now(),
        ).Scan(&blobID)
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to insert blob"})
            return
        }
    }

    tagStr := c.PostForm("tags")
    var tagArray []string
    if tagStr != "" {
        tagArray = strings.Split(tagStr, ",")
        for _, t := range tagArray {
            if !validateTag(t) {
                c.JSON(400, gin.H{"error": "invalid tag length"})
                return
            }
        }
    }

    folderIDStr := c.PostForm("folder_id")
    var folderID *int64
    if folderIDStr != "" {
        if fid, errConv := strconv.ParseInt(folderIDStr, 10, 64); errConv == nil {
            folderID = &fid
        }
    }

		previewAvailable := false
    if strings.HasPrefix(detected, "image/") ||
      detected == "application/pdf" ||
      strings.HasPrefix(detected, "text/") {
      previewAvailable = true
	  }

    var fileID int64
    err = db.Pool.QueryRow(
    c,
    "INSERT INTO files (blob_id, owner_id, filename, mime_type, size, created_at, tags, folder_id, preview_available) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id",
    blobID, userID, filename, detected, size, time.Now(), pq.Array(tagArray), folderID, previewAvailable,
    ).Scan(&fileID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to insert file"})
        return
    }

    var existingFileID int64
    err = db.Pool.QueryRow(
        c,
        "SELECT id FROM files WHERE owner_id=$1 AND filename=$2 AND folder_id IS NOT DISTINCT FROM $3 AND trashed=false LIMIT 1",
        userID, filename, folderID,
    ).Scan(&existingFileID)

    if err == nil && existingFileID != 0 && existingFileID != fileID {
        var maxVersion int
        _ = db.Pool.QueryRow(c, "SELECT COALESCE(MAX(version),0) FROM file_versions WHERE file_id=$1", existingFileID).Scan(&maxVersion)
        _, _ = db.Pool.Exec(c,
            "INSERT INTO file_versions (file_id, version, blob_id, created_at) VALUES ($1,$2,$3,$4)",
            existingFileID, maxVersion+1, blobID, time.Now(),
        )
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        userID, "upload_file", "file", fileID, fmt.Sprintf(`{\"filename\":\"%s\"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":     "file_uploaded",
        "file_id":   fileID,
        "filename":  filename,
        "size":      size,
        "mime":      detected,
        "tags":      tagArray,
        "folder_id": folderID,
        "user":      userID,
        "timestamp": time.Now(),
    })

    c.JSON(200, gin.H{
        "file_id":   fileID,
        "blob_id":   blobID,
        "filename":  filename,
        "size":      size,
        "hash":      hash,
        "mime":      detected,
        "tags":      tagArray,
        "folder_id": folderID,
        "status":    "uploaded",
    })
}

func (h *Handler) MultiUploadHandler(c *gin.Context) {
    maxSizeMB, _ := strconv.ParseInt(os.Getenv("MAX_FILE_SIZE_MB"), 10, 64)
    if maxSizeMB > 0 {
        maxSize := maxSizeMB * 1024 * 1024
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
    }

    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(400, gin.H{"error": "failed to parse form"})
        return
    }
    files := form.File["files"]
    userID := c.GetInt64("user_id")

    tagStr := c.PostForm("tags")
    var tagArray []string
    if tagStr != "" {
        tagArray = strings.Split(tagStr, ",")
        for _, t := range tagArray {
            if !validateTag(t) {
                c.JSON(400, gin.H{"error": "invalid tag length"})
                return
            }
        }
    }

    folderIDStr := c.PostForm("folder_id")
    var folderID *int64
    if folderIDStr != "" {
        if fid, errConv := strconv.ParseInt(folderIDStr, 10, 64); errConv == nil {
            folderID = &fid
        }
    }

    var quota int64
    err = db.Pool.QueryRow(c, "SELECT quota FROM users WHERE id=$1", userID).Scan(&quota)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to get quota"})
        return
    }

    var results []gin.H
    for _, f := range files {
        file, err := f.Open()
        if err != nil {
            results = append(results, gin.H{
                "filename": f.Filename,
                "status":   "rejected",
                "reason":   "failed to open file",
            })
            continue
        }

        func() {
            defer file.Close()

            safeName := filepath.Base(f.Filename)
            safeName = strings.ReplaceAll(safeName, "..", "")
            if !validateFilename(safeName) {
                results = append(results, gin.H{"filename": safeName, "status": "rejected", "reason": "invalid filename"})
                return
            }

            detected, valid := detectAndValidateMIME(file, f.Header.Get("Content-Type"))
            if !valid {
                results = append(results, gin.H{
                    "filename": safeName,
                    "status":   "rejected",
                    "reason":   "MIME mismatch",
                })
                return
            }
            file.Seek(0, 0)

            hash, size, err := computeSHA256(file)
            if err != nil {
                results = append(results, gin.H{
                    "filename": safeName,
                    "status":   "rejected",
                    "reason":   "failed to hash file",
                })
                return
            }
            file.Seek(0, 0)

            var usedNow int64
            _ = db.Pool.QueryRow(c, "SELECT COALESCE(SUM(size),0) FROM files WHERE owner_id=$1 AND trashed=false", userID).Scan(&usedNow)
            if quota > 0 && usedNow+size > quota {
                results = append(results, gin.H{
                    "filename": safeName,
                    "status":   "rejected",
                    "reason":   "quota exceeded",
                })
                return
            }

            blobPath := filepath.Join(h.StoragePath, hash)
            var blobID int64
            err = db.Pool.QueryRow(c, "SELECT id FROM blobs WHERE hash=$1", hash).Scan(&blobID)
            if err == nil {
                _, _ = db.Pool.Exec(c, "UPDATE blobs SET ref_count = ref_count + 1 WHERE id=$1", blobID)
            } else {
                out, err := os.Create(blobPath)
                if err != nil {
                    results = append(results, gin.H{"filename": safeName, "status": "rejected", "reason": "failed to save file"})
                    return
                }
                io.Copy(out, file)
                out.Close()

                _ = db.Pool.QueryRow(c,
                    "INSERT INTO blobs (hash, size, path, ref_count, created_at) VALUES ($1,$2,$3,$4,$5) RETURNING id",
                    hash, size, blobPath, 1, time.Now(),
                ).Scan(&blobID)
            }

            previewAvailable := false
            if strings.HasPrefix(detected, "image/") ||
                detected == "application/pdf" ||
                strings.HasPrefix(detected, "text/") {
                previewAvailable = true
            }

            var fileID int64
            _ = db.Pool.QueryRow(c,
                "INSERT INTO files (blob_id, owner_id, filename, mime_type, size, created_at, tags, folder_id, preview_available) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id",
                blobID, userID, safeName, detected, size, time.Now(), pq.Array(tagArray), folderID, previewAvailable,
            ).Scan(&fileID)

            var existingFileID int64
            _ = db.Pool.QueryRow(c,
                "SELECT id FROM files WHERE owner_id=$1 AND filename=$2 AND folder_id IS NOT DISTINCT FROM $3 AND trashed=false LIMIT 1",
                userID, safeName, folderID,
            ).Scan(&existingFileID)

            if existingFileID != 0 && existingFileID != fileID {
                var maxVersion int
                _ = db.Pool.QueryRow(c, "SELECT COALESCE(MAX(version),0) FROM file_versions WHERE file_id=$1", existingFileID).Scan(&maxVersion)
                _, _ = db.Pool.Exec(c,
                    "INSERT INTO file_versions (file_id, version, blob_id, created_at) VALUES ($1,$2,$3,$4)",
                    existingFileID, maxVersion+1, blobID, time.Now(),
                )
            }

            results = append(results, gin.H{
                "file_id":   fileID,
                "filename":  safeName,
                "hash":      hash,
                "size":      size,
                "mime":      detected,
                "tags":      tagArray,
                "folder_id": folderID,
                "status":    "uploaded",
            })

            _, _ = db.Pool.Exec(c,
                "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
                userID, "upload_file", "file", fileID, fmt.Sprintf(`{\"filename\":\"%s\"}`, safeName),
            )

            broadcastUpdate(gin.H{
                "event":     "file_uploaded",
                "file_id":   fileID,
                "filename":  safeName,
                "size":      size,
                "mime":      detected,
                "tags":      tagArray,
                "folder_id": folderID,
                "user":      userID,
                "timestamp": time.Now(),
            })
        }()
    }

    c.JSON(200, gin.H{"results": results})
}

func detectAndValidateMIME(file multipart.File, declared string) (string, bool) {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	file.Seek(0, 0)

	detected := http.DetectContentType(buf[:n])

	if detected == "application/octet-stream" {
		if isTextFile(buf[:n]) {
			detected = "text/plain; charset=utf-8"
		}
	}

	if declared == "" {
		return detected, true 
	}

	if strings.HasPrefix(detected, declared) || strings.HasPrefix(declared, detected) {
		return detected, true
	}

	declaredBase := strings.Split(declared, ";")[0]
	detectedBase := strings.Split(detected, ";")[0]
	return detected, detectedBase == declaredBase
}

func isTextFile(data []byte) bool {
	for _, b := range data {
		if b < 9 || (b > 13 && b < 32) {
			return false
		}
	}
	return true
}

func (h *Handler) ListFilesHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")

	rows, err := db.Pool.Query(
    c,
    `SELECT f.id, f.filename, f.size, f.created_at, f.download_count, f.is_public, f.preview_available, b.hash
     FROM files f
     JOIN blobs b ON f.blob_id = b.id
     WHERE f.owner_id = $1 AND f.trashed=false
     ORDER BY f.created_at DESC`,
    userID,
)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to query files"})
		return
	}
	defer rows.Close()

	var files []gin.H
	for rows.Next() {
		var (
    id            int64
    filename      string
    size          int64
    createdAt     time.Time
    downloadCount int64
    isPublic      bool
    previewAvail  bool
    hash          string
)
if err := rows.Scan(&id, &filename, &size, &createdAt, &downloadCount, &isPublic, &previewAvail, &hash); err != nil {
    c.JSON(500, gin.H{"error": "failed to scan row"})
    return
}
files = append(files, gin.H{
    "id":               id,
    "filename":         filename,
    "size":             size,
    "created_at":       createdAt,
    "download_count":   downloadCount,
    "is_public":        isPublic,
    "preview_available": previewAvail,
    "hash":             hash,
})
	}

	c.JSON(200, gin.H{"files": files})
}

func (h *Handler) DownloadHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var (
        filename string
        blobPath string
        ownerID  int64
        mimeType string
    )
    err := db.Pool.QueryRow(
        c,
        `SELECT f.filename, b.path, f.owner_id, f.mime_type
         FROM files f
         JOIN blobs b ON f.blob_id = b.id
         WHERE f.id=$1`,
        id,
    ).Scan(&filename, &blobPath, &ownerID, &mimeType)

    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    var isEditor bool
    _ = db.Pool.QueryRow(
        c,
        "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2",
        id, userID,
    ).Scan(&isEditor)

    var isPublic bool
    _ = db.Pool.QueryRow(
        c,
        "SELECT is_public FROM files WHERE id=$1",
        id,
    ).Scan(&isPublic)

    if userID != ownerID && !isEditor && !isPublic {
        c.JSON(403, gin.H{"error": "permission denied"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "UPDATE files SET download_count = download_count + 1 WHERE id=$1",
        id,
    )

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        userID, "download_file", "file", id, fmt.Sprintf(`{"filename":"%s"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":    "download",
        "file_id":  id,
        "filename": filename,
        "user_id":  userID,
        "ts":       time.Now(),
    })

    serveFileWithRange(c, blobPath, filename, mimeType, true)
}

func (h *Handler) DeleteHandler(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetInt64("user_id")

	var ownerID int64
	err := db.Pool.QueryRow(
		c,
		"SELECT owner_id FROM files WHERE id=$1",
		id,
	).Scan(&ownerID)

	if err != nil {
		c.JSON(404, gin.H{"error": "file not found"})
		return
	}


	if ownerID != userID {
		var canEdit bool
		_ = db.Pool.QueryRow(c, "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2", id, userID).Scan(&canEdit)
		if !canEdit {
			c.JSON(403, gin.H{"error": "not authorized"})
			return
		}
	}

	_, err = db.Pool.Exec(
		c,
		"UPDATE files SET trashed=true, trashed_at=$1 WHERE id=$2",
		time.Now(), id,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to move file to trash"})
		return
	}

	meta := map[string]interface{}{"file_id": id}
	metaJson, _ := json.Marshal(meta)
	_, _ = db.Pool.Exec(
		c,
		"INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
		userID, "trash_file", "file", id, string(metaJson),
	)

	broadcastUpdate(gin.H{
        "event":   "file_trashed",
        "file_id": id,
        "user":    userID,
    })

	c.JSON(200, gin.H{"message": "file moved to trash", "file_id": id})
}

func (h *Handler) StatsHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var totalUsed int64
	_ = db.Pool.QueryRow(
		c,
		"SELECT COALESCE(SUM(size),0) FROM files WHERE owner_id=$1 AND trashed=false",
		userID,
	).Scan(&totalUsed)

	var quota int64
	_ = db.Pool.QueryRow(
		c,
		"SELECT quota FROM users WHERE id=$1",
		userID,
	).Scan(&quota)

	rows, _ := db.Pool.Query(
		c,
		"SELECT COALESCE(mime_type,'unknown'), SUM(size) FROM files WHERE owner_id=$1 AND trashed=false GROUP BY mime_type",
		userID,
	)
	if rows != nil {
		defer rows.Close()
	}

	breakdown := make(map[string]int64)
	if rows != nil {
		for rows.Next() {
			var mime string
			var size int64
			_ = rows.Scan(&mime, &size)
			breakdown[mime] = size
		}
	}

	var trashSize int64
	_ = db.Pool.QueryRow(
		c,
		"SELECT COALESCE(SUM(size),0) FROM files WHERE owner_id=$1 AND trashed=true",
		userID,
	).Scan(&trashSize)

	var totalBlobSize, totalFileSize int64
	_ = db.Pool.QueryRow(c, "SELECT COALESCE(SUM(size),0) FROM blobs").Scan(&totalBlobSize)
	_ = db.Pool.QueryRow(c, "SELECT COALESCE(SUM(size),0) FROM files").Scan(&totalFileSize)

	savings := totalFileSize - totalBlobSize

	usedPercent := 0.0
	if quota > 0 {
		usedPercent = float64(totalUsed) / float64(quota) * 100
	}

	c.JSON(200, gin.H{
		"total_used":   totalUsed,
		"quota":        quota,
		"used_percent": usedPercent,
		"breakdown":    breakdown,
		"trash_size":   trashSize,
		"savings":      savings,
	})
}

func (h *Handler) SignupHandler(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	if strings.TrimSpace(body.Username) == "" || strings.TrimSpace(body.Password) == "" {
		c.JSON(400, gin.H{"error": "username and password required"})
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}

	var userID int64
	err = db.Pool.QueryRow(
		c,
		"INSERT INTO users (username, email, password_hash) VALUES ($1,$2,$3) RETURNING id",
		body.Username, body.Email, hash,
	).Scan(&userID)

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create user"})
		return
	}
	_, _ = db.Pool.Exec(c,
		"INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
		userID, "signup", "user", userID, nil,
	)

	c.JSON(200, gin.H{"message": "signup successful", "user_id": userID})
}

func (h *Handler) LoginHandler(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	var userID int64
	var hash, email string
	err := db.Pool.QueryRow(
		c,
		"SELECT id, password_hash, email FROM users WHERE username=$1",
		body.Username,
	).Scan(&userID, &hash, &email)

	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	if !auth.CheckPasswordHash(body.Password, hash) {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateJWT(userID, body.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	_, _ = db.Pool.Exec(c,
		"INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
		userID, "login", "user", userID,
	)

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":       userID,
			"username": body.Username,
			"email":    email,
		},
	})
}

func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(401, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	tokenStr := authHeader
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		tokenStr = strings.TrimSpace(authHeader[7:])
	}

	userID, _, err := auth.ParseJWT(tokenStr)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	c.Set("user_id", userID)
	c.Next()
}

func generateToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *Handler) CreateShareHandler(c *gin.Context) {
    id := c.Param("id")
    var expiresAt *time.Time

    expiry := c.Query("expiry")
    if expiry != "" {
        hrs, _ := time.ParseDuration(expiry + "h")
        t := time.Now().Add(hrs)
        expiresAt = &t
    }

    allowDownload := c.Query("download") == "true"

    var fileID int64
    err := db.Pool.QueryRow(c,
        "SELECT id FROM files WHERE id=$1 AND owner_id=$2",
        id, c.GetInt64("user_id"),
    ).Scan(&fileID)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    token, err := generateToken()
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to generate token"})
        return
    }

    _, err = db.Pool.Exec(c,
        "INSERT INTO shares (file_id, token, expires_at, allow_download) VALUES ($1,$2,$3,$4)",
        fileID, token, expiresAt, allowDownload,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to create share"})
        return
    }

    shareURL := "http://localhost:8080/s/" + token

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        c.GetInt64("user_id"), "create_share", "file", fileID,
        fmt.Sprintf(`{"share_url":"%s"}`, shareURL),
    )

    broadcastUpdate(gin.H{
        "event":          "file_shared",
        "file_id":        fileID,
        "share_url":      shareURL,
        "allow_download": allowDownload,
        "expires_at":     expiresAt,
        "timestamp":      time.Now(),
    })

    c.JSON(200, gin.H{
        "share_url":      shareURL,
        "expires_at":     expiresAt,
        "allow_download": allowDownload,
    })
}

// Download shared file
func (h *Handler) DownloadShareHandler(c *gin.Context) {
    token := c.Param("token")

    var (
        fileID        int64
        blobPath      string
        filename      string
        expiresAt     *time.Time
        allowDownload bool
        mimeType      string
    )

    err := db.Pool.QueryRow(
        c,
        `SELECT f.id, b.path, f.filename, s.expires_at, s.allow_download, f.mime_type
         FROM shares s
         JOIN files f ON s.file_id = f.id
         JOIN blobs b ON f.blob_id = b.id
         WHERE s.token=$1`,
        token,
    ).Scan(&fileID, &blobPath, &filename, &expiresAt, &allowDownload, &mimeType)

    if err != nil {
        c.JSON(404, gin.H{"error": "invalid or expired link"})
        return
    }
    if expiresAt != nil && time.Now().After(*expiresAt) {
        c.JSON(410, gin.H{"error": "link expired"})
        return
    }
    if !allowDownload {
        c.JSON(403, gin.H{"error": "download not allowed"})
        return
    }

    _, _ = db.Pool.Exec(c, "UPDATE files SET download_count = download_count + 1 WHERE id=$1", fileID)
    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        0, "download_file_public", "file", fileID, fmt.Sprintf(`{"filename":"%s"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":    "download",
        "file_id":  fileID,
        "filename": filename,
        "public":   true,
        "ts":       time.Now(),
    })

    serveFileWithRange(c, blobPath, filename, mimeType, true)
}

// Preview shared file
func (h *Handler) PreviewShareHandler(c *gin.Context) {
    token := c.Param("token")

    var (
        fileID           int64
        blobPath         string
        filename         string
        mimeType         string
        expiresAt        *time.Time
        previewAvailable bool
    )

    err := db.Pool.QueryRow(
        c,
        `SELECT f.id, b.path, f.filename, f.mime_type, s.expires_at, f.preview_available
         FROM shares s
         JOIN files f ON s.file_id = f.id
         JOIN blobs b ON f.blob_id = b.id
         WHERE s.token=$1`,
        token,
    ).Scan(&fileID, &blobPath, &filename, &mimeType, &expiresAt, &previewAvailable)

    if err != nil {
        c.JSON(404, gin.H{"error": "invalid or expired link"})
        return
    }
    if expiresAt != nil && time.Now().After(*expiresAt) {
        c.JSON(410, gin.H{"error": "link expired"})
        return
    }
    if !previewAvailable {
        c.JSON(403, gin.H{"error": "preview not available"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        0, "preview_file_public", "file", fileID, fmt.Sprintf(`{"filename":"%s"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":    "preview",
        "file_id":  fileID,
        "filename": filename,
        "public":   true,
        "ts":       time.Now(),
    })

    serveFileWithRange(c, blobPath, filename, mimeType, false)
}

func (h *Handler) AddEditorHandler(c *gin.Context) {
	fileID := c.Param("id")
	userID := c.GetInt64("user_id")

	var ownerID int64
	err := db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", fileID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		c.JSON(403, gin.H{"error": "only owner can add editors"})
		return
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	var editorID int64
	err = db.Pool.QueryRow(c, "SELECT id FROM users WHERE email=$1", body.Email).Scan(&editorID)
	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	_, err = db.Pool.Exec(c, "INSERT INTO file_permissions (file_id, user_id, can_edit) VALUES ($1,$2,$3)", fileID, editorID, true)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to add editor"})
		return
	}

	_, _ = db.Pool.Exec(c,
		"INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
		userID, "add_editor", "file", fileID, fmt.Sprintf(`{"editor_id":%d}`, editorID),
	)

	c.JSON(200, gin.H{"message": "editor added", "file_id": fileID, "editor_id": editorID})
}

func (h *Handler) RemoveEditorHandler(c *gin.Context) {
	fileID := c.Param("id")
	userID := c.GetInt64("user_id")

	var ownerID int64
	err := db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", fileID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		c.JSON(403, gin.H{"error": "only owner can remove editors"})
		return
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	var editorID int64
	err = db.Pool.QueryRow(c, "SELECT id FROM users WHERE email=$1", body.Email).Scan(&editorID)
	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	_, err = db.Pool.Exec(c, "DELETE FROM file_permissions WHERE file_id=$1 AND user_id=$2", fileID, editorID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to remove editor"})
		return
	}

	_, _ = db.Pool.Exec(c,
		"INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
		userID, "remove_editor", "file", fileID, fmt.Sprintf(`{"editor_id":%d}`, editorID),
	)

	c.JSON(200, gin.H{"message": "editor removed", "file_id": fileID, "editor_id": editorID})
}

func (h *Handler) SearchFilesHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")
	name := c.Query("name")
	mime := c.Query("mime")
	from := c.Query("from")
	to := c.Query("to")
	minSize := c.Query("min_size")
	maxSize := c.Query("max_size")
	tags := c.Query("tags")
	sortBy := c.DefaultQuery("sort_by", "date")
	order := strings.ToUpper(c.DefaultQuery("order", "DESC"))
	folderIDStr := c.Query("folder_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	where := "WHERE f.owner_id=$1 AND f.trashed=false"
	args := []interface{}{userID}
	argIdx := 2

	if name != "" {
		where += " AND f.filename ILIKE $" + strconv.Itoa(argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}
	if mime != "" {
		where += " AND f.mime_type=$" + strconv.Itoa(argIdx)
		args = append(args, mime)
		argIdx++
	}
	if from != "" {
		where += " AND f.created_at >= $" + strconv.Itoa(argIdx)
		args = append(args, from)
		argIdx++
	}
	if to != "" {
		where += " AND f.created_at <= $" + strconv.Itoa(argIdx)
		args = append(args, to)
		argIdx++
	}
	if minSize != "" {
		where += " AND f.size >= $" + strconv.Itoa(argIdx)
		args = append(args, minSize)
		argIdx++
	}
	if maxSize != "" {
		where += " AND f.size <= $" + strconv.Itoa(argIdx)
		args = append(args, maxSize)
		argIdx++
	}
	if folderIDStr != "" {
		fid, errConv := strconv.ParseInt(folderIDStr, 10, 64)
		if errConv == nil {
			where += " AND f.folder_id = $" + strconv.Itoa(argIdx)
			args = append(args, fid)
			argIdx++
		}
	}
	if tags != "" {
		tagList := strings.Split(tags, ",")
		for _, t := range tagList {
			where += " AND f.tags IS NOT NULL AND EXISTS (SELECT 1 FROM unnest(f.tags) tag WHERE tag ILIKE $" + strconv.Itoa(argIdx) + ")"
			args = append(args, "%"+t+"%")
			argIdx++
		}
	}

	countQuery := "SELECT COUNT(*) FROM files f " + where
	var totalCount int
	err := db.Pool.QueryRow(c, countQuery, args...).Scan(&totalCount)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to count results"})
		return
	}

	query := `SELECT f.id, f.filename, f.size, f.created_at, b.hash, f.mime_type, f.preview_available, COALESCE(f.tags, '{}')
          FROM files f
          JOIN blobs b ON f.blob_id = b.id ` + where

	switch sortBy {
	case "name":
		query += " ORDER BY f.filename " + order
	case "size":
		query += " ORDER BY f.size " + order
	case "mime":
		query += " ORDER BY f.mime_type " + order
	default:
		query += " ORDER BY f.created_at " + order
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := db.Pool.Query(c, query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to query"})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var id, size int64
var filename, hash, mimeType string
var created time.Time
var previewAvail bool
var fileTags []string
rows.Scan(&id, &filename, &size, &created, &hash, &mimeType, &previewAvail, pq.Array(&fileTags))

		var tags interface{}
		if len(fileTags) == 0 {
			tags = nil
		} else {
			tags = fileTags
		}

		results = append(results, gin.H{
    "id":               id,
    "filename":         filename,
    "size":             size,
    "created_at":       created,
    "hash":             hash,
    "mime_type":        mimeType,
    "tags":             tags,
    "preview_available": previewAvail,
})
	}

	c.JSON(200, gin.H{
		"page":        page,
		"limit":       limit,
		"total_count": totalCount,
		"results":     results,
	})
}

func (h *Handler) UpdateTagsHandler(c *gin.Context) {
	fileID := c.Param("id")
	userID := c.GetInt64("user_id")

	var body struct {
		Tags []string `json:"tags"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	for _, t := range body.Tags {
		if len(t) > 50 {
			c.JSON(400, gin.H{"error": "tag too long"})
			return
		}
	}

	var ownerID int64
	err := db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", fileID).Scan(&ownerID)
	if err != nil {
		c.JSON(404, gin.H{"error": "file not found"})
		return
	}

	if ownerID != userID {
		var canEdit bool
		err = db.Pool.QueryRow(c, "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2", fileID, userID).Scan(&canEdit)
		if err != nil || !canEdit {
			c.JSON(403, gin.H{"error": "not authorized"})
			return
		}
	}

	_, err = db.Pool.Exec(c, "UPDATE files SET tags=$1 WHERE id=$2", pq.Array(body.Tags), fileID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update tags"})
		return
	}

	_, _ = db.Pool.Exec(c,
		"INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
		userID, "update_tags", "file", fileID, fmt.Sprintf(`{"tags": %q}`, body.Tags),
	)

	c.JSON(200, gin.H{"message": "tags updated", "file_id": fileID, "tags": body.Tags})
}

func (h *Handler) GetTagsHandler(c *gin.Context) {
	fileID := c.Param("id")
	userID := c.GetInt64("user_id")

	var ownerID int64
	var tags []string
	err := db.Pool.QueryRow(c, "SELECT owner_id, tags FROM files WHERE id=$1", fileID).Scan(&ownerID, pq.Array(&tags))
	if err != nil {
		c.JSON(404, gin.H{"error": "file not found"})
		return
	}

	if ownerID != userID {
		var canEdit bool
		err = db.Pool.QueryRow(c, "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2", fileID, userID).Scan(&canEdit)
		if err != nil {
			c.JSON(403, gin.H{"error": "not authorized"})
			return
		}
	}

	c.JSON(200, gin.H{"file_id": fileID, "tags": tags})
}

func (h *Handler) CreateFolderHandler(c *gin.Context) {
    userID := c.GetInt64("user_id")
    var body struct {
        Name     string `json:"name"`
        ParentID *int64 `json:"parent_id"`
    }
    if err := c.BindJSON(&body); err != nil || body.Name == "" {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    var folderID int64
    err := db.Pool.QueryRow(
        c,
        "INSERT INTO folders (owner_id, name, parent_id) VALUES ($1,$2,$3) RETURNING id",
        userID, body.Name, body.ParentID,
    ).Scan(&folderID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to create folder"})
        return
    }

    broadcastUpdate(gin.H{
        "event":     "folder_created",
        "folder_id": folderID,
        "name":      body.Name,
        "parent_id": body.ParentID,
        "user":      userID,
    })

    c.JSON(200, gin.H{"message": "folder created", "folder_id": folderID})
}

func (h *Handler) ListFoldersHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")

	rows, err := db.Pool.Query(c, "SELECT id, name, parent_id, created_at FROM folders WHERE owner_id=$1 ORDER BY created_at DESC", userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list folders"})
		return
	}
	defer rows.Close()

	var folders []gin.H
	for rows.Next() {
		var id int64
		var name string
		var parentID *int64
		var created time.Time
		rows.Scan(&id, &name, &parentID, &created)
		folders = append(folders, gin.H{"id": id, "name": name, "parent_id": parentID, "created": created})
	}

	c.JSON(200, gin.H{"folders": folders})
}

func (h *Handler) RenameFolderHandler(c *gin.Context) {
    folderID := c.Param("id")
    userID := c.GetInt64("user_id")

    var body struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&body); err != nil || body.Name == "" {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    res, err := db.Pool.Exec(c,
        "UPDATE folders SET name=$1 WHERE id=$2 AND owner_id=$3",
        body.Name, folderID, userID,
    )
    if err != nil || res.RowsAffected() == 0 {
        c.JSON(404, gin.H{"error": "folder not found or not owned"})
        return
    }

    broadcastUpdate(gin.H{
        "event":     "folder_renamed",
        "folder_id": folderID,
        "new_name":  body.Name,
        "user":      userID,
    })

    c.JSON(200, gin.H{"message": "folder renamed"})
}

func (h *Handler) DeleteFolderHandler(c *gin.Context) {
	folderID := c.Param("id")
	userID := c.GetInt64("user_id")

	res, err := db.Pool.Exec(c, "DELETE FROM folders WHERE id=$1 AND owner_id=$2", folderID, userID)
	if err != nil || res.RowsAffected() == 0 {
		c.JSON(404, gin.H{"error": "folder not found or not owned"})
		return
	}

	_, _ = db.Pool.Exec(c, "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)", userID, "delete_folder", "folder", folderID)

	c.JSON(200, gin.H{"message": "folder deleted"})
}

func (h *Handler) MoveFileHandler(c *gin.Context) {
	fileID := c.Param("id")
	userID := c.GetInt64("user_id")

	var body struct {
		FolderID *int64 `json:"folder_id"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	res, err := db.Pool.Exec(c, "UPDATE files SET folder_id=$1 WHERE id=$2 AND owner_id=$3", body.FolderID, fileID, userID)
	if err != nil || res.RowsAffected() == 0 {
		c.JSON(404, gin.H{"error": "file not found or not owned"})
		return
	}

	_, _ = db.Pool.Exec(c, "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)", userID, "move_file", "file", fileID, fmt.Sprintf(`{"folder_id":%v}`, body.FolderID))

	c.JSON(200, gin.H{"message": "file moved", "folder_id": body.FolderID})
}

func (h *Handler) ListFolderFilesHandler(c *gin.Context) {
    folderID := c.Param("id")
    userID := c.GetInt64("user_id")

    rows, err := db.Pool.Query(c,
        `SELECT f.id, f.filename, f.size, f.mime_type, f.created_at
         FROM files f
         WHERE f.owner_id=$1 AND f.folder_id=$2 AND f.trashed=false
         ORDER BY f.created_at DESC`,
        userID, folderID,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to query files"})
        return
    }
    defer rows.Close()

    var files []gin.H
    for rows.Next() {
        var id, size int64
        var filename, mime string
        var created time.Time
        if err := rows.Scan(&id, &filename, &size, &mime, &created); err != nil {
            c.JSON(500, gin.H{"error": "failed to scan file row"})
            return
        }
        files = append(files, gin.H{
            "id":         id,
            "filename":   filename,
            "size":       size,
            "mime":       mime,
            "created_at": created,
        })
    }

    c.JSON(200, gin.H{"files": files})
}

func (h *Handler) GetFolderTreeHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")

	rows, err := db.Pool.Query(c, "SELECT id, name, parent_id FROM folders WHERE owner_id=$1 ORDER BY created_at ASC", userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch folders"})
		return
	}
	defer rows.Close()

	type Folder struct {
		ID       int64     `json:"id"`
		Name     string    `json:"name"`
		ParentID *int64    `json:"parent_id"`
		Children []*Folder `json:"children"`
	}

	folderMap := make(map[int64]*Folder)
	var roots []*Folder

	for rows.Next() {
		var id int64
		var name string
		var parentID *int64
		rows.Scan(&id, &name, &parentID)

		f := &Folder{ID: id, Name: name, ParentID: parentID}
		folderMap[id] = f
	}

	for _, f := range folderMap {
		if f.ParentID == nil {
			roots = append(roots, f)
		} else {
			if parent, ok := folderMap[*f.ParentID]; ok {
				parent.Children = append(parent.Children, f)
			}
		}
	}

	c.JSON(200, gin.H{"tree": roots})
}

func (h *Handler) MoveFolderHandler(c *gin.Context) {
    userID := c.GetInt64("user_id")
    folderID := c.Param("id")

    var body struct {
        NewParentID *int64 `json:"new_parent_id"`
    }
    if err := c.BindJSON(&body); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    var ownerID int64
    err := db.Pool.QueryRow(
        c,
        "SELECT owner_id FROM folders WHERE id=$1",
        folderID,
    ).Scan(&ownerID)

    if err != nil {
        c.JSON(404, gin.H{"error": "folder not found"})
        return
    }
    if ownerID != userID {
        c.JSON(403, gin.H{"error": "not authorized"})
        return
    }

    if body.NewParentID != nil {
        if strconv.FormatInt(*body.NewParentID, 10) == folderID {
            c.JSON(400, gin.H{"error": "cannot move folder into itself"})
            return
        }
    }

    _, err = db.Pool.Exec(
        c,
        "UPDATE folders SET parent_id=$1 WHERE id=$2",
        body.NewParentID, folderID,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to move folder"})
        return
    }

    broadcastUpdate(gin.H{
        "event":       "folder_moved",
        "folder_id":   folderID,
        "new_parent":  body.NewParentID,
        "user":        userID,
    })

    c.JSON(200, gin.H{
        "message":       "folder moved",
        "folder_id":     folderID,
        "new_parent_id": body.NewParentID,
    })
}

func (h *Handler) TrashFolderHandler(c *gin.Context) {
    folderID := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    err := db.Pool.QueryRow(c, "SELECT owner_id FROM folders WHERE id=$1", folderID).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "folder not found"})
        return
    }
    if ownerID != userID {
        c.JSON(403, gin.H{"error": "not authorized"})
        return
    }

    _, err = db.Pool.Exec(c, "UPDATE folders SET trashed=true, trashed_at=$1 WHERE id=$2", time.Now(), folderID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to trash folder"})
        return
    }

    _, _ = db.Pool.Exec(c, "UPDATE files SET trashed=true, trashed_at=$1 WHERE folder_id=$2 AND owner_id=$3", time.Now(), folderID, userID)

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
        userID, "trash_folder", "folder", folderID,
    )

    broadcastUpdate(gin.H{
    "event":     "folder_trashed",
    "folder_id": folderID,
    "user":      userID,
    "timestamp": time.Now(),
})

    c.JSON(200, gin.H{"message": "folder moved to trash", "folder_id": folderID})
}

func (h *Handler) RestoreFolderHandler(c *gin.Context) {
    folderID := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    err := db.Pool.QueryRow(c, "SELECT owner_id FROM folders WHERE id=$1", folderID).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "folder not found"})
        return
    }
    if ownerID != userID {
        c.JSON(403, gin.H{"error": "not authorized"})
        return
    }

    _, err = db.Pool.Exec(c, "UPDATE folders SET trashed=false, trashed_at=NULL WHERE id=$1", folderID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to restore folder"})
        return
    }

    _, _ = db.Pool.Exec(c, "UPDATE files SET trashed=false, trashed_at=NULL WHERE folder_id=$1 AND owner_id=$2", folderID, userID)

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
        userID, "restore_folder", "folder", folderID,
    )

    broadcastUpdate(gin.H{
    "event":     "folder_restored",
    "folder_id": folderID,
    "user":      userID,
    "timestamp": time.Now(),
})

    c.JSON(200, gin.H{"message": "folder restored", "folder_id": folderID})
}

func (h *Handler) PermanentlyDeleteFolderHandler(c *gin.Context) {
    folderID := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    err := db.Pool.QueryRow(c, "SELECT owner_id FROM folders WHERE id=$1", folderID).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "folder not found"})
        return
    }
    if ownerID != userID {
        c.JSON(403, gin.H{"error": "not authorized"})
        return
    }

    tx, _ := db.Pool.Begin(c)

    rows, _ := tx.Query(c, `SELECT f.id, b.id, b.path 
        FROM files f 
        JOIN blobs b ON f.blob_id=b.id 
        WHERE f.folder_id=$1 AND f.owner_id=$2 AND f.trashed=true`, folderID, userID)

    for rows.Next() {
        var fid, bid int64
        var bpath string
        rows.Scan(&fid, &bid, &bpath)

        _, _ = tx.Exec(c, "DELETE FROM files WHERE id=$1", fid)
        var refCount int
        _ = tx.QueryRow(c, "UPDATE blobs SET ref_count = ref_count - 1 WHERE id=$1 RETURNING ref_count", bid).Scan(&refCount)
        if refCount <= 0 {
            _, _ = tx.Exec(c, "DELETE FROM blobs WHERE id=$1", bid)
            os.Remove(bpath)
        }
    }
    rows.Close()

    _, err = tx.Exec(c, "DELETE FROM folders WHERE id=$1 AND owner_id=$2 AND trashed=true", folderID, userID)
    if err != nil {
        _ = tx.Rollback(c)
        c.JSON(500, gin.H{"error": "failed to permanently delete folder"})
        return
    }

    _ = tx.Commit(c)

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
        userID, "permanent_delete_folder", "folder", folderID,
    )

    broadcastUpdate(gin.H{
    "event":     "folder_deleted",
    "folder_id": folderID,
    "user":      userID,
    "timestamp": time.Now(),
})

    c.JSON(200, gin.H{"message": "folder permanently deleted", "folder_id": folderID})
}

func (h *Handler) ListTrashHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")

	filesRows, err := db.Pool.Query(c,
		`SELECT f.id, f.filename, f.size, f.mime_type, f.trashed_at, b.hash
		 FROM files f
		 JOIN blobs b ON f.blob_id = b.id
		 WHERE f.owner_id=$1 AND f.trashed=true
		 ORDER BY f.trashed_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch trashed files"})
		return
	}
	defer filesRows.Close()

	var trashedFiles []gin.H
	for filesRows.Next() {
		var id, size int64
		var filename, mime, hash string
		var trashedAt time.Time
		if err := filesRows.Scan(&id, &filename, &size, &mime, &trashedAt, &hash); err != nil {
			continue
		}
		trashedFiles = append(trashedFiles, gin.H{"id": id, "filename": filename, "size": size, "mime_type": mime, "hash": hash, "trashed_at": trashedAt, "type": "file"})
	}

	folderRows, err := db.Pool.Query(c, `SELECT id, name, trashed_at FROM folders WHERE owner_id=$1 AND trashed=true ORDER BY trashed_at DESC`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch trashed folders"})
		return
	}
	defer folderRows.Close()

	var trashedFolders []gin.H
	for folderRows.Next() {
		var id int64
		var name string
		var trashedAt time.Time
		if err := folderRows.Scan(&id, &name, &trashedAt); err != nil {
			continue
		}
		trashedFolders = append(trashedFolders, gin.H{"id": id, "name": name, "trashed_at": trashedAt, "type": "folder"})
	}

	c.JSON(200, gin.H{"files": trashedFiles, "folders": trashedFolders})
}

// empty trash
func (h *Handler) EmptyTrashHandler(c *gin.Context) {
    userID := c.GetInt64("user_id")

    tx, err := db.Pool.Begin(c)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to start transaction"})
        return
    }
    defer tx.Rollback(c)

    fileRows, err := tx.Query(c,
        `SELECT f.id, b.id, b.path 
         FROM files f 
         JOIN blobs b ON f.blob_id=b.id
         WHERE f.owner_id=$1 AND f.trashed=true`, userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to fetch trashed files"})
        return
    }
    defer fileRows.Close()

    var deletedFiles []int64
    for fileRows.Next() {
        var fid, bid int64
        var bpath string
        if err := fileRows.Scan(&fid, &bid, &bpath); err != nil {
            continue
        }

        deletedFiles = append(deletedFiles, fid)

        _, _ = tx.Exec(c, "DELETE FROM files WHERE id=$1", fid)

        var refCount int
        _ = tx.QueryRow(c, "UPDATE blobs SET ref_count=ref_count-1 WHERE id=$1 RETURNING ref_count", bid).Scan(&refCount)
        if refCount <= 0 {
            _, _ = tx.Exec(c, "DELETE FROM blobs WHERE id=$1", bid)
            os.Remove(bpath)
        }
    }

    _, _ = tx.Exec(c, "DELETE FROM folders WHERE owner_id=$1 AND trashed=true", userID)

    if err := tx.Commit(c); err != nil {
        c.JSON(500, gin.H{"error": "failed to empty trash"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        userID, "empty_trash", "system", userID, fmt.Sprintf(`{"deleted_files":%v}`, deletedFiles),
    )

    broadcastUpdate(gin.H{
        "event": "trash_emptied",
        "user":  userID,
    })

    c.JSON(200, gin.H{"message": "trash emptied"})
}

func (h *Handler) AdminListFiles(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	rows, err := db.Pool.Query(c,
		`SELECT f.id, f.filename, f.size, u.username, f.created_at, f.download_count
     FROM files f
     JOIN users u ON f.owner_id=u.id
     ORDER BY f.created_at DESC`)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to query files"})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var id, size, downloadCount int64
		var filename, username string
		var created time.Time
		rows.Scan(&id, &filename, &size, &username, &created, &downloadCount)
		results = append(results, gin.H{
			"id":             id,
			"filename":       filename,
			"size":           size,
			"owner":          username,
			"created_at":     created,
			"download_count": downloadCount,
		})
	}

	c.JSON(200, gin.H{"files": results})
}

func (h *Handler) AdminStats(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	var totalStorage int64
	_ = db.Pool.QueryRow(c, "SELECT COALESCE(SUM(size),0) FROM files").Scan(&totalStorage)

	var totalUsers int64
	_ = db.Pool.QueryRow(c, "SELECT COUNT(*) FROM users").Scan(&totalUsers)

	var totalDownloads int64
	_ = db.Pool.QueryRow(c, "SELECT COALESCE(SUM(download_count),0) FROM files").Scan(&totalDownloads)

	c.JSON(200, gin.H{
		"total_storage":   totalStorage,
		"total_users":     totalUsers,
		"total_downloads": totalDownloads,
	})
}

func (h *Handler) GetAuditLogsHandler(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "unauthenticated"})
		return
	}

	rows, err := db.Pool.Query(c,
		`SELECT action, object_type, object_id, meta, created_at
     FROM audit_logs
     WHERE user_id=$1
     ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch logs"})
		return
	}
	defer rows.Close()

	var logs []gin.H
	for rows.Next() {
		var action, objectType string
		var objectID int64
		var meta *string
		var created time.Time
		rows.Scan(&action, &objectType, &objectID, &meta, &created)

		logs = append(logs, gin.H{
			"action":      action,
			"object_type": objectType,
			"object_id":   objectID,
			"meta":        meta,
			"created_at":  created,
		})
	}

	c.JSON(200, gin.H{"logs": logs})
}

// Move file to trash
func (h *Handler) TrashFileHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    err := db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", id).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    if ownerID != userID {
        var canEdit bool
        _ = db.Pool.QueryRow(c, "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2", id, userID).Scan(&canEdit)
        if !canEdit {
            c.JSON(403, gin.H{"error": "not authorized"})
            return
        }
    }

    _, err = db.Pool.Exec(c, "UPDATE files SET trashed=true, trashed_at=$1 WHERE id=$2", time.Now(), id)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to trash file"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
        userID, "trash_file", "file", id,
    )

    broadcastUpdate(gin.H{
        "event":     "file_trashed",
        "file_id":   id,
        "user":      userID,
        "timestamp": time.Now(),
    })

    c.JSON(200, gin.H{"message": "file moved to trash", "file_id": id})
}

// Restore file from trash
func (h *Handler) RestoreFileHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    err := db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", id).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    if ownerID != userID {
        c.JSON(403, gin.H{"error": "not authorized"})
        return
    }

    _, err = db.Pool.Exec(c, "UPDATE files SET trashed=false, trashed_at=NULL WHERE id=$1", id)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to restore file"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
        userID, "restore_file", "file", id,
    )

    broadcastUpdate(gin.H{
        "event":     "file_restored",
        "file_id":   id,
        "user":      userID,
        "timestamp": time.Now(),
    })

    c.JSON(200, gin.H{"message": "file restored", "file_id": id})
}

// Permanently delete file
func (h *Handler) PermanentlyDeleteFileHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    var blobID int64
    var blobPath string

    err := db.Pool.QueryRow(c, "SELECT owner_id, blob_id FROM files WHERE id=$1", id).Scan(&ownerID, &blobID)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    if ownerID != userID {
        c.JSON(403, gin.H{"error": "not authorized"})
        return
    }

    _ = db.Pool.QueryRow(c, "SELECT path FROM blobs WHERE id=$1", blobID).Scan(&blobPath)

    _, _ = db.Pool.Exec(c, "DELETE FROM files WHERE id=$1", id)

    var refCount int
    _ = db.Pool.QueryRow(c, "UPDATE blobs SET ref_count=ref_count-1 WHERE id=$1 RETURNING ref_count", blobID).Scan(&refCount)
    if refCount <= 0 {
        _, _ = db.Pool.Exec(c, "DELETE FROM blobs WHERE id=$1", blobID)
        if blobPath != "" {
            _ = os.Remove(blobPath)
        }
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id) VALUES ($1,$2,$3,$4)",
        userID, "delete_file", "file", id,
    )

    broadcastUpdate(gin.H{
        "event":     "file_deleted",
        "file_id":   id,
        "user":      userID,
        "timestamp": time.Now(),
    })

    c.JSON(200, gin.H{"message": "file permanently deleted", "file_id": id})
}

func (h *Handler) ListFileVersionsHandler(c *gin.Context) {
    fileID := c.Param("id")
    userID := c.GetInt64("user_id")

    var ownerID int64
    err := db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", fileID).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    if ownerID != userID {
        var canEdit bool
        _ = db.Pool.QueryRow(c,
            "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2",
            fileID, userID,
        ).Scan(&canEdit)
        if !canEdit {
            c.JSON(403, gin.H{"error": "not authorized"})
            return
        }
    }

    rows, err := db.Pool.Query(c,
        `SELECT v.version, v.created_at, b.hash, b.size
         FROM file_versions v
         JOIN blobs b ON v.blob_id = b.id
         WHERE v.file_id=$1
         ORDER BY v.version DESC`, fileID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to fetch versions"})
        return
    }
    defer rows.Close()

    var versions []gin.H
    for rows.Next() {
        var version int
        var created time.Time
        var hash string
        var size int64
        if err := rows.Scan(&version, &created, &hash, &size); err != nil {
            continue
        }
        versions = append(versions, gin.H{
            "version":    version,
            "created_at": created,
            "hash":       hash,
            "size":       size,
        })
    }

    c.JSON(200, gin.H{
        "file_id":  fileID,
        "versions": versions,
    })
}

func (h *Handler) RestoreFileVersionHandler(c *gin.Context) {
    fileID := c.Param("id")
    versionStr := c.Param("version")
    userID := c.GetInt64("user_id")

    version, err := strconv.Atoi(versionStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid version"})
        return
    }

    var ownerID int64
    err = db.Pool.QueryRow(c, "SELECT owner_id FROM files WHERE id=$1", fileID).Scan(&ownerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    if ownerID != userID {
        var canEdit bool
        _ = db.Pool.QueryRow(c,
            "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2",
            fileID, userID,
        ).Scan(&canEdit)
        if !canEdit {
            c.JSON(403, gin.H{"error": "not authorized"})
            return
        }
    }

    var blobID int64
    err = db.Pool.QueryRow(c,
        "SELECT blob_id FROM file_versions WHERE file_id=$1 AND version=$2",
        fileID, version,
    ).Scan(&blobID)
    if err != nil {
        c.JSON(404, gin.H{"error": "version not found"})
        return
    }

    _, err = db.Pool.Exec(c, "UPDATE files SET blob_id=$1 WHERE id=$2", blobID, fileID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to restore version"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        userID, "restore_version", "file", fileID, fmt.Sprintf(`{"version":%d}`, version),
    )

    c.JSON(200, gin.H{
        "message": "version restored",
        "file_id": fileID,
        "version": version,
    })
}

func (h *Handler) GetPreviewHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var (
        filename   string
        blobPath   string
        ownerID    int64
        mimeType   string
        isPublic   bool
    )
    err := db.Pool.QueryRow(c,
        `SELECT f.filename, b.path, f.owner_id, f.mime_type, f.is_public
         FROM files f
         JOIN blobs b ON f.blob_id = b.id
         WHERE f.id=$1`,
        id,
    ).Scan(&filename, &blobPath, &ownerID, &mimeType, &isPublic)

    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    var isEditor bool
    _ = db.Pool.QueryRow(c,
        "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2",
        id, userID,
    ).Scan(&isEditor)

    if userID != ownerID && !isEditor && !isPublic {
        c.JSON(403, gin.H{"error": "permission denied"})
        return
    }

    if strings.HasPrefix(mimeType, "image/") {
        c.File(blobPath)
        return
    }
    if mimeType == "application/pdf" {
        c.File(blobPath)
        return
    }
    if strings.HasPrefix(mimeType, "text/") {
        data, err := os.ReadFile(blobPath)
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to read file"})
            return
        }
        maxLen := 2048
        if len(data) < maxLen {
            maxLen = len(data)
        }
        c.Data(200, "text/plain; charset=utf-8", data[:maxLen])
        return
    }

    c.JSON(415, gin.H{"error": "preview not supported"})
}


func (h *Handler) PreviewFileHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var (
        filename         string
        blobPath         string
        ownerID          int64
        mimeType         string
        previewAvailable bool
    )

    err := db.Pool.QueryRow(
        c,
        `SELECT f.filename, b.path, f.owner_id, f.mime_type, f.preview_available
         FROM files f
         JOIN blobs b ON f.blob_id = b.id
         WHERE f.id=$1 AND f.trashed=false`,
        id,
    ).Scan(&filename, &blobPath, &ownerID, &mimeType, &previewAvailable)
    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    if userID != ownerID {
        var canEdit bool
        _ = db.Pool.QueryRow(c,
            "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2",
            id, userID,
        ).Scan(&canEdit)

        var isPublic bool
        _ = db.Pool.QueryRow(c,
            "SELECT is_public FROM files WHERE id=$1",
            id,
        ).Scan(&isPublic)

        if !canEdit && !isPublic {
            c.JSON(403, gin.H{"error": "permission denied"})
            return
        }
    }

    if !previewAvailable {
        c.JSON(400, gin.H{"error": "preview not available for this file"})
        return
    }


    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        userID, "preview_file", "file", id, fmt.Sprintf(`{"filename":"%s"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":    "preview",
        "file_id":  id,
        "filename": filename,
        "public":   false,
        "ts":       time.Now(),
    })

    serveFileWithRange(c, blobPath, filename, mimeType, false)
}

func (h *Handler) AccessShareHandler(c *gin.Context) {
    token := c.Param("token")

    var (
        fileID           int64
        filename         string
        expiresAt        *time.Time
        allowDownload    bool
        mimeType         string
        previewAvailable bool
    )

    err := db.Pool.QueryRow(
        c,
        `SELECT f.id, f.filename, s.expires_at, s.allow_download, f.mime_type, f.preview_available
         FROM shares s
         JOIN files f ON s.file_id = f.id
         WHERE s.token=$1`,
        token,
    ).Scan(&fileID, &filename, &expiresAt, &allowDownload, &mimeType, &previewAvailable)

    if err != nil {
        c.JSON(404, gin.H{"error": "invalid or expired link"})
        return
    }

    if expiresAt != nil && time.Now().After(*expiresAt) {
        c.JSON(410, gin.H{"error": "link expired"})
        return
    }

    baseURL := "http://localhost:8080/s/" + token

    c.JSON(200, gin.H{
        "file_id":           fileID,
        "filename":          filename,
        "mime_type":         mimeType,
        "allow_download":    allowDownload,
        "preview_available": previewAvailable,
        "download_url":      baseURL + "/download",
        "preview_url":       baseURL + "/preview",
    })
}

func serveFileWithRange(c *gin.Context, filePath, filename, mimeType string, asAttachment bool) {
    file, err := os.Open(filePath)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to open file"})
        return
    }
    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to stat file"})
        return
    }
    size := stat.Size()

    c.Header("Accept-Ranges", "bytes")
    rangeHeader := c.GetHeader("Range")

    if rangeHeader == "" {
        c.Header("Content-Length", fmt.Sprintf("%d", size))
        if mimeType != "" {
            c.Header("Content-Type", mimeType)
        }
        disposition := "inline"
        if asAttachment {
            disposition = "attachment"
        }
        c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, filename))
        http.ServeContent(c.Writer, c.Request, filename, stat.ModTime(), file)
        return
    }

    var start, end int64
    if strings.HasPrefix(rangeHeader, "bytes=") {
        parts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
        start, _ = strconv.ParseInt(parts[0], 10, 64)
        if len(parts) > 1 && parts[1] != "" {
            end, _ = strconv.ParseInt(parts[1], 10, 64)
        } else {
            end = size - 1
        }
    }

    if start < 0 || end < start || end >= size {
        c.JSON(416, gin.H{"error": "invalid range"})
        return
    }

    length := end - start + 1
    c.Status(http.StatusPartialContent)
    c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
    c.Header("Content-Length", fmt.Sprintf("%d", length))
    if mimeType != "" {
        c.Header("Content-Type", mimeType)
    }
    disposition := "inline"
    if asAttachment {
        disposition = "attachment"
    }
    c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, filename))

    file.Seek(start, io.SeekStart)
    io.CopyN(c.Writer, file, length)
}

func (h *Handler) PreviewHandler(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetInt64("user_id")

    var (
        filename         string
        blobPath         string
        ownerID          int64
        mimeType         string
        previewAvailable bool
    )
    err := db.Pool.QueryRow(
        c,
        `SELECT f.filename, b.path, f.owner_id, f.mime_type, f.preview_available
         FROM files f
         JOIN blobs b ON f.blob_id = b.id
         WHERE f.id=$1`,
        id,
    ).Scan(&filename, &blobPath, &ownerID, &mimeType, &previewAvailable)

    if err != nil {
        c.JSON(404, gin.H{"error": "file not found"})
        return
    }

    var isEditor bool
    _ = db.Pool.QueryRow(
        c,
        "SELECT can_edit FROM file_permissions WHERE file_id=$1 AND user_id=$2",
        id, userID,
    ).Scan(&isEditor)

    var isPublic bool
    _ = db.Pool.QueryRow(
        c,
        "SELECT is_public FROM files WHERE id=$1",
        id,
    ).Scan(&isPublic)

    if userID != ownerID && !isEditor && !isPublic {
        c.JSON(403, gin.H{"error": "permission denied"})
        return
    }

    if !previewAvailable {
        c.JSON(400, gin.H{"error": "preview not available for this file"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        userID, "preview_file", "file", id, fmt.Sprintf(`{"filename":"%s"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":    "preview",
        "file_id":  id,
        "filename": filename,
        "user_id":  userID,
        "ts":       time.Now(),
    })

    serveFileWithRange(c, blobPath, filename, mimeType, false)
}


func (h *Handler) AccessSharePreviewHandler(c *gin.Context) {
    token := c.Param("token")

    var (
        fileID           int64
        blobPath         string
        filename         string
        expiresAt        *time.Time
        mimeType         string
        previewAvailable bool
    )

    err := db.Pool.QueryRow(
        c,
        `SELECT f.id, b.path, f.filename, s.expires_at, f.mime_type, f.preview_available
         FROM shares s
         JOIN files f ON s.file_id = f.id
         JOIN blobs b ON f.blob_id = b.id
         WHERE s.token=$1`,
        token,
    ).Scan(&fileID, &blobPath, &filename, &expiresAt, &mimeType, &previewAvailable)

    if err != nil {
        c.JSON(404, gin.H{"error": "invalid or expired link"})
        return
    }

    if expiresAt != nil && time.Now().After(*expiresAt) {
        c.JSON(410, gin.H{"error": "link expired"})
        return
    }

    if !previewAvailable {
        c.JSON(403, gin.H{"error": "preview not available for this file"})
        return
    }

    _, _ = db.Pool.Exec(c,
        "INSERT INTO audit_logs (user_id, action, object_type, object_id, meta) VALUES ($1,$2,$3,$4,$5)",
        0, "preview_file_public", "file", fileID,
        fmt.Sprintf(`{"filename":"%s"}`, filename),
    )

    broadcastUpdate(gin.H{
        "event":     "preview",
        "file_id":   fileID,
        "filename":  filename,
        "public":    true,
        "timestamp": time.Now(),
    })

    serveFileWithRange(c, blobPath, filename, mimeType, false)
}

func (h *Handler) BulkMoveFilesHandler(c *gin.Context) {
    userID := c.GetInt64("user_id")

    var body struct {
        FileIDs   []int64 `json:"file_ids"`
        FolderID  *int64  `json:"folder_id"`
    }
    if err := c.BindJSON(&body); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if len(body.FileIDs) == 0 {
        c.JSON(400, gin.H{"error": "no files provided"})
        return
    }

    tx, err := db.Pool.Begin(c)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to start transaction"})
        return
    }
    defer tx.Rollback(c)

    for _, fid := range body.FileIDs {
        res, err := tx.Exec(c,
            "UPDATE files SET folder_id=$1 WHERE id=$2 AND owner_id=$3",
            body.FolderID, fid, userID,
        )
        if err != nil || res.RowsAffected() == 0 {
            c.JSON(404, gin.H{"error": fmt.Sprintf("file %d not found or not owned", fid)})
            return
        }
    }

    if err := tx.Commit(c); err != nil {
        c.JSON(500, gin.H{"error": "failed to commit bulk move"})
        return
    }

    broadcastUpdate(gin.H{
        "event":     "bulk_file_moved",
        "file_ids":  body.FileIDs,
        "folder_id": body.FolderID,
        "user":      userID,
        "timestamp": time.Now(),
    })

    c.JSON(200, gin.H{"message": "files moved", "file_ids": body.FileIDs, "folder_id": body.FolderID})
}