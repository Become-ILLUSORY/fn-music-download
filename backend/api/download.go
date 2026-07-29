package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fn-music-dl/pkg"

	"github.com/gin-gonic/gin"
	"github.com/guohuiyuan/music-lib/model"
)

func handleDownload(c *gin.Context) {
	var req struct {
		ID         string   `json:"id"`
		Source     string   `json:"source"`
		Name       string   `json:"name"`
		Artist     string   `json:"artist"`
		Album      string   `json:"album"`
		Cover      string   `json:"cover"`
		Duration   int      `json:"duration"`
		WithCover  bool     `json:"withCover"`
		WithLyrics bool     `json:"withLyrics"`
		SaveLocal  bool     `json:"saveLocal"`
		IDs        []string `json:"ids"` // batch download
		Sources    []string `json:"sources"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	settings := pkg.GetWebSettings()

	// Single song download
	if req.ID != "" && req.Source != "" {
		song := &model.Song{
			ID:       req.ID,
			Source:   req.Source,
			Name:     req.Name,
			Artist:   req.Artist,
			Album:    req.Album,
			Cover:    req.Cover,
			Duration: req.Duration,
		}

		withCover := req.WithCover
		withLyrics := req.WithLyrics

		if req.SaveLocal {
			result, err := pkg.SaveSongToFile(song, settings.DownloadDir, withCover, withLyrics)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"filename": result.Filename,
				"saved":    result.SavedPath,
				"skipped":  result.Skipped,
				"warning":  result.Warning,
			})
			return
		}

		// Send as download response
		result, err := pkg.DownloadSongData(song, withCover, withLyrics)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if result.Warning != "" {
			c.Header("X-Download-Warning", result.Warning)
		}
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
		c.Data(http.StatusOK, result.ContentType, result.Data)
		return
	}

	// Batch download
	if len(req.IDs) > 0 && len(req.Sources) > 0 {
		if len(req.IDs) != len(req.Sources) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ids and sources length mismatch"})
			return
		}
		dedupSet, _ := pkg.LoadDownloadDedupSet()
		var results []gin.H
		for i := range req.IDs {
			song := &model.Song{
				ID:       req.IDs[i],
				Source:   req.Sources[i],
				Name:     req.Name,
				Artist:   req.Artist,
			}
			result, err := pkg.DownloadWithDedupCheck(song, settings.DownloadDir, req.WithCover, req.WithLyrics, dedupSet)
			if err != nil {
				results = append(results, gin.H{
					"id":     req.IDs[i],
					"source": req.Sources[i],
					"error":  err.Error(),
				})
			} else {
				results = append(results, gin.H{
					"id":       req.IDs[i],
					"source":   req.Sources[i],
					"filename": result.Filename,
					"saved":    result.SavedPath,
					"skipped":  result.Skipped,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{"results": results})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "missing id/source or ids/sources"})
}

func handleStream(c *gin.Context) {
	// Stream audio for preview
	id := strings.TrimSpace(c.Query("id"))
	source := strings.TrimSpace(c.Query("source"))

	if id == "" || source == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	song := &model.Song{ID: id, Source: source}

	// Handle soda decryption
	if source == "soda" {
		data, err := pkg.FetchDecryptedSodaAudio(song)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ext := pkg.DetectAudioExt(data)
		ct := pkg.AudioMimeByExt(ext)
		c.Header("Content-Type", ct)
		c.Header("Content-Length", strconv.Itoa(len(data)))
		http.ServeContent(c.Writer, c.Request, "audio."+ext, songMetaModTime(), bytes.NewReader(data))
		return
	}

	dlFunc := pkg.GetDownloadFunc(source)
	if dlFunc == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	urlStr, err := dlFunc(song)
	if err != nil || urlStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get download URL"})
		return
	}

	// Use Range fetch for streaming
	if fetch, handled, err := pkg.NewSourceRangeFetch(urlStr, source, ""); handled && err == nil {
		ext := fetch.Ext
		if ext == "" {
			ext = "mp3"
		}
		ct := pkg.AudioMimeByExt(ext)
		c.Header("Content-Type", ct)
		c.Header("Content-Length", strconv.FormatInt(fetch.ContentLength, 10))
		c.Header("Accept-Ranges", "bytes")
		_ = fetch.WriteTo(c.Writer)
		return
	}

	// Fallback: proxy the stream
	req, err := pkg.BuildSourceRequest("GET", urlStr, source, "")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "audio/mpeg"
	}
	c.Header("Content-Type", ct)
	io.Copy(c.Writer, resp.Body)
}

func songMetaModTime() time.Time {
	return time.Time{}
}
