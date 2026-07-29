import { useState, useEffect } from 'react'

interface Settings {
  downloadDir: string
  downloadFilenameTemplate: string
  embedDownload: boolean
  webPageSize: number
  downloadConcurrency: number
  autoCacheOnPlay: boolean
}

interface Cookies {
  [key: string]: string
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings>({
    downloadDir: 'data/downloads',
    downloadFilenameTemplate: '{artist} - {name}',
    embedDownload: true,
    webPageSize: 200,
    downloadConcurrency: 3,
    autoCacheOnPlay: true,
  })
  const [cookies, setCookies] = useState<Cookies>({})
  const [cookieInput, setCookieInput] = useState('')
  const [cookieSource, setCookieSource] = useState('netease')
  const [sources, setSources] = useState<string[]>([])
  const [saved, setSaved] = useState(false)

  const apiBase = '/app/music-dl/api'

  useEffect(() => {
    fetch(`${apiBase}/sources`)
      .then(r => r.json())
      .then(data => setSources((data.sources || []).filter((s: string) => s !== 'local')))
      .catch(() => {})

    fetch(`${apiBase}/settings`)
      .then(r => r.json())
      .then(data => setSettings(data))
      .catch(() => {})

    fetch(`${apiBase}/cookies`)
      .then(r => r.json())
      .then(data => setCookies(data || {}))
      .catch(() => {})
  }, [])

  const saveSettings = async () => {
    try {
      await fetch(`${apiBase}/settings`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(settings),
      })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {}
  }

  const saveCookie = async () => {
    const newCookies = { ...cookies, [cookieSource]: cookieInput }
    try {
      await fetch(`${apiBase}/cookies`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newCookies),
      })
      setCookies(newCookies)
      setCookieInput('')
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {}
  }

  const removeCookie = async (source: string) => {
    const newCookies = { ...cookies }
    delete newCookies[source]
    try {
      await fetch(`${apiBase}/cookies`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newCookies),
      })
      setCookies(newCookies)
    } catch {}
  }

  return (
    <div className="page settings-page">
      <h2>设置</h2>

      {saved && <div className="success-msg">已保存</div>}

      <section className="settings-section">
        <h3>下载设置</h3>
        <div className="setting-item">
          <label>下载目录</label>
          <input
            type="text"
            value={settings.downloadDir}
            onChange={e => setSettings({ ...settings, downloadDir: e.target.value })}
            className="input"
          />
        </div>
        <div className="setting-item">
          <label>文件名模板</label>
          <input
            type="text"
            value={settings.downloadFilenameTemplate}
            onChange={e => setSettings({ ...settings, downloadFilenameTemplate: e.target.value })}
            className="input"
          />
          <span className="hint">{`{name} {artist} {album} {source} {id} {ext}`}</span>
        </div>
        <div className="setting-item">
          <label>
            <input
              type="checkbox"
              checked={settings.embedDownload}
              onChange={e => setSettings({ ...settings, embedDownload: e.target.checked })}
            />
            下载时内嵌元数据（封面/歌词）
          </label>
        </div>
        <button className="btn-primary" onClick={saveSettings}>保存设置</button>
      </section>

      <section className="settings-section">
        <h3>平台 Cookie</h3>
        <p className="hint">设置 Cookie 可获取更高音质和个人歌单访问权限</p>
        <div className="cookie-form">
          <select value={cookieSource} onChange={e => {
            setCookieSource(e.target.value)
            setCookieInput(cookies[e.target.value] || '')
          }}>
            {sources.map(s => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
          <input
            type="text"
            placeholder="Cookie 值"
            value={cookieInput}
            onChange={e => setCookieInput(e.target.value)}
            className="input"
          />
          <button className="btn-primary" onClick={saveCookie}>保存</button>
        </div>
        <div className="cookie-list">
          {Object.entries(cookies).filter(([, v]) => v).map(([src, val]) => (
            <div key={src} className="cookie-item">
              <span className="cookie-source">{src}</span>
              <span className="cookie-value">{val.substring(0, 40)}...</span>
              <button className="btn-delete" onClick={() => removeCookie(src)}>✕</button>
            </div>
          ))}
        </div>
      </section>

      <section className="settings-section">
        <h3>关于</h3>
        <p>fn-music-dl v1.0.0</p>
        <p className="hint">基于 go-music-dl 构建的飞牛 FNOS 原生音乐下载应用</p>
      </section>
    </div>
  )
}
