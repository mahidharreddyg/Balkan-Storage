"use client"

import { useState } from "react"
import {
  File,
  ImageIcon,
  Video,
  Music,
  Archive,
  FileText,
  Download,
  Share2,
  Eye,
  MoreHorizontal,
  Trash2,
  User,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
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

interface FileCardProps {
  file: FileData
  isSelected?: boolean
  onSelect?: (fileId: string, isMultiSelect: boolean) => void
  onDragStart?: (fileIds: string[]) => void
  onDragEnd?: () => void
  onPreview?: (file: FileData) => void
  onDownload?: (file: FileData) => void
  onShare?: (file: FileData) => void
  onDelete?: (file: FileData) => void
}

const getFileIcon = (type: string) => {
  if (type.startsWith("image/")) return ImageIcon
  if (type.startsWith("video/")) return Video
  if (type.startsWith("audio/")) return Music
  if (type.includes("zip") || type.includes("rar")) return Archive
  if (type.includes("text") || type.includes("document")) return FileText
  return File
}

const getFileTypeColor = (type: string) => {
  if (type.startsWith("image/")) return "text-green-500"
  if (type.startsWith("video/")) return "text-purple-500"
  if (type.startsWith("audio/")) return "text-orange-500"
  if (type.includes("zip") || type.includes("rar")) return "text-red-500"
  if (type.includes("text") || type.includes("document")) return "text-blue-500"
  return "text-gray-400"
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 Bytes"
  const k = 1024
  const sizes = ["Bytes", "KB", "MB", "GB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
}

const formatDate = (dateString: string) =>
  new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  })

export function FileCard({
  file,
  isSelected = false,
  onSelect,
  onDragStart,
  onDragEnd,
  onPreview,
  onDownload,
  onShare,
  onDelete,
}: FileCardProps) {
  const [isHovered, setIsHovered] = useState(false)
  const { token } = useAuth()
  const FileIcon = getFileIcon(file.type)
  const fileTypeColor = getFileTypeColor(file.type)

  const handleDownload = async () => {
    if (onDownload) return onDownload(file)
    if (!token) return

    const res = await fetch(`http://localhost:8080/files/${file.id}/download`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    const blob = await res.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = file.name
    a.click()
  }

  const handleDelete = async () => {
    if (onDelete) return onDelete(file)
    if (!token) return

    await fetch(`http://localhost:8080/files/${file.id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    })
  }

  return (
    <div
      className={`glass-card p-4 rounded-2xl transition-all duration-300 group cursor-pointer relative select-none transform ${
        isSelected
          ? "ring-4 ring-blue-500 border-2 border-blue-400 bg-gradient-to-br from-blue-50/80 to-blue-100/80 shadow-lg shadow-blue-400/40"
          : "hover:scale-[1.02] hover:shadow-lg hover:shadow-black/10"
      }`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={(e) => {
        if ((e.target as HTMLElement).closest("button") || (e.target as HTMLElement).closest('[role="menuitem"]'))
          return
        const isMultiSelect = e.ctrlKey || e.metaKey
        onSelect?.(file.id, isMultiSelect)
      }}
      draggable
      onDragStart={(e) => {
        e.dataTransfer.effectAllowed = "move"
        e.dataTransfer.setData("text/plain", file.id)
        onDragStart?.([file.id])
      }}
      onDragEnd={onDragEnd}
    >
      <div className="relative mb-4">
        {file.thumbnail ? (
          <div className="relative">
            <img
              src={file.thumbnail}
              alt={file.name}
              className="w-full h-40 object-cover rounded-xl shadow-inner"
            />
          </div>
        ) : (
          <div className="w-full h-40 bg-gradient-to-br from-purple-50/40 to-blue-50/40 rounded-xl flex items-center justify-center relative">
            <FileIcon className={`w-16 h-16 ${fileTypeColor}`} />
          </div>
        )}
      </div>

      <div className="space-y-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <FileIcon className={`w-4 h-4 ${fileTypeColor}`} />
            <h3 className="font-semibold truncate text-sm">{file.name}</h3>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0 glass">
                <MoreHorizontal className="w-4 h-4 text-gray-500" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass-card border-white/20">
              <DropdownMenuItem onClick={() => onPreview?.(file)}>
                <Eye className="w-4 h-4 mr-2" /> Preview
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleDownload}>
                <Download className="w-4 h-4 mr-2" /> Download
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onShare?.(file)}>
                <Share2 className="w-4 h-4 mr-2" /> Share
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleDelete} className="text-red-600 focus:text-red-700">
                <Trash2 className="w-4 h-4 mr-2" /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span className="font-medium">{formatFileSize(file.size)}</span>
          <span>{formatDate(file.created_at)}</span>
        </div>

        {file.owner && (
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <User className="w-3 h-3" /> <span>{file.owner}</span>
          </div>
        )}

        {file.tags && file.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {file.tags.slice(0, 2).map((tag) => (
              <span
                key={tag}
                className="px-2 py-1 text-xs bg-gradient-to-r from-blue-500/20 to-purple-500/20 text-blue-700 rounded-md font-medium"
              >
                {tag}
              </span>
            ))}
            {file.tags.length > 2 && (
              <span className="px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded-md font-medium">
                +{file.tags.length - 2}
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  )
}