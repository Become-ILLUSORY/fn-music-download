import { useState, useEffect, useCallback } from 'react'

interface SourceDesc {
  [key: string]: string
}

interface SearchBarProps {
  onSearch: (query: string, sources: string[], type: string) => void
  loading?: boolean
  initialQuery?: string
  searchType?: string
  initialSources?: string[]
}

export default function SearchBar({ onSearch, loading, initialQuery, searchType: st, initialSources }: SearchBarProps) {
  const [query, setQuery] = useState(initialQuery || '')
  const [searchType, setSearchType] = useState(st || 'song')
  const [sources, setSources] = useState<string[]>([])
  const [allSources, setAllSources] = useState<string[]>([])
  const [descriptions, setDescriptions] = useState<SourceDesc>({})
  const [showSourcePicker, setShowSourcePicker] = useState(false)

  const apiBase = '/app/music-dl/api'

  useEffect(() => {
    fetch(`${apiBase}/sources`)
      .then(r => r.json())
      .then(data => {
        setAllSources(data.sources || [])
        setDescriptions(data.descriptions || {})
        // Use initialSources if provided, otherwise use defaults
        if (initialSources && initialSources.length > 0) {
          setSources(initialSources)
        } else {
          setSources(data.defaults || [])
        }
      })
      .catch(() => {})
  }, []) // only once on mount

  useEffect(() => {
    if (initialQuery) setQuery(initialQuery)
  }, [initialQuery])

  const handleSubmit = useCallback((e?: React.FormEvent) => {
    e?.preventDefault()
    if (!query.trim()) return
    onSearch(query.trim(), sources, searchType)
  }, [query, sources, searchType, onSearch])

  const toggleSource = (s: string) => {
    setSources(prev =>
      prev.includes(s) ? prev.filter(x => x !== s) : [...prev, s]
    )
  }

  return (
    <div className="search-bar">
      <form onSubmit={handleSubmit} className="search-form">
        <div className="search-type-tabs">
          {['song', 'playlist', 'album'].map(t => (
            <button
              key={t}
              type="button"
              className={`type-tab ${searchType === t ? 'active' : ''}`}
              onClick={() => setSearchType(t)}
            >
              {t === 'song' ? '单曲' : t === 'playlist' ? '歌单' : '专辑'}
            </button>
          ))}
        </div>
        <div className="search-input-row">
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder={searchType === 'song' ? '搜索歌曲、歌手，或粘贴分享链接' : searchType === 'playlist' ? '搜索歌单、创建者，或粘贴歌单链接' : '搜索专辑、歌手，或粘贴专辑链接'}
            className="search-input"
          />
          <button type="submit" className="search-btn" disabled={loading || !query.trim()}>
            {loading ? '搜索中...' : '搜索'}
          </button>
          <button type="button" className="source-btn" onClick={() => setShowSourcePicker(!showSourcePicker)}>
            源 {sources.length}/{allSources.length}
          </button>
        </div>
      </form>

      {showSourcePicker && (
        <div className="source-picker">
          {allSources.filter(s => s !== 'local').map(s => (
            <label key={s} className="source-chip">
              <input
                type="checkbox"
                checked={sources.includes(s)}
                onChange={() => toggleSource(s)}
              />
              <span>{descriptions[s] || s}</span>
            </label>
          ))}
        </div>
      )}
    </div>
  )
}
