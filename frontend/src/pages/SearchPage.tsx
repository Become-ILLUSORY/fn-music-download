import { useState, useCallback, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { usePlayer } from '../hooks/usePlayer'
import SearchBar from '../components/SearchBar'
import SongList from '../components/SongList'
import PlaylistGrid from '../components/PlaylistGrid'

interface Song {
  id: string
  name: string
  artist: string
  album: string
  duration: number
  source: string
  cover: string
  ext?: string
  size?: number
}

interface Playlist {
  id: string
  name: string
  description?: string
  cover?: string
  creator?: string
  source: string
  trackCount?: number
  link?: string
}

const apiBase = '/app/music-dl/api'
const STORAGE_KEY = 'music-dl-search-results'

export default function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [loading, setLoading] = useState(false)
  const [songs, setSongs] = useState<Song[]>([])
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [resultType, setResultType] = useState<string>('')
  const [error, setError] = useState('')
  const [detailSongs, setDetailSongs] = useState<Song[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const player = usePlayer()

  // Restore from URL params
  const initialQuery = searchParams.get('q') || ''
  const initialType = searchParams.get('type') || 'song'
  const initialSources = searchParams.get('sources')?.split(',').filter(Boolean) || []

  // Restore search results from sessionStorage on mount
  useEffect(() => {
    try {
      const saved = sessionStorage.getItem(STORAGE_KEY)
      if (saved) {
        const data = JSON.parse(saved)
        if (data.songs?.length || data.playlists?.length) {
          setSongs(data.songs || [])
          setPlaylists(data.playlists || [])
          setResultType(data.resultType || '')
          setDetailSongs(data.detailSongs || [])
        }
      }
    } catch {}
  }, [])

  // Save results to sessionStorage
  const persistResults = (data: {
    songs?: Song[]
    playlists?: Playlist[]
    resultType: string
    detailSongs?: Song[]
  }) => {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(data))
    } catch {}
  }

  const handlePlay = useCallback((song: Song) => {
    player.play(song)
  }, [player])

  const autoFixSongs = useCallback(async (songList: Song[]): Promise<Song[]> => {
    if (songList.length === 0) return songList

    try {
      // 1. Validate all songs
      const valRes = await fetch(`${apiBase}/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ songs: songList.map(s => ({ id: s.id, source: s.source })) }),
      })
      const valData = await valRes.json()
      if (!valData.results) return songList

      const badKeys = new Set<string>()
      valData.results.forEach((r: any) => {
        if (!r.playable) badKeys.add(`${r.source}:${r.id}`)
      })

      if (badKeys.size === 0) return songList

      // 2. Find alt sources for bad songs
      const badSongs = songList.filter(s => badKeys.has(`${s.source}:${s.id}`))
      const retryRes = await fetch(`${apiBase}/retry`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          songs: badSongs.map(s => ({ id: s.id, source: s.source, name: s.name, artist: s.artist })),
        }),
      })
      const retryData = await retryRes.json()
      if (!retryData.results) return songList

      // 3. Build replacement map
      const replacements = new Map<string, Song>()
      retryData.results.forEach((r: any) => {
        if (r.altId && r.altSource) {
          const original = songList.find(s => s.id === r.id && s.source === r.source)
          replacements.set(`${r.source}:${r.id}`, {
            id: r.altId,
            source: r.altSource,
            name: r.altName || original?.name || '',
            artist: r.altArtist || original?.artist || '',
            album: original?.album || '',
            duration: original?.duration || 0,
            cover: original?.cover || '',
          })
        }
      })

      // 4. Apply replacements
      const result = songList.map(s => {
        const key = `${s.source}:${s.id}`
        if (badKeys.has(key) && replacements.has(key)) {
          return replacements.get(key)!
        }
        return s
      }).filter(s => {
        const key = `${s.source}:${s.id}`
        return !badKeys.has(key) || replacements.has(key)
      })

      return result
    } catch {
      return songList
    }
  }, [])

  const handleSearch = useCallback(async (query: string, sources: string[], searchType: string) => {
    setLoading(true)
    setError('')
    setDetailSongs([])
    setSelected(new Set())

    setSearchParams({ q: query, type: searchType, sources: sources.join(',') }, { replace: true })

    try {
      const qs = new URLSearchParams({ q: query, type: searchType, sources: sources.join(',') })
      const res = await fetch(`${apiBase}/search?${qs}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()

      if (data.type === 'playlist_detail' || data.type === 'detail') {
        setPlaylists(data.playlist ? [data.playlist] : [])
        setSongs(data.songs || [])
        setDetailSongs(data.songs || [])
        setResultType('detail')
        persistResults({ songs: data.songs, playlists: data.playlist ? [data.playlist] : [], resultType: 'detail', detailSongs: data.songs })
      } else if (data.type === 'playlist' || data.type === 'album') {
        setPlaylists(data.playlists || [])
        setSongs([])
        setResultType(data.type)
        persistResults({ playlists: data.playlists, resultType: data.type })
      } else {
        // Song search — auto-fix invalid sources
        const fixed = await autoFixSongs(data.songs || [])
        setSongs(fixed)
        setPlaylists([])
        setResultType('song')
        persistResults({ songs: fixed, resultType: 'song' })
      }
    } catch (err: any) {
      setError(err.message || '搜索失败')
      setSongs([])
      setPlaylists([])
    } finally {
      setLoading(false)
    }
  }, [setSearchParams, autoFixSongs])

  const handleDownload = async (song: Song) => {
    try {
      const res = await fetch(`${apiBase}/download/queue`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: song.id, source: song.source,
          name: song.name, artist: song.artist,
          album: song.album, cover: song.cover,
          duration: song.duration,
          withCover: false, withLyrics: false,
        }),
      })
      const data = await res.json()
      if (data.task) {
        alert(`已加入下载队列: ${song.name}`)
      } else if (data.error) alert(`下载失败: ${data.error}`)
    } catch (err: any) {
      alert(`下载失败: ${err.message}`)
    }
  }

  const handlePlaylistSelect = async (pl: Playlist) => {
    setLoading(true)
    setError('')
    try {
      const qs = new URLSearchParams({ url: pl.link || `${pl.source}:${pl.id}` })
      const res = await fetch(`${apiBase}/parse?${qs}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      if (data.songs) {
        const fixed = await autoFixSongs(data.songs)
        setSongs(fixed)
        setDetailSongs(fixed)
        setResultType('detail')
        setPlaylists([pl])
        persistResults({ songs: fixed, playlists: [pl], resultType: 'detail', detailSongs: fixed })
      }
    } catch (err: any) {
      setError(err.message || '获取详情失败')
    } finally {
      setLoading(false)
    }
  }

  // Auto-search on mount if no restored results
  useEffect(() => {
    if (initialQuery && initialSources.length > 0 && songs.length === 0 && playlists.length === 0) {
      handleSearch(initialQuery, initialSources, initialType)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="page search-page">
      <SearchBar onSearch={handleSearch} loading={loading} initialQuery={initialQuery} searchType={initialType} initialSources={initialSources} />

      {error && <div className="error-msg">{error}</div>}
      {loading && <div className="loading">搜索中...</div>}

      {!loading && resultType === 'detail' && playlists.length > 0 && (
        <div className="detail-header">
          <h2>{playlists[0].name}</h2>
          {playlists[0].creator && <span className="detail-creator">{playlists[0].creator}</span>}
          <span className="detail-source">{playlists[0].source}</span>
        </div>
      )}

      {!loading && playlists.length > 0 && resultType !== 'detail' && (
        <PlaylistGrid playlists={playlists} onSelect={handlePlaylistSelect} />
      )}

      {!loading && (songs.length > 0 || detailSongs.length > 0) && (
        <SongList
          songs={detailSongs.length > 0 ? detailSongs : songs}
          onPlay={handlePlay}
          onDownload={handleDownload}
          selected={selected}
          onToggleSelect={(key) => {
            setSelected(prev => {
              const next = new Set(prev)
              if (next.has(key)) next.delete(key); else next.add(key)
              return next
            })
          }}
        />
      )}
    </div>
  )
}
