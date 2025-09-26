"use client"

import type React from "react"
import { useState } from "react"
import { useAuth } from "@/lib/auth-context"
import { Button } from "@/components/ui/button"
import Link from "next/link"
import { Eye, EyeOff, LogIn } from "lucide-react"
import { ThemeToggle } from "@/components/theme-toggle"

export default function LoginPage() {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState("")
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError("")

    const success = await login(username, password)
    if (!success) {
      setError("Invalid credentials. Please try again.")
    }

    setIsLoading(false)
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative">
      <div className="absolute top-4 right-4 z-10">
        <ThemeToggle />
      </div>

      <div
        className="flex items-center justify-center p-8 fixed top-0 h-full w-[550px] left-[calc(100%-700px)] rounded-2xl backdrop-blur-lg border"
        style={{
          background: "linear-gradient(135deg, rgba(255,255,255,0.08), rgba(255,255,255,0))",
          borderColor: "#7C3AED",
          boxShadow:
            "0 0 40px rgba(124,58,237,0.35), 0 0 80px rgba(124,58,237,0.25), 0 0 120px rgba(124,58,237,0.2)",
        }}
      >
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <div
              className="w-16 h-16 rounded-2xl mx-auto mb-4 flex items-center justify-center shadow-lg"
              style={{
                background: "linear-gradient(to bottom right, #7C3AED, #4C1D95)",
              }}
            >
              <LogIn className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
              Welcome Back
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Sign in to your Balkan Storage account
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label
                htmlFor="username"
                className="text-purple-700 dark:text-purple-300 text-lg font-semibold"
              >
                Username
              </label>
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                className="block w-full h-12 rounded-md border-2 bg-transparent px-3 outline-none transition-all placeholder-gray-400 text-purple-700 dark:text-white"
                style={{ borderColor: "#7C3AED" }}
                onFocus={(e) => (e.currentTarget.style.borderColor = "#4C1D95")}
                onBlur={(e) => (e.currentTarget.style.borderColor = "#7C3AED")}
                placeholder="Enter your username"
              />
            </div>

            <div className="space-y-2">
              <label
                htmlFor="password"
                className="text-purple-700 dark:text-purple-300 text-lg font-semibold"
              >
                Password
              </label>
              <div className="relative">
                <input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="block w-full h-12 rounded-md border-2 bg-transparent px-3 outline-none transition-all pr-10 placeholder-gray-400 text-purple-700 dark:text-white"
                  style={{ borderColor: "#7C3AED" }}
                  onFocus={(e) => (e.currentTarget.style.borderColor = "#4C1D95")}
                  onBlur={(e) => (e.currentTarget.style.borderColor = "#7C3AED")}
                  placeholder="Enter your password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                  style={{ color: "#7C3AED" }}
                >
                  {showPassword ? (
                    <EyeOff className="w-6 h-6" />
                  ) : (
                    <Eye className="w-6 h-6" />
                  )}
                </button>
              </div>
            </div>

            {error && (
              <div className="text-red-400 text-sm text-center bg-red-500/10 p-3 rounded-lg border border-red-500/20">
                {error}
              </div>
            )}

            <Button
              type="submit"
              disabled={isLoading}
              className="w-full h-12 font-semibold rounded-lg text-lg transition-all duration-200 hover:scale-[1.02] text-white"
              style={{
                background: "linear-gradient(to right, #7C3AED, #4C1D95)",
              }}
            >
              {isLoading ? (
                <div className="flex items-center gap-2">
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                  Signing in...
                </div>
              ) : (
                "Sign In"
              )}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-gray-600 dark:text-gray-400">
              Don't have an account?{" "}
              <Link href="/signup" className="font-medium" style={{ color: "#4C1D95" }}>
                Sign up
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}