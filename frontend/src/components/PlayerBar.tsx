import { usePlayer } from '../hooks/usePlayer'

const PlayIcon = () => (
  <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
    <polygon points="7,5 19,12 7,19" />
  </svg>
)

const PauseIcon = () => (
  <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
    <rect x="6" y="4" width="4" height="16" rx="1" />
    <rect x="14" y="4" width="4" height="16" rx="1" />
  </svg>
)

const StopIcon = () => (
  <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
    <rect x="6" y="6" width="12" height="12" rx="1.5" />
  </svg>
)

export default function PlayerBar() {
  const { currentSong, playing, togglePlay, stop } = usePlayer()

  if (!currentSong) return null

  return (
    <div className="player-bar">
      <div className="player-info">
        {currentSong.cover && (
          <img src={currentSong.cover} alt="" className="player-cover" crossOrigin="anonymous" />
        )}
        <div className="player-text">
          <div className="player-title">{currentSong.name}</div>
          <div className="player-artist">{currentSong.artist}</div>
        </div>
      </div>
      <div className="player-controls">
        <button className="btn-action btn-play" onClick={togglePlay} title={playing ? '暂停' : '播放'}>
          {playing ? <PauseIcon /> : <PlayIcon />}
        </button>
        <button className="btn-action btn-dl" onClick={stop} title="停止">
          <StopIcon />
        </button>
      </div>
    </div>
  )
}
