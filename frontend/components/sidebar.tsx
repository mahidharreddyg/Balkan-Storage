"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import {
  Menu,
  X,
  Home,
  Clock,
  Trash2,
  Star,
  Folder,
} from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { useAuth } from "../lib/auth-context"

interface FolderType {
  id: number
  name: string
}

const staticItems = [
  { name: "Home", icon: Home, href: "/dashboard" },
  { name: "Recent", icon: Clock, href: "/recent" },
  { name: "Favorites", icon: Star, href: "/favorites" },
  { name: "Trash", icon: Trash2, href: "/trash" },
]

export function Sidebar() {
  const [isOpen, setIsOpen] = useState(true)
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [folders, setFolders] = useState<FolderType[]>([])
  const [stats, setStats] = useState<{ used: number; quota: number } | null>(null)
  const pathname = usePathname()
  const { token } = useAuth()

  const COLLAPSED_WIDTH = 72
  const EXPANDED_WIDTH = 260
  const ICON_COL = COLLAPSED_WIDTH

  useEffect(() => {
  if (!token) return
  const fetchData = async () => {
    try {
      const [foldersRes, statsRes] = await Promise.all([
        fetch("http://localhost:8080/folders", {
          headers: { Authorization: `Bearer ${token}` },
        }).then(r => r.json()),
        fetch("http://localhost:8080/stats", {
          headers: { Authorization: `Bearer ${token}` },
        }).then(r => r.json()),
      ])
      setFolders(foldersRes?.folders || [])
      if (statsRes?.used !== undefined && statsRes?.quota !== undefined) {
        setStats({ used: statsRes.used, quota: statsRes.quota })
      }
    } catch (err) {
      console.error("Sidebar fetch error:", err)
    }
  }

  fetchData()

  const handleFolderCreated = () => fetchData()
  window.addEventListener("folderCreated", handleFolderCreated)

  return () => {
    window.removeEventListener("folderCreated", handleFolderCreated)
  }
}, [token])

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        aria-label={isOpen ? "Close sidebar" : "Open sidebar"}
        onClick={() => setIsOpen(!isOpen)}
        className="fixed top-4 left-4 z-50 md:hidden w-12 h-12 rounded-xl glass-bg flex justify-center items-center text-gray-700 dark:text-white"
      >
        {isOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
      </Button>

      <aside
        style={{
          width: isCollapsed ? COLLAPSED_WIDTH : EXPANDED_WIDTH,
          transition: "width 360ms cubic-bezier(0.2,0.9,0.18,1)",
        }}
        className={`
          relative fixed inset-y-0 left-0 z-40 flex flex-col
          glass-bg border border-gray-300 dark:border-white/20
          rounded-tr-lg rounded-br-lg
          md:static md:translate-x-0
          ${isOpen ? "translate-x-0" : "-translate-x-full"}
          overflow-hidden
        `}
      >
        <div className="w-full border-b border-gray-300 dark:border-white/20 z-10" />

        <nav className="relative z-20 flex flex-col flex-1 py-6 gap-2 select-none">
          {staticItems.map((item, index) => {
            const isActive = pathname === item.href
            return (
              <div key={item.name}>
                {index === 0 && (
                  <div className="w-full border-t border-gray-300 dark:border-white/20 my-6" />
                )}
                <Link
                  href={item.href}
                  className={`
                    relative flex items-center h-12 transition-colors duration-220 ease-in-out
                    ${
                      isActive && !isCollapsed
                        ? "bg-blue-800/70 text-white"
                        : isActive && isCollapsed
                        ? "text-blue-600"
                        : "text-gray-700 dark:text-white/70 hover:bg-gray-200/60 dark:hover:bg-white/10 hover:text-gray-900 dark:hover:text-white"
                    }
                  `}
                >
                  <div
                    className="icon-col flex items-center justify-center flex-shrink-0 relative"
                    style={{ width: ICON_COL }}
                  >
                    <span
                      className={`
                        active-square absolute transition-transform duration-260 ease-out
                        ${isActive && isCollapsed ? "scale-100 opacity-100" : "scale-0 opacity-0"}
                      `}
                      aria-hidden="true"
                    />
                    <item.icon
                      size={22}
                      className={`relative z-10 ${
                        isActive ? "text-white" : ""
                      }`}
                    />
                  </div>
                  {!isCollapsed && (
                    <span
                      className={`ml-2 text-sm font-medium ${
                        isActive ? "text-white" : ""
                      }`}
                    >
                      {item.name}
                    </span>
                  )}
                  {isActive && !isCollapsed && (
                    <span
                      className="active-bar absolute left-0 top-1/2 -translate-y-1/2"
                      aria-hidden="true"
                      style={{
                        width: 6,
                        height: 40,
                        borderRadius: "0 8px 8px 0",
                        background: "rgba(59,130,246,0.95)",
                      }}
                    />
                  )}
                </Link>
              </div>
            )
          })}

          {folders.length > 0 && (
            <>
              {!isCollapsed && (
                <div className="mt-6 mb-2 px-4 text-xs uppercase tracking-wide text-gray-500 dark:text-white/60">
                  Folders
                </div>
              )}
              {folders.map(folder => {
                const isActive = pathname === `/folders/${folder.id}`
                return (
                  <Link
                    key={folder.id}
                    href={`/folders/${folder.id}`}
                    className={`
                      relative flex items-center h-12 transition-colors duration-220 ease-in-out
                      ${
                        isActive && !isCollapsed
                          ? "bg-blue-800/70 text-white"
                          : "text-gray-700 dark:text-white/70 hover:bg-gray-200/60 dark:hover:bg-white/10 hover:text-gray-900 dark:hover:text-white"
                      }
                    `}
                  >
                    <div
                      className="icon-col flex items-center justify-center flex-shrink-0 relative"
                      style={{ width: ICON_COL }}
                    >
                      <span
                        className={`
                          active-square absolute transition-transform duration-260 ease-out
                          ${isActive && isCollapsed ? "scale-100 opacity-100" : "scale-0 opacity-0"}
                        `}
                        aria-hidden="true"
                      />
                      <Folder
                        size={20}
                        className={`relative z-10 ${
                          isActive ? "text-white" : ""
                        }`}
                      />
                    </div>
                    {!isCollapsed && (
                      <span
                        className={`ml-2 text-sm font-medium ${
                          isActive ? "text-white" : ""
                        }`}
                      >
                        {folder.name}
                      </span>
                    )}
                    {isActive && !isCollapsed && (
                      <span
                        className="active-bar absolute left-0 top-1/2 -translate-y-1/2"
                        aria-hidden="true"
                        style={{
                          width: 6,
                          height: 40,
                          borderRadius: "0 8px 8px 0",
                          background: "rgba(59,130,246,0.95)",
                        }}
                      />
                    )}
                  </Link>
                )
              })}
            </>
          )}

          {stats && !isCollapsed && (
            <div className="mt-auto p-4 text-xs text-gray-500 dark:text-white/60">
              Used: {Math.round((stats.used / stats.quota) * 100)}%
            </div>
          )}
        </nav>

        <div className="w-full border-t border-gray-300 dark:border-white/20 z-10" />

        <div className="p-4 flex justify-center z-20">
          <Button
            variant="ghost"
            size="sm"
            aria-label={isCollapsed ? "Expand sidebar" : "Collapse sidebar"}
            onClick={() => setIsCollapsed(!isCollapsed)}
            className="w-10 h-10 rounded-lg glass-bg hover:bg-gray-200/60 dark:hover:bg-white/20 text-gray-700 dark:text-white flex justify-center items-center"
          >
            {isCollapsed ? <Menu className="w-5 h-5" /> : <X className="w-5 h-5" />}
          </Button>
        </div>
      </aside>

      {isOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/60 md:hidden backdrop-blur-sm"
          aria-hidden="true"
          onClick={() => setIsOpen(false)}
        />
      )}

      <style jsx global>{`
        .glass-bg {
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0));
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(255, 255, 255, 0.18);
          box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.37);
        }
        .icon-col {
          min-width: 72px;
          max-width: 72px;
        }
        .active-square {
          left: 50%;
          top: 50%;
          transform: translate(-50%, -50%) scale(0);
          width: 40px;
          height: 40px;
          background: rgba(30, 64, 175, 0.75);
          border-radius: 6px;
          box-shadow: 0 6px 18px rgba(30, 64, 175, 0.3);
          z-index: 0;
          opacity: 0;
        }
        .active-square.scale-100 {
          transform: translate(-50%, -50%) scale(1);
          opacity: 1;
        }
        .active-bar {
          box-shadow: 0 8px 18px rgba(59, 130, 246, 0.15);
          z-index: 0;
        }
      `}</style>
    </>
  )
}