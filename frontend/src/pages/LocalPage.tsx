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

const apiBase = '/app/music-dl/api'

export default function LocalPage() {
  const [songs, setSongs] = useState<LocalSong[]>([])
  const [loading, setLoading] = useState(false)
  const [currentSong, setCurrentSong] = useState<Song | null>(null)
  const [dir, setDir] = useState('')
  const [renaming, setRenaming] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

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

  // Extend SongList with action buttons for each local song
  const songsWithActions = songs.map(s => ({
    ...s,
    _actions: (
      <div className="action-btns" style={{ display: 'flex', gap: 4 }}>
        {renaming === s.path ? (
          <>
            <input
              type="text"
              value={renameValue}
              onChange={e => setRenameValue(e.target.value)}
              className="rename-input"
              autoFocus
              onKeyDown={e => { if (e.key === 'Enter') handleRename(s); if (e.key === 'Escape') setRenaming(null) }}
            />
            <button className="btn-icon" onClick={() => handleRename(s)} title="确认">✓</button>
            <button className="btn-icon" onClick={() => setRenaming(null)} title="取消">✕</button>
          </>
        ) : (
          <>
            <button className="btn-icon" onClick={() => startRename(s)} title="重命名">✏️</button>
            <button className="btn-icon" onClick={() => handleDelete(s)} title="删除">🗑️</button>
          </>
        )}
      </div>
    ),
  })) as any[]

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
          {/* Desktop table */}
          <table className="song-table">
            <thead>
              <tr>
                <th className="col-name">文件名</th>
                <th className="col-size">大小</th>
                <th className="col-modified">修改时间</th>
                <th className="col-actions">操作</th>
              </tr>
            </thead>
            <tbody>
              {songs.map(s => (
                <tr key={s.path} className="song-row">
                  <td className="col-name">
                    <div className="local-name">
                      <button className="btn-link" onClick={() => setCurrentSong(s)} title="播放">
                        {s.name}.{s.ext}
                      </button>
                    </div>
                  </td>
                  <td className="col-size">{(s.size / 1024 / 1024).toFixed(1)} MB</td>
                  <td className="col-modified">{new Date(s.modified).toLocaleString()}</td>
                  <td className="col-actions">
                    {renaming === s.path ? (
                      <div className="rename-row">
                        <input
                          type="text"
                          value={renameValue}
                          onChange={e => setRenameValue(e.target.value)}
                          className="rename-input"
                          autoFocus
                          onKeyDown={e => { if (e.key === 'Enter') handleRename(s); if (e.key === 'Escape') setRenaming(null) }}
                        />
                        <button className="btn-icon" onClick={() => handleRename(s)}>✓</button>
                        <button className="btn-icon" onClick={() => setRenaming(null)}>✕</button>
                      </div>
                    ) : (
                      <div className="action-btns">
                        <button className="btn-icon" onClick={() => setCurrentSong(s)} title="播放">▶</button>
                        <button className="btn-icon" onClick={() => startRename(s)} title="重命名">✏️</button>
                        <button className="btn-icon" onClick={() => handleDelete(s)} title="删除">🗑️</button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Mobile cards */}
          <div className="song-cards">
            {songs.map(s => (
              <div key={s.path} className="song-card">
                <div className="card-body" style={{ flex: 1 }}>
                  <div className="card-title">{s.name}.{s.ext}</div>
                  <div className="card-meta">
                    <span>{(s.size / 1024 / 1024).toFixed(1)} MB</span>
                    <span>{new Date(s.modified).toLocaleString()}</span>
                  </div>
                  {renaming === s.path ? (
                    <div className="rename-row">
                      <input
                        type="text"
                        value={renameValue}
                        onChange={e => setRenameValue(e.target.value)}
                        className="rename-input"
                        autoFocus
                        onKeyDown={e => { if (e.key === 'Enter') handleRename(s); if (e.key === 'Escape') setRenaming(null) }}
                      />
                      <button className="btn-icon" onClick={() => handleRename(s)}>✓</button>
                      <button className="btn-icon" onClick={() => setRenaming(null)}>✕</button>
                    </div>
                  ) : (
                    <div className="action-btns" style={{ marginTop: 4 }}>
                      <button className="btn-icon" onClick={() => setCurrentSong(s)} title="播放">▶</button>
                      <button className="btn-icon" onClick={() => startRename(s)} title="重命名">✏️</button>
                      <button className="btn-icon" onClick={() => handleDelete(s)} title="删除">🗑️</button>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : !loading && (
        <div className="empty-state">
          <p>暂无本地音乐</p>
          <p className="hint">下载目录: {dir}</p>
        </div>
      )}

      <PlayerBar currentSong={currentSong} />
    </div>
  )
}
