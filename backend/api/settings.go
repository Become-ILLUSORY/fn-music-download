package api

import (
	"net/http"
	"strings"

	"fn-music-dl/pkg"

	"github.com/gin-gonic/gin"
)

func handleGetSettings(c *gin.Context) {
	settings := pkg.GetWebSettings()
	c.JSON(http.StatusOK, settings)
}

func handleSaveSettings(c *gin.Context) {
	var req pkg.WebSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings"})
		return
	}
	if err := pkg.SaveWebSettings(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pkg.GetWebSettings())
}

func handleGetCookies(c *gin.Context) {
	cookies := pkg.CM.GetAll()
	c.JSON(http.StatusOK, cookies)
}

func handleSaveCookies(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cookies"})
		return
	}
	pkg.CM.SetAll(req)
	pkg.CM.Save()
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleGetDownloads(c *gin.Context) {
	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		if n, err := parseInt(p); err == nil && n > 0 {
			page = n
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if n, err := parseInt(ps); err == nil && n > 0 {
			pageSize = n
		}
	}
	records, total, err := pkg.GetDownloadRecordPage(page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"records": []pkg.DownloadRecord{}, "total": 0})
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records, "total": total})
}

func handleClearDownloads(c *gin.Context) {
	_ = pkg.ClearDownloadRecords()
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func parseInt(s string) (int, error) {
	var n int
	for _, d := range strings.TrimSpace(s) {
		if d < '0' || d > '9' {
			return 0, nil
		}
		n = n*10 + int(d-'0')
	}
	return n, nil
}
