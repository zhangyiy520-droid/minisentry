import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Project } from '@/types/api'
import { Select, Loading } from '@/components/ui'
import { apiClient } from '@/lib/api'

interface ProjectSelectorProps {
  organizationId: string
  selectedProjectSlug?: string
  onProjectChange: (project: Project | null) => void
  placeholder?: string
  className?: string
}

export const ProjectSelector = ({
  organizationId,
  selectedProjectSlug,
  onProjectChange,
  placeholder = "Select a project",
  className
}: ProjectSelectorProps) => {
  const {
    data: projects = [],
    isLoading,
    error
  } = useQuery({
    queryKey: ['projects', organizationId],
    queryFn: () => apiClient.getProjects(organizationId),
    enabled: !!organizationId
  })

  const selectedProject = projects.find(p => p.slug === selectedProjectSlug)

  useEffect(() => {
    if (selectedProjectSlug && selectedProject) {
      onProjectChange(selectedProject)
    }
  }, [selectedProject, selectedProjectSlug, onProjectChange])

  const handleChange = (projectSlug: string) => {
    if (projectSlug === '') {
      onProjectChange(null)
    } else {
      const project = projects.find(p => p.slug === projectSlug)
      onProjectChange(project || null)
    }
  }

  if (isLoading) {
    return (
      <div className={`flex items-center space-x-2 ${className}`}>
        <Loading />
        <span className="text-sm text-gray-500">Loading projects...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`text-sm text-red-600 ${className}`}>
        Error loading projects
      </div>
    )
  }

  return (
    <Select
      value={selectedProjectSlug || ''}
      onChange={(e) => handleChange(e.target.value)}
      className={className}
    >
      <option value="">{placeholder}</option>
      {projects.map((project) => (
        <option key={project.id} value={project.slug}>
          {project.name}
        </option>
      ))}
    </Select>
  )
}

export default ProjectSelector