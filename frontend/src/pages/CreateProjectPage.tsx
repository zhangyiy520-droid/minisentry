import { useState } from 'react'
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query'
import { useNavigate, useParams } from '@tanstack/react-router'
import { CreateProjectRequest, ProjectPlatform } from '@/types/api'
import { apiClient } from '@/lib/api'
import { 
  Button, 
  Input, 
  Select,
  Card,
  Alert 
} from '@/components/ui'
import { 
  ArrowLeftIcon,
  FolderIcon 
} from '@heroicons/react/24/outline'

const PLATFORMS: { value: ProjectPlatform; label: string }[] = [
  { value: 'javascript', label: 'JavaScript' },
  { value: 'python', label: 'Python' },
  { value: 'go', label: 'Go' },
  { value: 'java', label: 'Java' },
  { value: 'dotnet', label: '.NET' },
  { value: 'php', label: 'PHP' },
  { value: 'ruby', label: 'Ruby' },
]

export const CreateProjectPage = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  
  // Extract orgSlug from URL path
  const orgSlug = window.location.pathname.split('/')[2]
  
  // Load organizations to get the organization ID
  const { data: organizations = [] } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => apiClient.getOrganizations()
  })
  
  const organization = organizations.find(org => org.slug === orgSlug)
  
  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    slug: '',
    platform: 'javascript',
    description: ''
  })
  const [error, setError] = useState('')

  // Create project mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateProjectRequest) => {
      if (!organization?.id) {
        throw new Error('Organization not found')
      }
      return apiClient.createProject(organization.id, data)
    },
    onSuccess: (newProject) => {
      // Update cache
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      
      // Navigate to the dashboard
      navigate({ to: '/' })
    },
    onError: (error: any) => {
      setError(error.response?.data?.message || 'Failed to create project')
    }
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    
    // Basic validation
    if (!formData.name.trim()) {
      setError('Project name is required')
      return
    }
    
    if (!formData.slug.trim()) {
      setError('Project slug is required')
      return
    }
    
    if (!formData.platform) {
      setError('Platform selection is required')
      return
    }
    
    createMutation.mutate(formData)
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
    
    // Auto-generate slug from name if slug is empty
    if (name === 'name' && !formData.slug) {
      const slug = value.toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim()
      setFormData(prev => ({ ...prev, slug }))
    }
  }

  const handleGoBack = () => {
    navigate({ to: '/' })
  }

  const isFormValid = formData.name.trim() && formData.slug.trim() && formData.platform

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <div className="flex items-center space-x-4 mb-4">
          <Button
            variant="outline"
            size="sm"
            onClick={handleGoBack}
          >
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back to Dashboard
          </Button>
        </div>
        <div className="flex items-center space-x-3">
          <div className="p-2 bg-blue-100 rounded-lg">
            <FolderIcon className="h-6 w-6 text-blue-600" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Create New Project</h1>
            <p className="mt-1 text-sm text-gray-500">
              Add a new project to start monitoring errors and issues.
            </p>
          </div>
        </div>
      </div>

      {/* Create project form */}
      <Card className="max-w-2xl">
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <Alert variant="error" description={error} />
          )}
          
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            <Input
              label="Project Name"
              type="text"
              name="name"
              required
              value={formData.name}
              onChange={handleInputChange}
              placeholder="My Awesome Project"
            />
            
            <Input
              label="Project Slug"
              type="text"
              name="slug"
              required
              value={formData.slug}
              onChange={handleInputChange}
              placeholder="my-awesome-project"
              helperText="Used in URLs and must be unique within the organization"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Platform *
            </label>
            <Select
              name="platform"
              value={formData.platform}
              onChange={handleInputChange}
              className="w-full"
            >
              {PLATFORMS.map((platform) => (
                <option key={platform.value} value={platform.value}>
                  {platform.label}
                </option>
              ))}
            </Select>
            <p className="mt-1 text-xs text-gray-500">
              Select the primary platform/language for this project
            </p>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description (Optional)
            </label>
            <textarea
              name="description"
              rows={3}
              value={formData.description}
              onChange={handleInputChange}
              placeholder="Describe your project..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
            <Button
              type="button"
              variant="outline"
              onClick={handleGoBack}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={createMutation.isPending}
              disabled={!isFormValid}
            >
              Create Project
            </Button>
          </div>
        </form>
      </Card>
    </div>
  )
}

export default CreateProjectPage