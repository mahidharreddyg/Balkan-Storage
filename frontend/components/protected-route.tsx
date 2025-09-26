
"use client"

import type React from "react"
import { useAuth } from "@/lib/auth-context"
import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, token, logout, isLoading } = useAuth()
  const router = useRouter()
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    const verifySession = async () => {
      if (!token) {
        router.push("/login")
        return
      }

      try {
        const res = await fetch("http://localhost:8080/verify-token", {
          headers: { Authorization: `Bearer ${token}` },
        })

        if (!res.ok) {
          logout()
          router.push("/login")
        }
      } catch (err) {
        console.error("Session verification failed:", err)
        logout()
        router.push("/login")
      } finally {
        setChecking(false)
      }
    }

    if (!isLoading) {
      verifySession()
    }
  }, [token, isLoading, logout, router])

  if (isLoading || checking) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="glass-card p-8 rounded-2xl">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-500"></div>
        </div>
      </div>
    )
  }

  if (!user) {
    return null
  }

  return <>{children}</>
}