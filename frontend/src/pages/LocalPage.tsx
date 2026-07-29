import { useState, useEffect, useRef } from 'react'
import { usePlayer } from '../hooks/usePlayer'
import SongList from '../components/SongList'

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

const apiBase = '/app/music-dl/api'

export default function LocalPage() {
  const [songs, setSongs] = useState<LocalSong[]>([])
  const [loading, setLoading] = useState(false)
  const [dir, setDir] = useState('')
  const [renaming, setRenaming] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)
  const player = usePlayer()

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
    if (!confirm(`确定删除 ${song.name}？\n此操作会删除实际文件。`)) return
    try {
      const res = await fetch(`${apiBase}/local/music?path=${encodeURIComponent(song.path)}`, { method: 'DELETE' })
      const data = await res.json()
      if (data.status === 'deleted' || data.status === 'already_deleted') {
        scanLocal()
      }
    } catch {}
  }

  const handleRename = async (song: LocalSong) => {
    if (!renameValue.trim()) return
    try {
      const res = await fetch(`${apiBase}/local/rename`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: song.path, newName: renameValue.trim() }),
      })
      const data = await res.json()
      if (data.status === 'renamed') {
        setRenaming(null)
        scanLocal()
      } else if (data.error) {
        alert(`重命名失败: ${data.error}`)
      }
    } catch (err: any) {
      alert(`重命名失败: ${err.message}`)
    }
  }

  const startRename = (song: LocalSong) => {
    setRenaming(song.path)
    setRenameValue(song.name)
  }

  return (
    <div className="page local-page">
      <div className="page-header">
        <h2><span className="mobile-only">💿</span> 本地音乐</h2>
        <div className="header-actions">
          <span className="dir-label">{dir}</span>
          <input type="file" accept="audio/*" hidden ref={fileRef} onChange={handleUpload} />
          <button className="btn-primary" onClick={() => fileRef.current?.click()}>上传</button>
          <button className="btn-secondary" onClick={scanLocal} disabled={loading}>
            {loading ? '扫描中...' : '刷新'}
          </button>
        </div>
      </div>

      {songs.length > 0 ? (
        <div>
          <div className="list-header">
            <div className="result-count">
              共 <span className="count">{songs.length}</span> 首本地音乐
            </div>
          </div>

          <ul className="result-list">
            {songs.map(s => (
              <li key={s.path} className="song-card">
                <div className="cover-wrapper">
                  <div className="cover-placeholder">💿</div>
                </div>
                <div className="song-info">
                  <h3 className="song-name">
                    <button className="btn-link" onClick={() => player.play(s)} title="播放">
                      {s.name}.{s.ext}
                    </button>
                  </h3>
                  <div className="artist-line">
                    <span className="artist-text">{(s.size / 1024 / 1024).toFixed(1)} MB</span>
                    <span className="meta-separator">·</span>
                    <span className="artist-text">{new Date(s.modified).toLocaleString()}</span>
                  </div>
                  <div className="tags">
                    <span className="tag tag-src">本地</span>
                    <span className="tag">{s.ext?.toUpperCase()}</span>
                  </div>
                </div>
                <div className="actions">
                  {renaming === s.path ? (
                    <>
                      <div className="rename-inline">
                        <input
                          type="text"
                          value={renameValue}
                          onChange={e => setRenameValue(e.target.value)}
                          className="rename-input"
                          autoFocus
                          onKeyDown={e => { if (e.key === 'Enter') handleRename(s); if (e.key === 'Escape') setRenaming(null) }}
                        />
                        <button className="btn-circle btn-dl" onClick={() => handleRename(s)} title="确认">✓</button>
                        <button className="btn-circle btn-dl" onClick={() => setRenaming(null)} title="取消">✕</button>
                      </div>
                    </>
                  ) : (
                    <>
                      <button className="btn-circle btn-play" onClick={() => player.play(s)} title="播放">▶</button>
                      <button className="btn-circle btn-dl" onClick={() => startRename(s)} title="重命名">✏️</button>
                      <button className="btn-circle btn-delete" onClick={() => handleDelete(s)} title="删除">🗑️</button>
                    </>
                  )}
                </div>
              </li>
            ))}
          </ul>
        </div>
      ) : !loading && (
        <div className="empty-state">
          <p>暂无本地音乐</p>
          <p className="hint">下载目录: {dir}</p>
        </div>
      )}
    </div>
  )
}
