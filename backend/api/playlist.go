package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Collection represents a local custom playlist.
type Collection struct {
	ID          string    `json:"id" gorm:"primaryKey;size:128"`
	Name        string    `json:"name" gorm:"size:256;not null"`
	Description string    `json:"description" gorm:"size:1024"`
	Cover       string    `json:"cover" gorm:"size:512"`
	Creator     string    `json:"creator" gorm:"size:256"`
	Source      string    `json:"source" gorm:"size:64"`
	ExternalID  string    `json:"externalId" gorm:"size:256"`
	Kind        string    `json:"kind" gorm:"size:32;default:manual"` // manual or imported
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// CollectionSong is a song entry within a collection.
type CollectionSong struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CollectionID string   `json:"collectionId" gorm:"size:128;not null;index"`
	SongID      string    `json:"songId" gorm:"size:256;not null"`
	Source      string    `json:"source" gorm:"size:64;not null"`
	Name        string    `json:"name" gorm:"size:512"`
	Artist      string    `json:"artist" gorm:"size:512"`
	Album       string    `json:"album" gorm:"size:512"`
	Cover       string    `json:"cover" gorm:"size:512"`
	Duration    int       `json:"duration"`
	Order       int       `json:"order"`
	AddedAt     time.Time `json:"addedAt" gorm:"autoCreateTime"`
}

var collectionDB *gorm.DB
var collectionInit bool

func ensureCollectionDB() error {
	if collectionInit {
		return nil
	}
	if err := ensureConfigDB(); err != nil {
		return err
	}
	collectionDB = configDB
	collectionInit = true
	return collectionDB.AutoMigrate(&Collection{}, &CollectionSong{})
}

func handleGetPlaylists(c *gin.Context) {
	if err := ensureCollectionDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	kind := strings.TrimSpace(c.Query("kind"))
	var cols []Collection
	tx := collectionDB.Order("updated_at DESC")
	if kind != "" {
		tx = tx.Where("kind = ?", kind)
	}
	if err := tx.Find(&cols).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"playlists": cols})
}

func handleCreatePlaylist(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := ensureCollectionDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	col := Collection{
		ID:          generateID(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Kind:        "manual",
	}
	if err := collectionDB.Create(&col).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, col)
}

func handleGetPlaylistSongs(c *gin.Context) {
	id := c.Param("id")
	if err := ensureCollectionDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var col Collection
	if err := collectionDB.Where("id = ?", id).First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}
	var songs []CollectionSong
	collectionDB.Where("collection_id = ?", id).Order("\"order\" ASC, added_at ASC").Find(&songs)
	c.JSON(http.StatusOK, gin.H{"playlist": col, "songs": songs})
}

func handleAddSongToPlaylist(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		SongID   string `json:"songId"`
		Source   string `json:"source"`
		Name     string `json:"name"`
		Artist   string `json:"artist"`
		Album    string `json:"album"`
		Cover    string `json:"cover"`
		Duration int    `json:"duration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := ensureCollectionDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var col Collection
	if err := collectionDB.Where("id = ?", id).First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}
	var maxOrder int
	collectionDB.Model(&CollectionSong{}).Where("collection_id = ?", id).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)

	song := CollectionSong{
		CollectionID: id,
		SongID:       req.SongID,
		Source:       req.Source,
		Name:         req.Name,
		Artist:       req.Artist,
		Album:        req.Album,
		Cover:        req.Cover,
		Duration:     req.Duration,
		Order:        maxOrder + 1,
	}
	if err := collectionDB.Create(&song).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, song)
}

func handleRemoveSongFromPlaylist(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		SongID string `json:"songId"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := ensureCollectionDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	collectionDB.Where("collection_id = ? AND song_id = ? AND source = ?", id, req.SongID, req.Source).Delete(&CollectionSong{})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleDeletePlaylist(c *gin.Context) {
	id := c.Param("id")
	if err := ensureCollectionDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	collectionDB.Where("collection_id = ?", id).Delete(&CollectionSong{})
	collectionDB.Where("id = ?", id).Delete(&Collection{})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
