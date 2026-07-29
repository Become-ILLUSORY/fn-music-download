import { createContext, useContext, useState, useRef, useEffect, ReactNode } from 'react'

export interface PlayableSong {
  id: string
  name: string
  artist: string
  source: string
  cover?: string
}

interface PlayerContextType {
  currentSong: PlayableSong | null
  playing: boolean
  play: (song: PlayableSong) => void
  togglePlay: () => void
  stop: () => void
}

const PlayerContext = createContext<PlayerContextType>(null!)

export function PlayerProvider({ children }: { children: ReactNode }) {
  const [currentSong, setCurrentSong] = useState<PlayableSong | null>(null)
  const [playing, setPlaying] = useState(false)
  const audioRef = useRef<HTMLAudioElement | null>(null)

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (audioRef.current) {
        audioRef.current.pause()
        audioRef.current = null
      }
    }
  }, [])

  const play = (song: PlayableSong) => {
    // If same song, just toggle
    if (currentSong?.id === song.id && currentSong?.source === song.source) {
      if (audioRef.current) {
        if (playing) {
          audioRef.current.pause()
          setPlaying(false)
        } else {
          audioRef.current.play().then(() => setPlaying(true)).catch(() => {})
        }
      }
      return
    }

    // Stop current
    if (audioRef.current) {
      audioRef.current.pause()
    }

    setCurrentSong(song)

    const url = `/app/music-dl/api/stream?id=${song.id}&source=${song.source}`
    const audio = new Audio(url)
    audioRef.current = audio
    audio.play().then(() => setPlaying(true)).catch(() => setPlaying(false))
    audio.onended = () => setPlaying(false)
    audio.onerror = () => setPlaying(false)
  }

  const togglePlay = () => {
    if (!audioRef.current) return
    if (playing) {
      audioRef.current.pause()
      setPlaying(false)
    } else {
      audioRef.current.play().then(() => setPlaying(true)).catch(() => {})
    }
  }

  const stop = () => {
    if (audioRef.current) {
      audioRef.current.pause()
      audioRef.current = null
    }
    setCurrentSong(null)
    setPlaying(false)
  }

  return (
    <PlayerContext.Provider value={{ currentSong, playing, play, togglePlay, stop }}>
      {children}
    </PlayerContext.Provider>
  )
}

export function usePlayer() {
  return useContext(PlayerContext)
}
