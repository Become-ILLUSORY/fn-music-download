import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { PlayerProvider } from './hooks/usePlayer'
import Layout from './components/Layout'
import SearchPage from './pages/SearchPage'
import DownloadPage from './pages/DownloadPage'
import LocalPage from './pages/LocalPage'
import SettingsPage from './pages/SettingsPage'

export default function App() {
  return (
    <BrowserRouter basename="/app/music-dl">
      <PlayerProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Navigate to="/search" replace />} />
            <Route path="/search" element={<SearchPage />} />
            <Route path="/download" element={<DownloadPage />} />
            <Route path="/local" element={<LocalPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>
        </Routes>
      </PlayerProvider>
    </BrowserRouter>
  )
}
