import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { Organization, CreateOrganizationRequest } from '@/types/api'
import { apiClient } from '@/lib/api'
import { 
  Button, 
  Input, 
  Modal, 
  Loading, 
  Alert 
} from '@/components/ui'
import { OrganizationCard } from '@/components/organizations'
import { 
  PlusIcon,
  BuildingOfficeIcon 
} from '@heroicons/react/24/outline'

export const OrganizationsPage = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [createForm, setCreateForm] = useState<CreateOrganizationRequest>({
    name: '',
    slug: '',
    description: ''
  })
  const [createError, setCreateError] = useState('')

  // Load organizations
  const { 
    data: organizations = [], 
    isLoading, 
    error 
  } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => apiClient.getOrganizations()
  })

  // Create organization mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateOrganizationRequest) => apiClient.createOrganization(data),
    onSuccess: (newOrg) => {
      // Update cache
      queryClient.setQueryData(['organizations'], (old: Organization[] = []) => [...old, newOrg])
      
      // Close modal and reset form
      setShowCreateModal(false)
      setCreateForm({ name: '', slug: '', description: '' })
      setCreateError('')
      
      // Navigate to the new organization
      navigate({ to: `/organizations/${newOrg.slug}` })
    },
    onError: (error: any) => {
      setCreateError(error.response?.data?.message || 'Failed to create organization')
    }
  })

  const handleCreateSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setCreateError('')
    
    // Basic validation
    if (!createForm.name.trim()) {
      setCreateError('Organization name is required')
      return
    }
    
    if (!createForm.slug.trim()) {
      setCreateError('Organization slug is required')
      return
    }
    
    createMutation.mutate(createForm)
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target
    setCreateForm(prev => ({ ...prev, [name]: value }))
    
    // Auto-generate slug from name if slug is empty
    if (name === 'name' && !createForm.slug) {
      const slug = value.toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim()
      setCreateForm(prev => ({ ...prev, slug }))
    }
  }

  const handleOrganizationClick = (organization: Organization) => {
    navigate({ to: `/organizations/${organization.slug}` })
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Organizations</h1>
          <p className="mt-1 text-sm text-gray-500">Loading...</p>
        </div>
        <Loading />
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Organizations</h1>
          <p className="mt-1 text-sm text-red-600">Failed to load organizations</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Organizations</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage your organizations and team access.
          </p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          <PlusIcon className="h-4 w-4 mr-2" />
          Create Organization
        </Button>
      </div>

      {/* Organizations grid */}
      {organizations.length === 0 ? (
        <div className="text-center py-12">
          <BuildingOfficeIcon className="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No organizations found</h3>
          <p className="text-gray-600 mb-6">Create your first organization to get started.</p>
          <Button onClick={() => setShowCreateModal(true)}>
            <PlusIcon className="h-4 w-4 mr-2" />
            Create Organization
          </Button>
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {organizations.map((organization) => (
            <OrganizationCard
              key={organization.id}
              organization={organization}
              onClick={() => handleOrganizationClick(organization)}
            />
          ))}
        </div>
      )}

      {/* Create organization modal */}
      <Modal
        open={showCreateModal}
        onClose={() => {
          setShowCreateModal(false)
          setCreateForm({ name: '', slug: '', description: '' })
          setCreateError('')
        }}
        title="Create Organization"
      >
        <form onSubmit={handleCreateSubmit} className="space-y-4">
          {createError && (
            <Alert variant="error" description={createError} />
          )}
          
          <Input
            label="Organization Name"
            type="text"
            name="name"
            required
            value={createForm.name}
            onChange={handleInputChange}
            placeholder="Enter organization name"
          />
          
          <Input
            label="Organization Slug"
            type="text"
            name="slug"
            required
            value={createForm.slug}
            onChange={handleInputChange}
            placeholder="organization-slug"
            helperText="This will be used in URLs and must be unique"
          />
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description (Optional)
            </label>
            <textarea
              name="description"
              rows={3}
              value={createForm.description}
              onChange={handleInputChange}
              placeholder="Describe your organization..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setShowCreateModal(false)
                setCreateForm({ name: '', slug: '', description: '' })
                setCreateError('')
              }}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={createMutation.isPending}
              disabled={!createForm.name.trim() || !createForm.slug.trim()}
            >
              Create Organization
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}

export default OrganizationsPage