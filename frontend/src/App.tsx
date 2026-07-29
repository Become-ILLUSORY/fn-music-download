import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import SearchPage from './pages/SearchPage'
import PlaylistPage from './pages/PlaylistPage'
import LocalPage from './pages/LocalPage'
import SettingsPage from './pages/SettingsPage'

export default function App() {
  return (
    <BrowserRouter basename="/app/music-dl">
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/search" replace />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/search/:query" element={<SearchPage />} />
          <Route path="/playlist" element={<PlaylistPage />} />
          <Route path="/local" element={<LocalPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
