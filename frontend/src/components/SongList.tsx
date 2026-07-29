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

  const sourceLabel = (s: string) => {
    const map: Record<string, string> = {
      netease: '网易云', qq: 'QQ', kugou: '酷狗', kuwo: '酷我',
      migu: '咪咕', bilibili: 'B站', soda: '汽水', apple: 'Apple',
      fivesing: '5sing', jamendo: 'Jamendo', joox: 'JOOX', qianqian: '千千', local: '本地',
    }
    return map[s] || s
  }

  return (
    <div className="song-list-wrap">
      {songs.length > 0 && (
        <div className="list-header">
          <div className="result-count">
            找到 <span className="count">{songs.length}</span> 首歌曲
          </div>
        </div>
      )}

      <ul className="result-list">
        {songs.map(song => {
          const bad = isInvalid(song)
          return (
            <li key={songKey(song)} className={`song-card ${bad ? 'invalid' : ''}`}>
              {selected && (
                <div className="checkbox-wrapper">
                  <input
                    type="checkbox"
                    checked={selected.has(songKey(song))}
                    onChange={() => onToggleSelect?.(songKey(song))}
                  />
                </div>
              )}
              <div className="cover-wrapper">
                {song.cover ? (
                  <img src={song.cover} alt="" className="cover-img" loading="lazy" crossOrigin="anonymous" />
                ) : (
                  <div className="cover-placeholder">🎵</div>
                )}
              </div>
              <div className="song-info">
                <h3 className="song-name">
                  <span className="song-name-text" title={song.name}>{song.name}</span>
                  {bad && <span className="invalid-badge">无效</span>}
                </h3>
                <div className="artist-line">
                  <span className="artist-icon">👤</span>
                  <span className="artist-text">{song.artist || '未知歌手'}</span>
                  {song.album && (
                    <>
                      <span className="meta-separator">·</span>
                      <span className="album-text" title={song.album}>{song.album}</span>
                    </>
                  )}
                </div>
                <div className="tags">
                  <span className={`tag ${bad ? 'tag-fail' : 'tag-src'}`}>
                    {bad ? '失效' : sourceLabel(song.source)}
                  </span>
                  <span className="tag">{fmtDuration(song.duration)}</span>
                  {song.size ? <span className="tag tag-size">{fmtSize(song.size)}</span> : null}
                </div>
              </div>
              <div className="actions">
                {onPlay && (
                  <button className="btn-circle btn-play" onClick={() => onPlay(song)} title="试听">
                    ▶
                  </button>
                )}
                {onDownload && (
                  <button className="btn-circle btn-dl" onClick={() => onDownload(song)} title="下载">
                    ⬇
                  </button>
                )}
              </div>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
