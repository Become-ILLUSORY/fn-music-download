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
  invalidSources?: Set<string>
}

export default function SongList({ songs, onPlay, onDownload, selected, onToggleSelect, invalidSources }: SongListProps) {
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
  const isInvalid = (s: Song) => invalidSources?.has(songKey(s))

  return (
    <div className="song-list">
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
          {songs.map(song => {
            const bad = isInvalid(song)
            return (
              <tr key={songKey(song)} className={`song-row ${bad ? 'invalid' : ''}`}>
                {selected && (
                  <td className="col-check">
                    <input type="checkbox" checked={selected.has(songKey(song))} onChange={() => onToggleSelect?.(songKey(song))} />
                  </td>
                )}
                <td className="col-cover">
                  {song.cover ? <img src={song.cover} alt="" className="thumb" crossOrigin="anonymous" />
                    : <div className="thumb-placeholder">🎵</div>}
                </td>
                <td className="col-name">
                  {song.name}
                  {bad && <span className="invalid-badge">无效</span>}
                </td>
                <td className="col-artist">{song.artist}</td>
                <td className="col-album">{song.album}</td>
                <td className="col-source">
                  <span className={`source-badge ${bad ? 'bad' : ''}`}>{song.source}</span>
                </td>
                <td className="col-duration">{fmtDuration(song.duration)}</td>
                <td className="col-actions">
                  <div className="action-btns">
                    {onPlay && <button className="btn-icon" onClick={() => onPlay(song)} title="试听">▶</button>}
                    {onDownload && <button className="btn-icon" onClick={() => onDownload(song)} title="下载">⬇</button>}
                  </div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>

      <div className="song-cards">
        {songs.map(song => {
          const bad = isInvalid(song)
          return (
            <div key={songKey(song)} className={`song-card ${bad ? 'invalid' : ''}`}>
              <div className="card-body">
                <div className="card-title">{song.name} {bad && <span className="invalid-badge">无效</span>}</div>
                <div className="card-subtitle">
                  {song.artist} · {song.album || '未知'} · <span className={`source-badge ${bad ? 'bad' : ''}`}>{song.source}</span>
                </div>
                <div className="card-meta">
                  <span>{fmtDuration(song.duration)}</span>
                  {song.size ? <span>{fmtSize(song.size)}</span> : null}
                </div>
              </div>
              <div className="card-actions">
                {onPlay && <button className="btn-icon" onClick={() => onPlay(song)}>▶</button>}
                {onDownload && <button className="btn-icon" onClick={() => onDownload(song)}>⬇</button>}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
