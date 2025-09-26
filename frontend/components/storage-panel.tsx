"use client"

import { useEffect, useState } from "react"
import { HardDrive, FileText, ImageIcon, Video, Music, Archive, DollarSign } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/lib/auth-context"

interface StatsData {
  used: number
  quota: number
  categories: {
    name: string
    files: number
    size: number
    type: string
  }[]
}

export function StoragePanel() {
  const { token } = useAuth()
  const [stats, setStats] = useState<StatsData | null>(null)

  useEffect(() => {
    if (!token) return
    const fetchStats = async () => {
      try {
        const res = await fetch("http://localhost:8080/stats", {
          headers: { Authorization: `Bearer ${token}` },
        })
        const data = await res.json()
        setStats(data)
      } catch (err) {
        console.error("Failed to fetch stats:", err)
      }
    }
    fetchStats()
  }, [token])

  const categories = [
    { name: "Documents", icon: FileText, color: "text-blue-500" },
    { name: "Images", icon: ImageIcon, color: "text-green-500" },
    { name: "Videos", icon: Video, color: "text-purple-500" },
    { name: "Audio", icon: Music, color: "text-orange-500" },
    { name: "Archives", icon: Archive, color: "text-red-500" },
  ]

  const used = stats?.used || 0
  const total = stats?.quota || 1
  const usagePercentage = (used / total) * 100
  const circumference = 2 * Math.PI * 45
  const strokeDasharray = circumference
  const strokeDashoffset = circumference - (usagePercentage / 100) * circumference

  return (
    <div className="w-80 space-y-6 m-4 p-2">
      <div className="glass-card p-6 rounded-2xl flex flex-col items-center">
        <h3 className="text-lg font-semibold text-foreground mb-6 text-center">Storage usage</h3>
        <div className="flex items-center justify-center mb-6">
          <div className="relative w-36 h-36">
            <svg className="w-36 h-36 transform -rotate-90" viewBox="0 0 100 100">
              <circle
                cx="50"
                cy="50"
                r="45"
                stroke="currentColor"
                strokeWidth="8"
                fill="none"
                className="text-gray-300 dark:text-gray-700"
              />
              <circle
                cx="50"
                cy="50"
                r="45"
                stroke="url(#gradient)"
                strokeWidth="8"
                fill="none"
                strokeLinecap="round"
                strokeDasharray={strokeDasharray}
                strokeDashoffset={strokeDashoffset}
                className="transition-all duration-500"
              />
              <defs>
                <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stopColor="#7C3AED" />
                  <stop offset="100%" stopColor="#3B82F6" />
                </linearGradient>
              </defs>
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <div className="w-10 h-10 glass rounded-full flex items-center justify-center mb-2">
                <HardDrive className="w-5 h-5 text-purple-600 dark:text-purple-400" />
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-foreground">{used.toFixed(1)} GB</div>
                <div className="text-sm text-muted-foreground">of {total} GB</div>
              </div>
            </div>
          </div>
        </div>
        <div className="space-y-4 w-full">
          {categories.map((cat) => {
            const data = stats?.categories.find((c) => c.name === cat.name)
            if (!data) return null
            return (
              <div key={cat.name} className="flex items-center gap-3">
                <div className="w-10 h-10 glass rounded-lg flex items-center justify-center">
                  <cat.icon className={`w-5 h-5 ${cat.color}`} />
                </div>
                <div className="flex-1">
                  <div className="font-medium text-foreground">{cat.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {data.files} Files | {data.size.toFixed(1)} MB
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
      <div className="glass-card p-6 rounded-2xl text-center">
        <div className="w-12 h-12 glass rounded-full flex items-center justify-center mx-auto mb-4">
          <DollarSign className="w-6 h-6 text-purple-600 dark:text-purple-400" />
        </div>
        <h4 className="font-semibold text-foreground mb-2">Get more space for your files</h4>
        <p className="text-sm text-muted-foreground mb-4">
          Upgrade your account to pro to get more storage
        </p>
        <Button className="w-full glass bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-gray-900 dark:text-white font-medium">
          Upgrade to Pro
        </Button>
      </div>
      <style jsx global>{`
        .glass-card {
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.05));
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
          border: 1px solid rgba(255, 255, 255, 0.15);
          box-shadow: 0 6px 20px rgba(0, 0, 0, 0.3);
        }
        .glass {
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.05));
          backdrop-filter: blur(14px);
          -webkit-backdrop-filter: blur(14px);
          border: 1px solid rgba(255, 255, 255, 0.25);
        }
      `}</style>
    </div>
  )
}