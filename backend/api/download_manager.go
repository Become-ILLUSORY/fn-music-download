package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fn-music-dl/pkg"

	"github.com/guohuiyuan/music-lib/model"
)

// DownloadTask represents a single download with progress tracking.
type DownloadTask struct {
	ID         string  `json:"id"`
	SongID     string  `json:"songId"`
	Source     string  `json:"source"`
	Name       string  `json:"name"`
	Artist     string  `json:"artist"`
	Album      string  `json:"album"`
	Cover      string  `json:"cover"`
	Duration   int     `json:"duration"`
	Progress   float64 `json:"progress"` // 0-100
	Status     string  `json:"status"`   // queued, downloading, completed, failed
	SavedPath  string  `json:"savedPath,omitempty"`
	Filename   string  `json:"filename,omitempty"`
	FileSize   int64   `json:"fileSize,omitempty"`
	Error      string  `json:"error,omitempty"`
	StartedAt  int64   `json:"startedAt,omitempty"`
	FinishedAt int64   `json:"finishedAt,omitempty"`
}

// DownloadManager manages the download queue.
type DownloadManager struct {
	mu     sync.RWMutex
	tasks  map[string]*DownloadTask
	queue  []string    // ordered task IDs
	nextID int
}

var dm = &DownloadManager{
	tasks: make(map[string]*DownloadTask),
}

func generateTaskID() string {
	dm.nextID++
	return fmt.Sprintf("dl_%d_%d", time.Now().UnixMilli(), dm.nextID)
}

// AddTask adds a song to the download queue and starts downloading.
func (m *DownloadManager) AddTask(song *model.Song, withCover, withLyrics bool) *DownloadTask {
	m.mu.Lock()
	taskID := generateTaskID()
	task := &DownloadTask{
		ID:       taskID,
		SongID:   song.ID,
		Source:   song.Source,
		Name:     song.Name,
		Artist:   song.Artist,
		Album:    song.Album,
		Cover:    song.Cover,
		Duration: song.Duration,
		Status:   "queued",
	}
	m.tasks[taskID] = task
	m.queue = append(m.queue, taskID)
	m.mu.Unlock()

	go m.runDownload(taskID, song, withCover, withLyrics)
	return task
}

func (m *DownloadManager) runDownload(taskID string, song *model.Song, withCover, withLyrics bool) {
	m.mu.Lock()
	task := m.tasks[taskID]
	task.Status = "downloading"
	task.StartedAt = time.Now().UnixMilli()
	m.mu.Unlock()

	settings := pkg.GetWebSettings()
	result, err := pkg.SaveSongToFile(song, settings.DownloadDir, withCover, withLyrics)

	m.mu.Lock()
	defer m.mu.Unlock()

	task.FinishedAt = time.Now().UnixMilli()
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		task.Progress = 0
		return
	}

	task.Status = "completed"
	task.Progress = 100
	task.SavedPath = result.SavedPath
	task.Filename = result.Filename
	if info, err := os.Stat(result.SavedPath); err == nil {
		task.FileSize = info.Size()
	}

	// Record in download history
	_ = pkg.SaveDownloadRecord(song.Name, song.Artist, song.Source, pkg.DownloadStatusSuccess, "")
}

// GetAllTasks returns all download tasks ordered by creation time.
func (m *DownloadManager) GetAllTasks() []*DownloadTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*DownloadTask, 0, len(m.queue))
	for _, id := range m.queue {
		if t, ok := m.tasks[id]; ok {
			cp := *t
			result = append(result, &cp)
		}
	}
	return result
}

// GetActiveTasks returns only queued and downloading tasks.
func (m *DownloadManager) GetActiveTasks() []*DownloadTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DownloadTask
	for _, id := range m.queue {
		if t, ok := m.tasks[id]; ok && (t.Status == "queued" || t.Status == "downloading") {
			cp := *t
			result = append(result, &cp)
		}
	}
	return result
}

// GetCompletedTasks returns only completed tasks.
func (m *DownloadManager) GetCompletedTasks() []*DownloadTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DownloadTask
	for _, id := range m.queue {
		if t, ok := m.tasks[id]; ok && (t.Status == "completed" || t.Status == "failed") {
			cp := *t
			result = append(result, &cp)
		}
	}
	return result
}

// RemoveTask removes a task from the queue by ID and optionally deletes the file.
func (m *DownloadManager) RemoveTask(taskID string, deleteFile bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return false
	}

	if deleteFile && task.SavedPath != "" {
		os.Remove(task.SavedPath)
	}

	delete(m.tasks, taskID)
	var newQueue []string
	for _, id := range m.queue {
		if id != taskID {
			newQueue = append(newQueue, id)
		}
	}
	m.queue = newQueue
	return true
}

// GetCompletedDownloads returns the list of all completed downloads with file info.
func GetCompletedDownloads() []map[string]interface{} {
	records, _, _ := pkg.GetDownloadRecordPage(1, 500)
	var result []map[string]interface{}

	settings := pkg.GetWebSettings()
	dlDir := settings.DownloadDir

	for _, r := range records {
		entry := map[string]interface{}{
			"id":        fmt.Sprintf("rec_%d", r.ID),
			"name":      r.Name,
			"artist":    r.Artist,
			"source":    r.Source,
			"status":    r.Status,
			"error":     r.Error,
			"createdAt": r.CreatedAt,
		}
		// Try to find the file
		patterns := []string{
			filepath.Join(dlDir, fmt.Sprintf("%s - %s.*", r.Artist, r.Name)),
			filepath.Join(dlDir, fmt.Sprintf("%s - %s.mp3", r.Artist, r.Name)),
			filepath.Join(dlDir, fmt.Sprintf("%s - %s.flac", r.Artist, r.Name)),
			filepath.Join(dlDir, fmt.Sprintf("*/%s - %s.*", r.Artist, r.Name)),
		}
		for _, pattern := range patterns {
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				entry["savedPath"] = matches[0]
				if info, err := os.Stat(matches[0]); err == nil {
					entry["fileSize"] = info.Size()
				}
				break
			}
		}
		result = append(result, entry)
	}
	return result
}

// ValidatePlayable checks if a song is playable by probing its download URL.
func ValidatePlayable(song *model.Song) bool {
	return pkg.ValidatePlayable(song)
}

// FindAltSource tries to find the same song from another source.
func FindAltSource(song *model.Song) []map[string]string {
	keyword := song.Name
	if song.Artist != "" {
		keyword = song.Artist + " " + song.Name
	}

	var results []map[string]string
	for _, src := range pkg.GetDefaultSourceNames() {
		if src == song.Source {
			continue
		}
		fn := pkg.GetSearchFunc(src)
		if fn == nil {
			continue
		}

		cookie := pkg.CM.Get(src)
		_ = cookie // search uses CM internally

		songs, err := fn(keyword)
		if err != nil || len(songs) == 0 {
			continue
		}

		// Find best match
		for _, candidate := range songs {
			sim := pkg.CalcSongSimilarity(song.Name, song.Artist, candidate.Name, candidate.Artist)
			if sim > 0.6 {
				results = append(results, map[string]string{
					"id":     candidate.ID,
					"source": src,
					"name":   candidate.Name,
					"artist": candidate.Artist,
				})
				break
			}
		}
		if len(results) > 0 {
			break // return first source that has matches
		}
	}
	return results
}

// AnnounceDownloadProgress is a no-op in the current implementation.
// Progress tracking is done via the DownloadManager tasks.
