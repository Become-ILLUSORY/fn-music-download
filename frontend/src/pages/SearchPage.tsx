import { useState, useCallback, useEffect, useRef } from 'react'
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
const INVALID_KEY = 'music-dl-invalid-sources'

export default function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [loading, setLoading] = useState(false)
  const [songs, setSongs] = useState<Song[]>([])
  const [invalidSet, setInvalidSet] = useState<Set<string>>(new Set())
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [resultType, setResultType] = useState<string>('')
  const [error, setError] = useState('')
  const [detailSongs, setDetailSongs] = useState<Song[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const player = usePlayer()
  const fixRef = useRef(0) // used to cancel stale background fix

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
      // Restore invalid set
      const inv = sessionStorage.getItem(INVALID_KEY)
      if (inv) setInvalidSet(new Set(JSON.parse(inv)))
    } catch {}
  }, [])

  const persist = (s: Song[], rt: string, pl?: Playlist[], ds?: Song[]) => {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ songs: s, playlists: pl || [], resultType: rt, detailSongs: ds || [] }))
    } catch {}
  }

  const persistInvalid = (s: Set<string>) => {
    try { sessionStorage.setItem(INVALID_KEY, JSON.stringify([...s])) } catch {}
  }

  const handlePlay = useCallback((song: Song) => player.play(song), [player])

  // Background fix: validate → retry → update state without blocking UI
  const startBackgroundFix = useCallback((songList: Song[], fixId: number) => {
    if (songList.length === 0) return

    // 1. Validate in background
    fetch(`${apiBase}/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ songs: songList.map(s => ({ id: s.id, source: s.source })) }),
    })
      .then(r => r.json())
      .then(valData => {
        if (fixRef.current !== fixId) return // stale

        if (!valData.results) return
        const badKeys = new Set<string>()
        valData.results.forEach((r: any) => {
          if (!r.playable) badKeys.add(`${r.source}:${r.id}`)
        })

        if (badKeys.size === 0) return

        // Update invalid set immediately so UI shows red
        setInvalidSet(badKeys)
        persistInvalid(badKeys)

        // 2. Try to find replacements for bad songs (don't wait for this to complete for UI)
        const badSongs = songList.filter(s => badKeys.has(`${s.source}:${s.id}`))

        fetch(`${apiBase}/retry`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            songs: badSongs.map(s => ({ id: s.id, source: s.source, name: s.name, artist: s.artist })),
          }),
        })
          .then(r => r.json())
          .then(retryData => {
            if (fixRef.current !== fixId) return
            if (!retryData.results) return

            // Build replacement map
            const replacements = new Map<string, Song>()
            retryData.results.forEach((r: any) => {
              if (r.altId && r.altSource) {
                const orig = songList.find(s => s.id === r.id && s.source === r.source)
                replacements.set(`${r.source}:${r.id}`, {
                  id: r.altId,
                  source: r.altSource,
                  name: r.altName || orig?.name || '',
                  artist: r.altArtist || orig?.artist || '',
                  album: orig?.album || '',
                  duration: orig?.duration || 0,
                  cover: orig?.cover || '',
                })
              }
            })

            if (replacements.size === 0) return

            // Replace songs that got fixed, remove from invalid set
            setSongs(prev => {
              const newSongs = prev.map(s => {
                const key = `${s.source}:${s.id}`
                return replacements.has(key) ? replacements.get(key)! : s
              })
              persist(newSongs, resultType, playlists, detailSongs)
              return newSongs
            })

            // Remove replaced songs from invalid set
            const fixedKeys = new Set(replacements.keys())
            setInvalidSet(prev => {
              const next = new Set(prev)
              fixedKeys.forEach(k => next.delete(k))
              persistInvalid(next)
              return next
            })
          })
          .catch(() => {})
      })
      .catch(() => {})
  }, [resultType, playlists, detailSongs])

  const handleSearch = useCallback(async (query: string, sources: string[], searchType: string) => {
    setLoading(true)
    setError('')
    setDetailSongs([])
    setSelected(new Set())
    setInvalidSet(new Set())

    setSearchParams({ q: query, type: searchType, sources: sources.join(',') }, { replace: true })

    try {
      const qs = new URLSearchParams({ q: query, type: searchType, sources: sources.join(',') })
      const res = await fetch(`${apiBase}/search?${qs}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()

      if (data.type === 'playlist_detail' || data.type === 'detail') {
        const pls = data.playlist ? [data.playlist] : []
        const songsData = data.songs || []
        setPlaylists(pls)
        setSongs(songsData)
        setDetailSongs(songsData)
        setResultType('detail')
        persist(songsData, 'detail', pls, songsData)
        // Background fix
        const id = ++fixRef.current
        startBackgroundFix(songsData, id)
      } else if (data.type === 'playlist' || data.type === 'album') {
        setPlaylists(data.playlists || [])
        setSongs([])
        setResultType(data.type)
        persist([], data.type, data.playlists)
      } else {
        // Song search — render immediately, fix in background
        const songsData = data.songs || []
        setSongs(songsData)
        setPlaylists([])
        setResultType('song')
        persist(songsData, 'song')
        // Background fix
        const id = ++fixRef.current
        startBackgroundFix(songsData, id)
      }
    } catch (err: any) {
      setError(err.message || '搜索失败')
      setSongs([])
      setPlaylists([])
    } finally {
      setLoading(false)
    }
  }, [setSearchParams, startBackgroundFix])

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
        setSongs(data.songs)
        setDetailSongs(data.songs)
        setResultType('detail')
        setPlaylists([pl])
        persist(data.songs, 'detail', [pl], data.songs)
        // Background fix
        const id = ++fixRef.current
        startBackgroundFix(data.songs, id)
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

  const displaySongs = detailSongs.length > 0 ? detailSongs : songs

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

      {displaySongs.length > 0 && (
        <SongList
          songs={displaySongs}
          onPlay={handlePlay}
          onDownload={handleDownload}
          invalidSources={invalidSet}
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
