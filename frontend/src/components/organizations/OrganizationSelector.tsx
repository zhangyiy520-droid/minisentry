import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Organization } from '@/types/api'
import { Select, Loading } from '@/components/ui'
import { apiClient } from '@/lib/api'

interface OrganizationSelectorProps {
  selectedOrgSlug?: string
  onOrganizationChange: (organization: Organization | null) => void
  placeholder?: string
  className?: string
}

export const OrganizationSelector = ({
  selectedOrgSlug,
  onOrganizationChange,
  placeholder = "Select an organization",
  className
}: OrganizationSelectorProps) => {
  const {
    data: organizations = [],
    isLoading,
    error
  } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => apiClient.getOrganizations()
  })

  const selectedOrganization = organizations.find(org => org.slug === selectedOrgSlug)

  useEffect(() => {
    if (selectedOrgSlug && selectedOrganization) {
      onOrganizationChange(selectedOrganization)
    }
  }, [selectedOrganization, selectedOrgSlug, onOrganizationChange])

  const handleChange = (orgSlug: string) => {
    if (orgSlug === '') {
      onOrganizationChange(null)
    } else {
      const organization = organizations.find(org => org.slug === orgSlug)
      onOrganizationChange(organization || null)
    }
  }

  if (isLoading) {
    return (
      <div className={`flex items-center space-x-2 ${className}`}>
        <Loading />
        <span className="text-sm text-gray-500">Loading organizations...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`text-sm text-red-600 ${className}`}>
        Error loading organizations
      </div>
    )
  }

  return (
    <Select
      value={selectedOrgSlug || ''}
      onChange={(e) => handleChange(e.target.value)}
      className={className}
    >
      <option value="">{placeholder}</option>
      {organizations.map((org) => (
        <option key={org.id} value={org.slug}>
          {org.name}
        </option>
      ))}
    </Select>
  )
}

export default OrganizationSelector