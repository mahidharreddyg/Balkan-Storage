"use client"

import { useState, useEffect } from "react"
import { X, CheckCircle, AlertCircle, File } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { useAuth } from "@/lib/auth-context"

interface UploadNotification {
  id: string
  file: File
  progress: number
  status: "uploading" | "completed" | "error"
}

interface UploadProgressProps {
  uploads: UploadNotification[]
  onDismiss: (id: string) => void
}

export function UploadProgress({ uploads, onDismiss }: UploadProgressProps) {
  const [visibleUploads, setVisibleUploads] = useState<UploadNotification[]>([])
  const { token } = useAuth()

  useEffect(() => {
    setVisibleUploads(uploads)
  }, [uploads])

  const uploadFile = async (upload: UploadNotification) => {
    if (!token) return

    try {
      const formData = new FormData()
      formData.append("file", upload.file)

      const xhr = new XMLHttpRequest()
      xhr.open("POST", "http://localhost:8080/files/upload", true)
      xhr.setRequestHeader("Authorization", `Bearer ${token}`)

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const percent = (event.loaded / event.total) * 100
          setVisibleUploads((prev) =>
            prev.map((u) => (u.id === upload.id ? { ...u, progress: percent } : u))
          )
        }
      }

      xhr.onload = () => {
        if (xhr.status === 200) {
          setVisibleUploads((prev) =>
            prev.map((u) => (u.id === upload.id ? { ...u, progress: 100, status: "completed" } : u))
          )
        } else {
          setVisibleUploads((prev) =>
            prev.map((u) => (u.id === upload.id ? { ...u, status: "error" } : u))
          )
        }
      }

      xhr.onerror = () => {
        setVisibleUploads((prev) =>
          prev.map((u) => (u.id === upload.id ? { ...u, status: "error" } : u))
        )
      }

      xhr.send(formData)
    } catch {
      setVisibleUploads((prev) =>
        prev.map((u) => (u.id === upload.id ? { ...u, status: "error" } : u))
      )
    }
  }

  useEffect(() => {
    uploads.forEach((upload) => {
      if (upload.status === "uploading") {
        uploadFile(upload)
      }
    })
  }, [uploads])

  if (visibleUploads.length === 0) return null

  return (
    <div className="fixed bottom-6 right-6 z-40 space-y-3 max-w-sm">
      {visibleUploads.map((upload) => (
        <div
          key={upload.id}
          className="glass-card p-4 rounded-xl shadow-lg animate-slide-in"
        >
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-600 to-purple-400 flex items-center justify-center">
              <File className="w-4 h-4 text-white" />
            </div>
            <span className="flex-1 text-sm font-medium truncate">{upload.file.name}</span>
            {upload.status === "completed" && (
              <CheckCircle className="w-5 h-5 text-green-500 animate-bounce" />
            )}
            {upload.status === "error" && <AlertCircle className="w-5 h-5 text-red-500" />}
            <Button
              variant="ghost"
              size="icon"
              className="w-6 h-6 hover:bg-white/20"
              onClick={() => onDismiss(upload.id)}
            >
              <X className="w-3 h-3" />
            </Button>
          </div>
          {upload.status === "uploading" && (
            <Progress
              value={upload.progress}
              className="h-2 bg-gray-200 dark:bg-gray-700"
              style={{
                background: "linear-gradient(135deg, #7C3AED, #B86BFF)",
                transition: "width 0.4s ease",
              }}
            />
          )}
          <div className="text-xs text-muted-foreground mt-1">
            {upload.status === "completed"
              ? "Upload completed"
              : upload.status === "error"
              ? "Upload failed"
              : `Uploading... ${Math.round(upload.progress)}%`}
          </div>
        </div>
      ))}

      <style jsx global>{`
        .glass-card {
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.05));
          backdrop-filter: blur(14px);
          -webkit-backdrop-filter: blur(14px);
          border: 1px solid rgba(255, 255, 255, 0.18);
        }
        .animate-slide-in {
          animation: slideIn 0.4s ease forwards;
        }
        @keyframes slideIn {
          from {
            opacity: 0;
            transform: translateX(20px) translateY(20px);
          }
          to {
            opacity: 1;
            transform: translateX(0) translateY(0);
          }
        }
      `}</style>
    </div>
  )
}