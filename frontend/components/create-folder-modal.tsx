"use client"

import { useState } from "react"
import { X, Folder, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAuth } from "@/lib/auth-context"

interface CreateFolderModalProps {
  isOpen: boolean
  onClose: () => void
  onFolderCreated?: (folder: any) => void
}

export function CreateFolderModal({ isOpen, onClose, onFolderCreated }: CreateFolderModalProps) {
  const [folderName, setFolderName] = useState("")
  const [isCreating, setIsCreating] = useState(false)
  const { token } = useAuth()

  const gradientStart = "#7C3AED"
  const gradientEnd = "#B86BFF"

  const handleCreate = async () => {
    if (!folderName.trim() || !token) return
    setIsCreating(true)

    try {
      const res = await fetch("http://localhost:8080/folders", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: folderName.trim() }),
      })

      if (!res.ok) throw new Error("Failed to create folder")
      const data = await res.json()

      onFolderCreated?.(data)
      window.dispatchEvent(new Event("folderCreated"))

      setFolderName("")
      onClose()
    } catch (err) {
      console.error("Folder creation failed:", err)
    } finally {
      setIsCreating(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="glass-modal w-full max-w-md rounded-3xl border border-white/20 shadow-2xl overflow-hidden">
        <div className="flex items-center justify-between p-6 border-b border-white/10">
          <div>
            <h2
  className="text-xl font-bold inline-block bg-clip-text text-transparent"
  style={{
    background: `linear-gradient(135deg, #7C3AED, #B86BFF)`,
    WebkitBackgroundClip: "text",
  }}
>
  Create New Folder
</h2>
            <p className="text-muted-foreground text-sm mt-1">Enter a name for your new folder</p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="hover:bg-white/10">
            <X className="w-4 h-4" />
          </Button>
        </div>

        <div className="p-6">
          <div className="flex items-center gap-4 mb-6">
            <div
              className="w-16 h-16 rounded-2xl flex items-center justify-center shadow-md"
              style={{ background: `linear-gradient(135deg, ${gradientStart}, ${gradientEnd})` }}
            >
              <Folder className="w-8 h-8 text-white" />
            </div>
            <Input
              type="text"
              placeholder="Folder name"
              value={folderName}
              onChange={(e) => setFolderName(e.target.value)}
              className="text-lg font-medium glass-input"
              autoFocus
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="outline" onClick={onClose} disabled={isCreating}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!folderName.trim() || isCreating}
              className="gradient-btn"
            >
              {isCreating ? (
                <>
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="w-4 h-4 mr-2" />
                  Create Folder
                </>
              )}
            </Button>
          </div>
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
        .glass-input {
          background: rgba(255, 255, 255, 0.15);
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
          border: 1px solid rgba(255, 255, 255, 0.2);
        }
        .gradient-btn {
          background: linear-gradient(135deg, ${gradientStart}, ${gradientEnd});
          color: white;
          font-weight: 500;
          transition: transform 0.2s ease-in-out, opacity 0.2s ease-in-out;
        }
        .gradient-btn:hover {
          transform: scale(1.05);
          opacity: 0.95;
        }
      `}</style>
    </div>
  )
}