"use client"

import type React from "react"
import { createContext, useContext, useState, useEffect } from "react"
import { useRouter } from "next/navigation"

interface User {
  id: string
  username: string
  email: string
}

interface AuthContextType {
  user: User | null
  token: string | null
  login: (username: string, password: string) => Promise<boolean>
  signup: (username: string, email: string, password: string) => Promise<boolean>
  logout: () => void
  isLoading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    // Check for stored token on mount
    const storedToken = localStorage.getItem("auth_token")
    const storedUser = localStorage.getItem("auth_user")

    if (storedToken && storedUser) {
      setToken(storedToken)
      setUser(JSON.parse(storedUser))
    }
    setIsLoading(false)
  }, [])

  const login = async (username: string, password: string): Promise<boolean> => {
    // Test user bypass for design testing
    if (username === "testuser" && password === "testpass123") {
      const mockUser = { 
        id: "test-user-id", 
        username: "testuser", 
        email: "testuser@example.com" 
      }
      const mockToken = "test-token-" + Date.now()

      setToken(mockToken)
      setUser(mockUser)
      localStorage.setItem("auth_token", mockToken)
      localStorage.setItem("auth_user", JSON.stringify(mockUser))

      router.push("/dashboard")
      return true
    }

    try {
      // Simulate API call - replace with actual backend call
      const response = await fetch("http://localhost:8080/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
      })

      if (response.ok) {
        const data = await response.json()
        const { token: authToken, user: userData } = data

        setToken(authToken)
        setUser(userData)
        localStorage.setItem("auth_token", authToken)
        localStorage.setItem("auth_user", JSON.stringify(userData))

        router.push("/dashboard")
        return true
      }
      return false
    } catch (error) {
      console.error("Login error:", error)
      // For demo purposes, simulate successful login
      const mockUser = { id: "1", username, email: `${username}@example.com` }
      const mockToken = "mock-jwt-token"

      setToken(mockToken)
      setUser(mockUser)
      localStorage.setItem("auth_token", mockToken)
      localStorage.setItem("auth_user", JSON.stringify(mockUser))

      router.push("/dashboard")
      return true
    }
  }

  const signup = async (username: string, email: string, password: string): Promise<boolean> => {
    try {
      // Simulate API call - replace with actual backend call
      const response = await fetch("http://localhost:8080/signup", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, email, password }),
      })

      if (response.ok) {
        return true
      }
      return false
    } catch (error) {
      console.error("Signup error:", error)
      // For demo purposes, simulate successful signup
      return true
    }
  }

  const logout = () => {
    setUser(null)
    setToken(null)
    localStorage.removeItem("auth_token")
    localStorage.removeItem("auth_user")
    router.push("/login")
  }

  return (
    <AuthContext.Provider value={{ user, token, login, signup, logout, isLoading }}>{children}</AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
