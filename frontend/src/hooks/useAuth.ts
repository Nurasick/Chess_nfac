import { useState } from 'react'
import { login as apiLogin, logout as apiLogout, register as apiRegister } from '../lib/apiClient'
import type { User } from '../types/api'

function loadPersistedUser(): User | null {
  const token = localStorage.getItem('access_token')
  const raw = localStorage.getItem('user')
  if (!token || !raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(loadPersistedUser)
  const [isLoading, setIsLoading] = useState(false)

  async function login(email: string, password: string) {
    setIsLoading(true)
    try {
      const res = await apiLogin(email, password)
      const { user: loggedInUser, access_token, refresh_token } = res.data
      localStorage.setItem('access_token', access_token)
      localStorage.setItem('refresh_token', refresh_token)
      localStorage.setItem('user', JSON.stringify(loggedInUser))
      setUser(loggedInUser)
      return loggedInUser
    } finally {
      setIsLoading(false)
    }
  }

  async function register(username: string, email: string, password: string, city: string) {
    setIsLoading(true)
    try {
      const res = await apiRegister(username, email, password, city)
      const { user: registeredUser, access_token, refresh_token } = res.data
      localStorage.setItem('access_token', access_token)
      localStorage.setItem('refresh_token', refresh_token)
      localStorage.setItem('user', JSON.stringify(registeredUser))
      setUser(registeredUser)
      return registeredUser
    } finally {
      setIsLoading(false)
    }
  }

  async function logout() {
    setIsLoading(true)
    try {
      await apiLogout()
    } finally {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('user')
      setUser(null)
      setIsLoading(false)
    }
  }

  return { user, isLoading, login, register, logout }
}
