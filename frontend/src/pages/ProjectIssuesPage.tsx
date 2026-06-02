import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useNavigate } from '@tanstack/react-router'
import { apiClient } from '@/lib/api'
import { IssueFilters, Issue, BulkUpdateIssuesRequest } from '@/types/api'
import { 
  Button, 
  Alert,
  Loading 
} from '@/components/ui'
import { 
  IssueStats, 
  IssueList, 
  IssueFilters as IssueFiltersComponent 
} from '@/components/issues'
import { 
  ArrowLeftIcon,
  CogIcon
} from '@heroicons/react/24/outline'

export const ProjectIssuesPage = () => {
  const { orgSlug, projectSlug } = useParams({ strict: false }) as {
    orgSlug: string;
    projectSlug: string;
  }
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  
  const [selectedIssues, setSelectedIssues] = useState<string[]>([])
  const [issueFilters, setIssueFilters] = useState<IssueFilters>({
    page: 1,
    limit: 25,
    status: ['unresolved']
  })

  // Load project details
  const { 
    data: project, 
    isLoading: projectLoading 
  } = useQuery({
    queryKey: ['project', orgSlug, projectSlug],
    queryFn: () => apiClient.getProject(orgSlug, projectSlug)
  })

  // Load issue stats
  const { 
    data: issueStats, 
    isLoading: statsLoading 
  } = useQuery({
    queryKey: ['issue-stats', orgSlug, projectSlug],
    queryFn: () => apiClient.getIssueStats(orgSlug, projectSlug),
    enabled: !!project
  })

  // Load issues
  const { 
    data: issuesData, 
    isLoading: issuesLoading,
    error: issuesError
  } = useQuery({
    queryKey: ['issues', orgSlug, projectSlug, issueFilters],
    queryFn: () => apiClient.getIssues(orgSlug, projectSlug, issueFilters),
    enabled: !!project
  })

  // Bulk update mutation
  const bulkUpdateMutation = useMutation({
    mutationFn: (request: BulkUpdateIssuesRequest) => 
      apiClient.bulkUpdateIssues(orgSlug, projectSlug, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues', orgSlug, projectSlug] })
      queryClient.invalidateQueries({ queryKey: ['issue-stats', orgSlug, projectSlug] })
      setSelectedIssues([])
    }
  })

  const handleIssueClick = (issue: Issue) => {
    navigate({
      to: `/organizations/${orgSlug}/projects/${projectSlug}/issues/${issue.id}`
    })
  }

  const handleIssueSelect = (issueId: string, selected: boolean) => {
    setSelectedIssues(prev => 
      selected 
        ? [...prev, issueId]
        : prev.filter(id => id !== issueId)
    )
  }

  const handleBulkUpdate = async (request: BulkUpdateIssuesRequest) => {
    await bulkUpdateMutation.mutateAsync(request)
  }

  const goBack = () => {
    navigate({ to: `/organizations/${orgSlug}/projects/${projectSlug}` })
  }

  if (projectLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="sm" onClick={goBack}>
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <h1 className="text-2xl font-bold text-gray-900">Loading Project...</h1>
        </div>
        <Loading />
      </div>
    )
  }

  if (!project) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="sm" onClick={goBack}>
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <h1 className="text-2xl font-bold text-gray-900">Project Not Found</h1>
        </div>
        <Alert variant="error">
          The project you're looking for could not be found.
        </Alert>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="sm" onClick={goBack}>
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              {project.name} Issues
            </h1>
            <p className="text-sm text-gray-500">
              Monitor and manage errors for {project.name}
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate({ to: `/organizations/${orgSlug}/projects/${projectSlug}/settings` })}
          >
            <CogIcon className="h-4 w-4 mr-2" />
            Settings
          </Button>
        </div>
      </div>

      {/* Issue statistics */}
      <IssueStats 
        stats={issueStats} 
        isLoading={statsLoading} 
      />

      {/* Filters and Issues */}
      <div className="space-y-4">
        <IssueFiltersComponent
          filters={issueFilters}
          onFiltersChange={setIssueFilters}
        />

        {issuesError ? (
          <Alert variant="error">
            Failed to load issues. Please try again.
          </Alert>
        ) : (
          <IssueList
            issues={issuesData?.data || []}
            total={issuesData?.total || 0}
            currentPage={issuesData?.page || 1}
            totalPages={issuesData?.total_pages || 1}
            isLoading={issuesLoading}
            selectedIssues={selectedIssues}
            onPageChange={(page) => setIssueFilters(prev => ({ ...prev, page }))}
            onIssueClick={handleIssueClick}
            onIssueSelect={handleIssueSelect}
            onBulkUpdate={handleBulkUpdate}
          />
        )}
      </div>
    </div>
  )
}

export default ProjectIssuesPage