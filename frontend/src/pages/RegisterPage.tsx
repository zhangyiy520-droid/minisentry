import { useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { useAuth } from '@/lib/auth'
import { Button, Input, Card, Alert } from '@/components/ui'
import { RegisterRequest } from '@/types/api'

export const RegisterPage = () => {
  const navigate = useNavigate()
  const { register, loading, error, clearError } = useAuth()
  
  const [formData, setFormData] = useState<RegisterRequest>({
    email: '',
    password: '',
    name: ''
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()
    
    try {
      await register(formData)
      navigate({ to: '/' })
    } catch (err) {
      // Error is handled by the auth context
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  const isFormValid = formData.email && formData.password && formData.name && formData.password.length >= 8

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold text-gray-900">
            Create your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{' '}
            <Link
              to="/login"
              search={{ return: undefined }}
              className="font-medium text-primary-600 hover:text-primary-500"
            >
              sign in to your existing account
            </Link>
          </p>
        </div>
        
        <Card className="max-w-md mx-auto">
          <form className="space-y-6" onSubmit={handleSubmit}>
            {error && (
              <Alert variant="error" description={error} onClose={clearError} />
            )}
            
            <Input
              label="Full name"
              type="text"
              name="name"
              autoComplete="name"
              required
              value={formData.name}
              onChange={handleInputChange}
              placeholder="Enter your full name"
            />
            
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
              autoComplete="new-password"
              required
              value={formData.password}
              onChange={handleInputChange}
              placeholder="Choose a password (min. 8 characters)"
              helperText="Password must be at least 8 characters long"
            />

            <Button
              type="submit"
              className="w-full"
              loading={loading}
              disabled={!isFormValid}
            >
              Create account
            </Button>
          </form>
        </Card>
      </div>
    </div>
  )
}

export default RegisterPage