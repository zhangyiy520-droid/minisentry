import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useNavigate } from '@tanstack/react-router'
import { apiClient } from '@/lib/api'
import { IssueUpdateRequest, IssueCommentRequest } from '@/types/api'
import { formatRelativeTime, formatDate } from '@/lib/utils'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Badge,
  Input,
  Loading,
  Alert,
  Dropdown,
  DropdownItem
} from '@/components/ui'
import {
  ArrowLeftIcon,
  CheckIcon,
  EyeSlashIcon,
  ChatBubbleLeftIcon,
  ClockIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  ExclamationCircleIcon,
  BugAntIcon,
  ChevronDownIcon
} from '@heroicons/react/24/outline'

const levelIcons = {
  error: ExclamationCircleIcon,
  warning: ExclamationTriangleIcon,
  info: InformationCircleIcon,
  debug: BugAntIcon
}

const levelColors = {
  error: 'error',
  warning: 'warning',
  info: 'info',
  debug: 'default'
} as const

const statusColors = {
  unresolved: 'error',
  resolved: 'success',
  ignored: 'default'
} as const

export const IssueDetailPage = () => {
  const { orgSlug, projectSlug, issueId } = useParams({ strict: false }) as {
    orgSlug: string;
    projectSlug: string;
    issueId: string;
  }
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  
  const [newComment, setNewComment] = useState('')
  const [isCommentSubmitting, setIsCommentSubmitting] = useState(false)

  // Load issue details
  const { 
    data: issue, 
    isLoading: issueLoading, 
    error: issueError 
  } = useQuery({
    queryKey: ['issue', orgSlug, projectSlug, issueId],
    queryFn: () => apiClient.getIssue(orgSlug, projectSlug, issueId)
  })

  // Load issue comments
  const { 
    data: commentsData, 
    isLoading: commentsLoading 
  } = useQuery({
    queryKey: ['issue-comments', orgSlug, projectSlug, issueId],
    queryFn: () => apiClient.getIssueComments(orgSlug, projectSlug, issueId),
    enabled: !!issue
  })

  // Load issue activity
  const { 
    data: activityData, 
    isLoading: activityLoading 
  } = useQuery({
    queryKey: ['issue-activity', orgSlug, projectSlug, issueId],
    queryFn: () => apiClient.getIssueActivity(orgSlug, projectSlug, issueId),
    enabled: !!issue
  })

  // Update issue mutation
  const updateIssueMutation = useMutation({
    mutationFn: (data: IssueUpdateRequest) => 
      apiClient.updateIssue(orgSlug, projectSlug, issueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issue', orgSlug, projectSlug, issueId] })
      queryClient.invalidateQueries({ queryKey: ['issue-activity', orgSlug, projectSlug, issueId] })
    }
  })

  // Add comment mutation
  const addCommentMutation = useMutation({
    mutationFn: (data: IssueCommentRequest) => 
      apiClient.addIssueComment(orgSlug, projectSlug, issueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issue-comments', orgSlug, projectSlug, issueId] })
      setNewComment('')
      setIsCommentSubmitting(false)
    },
    onError: () => {
      setIsCommentSubmitting(false)
    }
  })

  const handleStatusChange = (status: 'resolved' | 'ignored' | 'unresolved') => {
    updateIssueMutation.mutate({ status })
  }

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newComment.trim()) return
    
    setIsCommentSubmitting(true)
    addCommentMutation.mutate({ content: newComment.trim() })
  }

  const goBack = () => {
    navigate({ to: `/organizations/${orgSlug}/projects/${projectSlug}/issues` })
  }

  if (issueLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="sm" onClick={goBack}>
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <h1 className="text-2xl font-bold text-gray-900">Loading Issue...</h1>
        </div>
        <Loading />
      </div>
    )
  }

  if (issueError || !issue) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="sm" onClick={goBack}>
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <h1 className="text-2xl font-bold text-gray-900">Issue Not Found</h1>
        </div>
        <Alert variant="error">
          The issue you're looking for could not be found.
        </Alert>
      </div>
    )
  }

  const LevelIcon = levelIcons[issue.level]
  const comments = commentsData?.data || []
  const activities = activityData?.data || []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="sm" onClick={goBack}>
            <ArrowLeftIcon className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900 truncate">
              {issue.title}
            </h1>
            <p className="text-sm text-gray-500">
              Issue #{issue.id.slice(-8)} • {formatRelativeTime(issue.created_at)}
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <Dropdown
            trigger={
              <Button 
                variant="outline" 
                size="sm"
                disabled={updateIssueMutation.isPending}
              >
                Actions
                <ChevronDownIcon className="h-4 w-4 ml-1" />
              </Button>
            }
          >
            <DropdownItem 
              onClick={() => handleStatusChange('resolved')}
              disabled={issue.status === 'resolved'}
            >
              <CheckIcon className="h-4 w-4 mr-2" />
              Mark as Resolved
            </DropdownItem>
            <DropdownItem 
              onClick={() => handleStatusChange('ignored')}
              disabled={issue.status === 'ignored'}
            >
              <EyeSlashIcon className="h-4 w-4 mr-2" />
              Ignore
            </DropdownItem>
            {issue.status !== 'unresolved' && (
              <DropdownItem onClick={() => handleStatusChange('unresolved')}>
                <ExclamationTriangleIcon className="h-4 w-4 mr-2" />
                Reopen
              </DropdownItem>
            )}
          </Dropdown>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Issue details */}
          <Card>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="flex items-center space-x-3">
                  <LevelIcon className={`h-6 w-6 ${
                    issue.level === 'error' ? 'text-red-500' :
                    issue.level === 'warning' ? 'text-yellow-500' :
                    issue.level === 'info' ? 'text-blue-500' :
                    'text-gray-500'
                  }`} />
                  <div>
                    <CardTitle className="text-lg">{issue.title}</CardTitle>
                    {issue.culprit && (
                      <p className="text-sm text-gray-600 mt-1">{issue.culprit}</p>
                    )}
                  </div>
                </div>
                <div className="flex flex-col items-end space-y-2">
                  <Badge variant={levelColors[issue.level]}>
                    {issue.level}
                  </Badge>
                  <Badge variant={statusColors[issue.status]}>
                    {issue.status}
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                <div>
                  <p className="text-sm text-gray-500">First seen</p>
                  <p className="font-medium">{formatDate(issue.first_seen)}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-500">Last seen</p>
                  <p className="font-medium">{formatDate(issue.last_seen)}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-500">Occurrences</p>
                  <p className="font-medium">{issue.times_seen.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-500">Assignee</p>
                  <p className="font-medium">{issue.assignee?.name || 'Unassigned'}</p>
                </div>
              </div>

              {/* Latest event details */}
              {issue.latest_event && (
                <div className="border-t pt-4">
                  <h4 className="font-medium text-gray-900 mb-3">Latest Event</h4>
                  <div className="space-y-3">
                    {issue.latest_event.message && (
                      <div>
                        <p className="text-sm text-gray-500">Message</p>
                        <p className="text-sm font-mono bg-gray-50 p-2 rounded">
                          {issue.latest_event.message}
                        </p>
                      </div>
                    )}
                    
                    {issue.latest_event.exception_type && (
                      <div>
                        <p className="text-sm text-gray-500">Exception</p>
                        <p className="text-sm font-mono bg-gray-50 p-2 rounded">
                          {issue.latest_event.exception_type}: {issue.latest_event.exception_value}
                        </p>
                      </div>
                    )}

                    <div className="flex flex-wrap gap-2">
                      {issue.latest_event.environment && (
                        <Badge variant="default">
                          env: {issue.latest_event.environment}
                        </Badge>
                      )}
                      {issue.latest_event.release_version && (
                        <Badge variant="default">
                          release: {issue.latest_event.release_version}
                        </Badge>
                      )}
                      {issue.latest_event.server_name && (
                        <Badge variant="default">
                          server: {issue.latest_event.server_name}
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Comments */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center">
                <ChatBubbleLeftIcon className="h-5 w-5 mr-2" />
                Comments ({comments.length})
              </CardTitle>
            </CardHeader>
            <CardContent>
              {/* Add comment form */}
              <form onSubmit={handleAddComment} className="mb-6">
                <div className="flex space-x-3">
                  <Input
                    placeholder="Add a comment..."
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    className="flex-1"
                  />
                  <Button 
                    type="submit" 
                    disabled={!newComment.trim() || isCommentSubmitting}
                  >
                    {isCommentSubmitting ? 'Adding...' : 'Comment'}
                  </Button>
                </div>
              </form>

              {/* Comments list */}
              {commentsLoading ? (
                <Loading />
              ) : comments.length === 0 ? (
                <p className="text-center text-gray-500 py-6">No comments yet</p>
              ) : (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.id} className="border-l-2 border-gray-200 pl-4">
                      <div className="flex items-center space-x-2 mb-2">
                        <span className="font-medium text-sm">{comment.user.name}</span>
                        <span className="text-xs text-gray-500">
                          {formatRelativeTime(comment.created_at)}
                        </span>
                      </div>
                      <p className="text-sm text-gray-700">{comment.content}</p>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Issue metadata */}
          <Card>
            <CardHeader>
              <CardTitle>Issue Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm text-gray-500">Type</p>
                <p className="font-medium">{issue.type}</p>
              </div>
              
              <div>
                <p className="text-sm text-gray-500">Fingerprint</p>
                <p className="font-mono text-xs bg-gray-50 p-1 rounded">
                  {issue.fingerprint}
                </p>
              </div>

              {issue.tags && Object.keys(issue.tags).length > 0 && (
                <div>
                  <p className="text-sm text-gray-500 mb-2">Tags</p>
                  <div className="space-y-1">
                    {Object.entries(issue.tags).map(([key, value]) => (
                      <div key={key} className="flex justify-between text-xs">
                        <span className="text-gray-600">{key}</span>
                        <span className="font-mono">{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Activity */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center">
                <ClockIcon className="h-5 w-5 mr-2" />
                Activity
              </CardTitle>
            </CardHeader>
            <CardContent>
              {activityLoading ? (
                <Loading />
              ) : activities.length === 0 ? (
                <p className="text-center text-gray-500 py-4">No activity yet</p>
              ) : (
                <div className="space-y-3">
                  {activities.slice(0, 10).map((activity) => (
                    <div key={activity.id} className="text-xs">
                      <div className="flex items-center space-x-1">
                        <span className="font-medium">
                          {activity.user?.name || 'System'}
                        </span>
                        <span className="text-gray-600">{activity.type}</span>
                      </div>
                      <p className="text-gray-500">
                        {formatRelativeTime(activity.created_at)}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export default IssueDetailPage