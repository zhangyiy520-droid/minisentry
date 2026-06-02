import { useState } from 'react'
import { Link, useNavigate, useSearch } from '@tanstack/react-router'
import { useAuth } from '@/lib/auth'
import { Button, Input, Card, Alert } from '@/components/ui'
import { LoginRequest } from '@/types/api'

export const LoginPage = () => {
  const navigate = useNavigate()
  const search = useSearch({ strict: false }) as { return?: string }
  const { login, loading, error, clearError } = useAuth()
  
  const [formData, setFormData] = useState<LoginRequest>({
    email: '',
    password: ''
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()
    
    try {
      await login(formData)
      const returnUrl = (search as any)?.return || '/'
      navigate({ to: returnUrl } as any)
    } catch (err) {
      // Error is handled by the auth context
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold text-gray-900">
            Sign in to MiniSentry
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{' '}
            <Link
              to="/register"
              className="font-medium text-primary-600 hover:text-primary-500"
            >
              create a new account
            </Link>
          </p>
        </div>
        
        <Card className="max-w-md mx-auto">
          <form className="space-y-6" onSubmit={handleSubmit}>
            {error && (
              <Alert variant="error" description={error} onClose={clearError} />
            )}
            
            <Input
              label="Email address"
              type="email"
              name="email"
              autoComplete="email"
              required
              value={formData.email}
              onChange={handleInputChange}
              placeholder="Enter your email"
            />
            
            <Input
              label="Password"
              type="password"
              name="password"
              autoComplete="current-password"
              required
              value={formData.password}
              onChange={handleInputChange}
              placeholder="Enter your password"
            />

            <Button
              type="submit"
              className="w-full"
              loading={loading}
              disabled={!formData.email || !formData.password}
            >
              Sign in
            </Button>
          </form>
        </Card>
      </div>
    </div>
  )
}

export default LoginPage