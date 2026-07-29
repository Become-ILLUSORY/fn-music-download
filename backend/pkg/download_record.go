package pkg

import (
	"strings"
	"time"

	"github.com/guohuiyuan/music-lib/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	DownloadStatusSuccess = "success"
	DownloadStatusSkipped = "skipped"
	DownloadStatusFailed  = "failed"
)

// DownloadRecord stores the user-visible download history.
type DownloadRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:512;not null;index"`
	Artist    string    `gorm:"size:512;not null;index"`
	Source    string    `gorm:"size:64;not null"`
	Status    string    `gorm:"size:32;not null;index"`
	Error     string    `gorm:"size:1024"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}

// DownloadDedupEntry prevents re-downloading the same song.
type DownloadDedupEntry struct {
	SongKey   string    `gorm:"primaryKey;size:1024"`
	Name      string    `gorm:"size:512;not null"`
	Artist    string    `gorm:"size:512;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func initDownloadRecordTable() error {
	if err := ensureConfigDB(); err != nil {
		return err
	}
	return configDB.AutoMigrate(&DownloadRecord{}, &DownloadDedupEntry{})
}

// SaveDownloadRecord persists a download outcome.
func SaveDownloadRecord(name, artist, source, status, errStr string) error {
	if err := initDownloadRecordTable(); err != nil {
		return err
	}
	record := DownloadRecord{
		Name:   cleanText(name),
		Artist: cleanText(artist),
		Source: cleanText(source),
		Status: cleanText(status),
		Error:  cleanText(errStr),
	}
	return configDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if record.Status != DownloadStatusSuccess {
			return nil
		}
		return saveDedupEntry(tx, record.Name, record.Artist)
	})
}

func saveDedupEntry(db *gorm.DB, name, artist string) error {
	name = cleanText(name)
	artist = cleanText(artist)
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&DownloadDedupEntry{
		SongKey: SongKeyStr(name, artist),
		Name:    name,
		Artist:  artist,
	}).Error
}

// SongKey generates a stable dedup key.
func SongKey(song *model.Song) string {
	n := cleanText(song.Name)
	a := cleanText(song.Artist)
	if a == "" {
		a = "Unknown"
	}
	if n == "" {
		n = "Unknown"
	}
	return a + " - " + n
}

// SongKeyStr generates a dedup key from name and artist strings.
func SongKeyStr(name, artist string) string {
	if artist == "" {
		artist = "Unknown"
	}
	if name == "" {
		name = "Unknown"
	}
	return artist + " - " + name
}

// LoadDownloadDedupSet loads the dedup index into a set.
func LoadDownloadDedupSet() (map[string]struct{}, error) {
	if err := initDownloadRecordTable(); err != nil {
		return nil, err
	}
	var entries []DownloadDedupEntry
	if err := configDB.Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		var records []DownloadRecord
		if err := configDB.Select("name, artist").Where("status = ?", DownloadStatusSuccess).Find(&records).Error; err != nil {
			return nil, err
		}
		if len(records) > 0 {
			for _, r := range records {
				_ = saveDedupEntry(configDB, r.Name, r.Artist)
			}
			return LoadDownloadDedupSet()
		}
	}
	set := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if k := cleanText(e.SongKey); k != "" {
			set[k] = struct{}{}
		}
	}
	return set, nil
}

// DownloadWithDedupCheck downloads a song only if not already downloaded.
func DownloadWithDedupCheck(song *model.Song, outDir string, withCover, withLyrics bool, dedupSet map[string]struct{}) (*DownloadedSong, error) {
	key := SongKey(song)
	if _, exists := dedupSet[key]; exists {
		_ = SaveDownloadRecord(song.Name, song.Artist, song.Source, DownloadStatusSkipped, "")
		return &DownloadedSong{Skipped: true, Filename: key}, nil
	}
	result, dlErr := SaveSongToFile(song, outDir, withCover, withLyrics)
	if dlErr != nil {
		_ = SaveDownloadRecord(song.Name, song.Artist, song.Source, DownloadStatusFailed, dlErr.Error())
		return result, dlErr
	}
	_ = SaveDownloadRecord(song.Name, song.Artist, song.Source, DownloadStatusSuccess, "")
	if dedupSet != nil {
		dedupSet[key] = struct{}{}
	}
	return result, nil
}

// GetDownloadRecords returns the most recent download records.
func GetDownloadRecords() ([]DownloadRecord, error) {
	r, _, err := GetDownloadRecordPage(1, 200)
	return r, err
}

// GetDownloadRecordPage returns a page of download records.
func GetDownloadRecordPage(page, pageSize int) ([]DownloadRecord, int64, error) {
	if err := initDownloadRecordTable(); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	var total int64
	if err := configDB.Model(&DownloadRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []DownloadRecord
	err := configDB.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error
	return records, total, err
}

// ClearDownloadRecords clears only the visible history.
func ClearDownloadRecords() error {
	if err := initDownloadRecordTable(); err != nil {
		return err
	}
	return configDB.Where("1 = 1").Delete(&DownloadRecord{}).Error
}

func cleanText(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		if r >= 0x20 && r != 0x7f {
			b.WriteRune(r)
		}
	}
	return b.String()
}
