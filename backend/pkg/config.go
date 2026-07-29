package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ConfigDBFile                    = "data/settings.db"
	DefaultWebDownloadDir           = "data/downloads"
	DefaultDownloadFilenameTemplate = "{artist} - {name}"
	DefaultWebPageSize              = 200
	DefaultCLIPageSize              = 20
	DefaultDownloadConcurrency      = 3
	webSettingsKey                  = "web_settings"
)

type configKV struct {
	Key       string `gorm:"primaryKey;size:128"`
	Value     string `gorm:"type:text;not null"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

// WebSettings represents the user-configurable application settings.
type WebSettings struct {
	DownloadDir              string `json:"downloadDir"`
	DownloadFilenameTemplate string `json:"downloadFilenameTemplate"`
	EmbedDownload            bool   `json:"embedDownload"`
	WebPageSize              int    `json:"webPageSize"`
	CliPageSize              int    `json:"cliPageSize"`
	DownloadConcurrency      int    `json:"downloadConcurrency"`
	AutoCacheOnPlay          bool   `json:"autoCacheOnPlay"`
}

var (
	configDB      *gorm.DB
	configInit    sync.Once
	configInitErr error
)

func configDBPath() string {
	if path := strings.TrimSpace(os.Getenv("MUSIC_DL_CONFIG_DB")); path != "" {
		return path
	}
	return ConfigDBFile
}

// ConfigDBPath returns the canonical SQLite file used by the app.
func ConfigDBPath() string {
	return configDBPath()
}

func ensureConfigDB() error {
	configInit.Do(func() {
		dbPath := filepath.Clean(configDBPath())
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			configInitErr = err
			return
		}
		db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=busy_timeout(5000)"), &gorm.Config{})
		if err != nil {
			configInitErr = err
			return
		}
		if err := db.AutoMigrate(&configKV{}, &cookieEntry{}, &DownloadRecord{}, &DownloadDedupEntry{}); err != nil {
			configInitErr = err
			return
		}
		configDB = db
	})
	return configInitErr
}

// EnsureConfigDB is an exported wrapper that reuses the existing unexported ensureConfigDB.
func EnsureConfigDB() error {
	return ensureConfigDB()
}

// GetConfigDB returns the config database instance.
func GetConfigDB() *gorm.DB {
	return configDB
}

func defaultWebSettings() WebSettings {
	return WebSettings{
		DownloadDir:              DefaultWebDownloadDir,
		DownloadFilenameTemplate: DefaultDownloadFilenameTemplate,
		EmbedDownload:            true,
		WebPageSize:              DefaultWebPageSize,
		CliPageSize:              DefaultCLIPageSize,
		DownloadConcurrency:      DefaultDownloadConcurrency,
		AutoCacheOnPlay:          true,
	}
}

func normalizeWebSettings(s WebSettings) WebSettings {
	s.DownloadDir = strings.TrimSpace(s.DownloadDir)
	if s.DownloadDir == "" {
		s.DownloadDir = DefaultWebDownloadDir
	}
	s.DownloadFilenameTemplate = strings.TrimSpace(s.DownloadFilenameTemplate)
	if s.DownloadFilenameTemplate == "" {
		s.DownloadFilenameTemplate = DefaultDownloadFilenameTemplate
	}
	if s.WebPageSize <= 0 {
		s.WebPageSize = DefaultWebPageSize
	}
	if s.CliPageSize <= 0 {
		s.CliPageSize = DefaultCLIPageSize
	}
	if s.DownloadConcurrency <= 0 {
		s.DownloadConcurrency = DefaultDownloadConcurrency
	}
	if s.DownloadConcurrency > 5 {
		s.DownloadConcurrency = 5
	}
	return s
}

// GetWebSettings reads settings from SQLite.
func GetWebSettings() WebSettings {
	s := defaultWebSettings()
	if err := ensureConfigDB(); err != nil {
		return s
	}
	var row configKV
	if err := configDB.Where("key = ?", webSettingsKey).Limit(1).Find(&row).Error; err != nil || row.Key == "" {
		return s
	}
	if err := json.Unmarshal([]byte(row.Value), &s); err != nil {
		return defaultWebSettings()
	}
	return normalizeWebSettings(s)
}

// SaveWebSettings persists settings to SQLite.
func SaveWebSettings(s WebSettings) error {
	if err := ensureConfigDB(); err != nil {
		return err
	}
	s = normalizeWebSettings(s)
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return configDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&configKV{
		Key:   webSettingsKey,
		Value: string(data),
	}).Error
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
