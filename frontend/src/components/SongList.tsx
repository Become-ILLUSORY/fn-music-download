/// <reference types="vite/client" />

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

interface SongListProps {
  songs: Song[]
  onPlay?: (song: Song) => void
  onDownload?: (song: Song) => void
  selected?: Set<string>
  onToggleSelect?: (key: string) => void
}

export default function SongList({ songs, onPlay, onDownload, selected, onToggleSelect }: SongListProps) {
  if (songs.length === 0) return null

  const fmtDuration = (sec: number) => {
    if (!sec) return '--:--'
    const m = Math.floor(sec / 60)
    const s = sec % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  const fmtSize = (bytes: number) => {
    if (!bytes) return ''
    return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  }

  const songKey = (s: Song) => `${s.source}:${s.id}`

  return (
    <div className="song-list">
      {/* Desktop table view */}
      <table className="song-table">
        <thead>
          <tr>
            {selected && <th className="col-check"></th>}
            <th className="col-cover"></th>
            <th className="col-name">歌曲</th>
            <th className="col-artist">歌手</th>
            <th className="col-album">专辑</th>
            <th className="col-source">来源</th>
            <th className="col-duration">时长</th>
            <th className="col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          {songs.map(song => (
            <tr key={songKey(song)} className="song-row">
              {selected && (
                <td className="col-check">
                  <input
                    type="checkbox"
                    checked={selected.has(songKey(song))}
                    onChange={() => onToggleSelect?.(songKey(song))}
                  />
                </td>
              )}
              <td className="col-cover">
                {song.cover ? (
                  <img src={song.cover} alt="" className="thumb" crossOrigin="anonymous" />
                ) : (
                  <div className="thumb-placeholder">🎵</div>
                )}
              </td>
              <td className="col-name">{song.name}</td>
              <td className="col-artist">{song.artist}</td>
              <td className="col-album">{song.album}</td>
              <td className="col-source">
                <span className="source-badge">{song.source}</span>
              </td>
              <td className="col-duration">{fmtDuration(song.duration)}</td>
              <td className="col-actions">
                <div className="action-btns">
                  {onPlay && (
                    <button className="btn-icon" onClick={() => onPlay(song)} title="试听">▶</button>
                  )}
                  {onDownload && (
                    <button className="btn-icon" onClick={() => onDownload(song)} title="下载">⬇</button>
                  )}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Mobile card view */}
      <div className="song-cards">
        {songs.map(song => (
          <div key={songKey(song)} className="song-card">
            <div className="card-left">
              {song.cover ? (
                <img src={song.cover} alt="" className="card-cover" crossOrigin="anonymous" />
              ) : (
                <div className="card-cover-placeholder">🎵</div>
              )}
            </div>
            <div className="card-body">
              <div className="card-title">{song.name}</div>
              <div className="card-subtitle">
                {song.artist} · {song.album || '未知专辑'} · <span className="source-badge">{song.source}</span>
              </div>
              <div className="card-meta">
                <span>{fmtDuration(song.duration)}</span>
                {song.size ? <span>{fmtSize(song.size)}</span> : null}
              </div>
            </div>
            <div className="card-actions">
              {onPlay && (
                <button className="btn-icon" onClick={() => onPlay(song)}>▶</button>
              )}
              {onDownload && (
                <button className="btn-icon" onClick={() => onDownload(song)}>⬇</button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
