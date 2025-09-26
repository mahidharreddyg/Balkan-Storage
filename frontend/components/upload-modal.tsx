"use client"

import { useState, useCallback, useRef } from "react"
import { X, Upload, File, CheckCircle, AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { useAuth } from "@/lib/auth-context"

interface UploadFile {
  id: string
  file: File
  progress: number
  status: "uploading" | "completed" | "error"
  error?: string
}

interface UploadModalProps {
  isOpen: boolean
  onClose: () => void
  onUploadComplete?: (files: any[]) => void
}

export function UploadModal({ isOpen, onClose, onUploadComplete }: UploadModalProps) {
  const [uploadFiles, setUploadFiles] = useState<UploadFile[]>([])
  const [isDragOver, setIsDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { token } = useAuth()

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)
    const files = Array.from(e.dataTransfer.files)
    handleFiles(files)
  }, [])

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    handleFiles(files)
  }, [])

  const handleFiles = useCallback((files: File[]) => {
    const newUploadFiles: UploadFile[] = files.map((file) => ({
      id: Math.random().toString(36).substr(2, 9),
      file,
      progress: 0,
      status: "uploading",
    }))

    setUploadFiles((prev) => [...prev, ...newUploadFiles])
    newUploadFiles.forEach((uploadFile) => uploadToServer(uploadFile))
  }, [])

  const uploadToServer = async (uploadFile: UploadFile) => {
    const formData = new FormData()
    formData.append("file", uploadFile.file)

    try {
      const xhr = new XMLHttpRequest()
      xhr.open("POST", "http://localhost:8080/upload")
      if (token) xhr.setRequestHeader("Authorization", `Bearer ${token}`)

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const percent = (event.loaded / event.total) * 100
          setUploadFiles((prev) =>
            prev.map((f) =>
              f.id === uploadFile.id ? { ...f, progress: percent } : f
            )
          )
        }
      }

      xhr.onload = () => {
        if (xhr.status === 200) {
          setUploadFiles((prev) =>
            prev.map((f) =>
              f.id === uploadFile.id ? { ...f, status: "completed", progress: 100 } : f
            )
          )
          if (onUploadComplete) {
            const response = JSON.parse(xhr.responseText)
            onUploadComplete([response])
          }
        } else {
          setUploadFiles((prev) =>
            prev.map((f) =>
              f.id === uploadFile.id
                ? { ...f, status: "error", error: xhr.statusText }
                : f
            )
          )
        }
      }

      xhr.onerror = () => {
        setUploadFiles((prev) =>
          prev.map((f) =>
            f.id === uploadFile.id
              ? { ...f, status: "error", error: "Network error" }
              : f
          )
        )
      }

      xhr.send(formData)
    } catch (err) {
      console.error("Upload failed:", err)
    }
  }

  const handleClose = useCallback(() => {
    setUploadFiles([])
    onClose()
  }, [onClose])

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes"
    const k = 1024
    const sizes = ["Bytes", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="glass-modal w-full max-w-3xl max-h-[85vh] overflow-hidden rounded-3xl border border-white/20 shadow-2xl animate-fade-in">
        <div className="flex items-center justify-between p-8 border-b border-white/10">
          <div>
            <h2 className="text-2xl font-bold inline-block bg-clip-text text-transparent"
              style={{
                background: "linear-gradient(135deg, #7C3AED, #B86BFF)",
                WebkitBackgroundClip: "text"
              }}>
              Upload Files
            </h2>
            <p className="text-muted-foreground mt-1">Drag and drop files or browse to upload</p>
          </div>
          <Button variant="ghost" size="icon" onClick={handleClose} className="hover:bg-white/10">
            <X className="w-5 h-5" />
          </Button>
        </div>

        <div className="p-8">
          <div
            className={`border-2 border-dashed rounded-2xl p-12 text-center transition-all duration-300 ${
              isDragOver
                ? "border-purple-400 bg-gradient-to-br from-purple-50/20 to-blue-50/20 scale-[1.02]"
                : "border-gray-300 hover:border-purple-400 hover:bg-gradient-to-br hover:from-purple-50/10 hover:to-blue-50/10"
            }`}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            <div className="w-20 h-20 bg-gradient-to-br from-purple-600 to-blue-600 rounded-full flex items-center justify-center mx-auto mb-6 animate-pulse">
              <Upload className="w-10 h-10 text-white" />
            </div>
            <h3 className="text-xl font-semibold text-foreground mb-3">Drop files here to upload</h3>
            <p className="text-muted-foreground mb-6 text-lg">or click the button below to browse</p>
            <Button
              onClick={() => fileInputRef.current?.click()}
              className="gradient-btn px-8 py-3 text-lg font-medium hover:scale-105 transition"
              size="lg"
            >
              Choose Files
            </Button>
            <p className="text-sm text-muted-foreground mt-4">Support for multiple files • Maximum file size: 100MB</p>
            <input ref={fileInputRef} type="file" multiple className="hidden" onChange={handleFileSelect} />
          </div>

          {uploadFiles.length > 0 && (
            <div className="mt-8 space-y-4 max-h-80 overflow-y-auto animate-slide-up">
              <div className="flex items-center justify-between">
                <h4 className="text-lg font-semibold text-foreground">Upload Progress</h4>
                <span className="text-sm text-muted-foreground">
                  {uploadFiles.filter((f) => f.status === "completed").length} of {uploadFiles.length} completed
                </span>
              </div>
              {uploadFiles.map((uploadFile) => (
                <div key={uploadFile.id} className="glass-card p-5 rounded-xl border border-white/10 animate-fade-in">
                  <div className="flex items-center gap-4 mb-3">
                    <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-blue-500 rounded-lg flex items-center justify-center animate-bounce-slow">
                      <File className="w-5 h-5 text-white" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <span className="text-sm font-medium text-foreground truncate block">{uploadFile.file.name}</span>
                      <span className="text-xs text-muted-foreground">{formatFileSize(uploadFile.file.size)}</span>
                    </div>
                    {uploadFile.status === "completed" && <CheckCircle className="w-5 h-5 text-green-500 animate-pop-in" />}
                    {uploadFile.status === "error" && <AlertCircle className="w-5 h-5 text-red-500" />}
                  </div>
                  <Progress value={uploadFile.progress} className="h-2 mb-2 transition-all duration-500" />
                  <div className="flex justify-between items-center">
                    <span className="text-xs text-muted-foreground">
                      {uploadFile.status === "completed"
                        ? "Upload completed"
                        : uploadFile.status === "error"
                          ? "Upload failed"
                          : `Uploading... ${Math.round(uploadFile.progress)}%`}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <style jsx global>{`
        .glass-modal {
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03));
          backdrop-filter: blur(16px);
          -webkit-backdrop-filter: blur(16px);
          border: 1px solid rgba(255, 255, 255, 0.15);
          box-shadow: 0 8px 28px rgba(0, 0, 0, 0.35);
        }
        .gradient-btn {
          background: linear-gradient(135deg, #7C3AED, #B86BFF);
          color: white;
          font-weight: 500;
          transition: transform 0.2s ease-in-out, opacity 0.2s ease-in-out;
        }
        .gradient-btn:hover {
          transform: scale(1.05);
          opacity: 0.95;
        }
        .animate-fade-in {
          animation: fadeIn 0.4s ease-in-out;
        }
        .animate-slide-up {
          animation: slideUp 0.4s ease-in-out;
        }
        .animate-pop-in {
          animation: popIn 0.3s ease-out;
        }
        .animate-bounce-slow {
          animation: bounce 2s infinite;
        }
        @keyframes fadeIn {
          from { opacity: 0; transform: translateY(10px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes slideUp {
          from { opacity: 0; transform: translateY(20px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes popIn {
          0% { transform: scale(0.8); opacity: 0; }
          100% { transform: scale(1); opacity: 1; }
        }
      `}</style>
    </div>
  )
}