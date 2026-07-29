import { usePlayer } from '../hooks/usePlayer'

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
        <button className="btn-circle btn-play" onClick={togglePlay}>
          {playing ? '⏸' : '▶'}
        </button>
        <button className="btn-circle btn-dl" onClick={stop}>⏹</button>
      </div>
    </div>
  )
}
