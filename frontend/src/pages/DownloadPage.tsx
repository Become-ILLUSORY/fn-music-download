import { useState, useEffect } from 'react'
import { usePlayer } from '../hooks/usePlayer'

interface DownloadTask {
  id: string
  songId: string
  source: string
  name: string
  artist: string
  album: string
  progress: number
  status: string
  savedPath?: string
  filename?: string
  fileSize?: number
  error?: string
  startedAt?: number
  finishedAt?: number
}

const apiBase = '/app/music-dl/api'

export default function DownloadPage() {
  const [activeTab, setActiveTab] = useState<'active' | 'completed'>('active')
  const [tasks, setTasks] = useState<DownloadTask[]>([])
  const [completed, setCompleted] = useState<any[]>([])
  const [pollId, setPollId] = useState<number>(0)
  const player = usePlayer()

  useEffect(() => {
    const fetchTasks = () => {
      fetch(`${apiBase}/download/tasks`)
        .then(r => r.json())
        .then(data => {
          setTasks(data.active || [])
        })
        .catch(() => {})
    }
    fetchTasks()
    const id = window.setInterval(fetchTasks, 1500)
    setPollId(id)
    return () => clearInterval(id)
  }, [])

  useEffect(() => {
    if (activeTab === 'completed') {
      fetch(`${apiBase}/download/completed`)
        .then(r => r.json())
        .then(data => setCompleted(data.records || []))
        .catch(() => {})
    }
  }, [activeTab, pollId])

  const deleteTask = (taskId: string, deleteFile: boolean) => {
    fetch(`${apiBase}/download/task?id=${taskId}&deleteFile=${deleteFile}`, { method: 'DELETE' })
      .then(() => {
        setTasks(prev => prev.filter(t => t.id !== taskId))
      })
      .catch(() => {})
  }

  const playFile = (path: string) => {
    player.play({ id: path, source: 'local', name: path.split('/').pop() || '', artist: '' })
  }

  const deleteFile = (path: string) => {
    if (!confirm('确定删除此文件？')) return
    fetch(`${apiBase}/local/music?path=${encodeURIComponent(path)}`, { method: 'DELETE' })
      .then(() => {
        setCompleted(prev => prev.filter(r => r.savedPath !== path))
      })
      .catch(() => {})
  }

  return (
    <div className="page download-page">
      <h2>下载管理</h2>

      <div className="sub-tabs">
        <button className={`sub-tab ${activeTab === 'active' ? 'active' : ''}`} onClick={() => setActiveTab('active')}>
          正在下载 {tasks.filter(t => t.status === 'queued' || t.status === 'downloading').length > 0 && `(${tasks.filter(t => t.status === 'queued' || t.status === 'downloading').length})`}
        </button>
        <button className={`sub-tab ${activeTab === 'completed' ? 'active' : ''}`} onClick={() => setActiveTab('completed')}>
          下载完成 {completed.length > 0 && `(${completed.length})`}
        </button>
      </div>

      {activeTab === 'active' && (
        <div className="download-list">
          {tasks.length === 0 ? (
            <div className="empty-state">暂无正在下载的任务</div>
          ) : (
            tasks.map(task => (
              <div key={task.id} className="download-item">
                <div className="dl-info">
                  <div className="dl-title">{task.name}</div>
                  <div className="dl-subtitle">{task.artist} · {task.source}</div>
                </div>
                <div className="dl-progress-wrap">
                  <div className="dl-progress-bar">
                    <div className="dl-progress-fill" style={{ width: `${Math.max(2, task.progress)}%` }} />
                  </div>
                  <span className="dl-status">
                    {task.status === 'queued' ? '排队中' :
                     task.status === 'downloading' ? `${Math.round(task.progress)}%` :
                     task.status === 'completed' ? '已完成' : '失败'}
                  </span>
                </div>
                <div className="dl-actions">
                  {task.status === 'failed' && (
                    <button className="btn-icon" onClick={() => deleteTask(task.id, false)} title="移除">✕</button>
                  )}
                  {task.status === 'completed' && task.savedPath && (
                    <>
                      <button className="btn-icon" onClick={() => playFile(task.savedPath!)} title="播放">▶</button>
                      <button className="btn-icon" onClick={() => deleteTask(task.id, true)} title="删除文件">🗑️</button>
                    </>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {activeTab === 'completed' && (
        <div className="download-list">
          {completed.length === 0 ? (
            <div className="empty-state">暂无下载记录</div>
          ) : (
            completed.map((rec, idx) => (
              <div key={rec.id || idx} className="download-item">
                <div className="dl-info">
                  <div className="dl-title">{rec.name || '未知'}</div>
                  <div className="dl-subtitle">{rec.artist || ''} · {rec.source} {rec.fileSize ? `· ${(rec.fileSize / 1024 / 1024).toFixed(1)} MB` : ''}</div>
                  {rec.error && <div className="dl-error">{rec.error}</div>}
                </div>
                <div className="dl-actions">
                  {rec.savedPath && (
                    <>
                      <button className="btn-icon" onClick={() => playFile(rec.savedPath)} title="播放">▶</button>
                      <button className="btn-icon" onClick={() => deleteFile(rec.savedPath)} title="删除">🗑️</button>
                    </>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
