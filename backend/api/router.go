package api

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

// RegisterRoutes sets up all API routes on the Gin engine.
func RegisterRoutes(r *gin.Engine) {
	// Serve embedded frontend static files
	staticSub, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/app/music-dl/assets", http.FS(staticSub))

	// API routes
	api := r.Group("/app/music-dl/api")
	{
		api.GET("/search", handleSearch)
		api.GET("/parse", handleParse)
		api.POST("/download", handleDownload)
		api.GET("/stream", handleStream)

		api.GET("/playlists", handleGetPlaylists)
		api.POST("/playlists", handleCreatePlaylist)
		api.GET("/playlists/:id", handleGetPlaylistSongs)
		api.POST("/playlists/:id/songs", handleAddSongToPlaylist)
		api.DELETE("/playlists/:id/songs", handleRemoveSongFromPlaylist)
		api.DELETE("/playlists/:id", handleDeletePlaylist)

		api.GET("/local/music", handleLocalMusic)
		api.POST("/local/upload", handleLocalUpload)
		api.DELETE("/local/music", handleLocalDelete)
		api.GET("/local/cover", handleLocalCover)

		api.GET("/settings", handleGetSettings)
		api.POST("/settings", handleSaveSettings)
		api.GET("/cookies", handleGetCookies)
		api.POST("/cookies", handleSaveCookies)

		api.GET("/downloads", handleGetDownloads)
		api.DELETE("/downloads", handleClearDownloads)

		api.GET("/sources", handleGetSources)
		api.GET("/recommend", handleRecommend)
	}

	// Serve the frontend SPA for all other /app/music-dl paths
	r.GET("/app/music-dl", serveFrontend)
	r.GET("/app/music-dl/*path", serveFrontend)

	// Health check (used by FNOS gateway)
	r.GET("/app/music-dl/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": "music-dl"})
	})
}

func serveFrontend(c *gin.Context) {
	// Try to serve from embedded static first
	up := strings.TrimPrefix(c.Request.URL.Path, "/app/music-dl")
	if up == "" || up == "/" {
		up = "/index.html"
	}
	data, err := staticFS.ReadFile("static" + up)
	if err != nil {
		// Fallback to index.html for SPA routing
		data, err = staticFS.ReadFile("static/index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
	}
	ext := path.Ext(up)
	mime := mimeByExt(ext)
	if mime != "" {
		c.Header("Content-Type", mime)
	}
	c.Data(http.StatusOK, mime, data)
}

func mimeByExt(ext string) string {
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff2":
		return "font/woff2"
	}
	return ""
}
