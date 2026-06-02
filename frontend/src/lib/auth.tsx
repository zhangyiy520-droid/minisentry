import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { User, LoginRequest, RegisterRequest } from '@/types/api'
import { apiClient } from './api'

interface AuthContextType {
  user: User | null
  loading: boolean
  error: string | null
  login: (credentials: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  logout: () => Promise<void>
  updateUser: (user: User) => void
  clearError: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: ReactNode
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Check if user is authenticated on app load
  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem('access_token')
      if (!token) {
        setLoading(false)
        return
      }

      try {
        const currentUser = await apiClient.getCurrentUser()
        setUser(currentUser)
      } catch (err) {
        // Token might be invalid, clear it
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        console.error('Failed to fetch current user:', err)
      } finally {
        setLoading(false)
      }
    }

    initializeAuth()
  }, [])

  const login = async (credentials: LoginRequest) => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiClient.login(credentials)
      setUser(response.user)
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || err.message || 'Login failed'
      setError(errorMessage)
      throw err
    } finally {
      setLoading(false)
    }
  }

  const register = async (data: RegisterRequest) => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiClient.register(data)
      setUser(response.user)
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || err.message || 'Registration failed'
      setError(errorMessage)
      throw err
    } finally {
      setLoading(false)
    }
  }

  const logout = async () => {
    try {
      await apiClient.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      setUser(null)
      // Clear any cached data if needed
    }
  }

  const updateUser = (updatedUser: User) => {
    setUser(updatedUser)
  }

  const clearError = () => {
    setError(null)
  }

  const value: AuthContextType = {
    user,
    loading,
    error,
    login,
    register,
    logout,
    updateUser,
    clearError
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

// Hook to check if user is authenticated
export const useIsAuthenticated = () => {
  const { user, loading } = useAuth()
  return { isAuthenticated: !!user, loading }
}

// Hook to require authentication (throws if not authenticated)
export const useRequireAuth = () => {
  const { user, loading } = useAuth()
  
  if (loading) {
    return { user: null, loading: true }
  }
  
  if (!user) {
    throw new Error('Authentication required')
  }
  
  return { user, loading: false }
}