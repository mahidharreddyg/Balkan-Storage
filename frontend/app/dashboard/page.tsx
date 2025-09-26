"use client"

import { useState, useCallback, useEffect } from "react"
import { ProtectedRoute } from "@/components/protected-route"
import { Sidebar } from "@/components/sidebar"
import { TopNavbar } from "@/components/top-navbar"
import { StoragePanel } from "@/components/storage-panel"
import { FileCard } from "@/components/file-card"
import { UploadModal } from "@/components/upload-modal"
import { UploadProgress } from "@/components/upload-progress"
import { Button } from "@/components/ui/button"
import { Grid3X3, List, Filter } from "lucide-react"
import { CreateFolderModal } from "@/components/create-folder-modal"
import { useAuth } from "@/lib/auth-context"

interface FileData {
  id: string
  name: string
  size: number
  type: string
  created_at: string
  tags?: string[]
  thumbnail?: string
  owner?: string
}

interface UploadNotification {
  id: string
  file: File  
  progress: number
  status: "uploading" | "completed" | "error"
}
export default function DashboardPage() {
  const { token } = useAuth()
  const [files, setFiles] = useState<FileData[]>([])
  const [searchQuery, setSearchQuery] = useState("")
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid")
  const [selectedFiles, setSelectedFiles] = useState<string[]>([])
  const [draggedFiles, setDraggedFiles] = useState<string[]>([])
  const [isDragOver, setIsDragOver] = useState(false)
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false)
  const [isCreateFolderModalOpen, setIsCreateFolderModalOpen] = useState(false)
  const [uploadNotifications, setUploadNotifications] = useState<UploadNotification[]>([])
  const [globalDragOver, setGlobalDragOver] = useState(false)

  const fetchFiles = useCallback(async () => {
    if (!token) return
    try {
      const res = await fetch("http://localhost:8080/files", {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await res.json()
      setFiles(data.files || [])
    } catch (err) {
      console.error("Failed to fetch files:", err)
    }
  }, [token])

  useEffect(() => {
    fetchFiles()
  }, [fetchFiles])

  const filteredFiles = files.filter(
    (file) =>
      file.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      file.tags?.some((tag) => tag.toLowerCase().includes(searchQuery.toLowerCase())),
  )

  const handleFileSelect = useCallback((fileId: string, isMultiSelect: boolean) => {
    setSelectedFiles((prev) => {
      if (isMultiSelect) {
        return prev.includes(fileId) ? prev.filter((id) => id !== fileId) : [...prev, fileId]
      } else {
        return prev.includes(fileId) && prev.length === 1 ? [] : [fileId]
      }
    })
  }, [])

  const handleSelectAll = useCallback(() => {
    setSelectedFiles(selectedFiles.length === filteredFiles.length ? [] : filteredFiles.map((f) => f.id))
  }, [selectedFiles.length, filteredFiles])

  const handleDeleteSelected = async () => {
    if (!token || selectedFiles.length === 0) return
    try {
      await Promise.all(
        selectedFiles.map((id) =>
          fetch(`http://localhost:8080/files/${id}`, {
            method: "DELETE",
            headers: { Authorization: `Bearer ${token}` },
          }),
        ),
      )
      fetchFiles()
      setSelectedFiles([])
    } catch (err) {
      console.error("Failed to delete files:", err)
    }
  }

  const handleMoveSelected = async (folderId: string | null = null) => {
    if (!token || selectedFiles.length === 0) return
    try {
      await fetch("http://localhost:8080/files/move", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ file_ids: selectedFiles, folder_id: folderId }),
      })
      fetchFiles()
      setSelectedFiles([])
    } catch (err) {
      console.error("Failed to move files:", err)
    }
  }

  const handleUpload = () => {
    setIsUploadModalOpen(true)
  }

  const handleCreateFolder = () => {
    setIsCreateFolderModalOpen(true)
  }

  const handleFolderCreated = useCallback(() => {
    fetchFiles()
  }, [fetchFiles])

  const handleUploadComplete = useCallback(
  (files: File[]) => {
    const notifications = files.map((file) => ({
      id: Math.random().toString(36).substr(2, 9),
      file,
      progress: 0,
      status: "uploading" as const,
    }))
    setUploadNotifications((prev) => [...prev, ...notifications])
    fetchFiles()
  },
  [fetchFiles],
)

  const handleDragStart = useCallback((fileIds: string[]) => {
    setDraggedFiles(fileIds)
  }, [])

  const handleDragEnd = useCallback(() => {
    setDraggedFiles([])
    setIsDragOver(false)
  }, [])

  const handleDropOnFolder = async (folderId: string) => {
    if (!token || draggedFiles.length === 0) return
    try {
      await fetch("http://localhost:8080/files/move", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ file_ids: draggedFiles, folder_id: folderId }),
      })
      fetchFiles()
      setDraggedFiles([])
    } catch (err) {
      console.error("Failed to move files via drag-and-drop:", err)
    }
  }

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "a") {
        e.preventDefault()
        handleSelectAll()
      }
    }

    document.addEventListener("keydown", handleKeyDown)
    return () => {
      document.removeEventListener("keydown", handleKeyDown)
    }
  }, [handleSelectAll])

  return (
    <ProtectedRoute>
      <div className="flex h-screen bg-background">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <TopNavbar
            onSearchResults={(res) => setFiles(res.files || [])}
            onUpload={handleUpload}
            onCreateFolder={handleCreateFolder}
          />
          <div className="flex-1 flex overflow-hidden">
            <main className="flex-1 overflow-y-auto p-6">
              <div className="mb-8">
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-2xl font-bold bg-white bg-clip-text text-transparent">
                    All Files
                  </h2>
                  {selectedFiles.length > 0 && (
                    <div className="flex items-center gap-2">
                      <Button variant="outline" size="sm" onClick={() => setSelectedFiles([])}>
                        Clear
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => handleMoveSelected(null)}>
                        Move
                      </Button>
                      <Button variant="outline" size="sm" onClick={handleDeleteSelected} className="text-red-500">
                        Delete
                      </Button>
                    </div>
                  )}
                </div>
                <div
                  className={`transition-all duration-200 ${
                    isDragOver ? "bg-purple-50/40 border-2 border-dashed border-purple-400 rounded-xl p-4" : ""
                  }`}
                  onDragOver={(e) => {
                    e.preventDefault()
                    setIsDragOver(true)
                  }}
                  onDragLeave={(e) => {
                    e.preventDefault()
                    setIsDragOver(false)
                  }}
                  onDrop={(e) => {
                    e.preventDefault()
                    setIsDragOver(false)
                  }}
                >
                  {viewMode === "grid" ? (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                      {filteredFiles.map((file) => (
                        <FileCard
                          key={file.id}
                          file={file}
                          isSelected={selectedFiles.includes(file.id)}
                          onSelect={handleFileSelect}
                          onDragStart={handleDragStart}
                          onDragEnd={handleDragEnd}
                        />
                      ))}
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {filteredFiles.map((file) => (
                        <div
                          key={file.id}
                          className={`glass-card p-4 rounded-xl flex items-center gap-4 cursor-pointer select-none ${
                            selectedFiles.includes(file.id)
                              ? "ring-2 ring-purple-500 bg-gradient-to-r from-purple-50/40 to-blue-50/40 border-2 border-purple-400"
                              : "border-2 border-transparent hover:bg-white/10"
                          }`}
                          onClick={(e) => {
                            const isMultiSelect = e.ctrlKey || e.metaKey
                            handleFileSelect(file.id, isMultiSelect)
                          }}
                          draggable
                          onDragStart={() =>
                            handleDragStart(selectedFiles.includes(file.id) ? selectedFiles : [file.id])
                          }
                          onDragEnd={handleDragEnd}
                        >
                          <div className="w-10 h-10 bg-gradient-to-br from-purple-500/20 to-blue-500/20 rounded-lg flex items-center justify-center">
                            <Grid3X3 className="w-5 h-5 text-purple-600" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-medium text-foreground truncate">{file.name}</p>
                            <p className="text-sm text-muted-foreground">
                              {new Date(file.created_at).toLocaleDateString()}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </main>
            <aside className="hidden xl:block">
              <StoragePanel />
            </aside>
          </div>
        </div>
      </div>
      <UploadModal
        isOpen={isUploadModalOpen}
        onClose={() => setIsUploadModalOpen(false)}
        onUploadComplete={handleUploadComplete}
      />
      <CreateFolderModal
        isOpen={isCreateFolderModalOpen}
        onClose={() => setIsCreateFolderModalOpen(false)}
        onFolderCreated={handleFolderCreated}
      />
      <UploadProgress
        uploads={uploadNotifications}
        onDismiss={(id) => setUploadNotifications((prev) => prev.filter((n) => n.id !== id))}
      />
      {globalDragOver && (
        <div className="fixed inset-0 bg-purple-500/20 backdrop-blur-sm z-40 flex items-center justify-center">
          <div className="glass-card p-8 rounded-2xl text-center">
            <div className="w-16 h-16 bg-gradient-to-br from-purple-600 to-blue-600 rounded-full flex items-center justify-center mx-auto mb-4">
              <Grid3X3 className="w-8 h-8 text-white" />
            </div>
            <h3 className="text-xl font-semibold text-foreground mb-2">Drop files to upload</h3>
            <p className="text-muted-foreground">Release to upload your files</p>
          </div>
        </div>
      )}
    </ProtectedRoute>
  )
}