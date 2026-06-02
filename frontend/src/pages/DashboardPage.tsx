import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useAuth } from '@/lib/auth'
import { useAppContext } from '@/lib/context'
import { apiClient } from '@/lib/api'
import { IssueFilters, Issue } from '@/types/api'
import { 
  Button, 
  Loading 
} from '@/components/ui'
import { 
  OrganizationSelector 
} from '@/components/organizations'
import { 
  ProjectSelector, 
  ProjectList 
} from '@/components/projects'
import { 
  IssueStats, 
  IssueList, 
  IssueFilters as IssueFiltersComponent 
} from '@/components/issues'
import { 
  PlusIcon,
  ExclamationTriangleIcon,
  FolderIcon,
  BuildingOfficeIcon
} from '@heroicons/react/24/outline'

export const DashboardPage = () => {
  const { user } = useAuth()
  const { selectedOrganization, selectedProject, setSelectedOrganization, setSelectedProject } = useAppContext()
  const navigate = useNavigate()
  
  const [issueFilters, setIssueFilters] = useState<IssueFilters>({
    page: 1,
    limit: 10,
    status: ['unresolved']
  })

  // Load organizations
  const { 
    data: organizations = [], 
    isLoading: organizationsLoading 
  } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => apiClient.getOrganizations()
  })

  // Load projects for selected organization
  const { 
    data: projects = [], 
    isLoading: projectsLoading 
  } = useQuery({
    queryKey: ['projects', selectedOrganization?.id],
    queryFn: () => apiClient.getProjects(selectedOrganization!.id),
    enabled: !!selectedOrganization
  })

  // Load issues for selected project
  const { 
    data: issuesData, 
    isLoading: issuesLoading 
  } = useQuery({
    queryKey: ['issues', selectedProject?.id, issueFilters],
    queryFn: () => apiClient.getIssues(selectedProject!.id, issueFilters),
    enabled: !!selectedProject
  })

  // Load issue stats for selected project
  const { 
    data: issueStats, 
    isLoading: statsLoading 
  } = useQuery({
    queryKey: ['issue-stats', selectedProject?.id],
    queryFn: () => apiClient.getIssueStats(selectedProject!.id),
    enabled: !!selectedProject
  })

  // Auto-select first organization and project if none selected
  useEffect(() => {
    if (!selectedOrganization && organizations && organizations.length > 0) {
      setSelectedOrganization(organizations[0])
    }
  }, [organizations, selectedOrganization, setSelectedOrganization])

  useEffect(() => {
    if (!selectedProject && projects && projects.length > 0) {
      setSelectedProject(projects[0])
    }
  }, [projects, selectedProject, setSelectedProject])

  const handleIssueClick = (issue: Issue) => {
    if (selectedOrganization && selectedProject) {
      navigate({
        to: `/organizations/${selectedOrganization.slug}/projects/${selectedProject.slug}/issues/${issue.id}`
      })
    }
  }

  const handleProjectClick = (project: any) => {
    if (selectedOrganization) {
      navigate({
        to: `/organizations/${selectedOrganization.slug}/projects/${project.slug}`
      })
    }
  }

  // Show loading state while organizations are loading
  if (organizationsLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">Loading...</p>
        </div>
        <Loading />
      </div>
    )
  }

  // Show create organization prompt if no organizations
  if (organizations.length === 0) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Welcome back, {user?.name}!
          </p>
        </div>

        <div className="text-center py-12">
          <BuildingOfficeIcon className="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No organizations found</h3>
          <p className="text-gray-600 mb-6">Create your first organization to get started.</p>
          <Button onClick={() => navigate({ to: '/organizations' } as any)}>
            <PlusIcon className="h-4 w-4 mr-2" />
            Create Organization
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Welcome back, {user?.name}!
          </p>
        </div>
        <div className="flex items-center space-x-4">
          <OrganizationSelector
            selectedOrgSlug={selectedOrganization?.slug}
            onOrganizationChange={setSelectedOrganization}
            className="w-48"
          />
          {selectedOrganization && (
            <ProjectSelector
              organizationId={selectedOrganization.id}
              selectedProjectSlug={selectedProject?.slug}
              onProjectChange={setSelectedProject}
              className="w-48"
            />
          )}
        </div>
      </div>

      {/* System-wide health overview */}
      <StatsOverview />

      {/* Main content based on selection */}
      {!selectedOrganization ? (
        <div className="text-center py-12">
          <p className="text-gray-500">Select an organization to continue</p>
        </div>
      ) : !selectedProject ? (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium text-gray-900">Projects</h2>
            <Button 
              onClick={() => navigate({ to: `/organizations/${selectedOrganization.slug}/projects/new` })}
              size="sm"
            >
              <PlusIcon className="h-4 w-4 mr-2" />
              New Project
            </Button>
          </div>
          
          {projectsLoading ? (
            <Loading />
          ) : projects.length === 0 ? (
            <div className="text-center py-12">
              <FolderIcon className="mx-auto h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No projects found</h3>
              <p className="text-gray-600 mb-6">Create your first project to start monitoring errors.</p>
              <Button 
                onClick={() => navigate({ to: `/organizations/${selectedOrganization.slug}/projects/new` })}
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                Create Project
              </Button>
            </div>
          ) : (
            <ProjectList 
              projects={projects} 
              onProjectClick={handleProjectClick}
            />
          )}
        </div>
      ) : (
        <div className="space-y-6">
          {/* Issue statistics */}
          <IssueStats 
            stats={issueStats} 
            isLoading={statsLoading} 
          />

          {/* Issues section */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-medium text-gray-900">Recent Issues</h2>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate({ 
                  to: `/organizations/${selectedOrganization.slug}/projects/${selectedProject.slug}/issues` 
                })}
              >
                View All Issues
              </Button>
            </div>

            <IssueFiltersComponent
              filters={issueFilters}
              onFiltersChange={setIssueFilters}
            />

            {issuesLoading ? (
              <Loading />
            ) : !issuesData || issuesData.data.length === 0 ? (
              <div className="text-center py-12">
                <ExclamationTriangleIcon className="mx-auto h-12 w-12 text-gray-400 mb-4" />
                <h3 className="text-lg font-medium text-gray-900 mb-2">No issues found</h3>
                <p className="text-gray-600">
                  {issueFilters.status?.includes('unresolved') 
                    ? "Great! No unresolved issues at the moment." 
                    : "No issues match your current filters."
                  }
                </p>
              </div>
            ) : (
              <IssueList
                issues={issuesData.data}
                total={issuesData.total}
                currentPage={issuesData.page}
                totalPages={issuesData.total_pages}
                onPageChange={(page) => setIssueFilters(prev => ({ ...prev, page }))}
                onIssueClick={handleIssueClick}
              />
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default DashboardPage