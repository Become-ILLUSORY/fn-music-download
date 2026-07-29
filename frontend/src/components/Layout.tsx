import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'

const tabs = [
  { key: 'search', label: '搜索', icon: '🔍' },
  { key: 'playlist', label: '歌单', icon: '📋' },
  { key: 'local', label: '本地', icon: '💿' },
  { key: 'settings', label: '设置', icon: '⚙️' },
]

export default function Layout() {
  const navigate = useNavigate()
  const location = useLocation()
  const currentTab = location.pathname.split('/')[1] || 'search'

  return (
    <div className="app-layout">
      {/* Desktop sidebar */}
      <nav className="desktop-sidebar">
        <div className="sidebar-header">
          <h1>🎵 音乐下载</h1>
        </div>
        <div className="sidebar-nav">
          {tabs.map(tab => (
            <button
              key={tab.key}
              className={`nav-item ${currentTab === tab.key ? 'active' : ''}`}
              onClick={() => navigate(`/${tab.key}`)}
            >
              <span className="nav-icon">{tab.icon}</span>
              <span className="nav-label">{tab.label}</span>
            </button>
          ))}
        </div>
      </nav>

      {/* Main content */}
      <main className="main-content">
        <Outlet />
      </main>

      {/* Mobile bottom nav */}
      <nav className="mobile-bottom-nav">
        {tabs.map(tab => (
          <button
            key={tab.key}
            className={`tab-item ${currentTab === tab.key ? 'active' : ''}`}
            onClick={() => navigate(`/${tab.key}`)}
          >
            <span className="tab-icon">{tab.icon}</span>
            <span className="tab-label">{tab.label}</span>
          </button>
        ))}
      </nav>
    </div>
  )
}
