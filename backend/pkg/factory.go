package pkg

import (
	"strings"

	"github.com/guohuiyuan/music-lib/apple"
	"github.com/guohuiyuan/music-lib/bilibili"
	"github.com/guohuiyuan/music-lib/fivesing"
	"github.com/guohuiyuan/music-lib/jamendo"
	"github.com/guohuiyuan/music-lib/joox"
	"github.com/guohuiyuan/music-lib/kugou"
	"github.com/guohuiyuan/music-lib/kuwo"
	"github.com/guohuiyuan/music-lib/migu"
	"github.com/guohuiyuan/music-lib/model"
	"github.com/guohuiyuan/music-lib/netease"
	"github.com/guohuiyuan/music-lib/qianqian"
	"github.com/guohuiyuan/music-lib/qq"
	"github.com/guohuiyuan/music-lib/soda"
)

// Factory function types
type (
	SearchFunc          func(keyword string) ([]model.Song, error)
	SearchPlaylistFunc  func(keyword string) ([]model.Playlist, error)
	PlaylistDetailFunc  func(id string) ([]model.Song, error)
	DownloadFunc        func(song *model.Song) (string, error)
	LyricFunc           func(song *model.Song) (string, error)
	ParseFunc           func(link string) (*model.Song, error)
	ParsePlaylistFunc   func(link string) (*model.Playlist, []model.Song, error)
	ParseAlbumFunc      func(link string) (*model.Playlist, []model.Song, error)
	RecommendFunc       func() ([]model.Playlist, error)
)

// GetSearchFunc returns the search function for the given source.
func GetSearchFunc(source string) SearchFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).Search
	case "qq":
		return qq.New(c).Search
	case "kugou":
		return kugou.New(c).Search
	case "kuwo":
		return kuwo.New(c).Search
	case "migu":
		return migu.New(c).Search
	case "bilibili":
		return bilibili.New(c).Search
	case "fivesing":
		return fivesing.New(c).Search
	case "jamendo":
		return jamendo.New(c).Search
	case "joox":
		return joox.New(c).Search
	case "qianqian":
		return qianqian.New(c).Search
	case "soda":
		return soda.New(c).Search
	case "apple":
		return apple.New(c).Search
	default:
		return nil
	}
}

// GetPlaylistSearchFunc returns the playlist search function.
func GetPlaylistSearchFunc(source string) SearchPlaylistFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).SearchPlaylist
	case "qq":
		return qq.New(c).SearchPlaylist
	case "kugou":
		return kugou.New(c).SearchPlaylist
	case "kuwo":
		return kuwo.New(c).SearchPlaylist
	case "migu":
		return migu.New(c).SearchPlaylist
	case "jamendo":
		return jamendo.New(c).SearchPlaylist
	case "joox":
		return joox.New(c).SearchPlaylist
	case "qianqian":
		return qianqian.New(c).SearchPlaylist
	case "bilibili":
		return bilibili.New(c).SearchPlaylist
	case "soda":
		return soda.New(c).SearchPlaylist
	case "fivesing":
		return fivesing.New(c).SearchPlaylist
	case "apple":
		return apple.New(c).SearchPlaylist
	default:
		return nil
	}
}

// GetAlbumSearchFunc returns the album search function.
func GetAlbumSearchFunc(source string) SearchPlaylistFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).SearchAlbum
	case "qq":
		return qq.New(c).SearchAlbum
	case "kugou":
		return kugou.New(c).SearchAlbum
	case "kuwo":
		return kuwo.New(c).SearchAlbum
	case "migu":
		return migu.New(c).SearchAlbum
	case "jamendo":
		return jamendo.New(c).SearchAlbum
	case "joox":
		return joox.New(c).SearchAlbum
	case "qianqian":
		return qianqian.New(c).SearchAlbum
	case "soda":
		return soda.New(c).SearchAlbum
	case "apple":
		return apple.New(c).SearchAlbum
	default:
		return nil
	}
}

// GetPlaylistDetailFunc returns the playlist detail function.
func GetPlaylistDetailFunc(source string) PlaylistDetailFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).GetPlaylistSongs
	case "qq":
		return qq.New(c).GetPlaylistSongs
	case "kugou":
		return kugou.New(c).GetPlaylistSongs
	case "kuwo":
		return kuwo.New(c).GetPlaylistSongs
	case "migu":
		return migu.New(c).GetPlaylistSongs
	case "jamendo":
		return jamendo.New(c).GetPlaylistSongs
	case "joox":
		return joox.New(c).GetPlaylistSongs
	case "qianqian":
		return qianqian.New(c).GetPlaylistSongs
	case "bilibili":
		return bilibili.New(c).GetPlaylistSongs
	case "soda":
		return soda.New(c).GetPlaylistSongs
	case "fivesing":
		return fivesing.New(c).GetPlaylistSongs
	case "apple":
		return apple.New(c).GetPlaylistSongs
	default:
		return nil
	}
}

// GetAlbumDetailFunc returns the album detail function.
func GetAlbumDetailFunc(source string) PlaylistDetailFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).GetAlbumSongs
	case "qq":
		return qq.New(c).GetAlbumSongs
	case "kugou":
		return kugou.New(c).GetAlbumSongs
	case "kuwo":
		return kuwo.New(c).GetAlbumSongs
	case "migu":
		return migu.New(c).GetAlbumSongs
	case "jamendo":
		return jamendo.New(c).GetAlbumSongs
	case "joox":
		return joox.New(c).GetAlbumSongs
	case "qianqian":
		return qianqian.New(c).GetAlbumSongs
	case "soda":
		return soda.New(c).GetAlbumSongs
	case "apple":
		return apple.New(c).GetAlbumSongs
	default:
		return nil
	}
}

// GetDownloadFunc returns the download URL function.
func GetDownloadFunc(source string) DownloadFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).GetDownloadURL
	case "qq":
		return qq.New(c).GetDownloadURL
	case "kugou":
		return kugou.New(c).GetDownloadURL
	case "kuwo":
		return kuwo.New(c).GetDownloadURL
	case "migu":
		return migu.New(c).GetDownloadURL
	case "soda":
		return soda.New(c).GetDownloadURL
	case "bilibili":
		return bilibili.New(c).GetDownloadURL
	case "fivesing":
		return fivesing.New(c).GetDownloadURL
	case "jamendo":
		return jamendo.New(c).GetDownloadURL
	case "joox":
		return joox.New(c).GetDownloadURL
	case "qianqian":
		return qianqian.New(c).GetDownloadURL
	case "apple":
		return apple.New(c).GetDownloadURL
	default:
		return nil
	}
}

// GetLyricFunc returns the lyrics function.
func GetLyricFunc(source string) LyricFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).GetLyrics
	case "qq":
		return qq.New(c).GetLyrics
	case "kugou":
		return kugou.New(c).GetLyrics
	case "kuwo":
		return kuwo.New(c).GetLyrics
	case "migu":
		return migu.New(c).GetLyrics
	case "soda":
		return soda.New(c).GetLyrics
	case "bilibili":
		return bilibili.New(c).GetLyrics
	case "fivesing":
		return fivesing.New(c).GetLyrics
	case "jamendo":
		return jamendo.New(c).GetLyrics
	case "joox":
		return joox.New(c).GetLyrics
	case "qianqian":
		return qianqian.New(c).GetLyrics
	case "apple":
		return apple.New(c).GetLyrics
	default:
		return nil
	}
}

// GetParseFunc returns the song link parser function.
func GetParseFunc(source string) ParseFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).Parse
	case "qq":
		return qq.New(c).Parse
	case "kugou":
		return kugou.New(c).Parse
	case "kuwo":
		return kuwo.New(c).Parse
	case "migu":
		return migu.New(c).Parse
	case "soda":
		return soda.New(c).Parse
	case "bilibili":
		return bilibili.New(c).Parse
	case "fivesing":
		return fivesing.New(c).Parse
	case "jamendo":
		return jamendo.New(c).Parse
	case "joox":
		return joox.New(c).Parse
	case "qianqian":
		return qianqian.New(c).Parse
	case "apple":
		return apple.New(c).Parse
	default:
		return nil
	}
}

// GetParsePlaylistFunc returns the playlist link parser function.
func GetParsePlaylistFunc(source string) ParsePlaylistFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).ParsePlaylist
	case "qq":
		return qq.New(c).ParsePlaylist
	case "kugou":
		return kugou.New(c).ParsePlaylist
	case "kuwo":
		return kuwo.New(c).ParsePlaylist
	case "migu":
		return migu.New(c).ParsePlaylist
	case "jamendo":
		return jamendo.New(c).ParsePlaylist
	case "joox":
		return joox.New(c).ParsePlaylist
	case "qianqian":
		return qianqian.New(c).ParsePlaylist
	case "bilibili":
		return bilibili.New(c).ParsePlaylist
	case "soda":
		return soda.New(c).ParsePlaylist
	case "fivesing":
		return fivesing.New(c).ParsePlaylist
	case "apple":
		return apple.New(c).ParsePlaylist
	default:
		return nil
	}
}

// GetParseAlbumFunc returns the album link parser function.
func GetParseAlbumFunc(source string) ParseAlbumFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).ParseAlbum
	case "qq":
		return qq.New(c).ParseAlbum
	case "kugou":
		return kugou.New(c).ParseAlbum
	case "kuwo":
		return kuwo.New(c).ParseAlbum
	case "migu":
		return migu.New(c).ParseAlbum
	case "jamendo":
		return jamendo.New(c).ParseAlbum
	case "joox":
		return joox.New(c).ParseAlbum
	case "qianqian":
		return qianqian.New(c).ParseAlbum
	case "soda":
		return soda.New(c).ParseAlbum
	case "apple":
		return apple.New(c).ParseAlbum
	default:
		return nil
	}
}

// GetRecommendFunc returns the recommended playlists function.
func GetRecommendFunc(source string) RecommendFunc {
	c := CM.Get(source)
	switch source {
	case "netease":
		return netease.New(c).GetRecommendedPlaylists
	case "qq":
		return qq.New(c).GetRecommendedPlaylists
	case "kugou":
		return kugou.New(c).GetRecommendedPlaylists
	case "kuwo":
		return kuwo.New(c).GetRecommendedPlaylists
	default:
		return nil
	}
}

// DetectSource identifies the music source from a URL.
func DetectSource(link string) string {
	if strings.Contains(link, "163.com") {
		return "netease"
	}
	if strings.Contains(link, "qq.com") {
		return "qq"
	}
	if strings.Contains(link, "5sing") {
		return "fivesing"
	}
	if strings.Contains(link, "kugou.com") {
		return "kugou"
	}
	if strings.Contains(link, "kuwo.cn") {
		return "kuwo"
	}
	if strings.Contains(link, "migu.cn") {
		return "migu"
	}
	if strings.Contains(link, "joox.com") {
		return "joox"
	}
	if strings.Contains(link, "bilibili.com") || strings.Contains(link, "b23.tv") {
		return "bilibili"
	}
	if strings.Contains(link, "douyin.com") || strings.Contains(link, "qishui") {
		return "soda"
	}
	if strings.Contains(link, "91q.com") {
		return "qianqian"
	}
	if strings.Contains(link, "jamendo.com") {
		return "jamendo"
	}
	if strings.Contains(link, "music.apple.com") || strings.Contains(link, "itunes.apple.com") {
		return "apple"
	}
	return ""
}

// GetAllSourceNames returns all available source names.
func GetAllSourceNames() []string {
	return []string{"netease", "qq", "kugou", "kuwo", "migu", "fivesing", "jamendo", "joox", "qianqian", "soda", "bilibili", "apple", "local"}
}

// GetDefaultSourceNames returns default search source names.
func GetDefaultSourceNames() []string {
	all := GetAllSourceNames()
	var def []string
	excluded := map[string]bool{"bilibili": true, "joox": true, "jamendo": true, "fivesing": true, "local": true}
	for _, s := range all {
		if !excluded[s] {
			def = append(def, s)
		}
	}
	return def
}

// GetPlaylistSourceNames returns sources that support playlist search.
func GetPlaylistSourceNames() []string {
	return []string{"netease", "qq", "kugou", "kuwo", "migu", "jamendo", "joox", "qianqian", "bilibili", "soda", "fivesing", "apple"}
}

// GetAlbumSourceNames returns sources that support album search.
func GetAlbumSourceNames() []string {
	return []string{"netease", "qq", "kugou", "kuwo", "migu", "jamendo", "joox", "qianqian", "soda", "apple"}
}

// GetSourceDescription returns the Chinese name for a source.
func GetSourceDescription(source string) string {
	desc := map[string]string{
		"netcase":  "网易云音乐",
		"qq":       "QQ音乐",
		"kugou":    "酷狗音乐",
		"kuwo":     "酷我音乐",
		"migu":     "咪咕音乐",
		"fivesing": "5sing",
		"jamendo":  "Jamendo (CC)",
		"joox":     "JOOX",
		"qianqian": "千千音乐",
		"soda":     "汽水音乐",
		"bilibili": "Bilibili",
		"apple":    "Apple Music",
		"local":    "本地音乐",
	}
	if d, ok := desc[source]; ok {
		return d
	}
	return "未知音乐源"
}

// GetOriginalLink returns the original web URL for a song/playlist/album.
func GetOriginalLink(source, id, typeStr string) string {
	switch source {
	case "netease":
		switch typeStr {
		case "album":
			return "https://music.163.com/#/album?id=" + id
		case "playlist":
			return "https://music.163.com/#/playlist?id=" + id
		default:
			return "https://music.163.com/#/song?id=" + id
		}
	case "qq":
		switch typeStr {
		case "album":
			return "https://y.qq.com/n/ryqq/albumDetail/" + id
		case "playlist":
			return "https://y.qq.com/n/ryqq/playlist/" + id
		default:
			return "https://y.qq.com/n/ryqq/songDetail/" + id
		}
	case "kugou":
		switch typeStr {
		case "album":
			return "https://www.kugou.com/album/" + id + ".html"
		case "playlist":
			return "https://www.kugou.com/yy/special/single/" + id + ".html"
		default:
			return "https://www.kugou.com/song/#hash=" + id
		}
	case "kuwo":
		switch typeStr {
		case "album":
			return "http://www.kuwo.cn/album_detail/" + id
		case "playlist":
			return "http://www.kuwo.cn/playlist_detail/" + id
		default:
			return "http://www.kuwo.cn/play_detail/" + id
		}
	case "migu":
		switch typeStr {
		case "album":
			return "https://music.migu.cn/v3/music/album/" + id
		default:
			return "https://music.migu.cn/v3/music/song/" + id
		}
	case "bilibili":
		return "https://www.bilibili.com/video/" + id
	case "soda":
		switch typeStr {
		case "album":
			return "https://www.qishui.com/share/album?album_id=" + id
		case "playlist":
			return "https://www.qishui.com/playlist/" + id
		}
	case "apple":
		switch typeStr {
		case "album":
			return "https://music.apple.com/album/" + id
		case "playlist":
			return "https://music.apple.com/playlist/" + id
		default:
			return "https://music.apple.com/song/" + id
		}
	}
	return ""
}

// CalcSongSimilarity computes a similarity score between two songs (0-1).
func CalcSongSimilarity(name, artist, candName, candArtist string) float64 {
	nameA := normalizeText(name)
	nameB := normalizeText(candName)
	if nameA == "" || nameB == "" {
		return 0
	}
	nameSim := similarityScore(nameA, nameB)
	artistA := normalizeText(artist)
	artistB := normalizeText(candArtist)
	if artistA == "" || artistB == "" {
		return nameSim
	}
	artistSim := similarityScore(artistA, artistB)
	return nameSim*0.7 + artistSim*0.3
}

func normalizeText(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r >= 0x4E00 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func similarityScore(a, b string) float64 {
	if a == b {
		return 1
	}
	if a == "" || b == "" {
		return 0
	}
	ra := []rune(a)
	rb := []rune(b)
	maxLen := len(ra)
	if len(rb) > maxLen {
		maxLen = len(rb)
	}
	if maxLen == 0 {
		return 0
	}
	dist := levenshteinDistance(ra, rb)
	if dist >= maxLen {
		return 0
	}
	return 1 - float64(dist)/float64(maxLen)
}

func levenshteinDistance(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = del
			if ins < cur[j] {
				cur[j] = ins
			}
			if sub < cur[j] {
				cur[j] = sub
			}
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}
