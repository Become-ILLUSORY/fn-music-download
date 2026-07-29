import { useState, useRef, useEffect } from 'react'

interface Song {
  id: string
  name: string
  artist: string
  source: string
  cover: string
}

interface PlayerBarProps {
  currentSong?: Song | null
  onNext?: () => void
  onPrev?: () => void
}

export default function PlayerBar({ currentSong, onNext, onPrev }: PlayerBarProps) {
  const [playing, setPlaying] = useState(false)
  const audioRef = useRef<HTMLAudioElement | null>(null)

  useEffect(() => {
    if (!currentSong) {
      if (audioRef.current) {
        audioRef.current.pause()
        audioRef.current = null
      }
      setPlaying(false)
      return
    }
    // Get stream URL
    const url = `/app/music-dl/api/stream?id=${currentSong.id}&source=${currentSong.source}`
    if (audioRef.current) {
      audioRef.current.src = url
      audioRef.current.play().then(() => setPlaying(true)).catch(() => setPlaying(false))
    } else {
      const audio = new Audio(url)
      audioRef.current = audio
      audio.play().then(() => setPlaying(true)).catch(() => setPlaying(false))
      audio.onended = () => setPlaying(false)
    }
  }, [currentSong])

  const togglePlay = () => {
    if (!audioRef.current) return
    if (playing) {
      audioRef.current.pause()
      setPlaying(false)
    } else {
      audioRef.current.play().then(() => setPlaying(true)).catch(() => {})
    }
  }

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
        {onPrev && <button className="btn-icon" onClick={onPrev}>⏮</button>}
        <button className="btn-icon btn-play" onClick={togglePlay}>
          {playing ? '⏸' : '▶'}
        </button>
        {onNext && <button className="btn-icon" onClick={onNext}>⏭</button>}
      </div>
    </div>
  )
}
