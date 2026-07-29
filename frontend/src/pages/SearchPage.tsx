import { useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
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

export default function SearchPage() {
  const { query: urlQuery } = useParams()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [songs, setSongs] = useState<Song[]>([])
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [resultType, setResultType] = useState<string>('')
  const [error, setError] = useState('')
  const [currentSong, setCurrentSong] = useState<Song | null>(null)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [lastQuery, setLastQuery] = useState('')
  const [lastSources, setLastSources] = useState<string[]>([])
  const [lastType, setLastType] = useState('song')
  const [detailSongs, setDetailSongs] = useState<Song[]>([])

  const apiBase = '/app/music-dl/api'

  const handleSearch = useCallback(async (query: string, sources: string[], searchType: string) => {
    setLoading(true)
    setError('')
    setDetailSongs([])
    setCurrentSong(null)
    setLastQuery(query)
    setLastSources(sources)
    setLastType(searchType)
    navigate(`/search/${encodeURIComponent(query)}`, { replace: true })

    try {
      const qs = new URLSearchParams({ q: query, type: searchType, sources: sources.join(',') })
      const res = await fetch(`${apiBase}/search?${qs}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()

      if (data.type === 'playlist_detail' || data.type === 'detail') {
        setPlaylists(data.playlist ? [data.playlist] : [])
        setSongs(data.songs || [])
        setResultType('detail')
        if (data.playlist) {
          setDetailSongs(data.songs || [])
        }
      } else if (data.type === 'playlist' || data.type === 'album') {
        setPlaylists(data.playlists || [])
        setSongs([])
        setResultType(data.type)
      } else {
        setSongs(data.songs || [])
        setPlaylists([])
        setResultType('song')
      }
    } catch (err: any) {
      setError(err.message || '搜索失败')
      setSongs([])
      setPlaylists([])
    } finally {
      setLoading(false)
    }
  }, [navigate])

  const handlePlay = (song: Song) => {
    setCurrentSong(song)
  }

  const handleDownload = async (song: Song) => {
    try {
      const res = await fetch(`${apiBase}/download`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: song.id,
          source: song.source,
          name: song.name,
          artist: song.artist,
          album: song.album,
          cover: song.cover,
          duration: song.duration,
          withCover: true,
          withLyrics: true,
          saveLocal: true,
        }),
      })
      const data = await res.json()
      if (data.success) {
        alert(`已下载: ${data.filename}`)
      } else if (data.error) {
        alert(`下载失败: ${data.error}`)
      }
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
        setSongs(data.songs)
        setDetailSongs(data.songs)
        setResultType('detail')
        setPlaylists([pl])
      }
    } catch (err: any) {
      setError(err.message || '获取歌单详情失败')
    } finally {
      setLoading(false)
    }
  }

  const toggleSelect = (key: string) => {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  return (
    <div className="page search-page">
      <SearchBar onSearch={handleSearch} loading={loading} searchType={lastType} />

      {error && <div className="error-msg">{error}</div>}

      {loading && <div className="loading">搜索中...</div>}

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
        />
      )}

      <PlayerBar currentSong={currentSong} />
    </div>
  )
}
