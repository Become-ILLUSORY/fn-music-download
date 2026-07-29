package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fn-music-dl/pkg"

	"github.com/gin-gonic/gin"
)

type localSong struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
	Size     int64  `json:"size"`
	Ext      string `json:"ext"`
	Cover    string `json:"cover"`
	Modified string `json:"modified"`
	Source   string `json:"source"`
}

var supportedAudioExts = map[string]bool{
	".mp3": true, ".flac": true, ".m4a": true, ".ogg": true,
	".wav": true, ".wma": true, ".aac": true,
}

func handleLocalMusic(c *gin.Context) {
	settings := pkg.GetWebSettings()
	dir := settings.DownloadDir
	if dir == "" {
		dir = pkg.DefaultWebDownloadDir
	}

	songs := scanLocalDir(dir)
	c.JSON(http.StatusOK, gin.H{
		"songs": songs,
		"total": len(songs),
		"dir":   dir,
	})
}

func handleLocalUpload(c *gin.Context) {
	settings := pkg.GetWebSettings()
	dir := settings.DownloadDir
	if dir == "" {
		dir = pkg.DefaultWebDownloadDir
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !supportedAudioExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported audio format: " + ext})
		return
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	outPath := filepath.Join(dir, header.Filename)
	// Handle name collision
	if _, err := os.Stat(outPath); err == nil {
		base := strings.TrimSuffix(header.Filename, ext)
		for i := 1; ; i++ {
			outPath = filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
			if _, err := os.Stat(outPath); os.IsNotExist(err) {
				break
			}
		}
	}

	out, err := os.Create(outPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"path":    outPath,
	})
}

func handleLocalDelete(c *gin.Context) {
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path"})
		return
	}

	settings := pkg.GetWebSettings()
	dlDir, _ := filepath.Abs(settings.DownloadDir)
	absPath, _ := filepath.Abs(path)

	// Safety check: file must be within download directory
	if !strings.HasPrefix(absPath, dlDir) {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete file outside download directory"})
		return
	}

	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"status": "already_deleted"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func handleLocalCover(c *gin.Context) {
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Check for image files with the same base name
	base := strings.TrimSuffix(path, filepath.Ext(path))
	for _, imgExt := range []string{".jpg", ".jpeg", ".png", ".webp", ".bmp", ".gif"} {
		imgPath := base + imgExt
		if _, err := os.Stat(imgPath); err == nil {
			c.File(imgPath)
			return
		}
	}
	c.Status(http.StatusNotFound)
}

func scanLocalDir(dir string) []localSong {
	var songs []localSong

	entries, err := os.ReadDir(dir)
	if err != nil {
		return songs
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Recurse one level for artist/album organization
			subSongs := scanLocalDir(filepath.Join(dir, entry.Name()))
			songs = append(songs, subSongs...)
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !supportedAudioExts[ext] {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ext)
		artist := "Unknown"
		album := ""

		// Try to extract from path: Artist/Album/Song.ext or just Song.ext
		relPath, _ := filepath.Rel(dir, filepath.Join(dir, entry.Name()))
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) >= 2 {
			artist = parts[0]
		}
		if len(parts) >= 3 {
			album = parts[1]
		}

		songs = append(songs, localSong{
			Path:     filepath.Join(dir, entry.Name()),
			Name:     name,
			Artist:   artist,
			Album:    album,
			Size:     info.Size(),
			Ext:      strings.TrimPrefix(ext, "."),
			Modified: info.ModTime().Format(time.RFC3339),
			Source:   "local",
		})
	}

	return songs
}
