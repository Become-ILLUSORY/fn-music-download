import { useState, useCallback, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import SearchBar from '../components/SearchBar'
import SongList from '../components/SongList'
import PlaylistGrid from '../components/PlaylistGrid'
import PlayerBar from '../components/PlayerBar'

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

export default function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [loading, setLoading] = useState(false)
  const [songs, setSongs] = useState<Song[]>([])
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [resultType, setResultType] = useState<string>('')
  const [error, setError] = useState('')
  const [currentSong, setCurrentSong] = useState<Song | null>(null)
  const [detailSongs, setDetailSongs] = useState<Song[]>([])
  const [invalidSongs, setInvalidSongs] = useState<Set<string>>(new Set())
  const [validating, setValidating] = useState(false)

  // Restore from URL params
  const initialQuery = searchParams.get('q') || ''
  const initialType = searchParams.get('type') || 'song'
  const initialSources = searchParams.get('sources')?.split(',').filter(Boolean) || []

  const handleSearch = useCallback(async (query: string, sources: string[], searchType: string) => {
    setLoading(true)
    setError('')
    setDetailSongs([])
    setCurrentSong(null)
    setInvalidSongs(new Set())

    setSearchParams({ q: query, type: searchType, sources: sources.join(',') }, { replace: true })

    try {
      const qs = new URLSearchParams({ q: query, type: searchType, sources: sources.join(',') })
      const res = await fetch(`${apiBase}/search?${qs}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()

      if (data.type === 'playlist_detail' || data.type === 'detail') {
        setPlaylists(data.playlist ? [data.playlist] : [])
        setSongs(data.songs || [])
        setResultType('detail')
        if (data.playlist) setDetailSongs(data.songs || [])
      } else if (data.type === 'playlist' || data.type === 'album') {
        setPlaylists(data.playlists || [])
        setSongs([])
        setResultType(data.type)
      } else {
        setSongs(data.songs || [])
        setPlaylists([])
        setResultType('song')
        // Auto-validate playable sources
        if (data.songs?.length > 0) {
          autoValidate(data.songs)
        }
      }
    } catch (err: any) {
      setError(err.message || '搜索失败')
      setSongs([])
      setPlaylists([])
    } finally {
      setLoading(false)
    }
  }, [setSearchParams])

  const autoValidate = async (songList: Song[]) => {
    setValidating(true)
    try {
      const res = await fetch(`${apiBase}/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ songs: songList.map(s => ({ id: s.id, source: s.source })) }),
      })
      const data = await res.json()
      if (data.results) {
        const bad = new Set<string>()
        data.results.forEach((r: any) => {
          if (!r.playable) bad.add(`${r.source}:${r.id}`)
        })
        setInvalidSongs(bad)
      }
    } catch {}
    setValidating(false)
  }

  // Auto-search on mount
  useEffect(() => {
    if (initialQuery && initialSources.length > 0) {
      handleSearch(initialQuery, initialSources, initialType)
    }
  }, [])

  const handlePlay = (song: Song) => setCurrentSong(song)

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
          withCover: true, withLyrics: true,
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

  const handleBatchRetry = async () => {
    if (invalidSongs.size === 0) return
    const badSongs = songs.filter(s => invalidSongs.has(`${s.source}:${s.id}`))
    try {
      const res = await fetch(`${apiBase}/retry`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ songs: badSongs.map(s => ({ id: s.id, source: s.source, name: s.name, artist: s.artist })) }),
      })
      const data = await res.json()
      if (data.results?.length > 0) {
        // Add alt sources to songs list
        const altMap: Record<string, Song> = {}
        data.results.forEach((r: any) => {
          if (r.altId && r.altSource) {
            altMap[`${r.source}:${r.id}`] = {
              id: r.altId, source: r.altSource,
              name: r.altName || badSongs.find(s => s.id === r.id)?.name || '',
              artist: r.altArtist || '',
              album: '', duration: 0, cover: '',
            }
          }
        })
        setSongs(prev => {
          const keys = new Set(prev.map(s => `${s.source}:${s.id}`))
          const additions = Object.values(altMap).filter(s => !keys.has(`${s.source}:${s.id}`))
          return [...prev, ...additions]
        })
        setInvalidSongs(new Set())
        alert(`已找到 ${data.results.length} 个替换音源`)
      } else {
        alert('未找到可替换的音源')
      }
    } catch {}
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
        setSongs(data.songs)
        setDetailSongs(data.songs)
        setResultType('detail')
        setPlaylists([pl])
      }
    } catch (err: any) {
      setError(err.message || '获取详情失败')
    } finally {
      setLoading(false)
    }
  }

  const toggleSelect = (key: string) => {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key); else next.add(key)
      return next
    })
  }

  const [selected, setSelected] = useState<Set<string>>(new Set())

  return (
    <div className="page search-page">
      <SearchBar onSearch={handleSearch} loading={loading} initialQuery={initialQuery} searchType={initialType} initialSources={initialSources} />

      {error && <div className="error-msg">{error}</div>}
      {loading && <div className="loading">搜索中...</div>}

      {!loading && invalidSongs.size > 0 && resultType === 'song' && (
        <div className="retry-bar">
          <span>{invalidSongs.size} 个音源不可用</span>
          <button className="btn-primary" onClick={handleBatchRetry} disabled={validating}>
            批量换源
          </button>
        </div>
      )}

      {!loading && !error && resultType === 'detail' && playlists.length > 0 && (
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
          onToggleSelect={toggleSelect}
          invalidSources={invalidSongs}
        />
      )}

      <PlayerBar currentSong={currentSong} />
    </div>
  )
}
