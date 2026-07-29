import { useState, useEffect } from 'react'
import SongList from '../components/SongList'
import PlayerBar from '../components/PlayerBar'

interface Playlist {
  id: string
  name: string
  description?: string
  cover?: string
  creator?: string
  kind: string
  createdAt: string
}

interface Song {
  id: string
  name: string
  artist: string
  source: string
}

interface CollectionSong extends Song {
  songId: string
  collectionId: string
  album: string
  duration: number
  cover: string
  addedAt: string
}

export default function PlaylistPage() {
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [selectedPlaylist, setSelectedPlaylist] = useState<Playlist | null>(null)
  const [songs, setSongs] = useState<CollectionSong[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [currentSong, setCurrentSong] = useState<Song | null>(null)

  const apiBase = '/app/music-dl/api'

  useEffect(() => {
    fetchPlaylists()
  }, [])

  const fetchPlaylists = () => {
    fetch(`${apiBase}/playlists`)
      .then(r => r.json())
      .then(data => setPlaylists(data.playlists || []))
      .catch(() => {})
  }

  const openPlaylist = (pl: Playlist) => {
    setSelectedPlaylist(pl)
    fetch(`${apiBase}/playlists/${pl.id}`)
      .then(r => r.json())
      .then(data => {
        const mapped = (data.songs || []).map((s: CollectionSong) => ({
          ...s,
          id: s.songId || s.id,
        }))
        setSongs(mapped)
      })
      .catch(() => {})
  }

  const createPlaylist = async () => {
    if (!newName.trim()) return
    try {
      await fetch(`${apiBase}/playlists`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newName, description: newDesc }),
      })
      setShowCreate(false)
      setNewName('')
      setNewDesc('')
      fetchPlaylists()
    } catch {}
  }

  const deletePlaylist = async (id: string) => {
    if (!confirm('确定删除此歌单？')) return
    await fetch(`${apiBase}/playlists/${id}`, { method: 'DELETE' })
    if (selectedPlaylist?.id === id) {
      setSelectedPlaylist(null)
      setSongs([])
    }
    fetchPlaylists()
  }

  return (
    <div className="page playlist-page">
      <div className="page-header">
        <h2>我的歌单</h2>
        <button className="btn-primary" onClick={() => setShowCreate(true)}>新建歌单</button>
      </div>

      {showCreate && (
        <div className="create-form">
          <input
            type="text"
            placeholder="歌单名称"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            className="input"
          />
          <input
            type="text"
            placeholder="描述（可选）"
            value={newDesc}
            onChange={e => setNewDesc(e.target.value)}
            className="input"
          />
          <div className="form-actions">
            <button className="btn-primary" onClick={createPlaylist}>创建</button>
            <button className="btn-secondary" onClick={() => setShowCreate(false)}>取消</button>
          </div>
        </div>
      )}

      {selectedPlaylist ? (
        <div>
          <div className="detail-header">
            <button className="btn-secondary" onClick={() => { setSelectedPlaylist(null); setSongs([]) }}>← 返回</button>
            <h3>{selectedPlaylist.name}</h3>
          </div>
          <SongList
            songs={songs}
            onPlay={s => setCurrentSong(s)}
          />
        </div>
      ) : (
        <div className="playlist-grid">
          {playlists.map(pl => (
            <div key={pl.id} className="playlist-card" onClick={() => openPlaylist(pl)}>
              <div className="pl-cover-wrap">
                <div className="pl-cover-placeholder">📋</div>
              </div>
              <div className="pl-info">
                <div className="pl-name">{pl.name}</div>
                {pl.description && <div className="pl-creator">{pl.description}</div>}
              </div>
              <button className="btn-delete" onClick={e => { e.stopPropagation(); deletePlaylist(pl.id) }}>✕</button>
            </div>
          ))}
          {playlists.length === 0 && <div className="empty-state">还没有歌单，点击上方创建</div>}
        </div>
      )}

      <PlayerBar currentSong={currentSong} />
    </div>
  )
}
