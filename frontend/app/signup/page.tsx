"use client"

import type React from "react"
import { useState } from "react"
import { useAuth } from "@/lib/auth-context"
import { Button } from "@/components/ui/button"
import Link from "next/link"
import { Eye, EyeOff, UserPlus } from "lucide-react"
import { useRouter } from "next/navigation"
import { ThemeToggle } from "@/components/theme-toggle"

export default function SignupPage() {
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const { signup } = useAuth()
  const router = useRouter()

  const validateForm = () => {
    const newErrors: Record<string, string> = {}
    if (username.length < 3) newErrors.username = "Username must be at least 3 characters"
    if (!/\S+@\S+\.\S+/.test(email)) newErrors.email = "Please enter a valid email address"
    if (password.length < 6) newErrors.password = "Password must be at least 6 characters"
    if (password !== confirmPassword) newErrors.confirmPassword = "Passwords do not match"
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateForm()) return
    setIsLoading(true)
    const success = await signup(username, email, password)
    if (success) {
      router.push("/login?message=Account created successfully")
    } else {
      setErrors({ general: "Failed to create account. Please try again." })
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
          borderColor: "#4C1D95",
          boxShadow:
            "0 0 40px rgba(76,29,149,0.35), 0 0 80px rgba(76,29,149,0.25), 0 0 120px rgba(76,29,149,0.2)",
        }}
      >
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <div
              className="w-16 h-16 rounded-2xl mx-auto mb-4 flex items-center justify-center shadow-lg"
              style={{ background: "linear-gradient(to bottom right, #4C1D95, #7C3AED)" }}
            >
              <UserPlus className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
              Create Account
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Join Balkan Storage and start managing your files
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <label htmlFor="username" className="text-purple-700 dark:text-purple-300 text-lg font-semibold">
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
                placeholder="Choose a username"
              />
              {errors.username && <p className="text-red-400 text-sm">{errors.username}</p>}
            </div>

            <div className="space-y-2">
              <label htmlFor="email" className="text-purple-700 dark:text-purple-300 text-lg font-semibold">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="block w-full h-12 rounded-md border-2 bg-transparent px-3 outline-none transition-all placeholder-gray-400 text-purple-700 dark:text-white"
                style={{ borderColor: "#7C3AED" }}
                onFocus={(e) => (e.currentTarget.style.borderColor = "#4C1D95")}
                onBlur={(e) => (e.currentTarget.style.borderColor = "#7C3AED")}
                placeholder="Enter your email"
              />
              {errors.email && <p className="text-red-400 text-sm">{errors.email}</p>}
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-purple-700 dark:text-purple-300 text-lg font-semibold">
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
                  placeholder="Create a password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                  style={{ color: "#7C3AED" }}
                >
                  {showPassword ? <EyeOff className="w-6 h-6" /> : <Eye className="w-6 h-6" />}
                </button>
              </div>
              {errors.password && <p className="text-red-400 text-sm">{errors.password}</p>}
            </div>

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-purple-700 dark:text-purple-300 text-lg font-semibold">
                Confirm Password
              </label>
              <div className="relative">
                <input
                  id="confirmPassword"
                  type={showConfirmPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                  className="block w-full h-12 rounded-md border-2 bg-transparent px-3 outline-none transition-all pr-10 placeholder-gray-400 text-purple-700 dark:text-white"
                  style={{ borderColor: "#7C3AED" }}
                  onFocus={(e) => (e.currentTarget.style.borderColor = "#4C1D95")}
                  onBlur={(e) => (e.currentTarget.style.borderColor = "#7C3AED")}
                  placeholder="Confirm your password"
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                  style={{ color: "#7C3AED" }}
                >
                  {showConfirmPassword ? <EyeOff className="w-6 h-6" /> : <Eye className="w-6 h-6" />}
                </button>
              </div>
              {errors.confirmPassword && <p className="text-red-400 text-sm">{errors.confirmPassword}</p>}
            </div>

            {errors.general && (
              <div className="text-red-400 text-sm text-center bg-red-500/10 p-3 rounded-lg border border-red-500/20">
                {errors.general}
              </div>
            )}

            <Button
              type="submit"
              disabled={isLoading}
              className="w-full h-12 font-semibold rounded-lg text-lg transition-all duration-200 hover:scale-[1.02] text-white"
              style={{ background: "linear-gradient(to right, #4C1D95, #7C3AED)" }}
            >
              {isLoading ? (
                <div className="flex items-center gap-2">
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                  Creating account...
                </div>
              ) : (
                "Create Account"
              )}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-gray-600 dark:text-gray-400">
              Already have an account?{" "}
              <Link href="/login" className="font-medium" style={{ color: "#4C1D95" }}>
                Sign in
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}