interface Playlist {
  id: string
  name: string
  description?: string
  cover?: string
  creator?: string
  source: string
  trackCount?: number
  link?: string
}

interface PlaylistGridProps {
  playlists: Playlist[]
  onSelect?: (pl: Playlist) => void
}

export default function PlaylistGrid({ playlists, onSelect }: PlaylistGridProps) {
  if (playlists.length === 0) return null

  return (
    <div className="playlist-grid">
      {playlists.map(pl => (
        <div
          key={`${pl.source}:${pl.id}`}
          className="playlist-card"
          onClick={() => onSelect?.(pl)}
        >
          <div className="pl-cover-wrap">
            {pl.cover ? (
              <img src={pl.cover} alt="" className="pl-cover" crossOrigin="anonymous" />
            ) : (
              <div className="pl-cover-placeholder">📋</div>
            )}
            <span className="pl-source-tag">{pl.source}</span>
          </div>
          <div className="pl-info">
            <div className="pl-name">{pl.name}</div>
            {pl.creator && <div className="pl-creator">{pl.creator}</div>}
            {pl.trackCount ? <div className="pl-count">{pl.trackCount} 首</div> : null}
          </div>
        </div>
      ))}
    </div>
  )
}
