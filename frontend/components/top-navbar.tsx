"use client"

import { useState } from "react"
import { Search, Upload, Folder, Bell, User, Sun, Moon, LogOut } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAuth } from "@/lib/auth-context"
import { useTheme } from "@/lib/theme-context"

interface TopNavbarProps {
  onUpload?: () => void
  onCreateFolder?: () => void
  onSearchResults?: (results: any) => void
}

export function TopNavbar({ onUpload, onCreateFolder, onSearchResults }: TopNavbarProps) {
  const { user, logout, token } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const [searchQuery, setSearchQuery] = useState("")
  const [menuOpen, setMenuOpen] = useState(false)

  const gradientStart = "#7C3AED"
  const gradientEnd = "#B86BFF"

  const handleSearch = async (query: string) => {
    setSearchQuery(query)
    if (!query || !token) return
    try {
      const res = await fetch(`http://localhost:8080/files/search?name=${encodeURIComponent(query)}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await res.json()
      if (onSearchResults) onSearchResults(data)
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <div className="sticky top-0 z-30 px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-white/6 dark:bg-gray-900/20 backdrop-blur-lg">
      <div className="flex items-center justify-between max-w-7xl mx-auto gap-8">
        <div className="flex flex-col items-start">
          <h1 className="text-2xl font-extrabold tracking-tight">Balkan Storage</h1>
          <span
            style={{
              marginTop: 4,
              fontSize: 13,
              fontWeight: 600,
              background: `linear-gradient(120deg, ${gradientStart}, ${gradientEnd})`,
              WebkitBackgroundClip: "text",
              backgroundClip: "text",
              color: "transparent",
            }}
          >
            Welcome back, {user?.username}
          </span>
        </div>

        <div className="flex-1 max-w-2xl mx-auto">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-300" />
            <Input
              placeholder="Search files and folders..."
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
              className="pl-10 glass rounded-lg py-3"
            />
          </div>
        </div>

        <div className="flex items-center gap-4">
          <Button onClick={onUpload} className="glass flex items-center gap-2 px-4 py-2 font-medium hover:scale-103">
            <Upload className="w-4 h-4 text-purple-600 dark:text-purple-400" />
            <span className="text-purple-600 dark:text-purple-400">Upload</span>
          </Button>

          <Button onClick={onCreateFolder} className="glass flex items-center gap-2 px-4 py-2 font-medium hover:scale-103">
            <Folder className="w-4 h-4 text-purple-600 dark:text-purple-400" />
            <span className="text-purple-600 dark:text-purple-400">New Folder</span>
          </Button>

          <Button variant="ghost" size="icon" onClick={toggleTheme} className="glass w-10 h-10 hover:scale-110">
            {theme === "dark" ? <Sun className="w-4 h-4 text-white" /> : <Moon className="w-4 h-4 text-gray-800" />}
          </Button>

          <Button variant="ghost" size="icon" className="glass w-10 h-10 hover:scale-110">
            <Bell className="w-4 h-4 text-gray-500 dark:text-gray-200" />
          </Button>

          <div className="relative">
            <Button
              onClick={() => setMenuOpen(!menuOpen)}
              className="glass flex items-center gap-3 px-3 py-2 hover:scale-105"
            >
              <div
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: 6,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontWeight: 700,
                  color: "#fff",
                  background: `linear-gradient(135deg, ${gradientStart}, ${gradientEnd})`,
                }}
              >
                {user?.username?.charAt(0).toUpperCase() || "U"}
              </div>
              <span
                style={{
                  fontWeight: 600,
                  background: `linear-gradient(120deg, ${gradientStart}, ${gradientEnd})`,
                  WebkitBackgroundClip: "text",
                  backgroundClip: "text",
                  color: "transparent",
                  fontSize: 15,
                }}
              >
                {user?.username}
              </span>
            </Button>

            {menuOpen && (
              <div className="absolute right-0 mt-2 w-52 frosted-glass rounded-md shadow-lg border border-white/10 p-2">
                <button className="w-full flex items-center gap-2 px-3 py-2 text-sm font-medium text-white dark:text-gray-200 hover:bg-white/10 dark:hover:bg-gray-700/40 rounded-md">
                  <User className="w-4 h-4 text-white" />
                  Profile Settings
                </button>
                <hr className="my-2 border-white/10" />
                <button
                  onClick={logout}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm font-medium text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 rounded-md"
                >
                  <LogOut className="w-4 h-4" />
                  Sign out
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <style jsx global>{`
        .glass {
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(255, 255, 255, 0.18);
          box-shadow: 0 8px 24px 0 rgba(0, 0, 0, 0.35);
          transition: all 0.2s ease-in-out;
        }
        .frosted-glass {
          background: rgba(40, 40, 40, 0.65);
          backdrop-filter: blur(18px) saturate(140%);
          -webkit-backdrop-filter: blur(18px) saturate(140%);
          border: 1px solid rgba(255, 255, 255, 0.12);
        }
        .hover\\:scale-103:hover {
          transform: scale(1.03);
        }
      `}</style>
    </div>
  )
}