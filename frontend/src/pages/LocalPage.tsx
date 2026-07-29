import { useState, useEffect, useRef } from 'react'
import SongList from '../components/SongList'
import PlayerBar from '../components/PlayerBar'

interface Song {
  id: string
  name: string
  artist: string
  source: string
}

interface LocalSong extends Song {
  path: string
  album: string
  duration: number
  size: number
  ext: string
  cover: string
  modified: string
}

export default function LocalPage() {
  const [songs, setSongs] = useState<LocalSong[]>([])
  const [loading, setLoading] = useState(false)
  const [currentSong, setCurrentSong] = useState<Song | null>(null)
  const [dir, setDir] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  const apiBase = '/app/music-dl/api'

  useEffect(() => {
    scanLocal()
  }, [])

  const scanLocal = () => {
    setLoading(true)
    fetch(`${apiBase}/local/music`)
      .then(r => r.json())
      .then(data => {
        setSongs((data.songs || []).map((s: LocalSong) => ({
          ...s,
          id: s.path || s.name,
          source: 'local',
        })))
        setDir(data.dir || '')
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const form = new FormData()
    form.append('file', file)
    try {
      await fetch(`${apiBase}/local/upload`, { method: 'POST', body: form })
      scanLocal()
    } catch {}
    if (fileRef.current) fileRef.current.value = ''
  }

  const handleDelete = async (song: LocalSong) => {
    if (!confirm(`确定删除 ${song.name}？\n此操作会实际删除文件。`)) return
    try {
      await fetch(`${apiBase}/local/music?path=${encodeURIComponent(song.path)}`, { method: 'DELETE' })
      scanLocal()
    } catch {}
  }

  return (
    <div className="page local-page">
      <div className="page-header">
        <h2>本地音乐</h2>
        <div className="header-actions">
          <span className="dir-label">{dir}</span>
          <input type="file" accept="audio/*" hidden ref={fileRef} onChange={handleUpload} />
          <button className="btn-primary" onClick={() => fileRef.current?.click()}>上传</button>
          <button className="btn-secondary" onClick={scanLocal} disabled={loading}>
            {loading ? '扫描中...' : '刷新'}
          </button>
        </div>
      </div>

      <SongList
        songs={songs}
        onPlay={s => setCurrentSong(s)}
        onDownload={() => {}}
      />

      {songs.length === 0 && !loading && (
        <div className="empty-state">
          暂无本地音乐，请先下载或上传音频文件
        </div>
      )}

      <PlayerBar currentSong={currentSong} />
    </div>
  )
}
