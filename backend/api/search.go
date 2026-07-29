package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"fn-music-dl/pkg"

	"github.com/gin-gonic/gin"
	"github.com/guohuiyuan/music-lib/model"
)

type searchResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
	Source   string `json:"source"`
	Cover    string `json:"cover"`
	Ext      string `json:"ext,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

func handleSearch(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("q"))
	searchType := strings.TrimSpace(c.Query("type"))
	sourcesParam := strings.TrimSpace(c.Query("sources"))

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter 'q'"})
		return
	}
	if searchType == "" {
		searchType = "song"
	}

	var sources []string
	if sourcesParam != "" {
		sources = strings.Split(sourcesParam, ",")
	} else {
		switch searchType {
		case "playlist":
			sources = pkg.GetPlaylistSourceNames()
		case "album":
			sources = pkg.GetAlbumSourceNames()
		default:
			sources = pkg.GetDefaultSourceNames()
		}
	}

	// Clean sources
	var cleanSources []string
	for _, s := range sources {
		s = strings.TrimSpace(s)
		if s != "" {
			cleanSources = append(cleanSources, s)
		}
	}

	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		results    []searchResult
		playlists  []model.Playlist
	)

	// Try parsing as a link first
	if source := pkg.DetectSource(keyword); source != "" {
		if pl, songs, err := tryParseLink(source, keyword); err == nil {
			if searchType == "playlist" && pl != nil {
				c.JSON(http.StatusOK, gin.H{
					"type":      "playlist_detail",
					"playlist":  pl,
					"songs":     toSearchResults(songs),
					"total":     len(songs),
				})
				return
			}
			if len(songs) > 0 {
				c.JSON(http.StatusOK, gin.H{
					"type":  "song",
					"songs": toSearchResults(songs),
					"total": len(songs),
				})
				return
			}
		}
	}

	switch searchType {
	case "playlist":
		for _, source := range cleanSources {
			if source == "local" {
				continue
			}
			source := source
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn := pkg.GetPlaylistSearchFunc(source)
				if fn == nil {
					return
				}
				pls, err := fn(keyword)
				if err != nil || len(pls) == 0 {
					return
				}
				mu.Lock()
				playlists = append(playlists, pls...)
				mu.Unlock()
			}()
		}
		wg.Wait()

		// Sort by source
		sort.Slice(playlists, func(i, j int) bool {
			return playlists[i].Source < playlists[j].Source
		})

		c.JSON(http.StatusOK, gin.H{
			"type":       "playlist",
			"playlists":  playlists,
			"total":      len(playlists),
		})

	case "album":
		for _, source := range cleanSources {
			source := source
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn := pkg.GetAlbumSearchFunc(source)
				if fn == nil {
					return
				}
				pls, err := fn(keyword)
				if err != nil || len(pls) == 0 {
					return
				}
				mu.Lock()
				playlists = append(playlists, pls...)
				mu.Unlock()
			}()
		}
		wg.Wait()

		sort.Slice(playlists, func(i, j int) bool {
			return playlists[i].Source < playlists[j].Source
		})

		c.JSON(http.StatusOK, gin.H{
			"type":      "album",
			"playlists": playlists,
			"total":     len(playlists),
		})

	default: // song
		for _, source := range cleanSources {
			source := source
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn := pkg.GetSearchFunc(source)
				if fn == nil {
					return
				}
				songs, err := fn(keyword)
				if err != nil || len(songs) == 0 {
					return
				}
				mu.Lock()
				for _, song := range songs {
					results = append(results, searchResult{
						ID:       song.ID,
						Name:     song.Name,
						Artist:   song.Artist,
						Album:    song.Album,
						Duration: song.Duration,
						Source:   source,
						Cover:    song.Cover,
						Ext:      song.Ext,
						Size:     song.Size,
					})
				}
				mu.Unlock()
			}()
		}
		wg.Wait()

		sort.Slice(results, func(i, j int) bool {
			if results[i].Source != results[j].Source {
				return results[i].Source < results[j].Source
			}
			return results[i].Name < results[j].Name
		})

		c.JSON(http.StatusOK, gin.H{
			"type":  "song",
			"songs": results,
			"total": len(results),
		})
	}
}

func handleParse(c *gin.Context) {
	link := strings.TrimSpace(c.Query("url"))
	if link == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url"})
		return
	}

	source := pkg.DetectSource(link)
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported link"})
		return
	}

	if pl, songs, err := tryParseLink(source, link); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"type":     "detail",
			"playlist": pl,
			"songs":    toSearchResults(songs),
			"total":    len(songs),
		})
		return
	}

	// Try parsing as a single song
	fn := pkg.GetParseFunc(source)
	if fn == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parse not supported for this source"})
		return
	}
	song, err := fn(link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if song == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "song not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"type": "song",
		"songs": toSearchResults([]model.Song{*song}),
		"total": 1,
	})
}

func handleRecommend(c *gin.Context) {
	source := strings.TrimSpace(c.Query("source"))
	if source == "" {
		source = "netease"
	}

	fn := pkg.GetRecommendFunc(source)
	if fn == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recommendations not supported for this source"})
		return
	}

	playlists, err := fn()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlists": playlists,
		"total":     len(playlists),
	})
}

func handleGetSources(c *gin.Context) {
	all := pkg.GetAllSourceNames()
	desc := make(map[string]string)
	for _, s := range all {
		desc[s] = pkg.GetSourceDescription(s)
	}
	c.JSON(http.StatusOK, gin.H{
		"sources":      all,
		"defaults":     pkg.GetDefaultSourceNames(),
		"descriptions": desc,
	})
}

func tryParseLink(source, link string) (*model.Playlist, []model.Song, error) {
	// Try playlist
	fn := pkg.GetParsePlaylistFunc(source)
	if fn != nil {
		if pl, songs, err := fn(link); err == nil && pl != nil {
			return pl, songs, nil
		}
	}
	// Try album
	afn := pkg.GetParseAlbumFunc(source)
	if afn != nil {
		if pl, songs, err := afn(link); err == nil && pl != nil {
			return pl, songs, nil
		}
	}
	return nil, nil, fmt.Errorf("no matching parser")
}

func toSearchResults(songs []model.Song) []searchResult {
	res := make([]searchResult, 0, len(songs))
	for _, s := range songs {
		res = append(res, searchResult{
			ID:       s.ID,
			Name:     s.Name,
			Artist:   s.Artist,
			Album:    s.Album,
			Duration: s.Duration,
			Source:   s.Source,
			Cover:    s.Cover,
			Ext:      s.Ext,
			Size:     s.Size,
		})
	}
	return res
}
