package pkg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"

	"github.com/dhowden/tag"
	"github.com/guohuiyuan/music-lib/model"
	"github.com/guohuiyuan/music-lib/soda"
	"github.com/guohuiyuan/music-lib/utils"
)

const (
	UA_Common    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	UA_Mobile    = "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46"
	Ref_Netease  = "http://music.163.com/"
	Ref_Bilibili = "https://www.bilibili.com/"
	Ref_Migu     = "http://music.migu.cn/"
)

var ErrFFmpegNotFound = errors.New("ffmpeg not found")

// DownloadedSong holds the result of a download operation.
type DownloadedSong struct {
	Data        []byte
	Ext         string
	ContentType string
	Filename    string
	SavedPath   string
	Warning     string
	Skipped     bool
}

// ======================== Download Pipeline ========================

// DownloadSongData downloads a song's audio data with optional cover/lyrics embedding.
func DownloadSongData(song *model.Song, withCover bool, withLyrics bool) (*DownloadedSong, error) {
	return DownloadSongDataWithTemplate(song, withCover, withLyrics, DefaultDownloadFilenameTemplate)
}

// DownloadSongDataWithTemplate is like DownloadSongData with a custom filename template.
func DownloadSongDataWithTemplate(song *model.Song, withCover bool, withLyrics bool, filenameTemplate string) (*DownloadedSong, error) {
	if song == nil {
		return nil, errors.New("song is nil")
	}
	if strings.TrimSpace(song.ID) == "" || strings.TrimSpace(song.Source) == "" {
		return nil, errors.New("missing song id or source")
	}

	normalized := *song
	normalized.Name = strings.TrimSpace(normalized.Name)
	normalized.Artist = strings.TrimSpace(normalized.Artist)
	normalized.Album = strings.TrimSpace(normalized.Album)
	if normalized.Name == "" {
		normalized.Name = "Unknown"
	}
	if normalized.Artist == "" {
		normalized.Artist = "Unknown"
	}

	audioData, contentType, err := fetchSongAudio(&normalized)
	if err != nil {
		return nil, err
	}

	sigExt := DetectAudioExtBySignature(audioData)
	ext := sigExt
	if ext == "" {
		ext = DetectAudioExtByContentType(contentType)
	}
	if ext == "" {
		ext = DetectAudioExt(audioData)
	}

	var lyric string
	if withLyrics {
		if lyricFn := GetLyricFunc(normalized.Source); lyricFn != nil {
			lyric, _ = lyricFn(&normalized)
		}
	}

	var coverData []byte
	var coverMime string
	if withCover && strings.TrimSpace(normalized.Cover) != "" {
		coverData, coverMime, _ = FetchBytesWithMime(normalized.Cover, normalized.Source)
	}

	finalData := audioData
	warning := ""
	if ext == "mp3" || ext == "flac" || ext == "m4a" || ext == "wma" {
		embeddedData, embedErr := EmbedSongMetadata(audioData, &normalized, lyric, coverData, coverMime)
		switch {
		case embedErr == nil:
			finalData = embeddedData
		case errors.Is(embedErr, ErrFFmpegNotFound):
			warning = "ffmpeg not found, metadata embedding skipped"
		default:
			warning = "metadata embedding failed, using original audio"
		}
	}
	if ext == "" {
		ext = DetectAudioExt(finalData)
	}

	return &DownloadedSong{
		Data:        finalData,
		Ext:         ext,
		ContentType: AudioMimeByExt(ext),
		Filename:    BuildDownloadFilename(&normalized, ext, filenameTemplate),
		Warning:     warning,
	}, nil
}

// SaveSongToFile downloads a song and saves it to disk.
func SaveSongToFile(song *model.Song, outDir string, withCover bool, withLyrics bool) (*DownloadedSong, error) {
	result, err := DownloadSongData(song, withCover, withLyrics)
	if err != nil {
		return nil, err
	}
	return saveDownloadedSongToFile(result, outDir)
}

func saveDownloadedSongToFile(result *DownloadedSong, outDir string) (*DownloadedSong, error) {
	if result == nil {
		return nil, errors.New("download result is nil")
	}
	targetDir := strings.TrimSpace(outDir)
	if targetDir == "" {
		targetDir = DefaultWebDownloadDir
	}
	targetDir = filepath.Clean(targetDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, err
	}
	fileName := SanitizeDownloadRelativePath(result.Filename)
	filePath := filepath.Join(targetDir, fileName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filePath, result.Data, 0644); err != nil {
		return nil, err
	}
	result.Filename = fileName
	result.SavedPath = filePath
	return result, nil
}

// ======================== Fetch Audio ========================

func fetchSongAudio(song *model.Song) ([]byte, string, error) {
	if song.Source == "soda" {
		return FetchDecryptedSodaAudio(song)
	}
	dlFunc := GetDownloadFunc(song.Source)
	if dlFunc == nil {
		return nil, "", fmt.Errorf("unsupported source: %s", song.Source)
	}
	urlStr, err := dlFunc(song)
	if err != nil {
		return nil, "", err
	}
	if urlStr == "" {
		return nil, "", errors.New("empty download url")
	}
	return FetchBytesWithMime(urlStr, song.Source)
}

// FetchDecryptedSodaAudio downloads and decrypts soda (qishui) encrypted audio.
func FetchDecryptedSodaAudio(song *model.Song) ([]byte, error) {
	cookie := CM.Get("soda")
	sodaInst := soda.New(cookie)
	info, err := sodaInst.GetDownloadInfo(song)
	if err != nil {
		return nil, err
	}
	encryptedData, _, err := FetchBytesWithMime(info.URL, "soda")
	if err != nil {
		return nil, err
	}
	return soda.DecryptAudio(encryptedData, info.PlayAuth)
}

// ======================== HTTP Helpers ========================

// BuildSourceRequest creates an HTTP request with UA/Referer/Cookie for a given source.
func BuildSourceRequest(method, urlStr, source, rangeHeader string) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	req.Header.Set("User-Agent", UA_Common)
	switch source {
	case "bilibili":
		req.Header.Set("Referer", Ref_Bilibili)
	case "netease":
		req.Header.Set("Referer", Ref_Netease)
	case "migu":
		req.Header.Set("User-Agent", UA_Mobile)
		req.Header.Set("Referer", Ref_Migu)
	case "qq":
		req.Header.Set("Referer", "http://y.qq.com")
	}
	if cookie := CM.Get(source); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	return req, nil
}

// FetchBytesWithMime fetches a URL and returns the body bytes + content type.
func FetchBytesWithMime(urlStr string, source string) ([]byte, string, error) {
	if fetch, handled, err := NewSourceRangeFetch(urlStr, source, ""); handled || err != nil {
		if err != nil {
			return nil, "", err
		}
		var buf bytes.Buffer
		if fetch.ContentLength > 0 && fetch.ContentLength <= int64(1<<(strconv.IntSize-1)-1) {
			buf.Grow(int(fetch.ContentLength))
		}
		if err := fetch.WriteTo(&buf); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), fetch.ContentType, nil
	}
	return fetchBytesSingle(urlStr, source)
}

func fetchBytesSingle(urlStr string, source string) ([]byte, string, error) {
	req, err := BuildSourceRequest("GET", urlStr, source, "")
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" && len(data) > 0 {
		contentType = http.DetectContentType(data)
	}
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	return data, contentType, nil
}

// ======================== Range Fetch ========================

// SourceRangeFetch handles ranged HTTP downloads with parallel chunk fetching.
type SourceRangeFetch struct {
	URL           string
	Source        string
	StatusCode    int
	ContentLength int64
	ContentRange  string
	ContentType   string
	Ext           string
	Start         int64
	End           int64
	Total         int64
}

// NewSourceRangeFetch probes a URL for Range support and returns a fetcher.
func NewSourceRangeFetch(urlStr string, source string, rangeHeader string) (*SourceRangeFetch, bool, error) {
	req, err := BuildSourceRequest("GET", urlStr, source, "bytes=0-3")
	if err != nil {
		return nil, false, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return nil, false, nil
	}
	total, ok := ParseContentRangeTotal(resp.Header.Get("Content-Range"))
	if !ok || total <= 0 {
		return nil, false, nil
	}
	if total > int64(1<<(strconv.IntSize-1)-1) {
		return nil, true, fmt.Errorf("download too large: %d bytes", total)
	}
	probeData, _ := io.ReadAll(resp.Body)
	ext := DetectAudioExtBySignature(probeData)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if ext == "" {
		ext = DetectAudioExtByContentType(contentType)
	}
	if ext != "" && (contentType == "" || strings.HasPrefix(strings.ToLower(contentType), "application/octet-stream")) {
		contentType = AudioMimeByExt(ext)
	}
	start, end, partial, ok := ResolveRangeHeader(rangeHeader, total)
	if !ok {
		return nil, true, fmt.Errorf("invalid range: %s", rangeHeader)
	}
	fetch := &SourceRangeFetch{
		URL:           urlStr,
		Source:        source,
		StatusCode:    200,
		ContentLength: end - start + 1,
		ContentType:   contentType,
		Ext:           ext,
		Start:         start,
		End:           end,
		Total:         total,
	}
	if partial {
		fetch.StatusCode = http.StatusPartialContent
		fetch.ContentRange = fmt.Sprintf("bytes %d-%d/%d", start, end, total)
	}
	return fetch, true, nil
}

// WriteTo writes the full range content to the writer.
func (f *SourceRangeFetch) WriteTo(w io.Writer) error {
	if f == nil {
		return errors.New("nil range fetch")
	}
	return writeParallelRange(w, f.URL, f.Source, f.Start, f.End)
}

type rangeChunkJob struct {
	index int
	start int64
	end   int64
}

type rangeChunkResult struct {
	index       int
	data        []byte
	contentType string
	err         error
}

const firstChunkSize int64 = 32 * 1024
const chunkSize int64 = 256 * 1024
const maxConcurrentChunks = 16

func writeParallelRange(w io.Writer, urlStr, source string, start, end int64) error {
	if end < start {
		return nil
	}
	var jobs []rangeChunkJob
	firstEnd := start + firstChunkSize - 1
	if firstEnd > end {
		firstEnd = end
	}
	jobs = append(jobs, rangeChunkJob{index: 0, start: start, end: firstEnd})
	for chunkStart := firstEnd + 1; chunkStart <= end; chunkStart += chunkSize {
		chunkEnd := chunkStart + chunkSize - 1
		if chunkEnd > end {
			chunkEnd = end
		}
		jobs = append(jobs, rangeChunkJob{index: len(jobs), start: chunkStart, end: chunkEnd})
	}

	sem := make(chan struct{}, maxConcurrentChunks)
	results := make(chan rangeChunkResult, len(jobs))
	for _, job := range jobs {
		job := job
		go func() {
			sem <- struct{}{}
			chunk, ct, err := fetchRangeChunk(urlStr, source, job.start, job.end)
			<-sem
			results <- rangeChunkResult{index: job.index, data: chunk, contentType: ct, err: err}
		}()
	}

	next := 0
	pending := make(map[int]rangeChunkResult)
	for next < len(jobs) {
		result := <-results
		if result.err != nil {
			return result.err
		}
		pending[result.index] = result
		for {
			ready, ok := pending[next]
			if !ok {
				break
			}
			if _, err := w.Write(ready.data); err != nil {
				return err
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			delete(pending, next)
			next++
		}
	}
	return nil
}

func fetchRangeChunk(urlStr, source string, start, end int64) ([]byte, string, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := BuildSourceRequest("GET", urlStr, source, fmt.Sprintf("bytes=%d-%d", start, end))
		if err != nil {
			return nil, "", err
		}
		client := &http.Client{Timeout: 90 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("range %d-%d returned status %d", start, end, resp.StatusCode)
			continue
		}
		if readErr != nil {
			lastErr = readErr
			continue
		}
		expected := int(end - start + 1)
		if len(data) != expected {
			lastErr = fmt.Errorf("range %d-%d returned %d bytes, want %d", start, end, len(data), expected)
			continue
		}
		return data, resp.Header.Get("Content-Type"), nil
	}
	return nil, "", lastErr
}

// ParseContentRangeTotal extracts the total size from a Content-Range header.
func ParseContentRangeTotal(value string) (int64, bool) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "*" {
		return 0, false
	}
	total, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	return total, err == nil
}

// ResolveRangeHeader parses a Range header and returns start/end/partial/ok.
func ResolveRangeHeader(value string, total int64) (int64, int64, bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, total - 1, false, true
	}
	if !strings.HasPrefix(strings.ToLower(value), "bytes=") {
		return 0, 0, false, false
	}
	spec := strings.TrimSpace(value[len("bytes="):])
	if strings.Contains(spec, ",") {
		return 0, 0, false, false
	}
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false, false
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	switch {
	case left == "" && right == "":
		return 0, 0, false, false
	case left == "":
		suffix, err := strconv.ParseInt(right, 10, 64)
		if err != nil || suffix <= 0 {
			return 0, 0, false, false
		}
		if suffix > total {
			suffix = total
		}
		return total - suffix, total - 1, true, true
	case right == "":
		start, err := strconv.ParseInt(left, 10, 64)
		if err != nil || start < 0 || start >= total {
			return 0, 0, false, false
		}
		return start, total - 1, true, true
	default:
		start, err := strconv.ParseInt(left, 10, 64)
		if err != nil || start < 0 {
			return 0, 0, false, false
		}
		end, err := strconv.ParseInt(right, 10, 64)
		if err != nil || end < start {
			return 0, 0, false, false
		}
		if start >= total {
			return 0, 0, false, false
		}
		if end >= total {
			end = total - 1
		}
		return start, end, true, true
	}
}

// ======================== Audio Detection ========================

// DetectAudioExt detects audio extension from data.
func DetectAudioExt(data []byte) string {
	if ext := DetectAudioExtBySignature(data); ext != "" {
		return ext
	}
	return "mp3"
}

// DetectAudioExtBySignature detects audio format from magic bytes.
func DetectAudioExtBySignature(data []byte) string {
	if len(data) >= 16 && bytes.Equal(data[:16], []byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}) {
		return "wma"
	}
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{'f', 'L', 'a', 'C'}) {
		return "flac"
	}
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{'I', 'D', '3'}) {
		return "mp3"
	}
	if len(data) >= 2 && data[0] == 0xFF && (data[1]&0xE0) == 0xE0 {
		return "mp3"
	}
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{'O', 'g', 'g', 'S'}) {
		return "ogg"
	}
	if len(data) >= 12 && bytes.Equal(data[4:8], []byte{'f', 't', 'y', 'p'}) {
		return "m4a"
	}
	return ""
}

// DetectAudioExtByContentType detects audio extension from Content-Type.
func DetectAudioExtByContentType(contentType string) string {
	ct := strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.Index(ct, ";"); idx >= 0 {
		ct = strings.TrimSpace(ct[:idx])
	}
	switch ct {
	case "audio/flac", "audio/x-flac":
		return "flac"
	case "audio/x-ms-wma", "audio/wma", "video/x-ms-asf", "application/vnd.ms-asf":
		return "wma"
	case "audio/mpeg", "audio/mp3", "audio/x-mp3":
		return "mp3"
	case "audio/ogg", "application/ogg":
		return "ogg"
	case "audio/mp4", "audio/x-m4a", "audio/aac", "audio/aacp":
		return "m4a"
	}
	return ""
}

// AudioMimeByExt returns the MIME type for an audio extension.
func AudioMimeByExt(ext string) string {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case "wma":
		return "audio/x-ms-wma"
	case "flac":
		return "audio/flac"
	case "ogg":
		return "audio/ogg"
	case "m4a":
		return "audio/mp4"
	default:
		return "audio/mpeg"
	}
}

// ======================== Filename Helpers ========================

// BuildDownloadFilename builds a safe filename from template.
func BuildDownloadFilename(song *model.Song, ext, filenameTemplate string) string {
	tmpl := strings.TrimSpace(filenameTemplate)
	if tmpl == "" {
		tmpl = DefaultDownloadFilenameTemplate
	}
	ext = strings.TrimSpace(strings.TrimPrefix(ext, "."))
	name := "Unknown"
	artist := "Unknown"
	album := ""
	source := ""
	id := ""
	if song != nil {
		if strings.TrimSpace(song.Name) != "" {
			name = strings.TrimSpace(song.Name)
		}
		if strings.TrimSpace(song.Artist) != "" {
			artist = strings.TrimSpace(song.Artist)
		}
		album = strings.TrimSpace(song.Album)
		source = strings.TrimSpace(song.Source)
		id = strings.TrimSpace(song.ID)
	}
	name = sanitizeTemplateValue(name, "Unknown")
	artist = sanitizeTemplateValue(artist, "Unknown")
	album = sanitizeTemplateValue(album, "")
	source = sanitizeTemplateValue(source, "")
	id = sanitizeTemplateValue(id, "")

	hasExt := strings.Contains(tmpl, "{ext}")
	rendered := strings.NewReplacer(
		"{name}", name, "{artist}", artist, "{album}", album,
		"{source}", source, "{id}", id, "{ext}", ext,
	).Replace(tmpl)
	rendered = strings.TrimSpace(rendered)
	if rendered == "" {
		rendered = strings.TrimSpace(DefaultDownloadFilenameTemplate)
		rendered = strings.NewReplacer("{name}", name, "{artist}", artist, "{album}", album, "{source}", source, "{id}", id, "{ext}", ext).Replace(rendered)
	}
	if !hasExt && ext != "" {
		rendered += "." + ext
	}
	return SanitizeDownloadRelativePath(rendered)
}

// SanitizeDownloadRelativePath sanitizes a download file path.
func SanitizeDownloadRelativePath(name string) string {
	name = strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	parts := strings.Split(name, "/")
	safe := make([]string, 0, len(parts))
	for _, part := range parts {
		p := sanitizePathSegment(part)
		if p == "" || p == "." || p == ".." {
			continue
		}
		safe = append(safe, p)
	}
	if len(safe) == 0 {
		return "download"
	}
	return filepath.Join(safe...)
}

func sanitizeTemplateValue(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	value = sanitizePathSegment(value)
	if value == "" {
		return fallback
	}
	return value
}

func sanitizePathSegment(value string) string {
	value = strings.Trim(value, " .")
	if value == "" {
		return ""
	}
	return strings.Trim(utils.SanitizeFilename(value), " .")
}

// ======================== ID3v2 Metadata Embedding ========================

// EmbedSongMetadata embeds metadata into audio data (ID3v2 for MP3, ffmpeg for others).
func EmbedSongMetadata(audioData []byte, song *model.Song, lyric string, coverData []byte, coverMime string) ([]byte, error) {
	if len(audioData) == 0 {
		return nil, errors.New("empty audio data")
	}
	ext := DetectAudioExt(audioData)
	if song != nil && song.Ext != "" {
		se := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(song.Ext, ".")))
		switch se {
		case "mp3", "flac", "m4a", "wma":
			ext = se
		}
	}
	title := ""
	artist := ""
	album := ""
	if song != nil {
		title = strings.TrimSpace(song.Name)
		artist = strings.TrimSpace(song.Artist)
		album = strings.TrimSpace(song.Album)
	}
	lyric = strings.TrimSpace(lyric)
	coverMime = normalizeCoverMime(coverMime)
	incomingCover := len(coverData) > 0

	if existing, err := tag.ReadFrom(bytes.NewReader(audioData)); err == nil {
		if t := strings.TrimSpace(existing.Title()); title == "" && t != "" {
			title = t
		}
		if a := strings.TrimSpace(existing.Artist()); artist == "" && a != "" {
			artist = a
		}
		if al := strings.TrimSpace(existing.Album()); album == "" && al != "" {
			album = al
		}
		if l := strings.TrimSpace(existing.Lyrics()); lyric == "" && l != "" {
			lyric = l
		}
		if ext == "mp3" && !incomingCover {
			if pic := existing.Picture(); pic != nil && len(pic.Data) > 0 {
				coverData = append([]byte(nil), pic.Data...)
				if pic.MIMEType != "" {
					coverMime = pic.MIMEType
				}
			}
		}
	}
	if ext != "mp3" && ext != "flac" && ext != "m4a" && ext != "wma" {
		return audioData, nil
	}
	if title == "" && artist == "" && album == "" && lyric == "" && len(coverData) == 0 {
		return audioData, nil
	}
	if ext == "mp3" {
		return embedMP3ID3v23Metadata(audioData, title, artist, album, lyric, coverData, coverMime)
	}
	return embedAudioMetadataByFFmpeg(audioData, ext, title, artist, album, lyric, coverData, coverMime)
}

func embedMP3ID3v23Metadata(audioData []byte, title, artist, album, lyric string, coverData []byte, coverMime string) ([]byte, error) {
	var frames bytes.Buffer
	replace := map[string]bool{}
	if title != "" {
		replace["TIT2"] = true
	}
	if artist != "" {
		replace["TPE1"] = true
	}
	if album != "" {
		replace["TALB"] = true
	}
	if lyric != "" {
		replace["USLT"] = true
	}
	if len(coverData) > 0 {
		replace["APIC"] = true
	}
	frames.Write(preservedID3v23Frames(audioData, replace))

	if title != "" {
		frames.Write(id3v23Frame("TIT2", id3TextFramePayload(title)))
	}
	if artist != "" {
		frames.Write(id3v23Frame("TPE1", id3TextFramePayload(artist)))
	}
	if album != "" {
		frames.Write(id3v23Frame("TALB", id3TextFramePayload(album)))
	}
	if lyric != "" {
		frames.Write(id3v23Frame("USLT", id3USLTPayload(lyric)))
	}
	if len(coverData) > 0 {
		frames.Write(id3v23Frame("APIC", id3APICPayload(coverData, coverMime)))
	}

	frameData := frames.Bytes()
	if len(frameData) == 0 {
		return audioData, nil
	}
	size := id3SynchsafeSize(len(frameData))
	out := make([]byte, 0, 10+len(frameData)+len(audioData))
	out = append(out, 'I', 'D', '3', 0x03, 0x00, 0x00)
	out = append(out, size[:]...)
	out = append(out, frameData...)
	out = append(out, stripID3v2Prefix(audioData)...)
	return out, nil
}

func stripID3v2Prefix(data []byte) []byte {
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return data
	}
	tagSize, ok := decodeID3SynchsafeSize(data[6:10])
	if !ok {
		return data
	}
	total := 10 + tagSize
	if data[5]&0x10 != 0 {
		total += 10
	}
	if total <= 0 || total > len(data) {
		return data
	}
	return data[total:]
}

func decodeID3SynchsafeSize(data []byte) (int, bool) {
	if len(data) < 4 {
		return 0, false
	}
	if data[0]&0x80 != 0 || data[1]&0x80 != 0 || data[2]&0x80 != 0 || data[3]&0x80 != 0 {
		return 0, false
	}
	return int(data[0])<<21 | int(data[1])<<14 | int(data[2])<<7 | int(data[3]), true
}

func id3SynchsafeSize(size int) [4]byte {
	return [4]byte{
		byte((size >> 21) & 0x7F),
		byte((size >> 14) & 0x7F),
		byte((size >> 7) & 0x7F),
		byte(size & 0x7F),
	}
}

func id3UTF16LEText(value string) []byte {
	units := utf16.Encode([]rune(value))
	data := make([]byte, 0, 2+len(units)*2)
	data = append(data, 0xFF, 0xFE)
	for _, u := range units {
		data = binary.LittleEndian.AppendUint16(data, u)
	}
	return data
}

func id3TextFramePayload(value string) []byte {
	payload := []byte{0x01}
	payload = append(payload, id3UTF16LEText(value)...)
	return payload
}

func id3USLTPayload(lyric string) []byte {
	payload := []byte{0x01, 'e', 'n', 'g'}
	payload = append(payload, id3UTF16LEText("")...)
	payload = append(payload, 0x00, 0x00)
	payload = append(payload, id3UTF16LEText(lyric)...)
	return payload
}

func id3APICPayload(coverData []byte, coverMime string) []byte {
	mime := normalizeCoverMime(coverMime)
	payload := []byte{0x00}
	payload = append(payload, []byte(mime)...)
	payload = append(payload, 0x00, 0x03, 0x00)
	payload = append(payload, coverData...)
	return payload
}

func id3v23Frame(id string, payload []byte) []byte {
	if id == "" || len(payload) == 0 {
		return nil
	}
	frame := make([]byte, 0, 10+len(payload))
	frame = append(frame, []byte(id)...)
	frame = binary.BigEndian.AppendUint32(frame, uint32(len(payload)))
	frame = append(frame, 0x00, 0x00)
	frame = append(frame, payload...)
	return frame
}

func isID3v23FrameID(id []byte) bool {
	if len(id) != 4 {
		return false
	}
	for _, c := range id {
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}

func preservedID3v23Frames(audioData []byte, replace map[string]bool) []byte {
	if len(audioData) < 10 || string(audioData[:3]) != "ID3" || audioData[3] != 0x03 {
		return nil
	}
	if audioData[5]&0x40 != 0 {
		return nil
	}
	tagSize, ok := decodeID3SynchsafeSize(audioData[6:10])
	if !ok {
		return nil
	}
	tagEnd := 10 + tagSize
	if tagEnd > len(audioData) {
		return nil
	}
	tagData := audioData[10:tagEnd]
	preserved := make([]byte, 0, len(tagData))
	for pos := 0; pos+10 <= len(tagData); {
		hdr := tagData[pos : pos+10]
		if bytes.Equal(hdr, make([]byte, 10)) {
			break
		}
		if !isID3v23FrameID(hdr[:4]) {
			break
		}
		frameSize := int(binary.BigEndian.Uint32(hdr[4:8]))
		if frameSize <= 0 || pos+10+frameSize > len(tagData) {
			break
		}
		fid := string(hdr[:4])
		if !replace[fid] {
			preserved = append(preserved, tagData[pos:pos+10+frameSize]...)
		}
		pos += 10 + frameSize
	}
	return preserved
}

func normalizeCoverMime(mime string) string {
	mime = strings.TrimSpace(strings.ToLower(mime))
	if mime == "" {
		return "image/jpeg"
	}
	switch {
	case strings.Contains(mime, "png"):
		return "image/png"
	case strings.Contains(mime, "webp"):
		return "image/webp"
	case strings.Contains(mime, "gif"):
		return "image/gif"
	}
	return "image/jpeg"
}

func embedAudioMetadataByFFmpeg(audioData []byte, ext, title, artist, album, lyric string, coverData []byte, coverMime string) ([]byte, error) {
	ffmpegPath, err := ResolveFFmpegPath()
	if err != nil {
		return nil, ErrFFmpegNotFound
	}
	inFile, err := os.CreateTemp("", "fnmusicdl-in-*."+ext)
	if err != nil {
		return nil, err
	}
	inPath := inFile.Name()
	defer os.Remove(inPath)
	if _, err := inFile.Write(audioData); err != nil {
		inFile.Close()
		return nil, err
	}
	inFile.Close()

	outFile, err := os.CreateTemp("", "fnmusicdl-out-*."+ext)
	if err != nil {
		return nil, err
	}
	outPath := outFile.Name()
	outFile.Close()
	defer os.Remove(outPath)

	args := []string{"-y", "-hide_banner", "-loglevel", "error", "-i", inPath}
	hasCover := len(coverData) > 0
	coverPath := ""
	if hasCover {
		coverExt := ".jpg"
		if strings.Contains(coverMime, "png") {
			coverExt = ".png"
		}
		cf, err := os.CreateTemp("", "fnmusicdl-cover-*"+coverExt)
		if err != nil {
			return nil, err
		}
		coverPath = cf.Name()
		defer os.Remove(coverPath)
		if _, err := cf.Write(coverData); err != nil {
			cf.Close()
			return nil, err
		}
		cf.Close()
		args = append(args, "-i", coverPath)
	}
	if hasCover {
		args = append(args, "-map", "0:a:0", "-map", "1:v:0")
	} else {
		args = append(args, "-map", "0")
	}
	args = append(args, "-map_metadata", "0")
	if hasCover {
		args = append(args, "-c:a", "copy", "-c:v", "copy", "-disposition:v:0", "attached_pic")
	} else {
		args = append(args, "-c", "copy")
	}
	if title != "" {
		args = append(args, "-metadata", "title="+title)
	}
	if artist != "" {
		args = append(args, "-metadata", "artist="+artist)
	}
	if album != "" {
		args = append(args, "-metadata", "album="+album)
	}
	if lyric != "" {
		args = append(args, "-metadata", "lyrics="+lyric)
	}
	if ext == "mp3" {
		args = append(args, "-id3v2_version", "3", "-write_id3v1", "1")
	}
	args = append(args, outPath)

	cmd := exec.Command(ffmpegPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg metadata embed failed: %v, output: %s", err, strings.TrimSpace(string(out)))
	}
	finalData, err := os.ReadFile(filepath.Clean(outPath))
	if err != nil {
		return nil, err
	}
	if len(finalData) == 0 {
		return nil, errors.New("embedded output is empty")
	}
	return finalData, nil
}

// ======================== Utility ========================

// FormatSize formats a byte count as a human-readable string.
func FormatSize(s int64) string {
	if s <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f MB", float64(s)/1024/1024)
}

// ValidatePlayable checks if a song URL is actually reachable.
func ValidatePlayable(song *model.Song) bool {
	if song == nil || song.ID == "" || song.Source == "" {
		return false
	}
	if song.Source == "soda" || song.Source == "fivesing" || song.Source == "local" || song.Source == "local-file" {
		return false
	}
	fn := GetDownloadFunc(song.Source)
	if fn == nil {
		return false
	}
	urlStr, err := fn(&model.Song{ID: song.ID, Source: song.Source})
	if err != nil || urlStr == "" {
		return false
	}
	req, err := BuildSourceRequest("GET", urlStr, song.Source, "bytes=0-1")
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200 || resp.StatusCode == 206
}
