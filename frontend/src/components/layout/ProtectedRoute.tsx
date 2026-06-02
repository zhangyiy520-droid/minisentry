import { ReactNode } from 'react'
import { useLocation, useNavigate } from '@tanstack/react-router'
import { useAuth } from '@/lib/auth'
import { Loading } from '@/components/ui'
import { useEffect } from 'react'

interface ProtectedRouteProps {
  children: ReactNode
  requireAuth?: boolean
  redirectTo?: string
}

export const ProtectedRoute = ({ 
  children, 
  requireAuth = true, 
  redirectTo = '/login' 
}: ProtectedRouteProps) => {
  const { user, loading } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    if (loading) return

    if (requireAuth && !user) {
      // Redirect to login with return URL
      const returnUrl = location.pathname !== '/' ? location.pathname : undefined
      const loginUrl = returnUrl ? `${redirectTo}?return=${encodeURIComponent(returnUrl)}` : redirectTo
      navigate({ to: loginUrl, replace: true })
      return
    }

    if (!requireAuth && user) {
      // User is logged in but accessing a guest-only route (like login/register)
      navigate({ to: '/', replace: true })
      return
    }
  }, [user, loading, requireAuth, location.pathname, navigate, redirectTo])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading size="lg" text="Loading..." />
      </div>
    )
  }

  if (requireAuth && !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading size="lg" text="Redirecting..." />
      </div>
    )
  }

  if (!requireAuth && user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading size="lg" text="Redirecting..." />
      </div>
    )
  }

  return <>{children}</>
}

export default ProtectedRoute