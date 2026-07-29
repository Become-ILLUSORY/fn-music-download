package api

import (
	"net/http"

	"fn-music-dl/pkg"

	"github.com/gin-gonic/gin"
	"github.com/guohuiyuan/music-lib/model"
)

func handleEnqueueDownload(c *gin.Context) {
	var req struct {
		ID         string `json:"id"`
		Source     string `json:"source"`
		Name       string `json:"name"`
		Artist     string `json:"artist"`
		Album      string `json:"album"`
		Cover      string `json:"cover"`
		Duration   int    `json:"duration"`
		WithCover  bool   `json:"withCover"`
		WithLyrics bool   `json:"withLyrics"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == "" || req.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	song := &model.Song{
		ID:       req.ID,
		Source:   req.Source,
		Name:     req.Name,
		Artist:   req.Artist,
		Album:    req.Album,
		Cover:    req.Cover,
		Duration: req.Duration,
	}

	task := dm.AddTask(song, req.WithCover, req.WithLyrics)
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func handleGetDownloadTasks(c *gin.Context) {
	all := dm.GetAllTasks()
	c.JSON(http.StatusOK, gin.H{
		"active":    dm.GetActiveTasks(),
		"completed": dm.GetCompletedTasks(),
		"all":       all,
	})
}

func handleDeleteDownloadTask(c *gin.Context) {
	taskID := c.Query("id")
	deleteFile := c.Query("deleteFile") == "true"
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}
	ok := dm.RemoveTask(taskID, deleteFile)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func handleGetCompletedDownloads(c *gin.Context) {
	completions := GetCompletedDownloads()
	c.JSON(http.StatusOK, gin.H{"records": completions})
}

func handleBatchValidate(c *gin.Context) {
	var req struct {
		Songs []struct {
			ID     string `json:"id"`
			Source string `json:"source"`
		} `json:"songs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	type result struct {
		ID       string `json:"id"`
		Source   string `json:"source"`
		Playable bool   `json:"playable"`
	}
	var results []result
	for _, s := range req.Songs {
		playable := pkg.ValidatePlayable(&model.Song{ID: s.ID, Source: s.Source})
		results = append(results, result{
			ID:       s.ID,
			Source:   s.Source,
			Playable: playable,
		})
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

func handleBatchRetry(c *gin.Context) {
	var req struct {
		Songs []struct {
			ID     string `json:"id"`
			Source string `json:"source"`
			Name   string `json:"name"`
			Artist string `json:"artist"`
		} `json:"songs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	type altResult struct {
		ID       string `json:"id"`
		Source   string `json:"source"`
		AltID    string `json:"altId"`
		AltSource string `json:"altSource"`
		AltName  string `json:"altName"`
		AltArtist string `json:"altArtist"`
	}
	var results []altResult
	for _, s := range req.Songs {
		alts := FindAltSource(&model.Song{
			ID:     s.ID,
			Source: s.Source,
			Name:   s.Name,
			Artist: s.Artist,
		})
		if len(alts) > 0 {
			alt := alts[0]
			results = append(results, altResult{
				ID:        s.ID,
				Source:    s.Source,
				AltID:     alt["id"],
				AltSource: alt["source"],
				AltName:   alt["name"],
				AltArtist: alt["artist"],
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
