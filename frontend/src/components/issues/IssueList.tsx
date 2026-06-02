import { useState } from 'react'
import { Issue, BulkUpdateIssuesRequest } from '@/types/api'
import { Button, Pagination, Dropdown, DropdownItem } from '@/components/ui'
import IssueCard from './IssueCard'
import { 
  CheckIcon, 
  EyeSlashIcon, 
  UserIcon,
  ChevronDownIcon
} from '@heroicons/react/24/outline'

interface IssueListProps {
  issues: Issue[]
  total: number
  currentPage: number
  totalPages: number
  isLoading?: boolean
  selectedIssues?: string[]
  onPageChange: (page: number) => void
  onIssueClick: (issue: Issue) => void
  onIssueSelect?: (issueId: string, selected: boolean) => void
  onBulkUpdate?: (request: BulkUpdateIssuesRequest) => void
  className?: string
}

export const IssueList = ({
  issues,
  total,
  currentPage,
  totalPages,
  isLoading,
  selectedIssues = [],
  onPageChange,
  onIssueClick,
  onIssueSelect,
  onBulkUpdate,
  className
}: IssueListProps) => {
  const [bulkUpdateLoading, setBulkUpdateLoading] = useState(false)

  const handleSelectAll = () => {
    if (!onIssueSelect) return
    
    const allSelected = issues.every(issue => selectedIssues.includes(issue.id))
    issues.forEach(issue => {
      onIssueSelect(issue.id, !allSelected)
    })
  }

  const handleBulkAction = async (action: 'resolve' | 'ignore' | 'unresolve' | 'assign', assigneeId?: string) => {
    if (!onBulkUpdate || selectedIssues.length === 0) return

    setBulkUpdateLoading(true)
    try {
      await onBulkUpdate({
        issue_ids: selectedIssues,
        action,
        assignee_id: assigneeId
      })
    } finally {
      setBulkUpdateLoading(false)
    }
  }

  const hasSelection = selectedIssues.length > 0
  const allSelected = issues.length > 0 && issues.every(issue => selectedIssues.includes(issue.id))

  if (isLoading) {
    return (
      <div className={`space-y-4 ${className}`}>
        {[...Array(5)].map((_, i) => (
          <div key={i} className="animate-pulse">
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <div className="flex space-x-4">
                <div className="w-8 h-8 bg-gray-200 rounded"></div>
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                  <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                  <div className="h-3 bg-gray-200 rounded w-1/4"></div>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (issues.length === 0) {
    return (
      <div className={`text-center py-12 ${className}`}>
        <div className="text-gray-500">
          <CheckIcon className="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No issues found</h3>
          <p>No issues match your current filters.</p>
        </div>
      </div>
    )
  }

  return (
    <div className={className}>
      {/* Bulk actions */}
      {onIssueSelect && (
        <div className="bg-white border border-gray-200 rounded-lg p-4 mb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={handleSelectAll}
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm text-gray-700">
                  {hasSelection ? `${selectedIssues.length} selected` : 'Select all'}
                </span>
              </label>
              
              {total > 0 && (
                <span className="text-sm text-gray-500">
                  Showing {issues.length} of {total} issues
                </span>
              )}
            </div>

            {hasSelection && onBulkUpdate && (
              <div className="flex items-center space-x-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleBulkAction('resolve')}
                  disabled={bulkUpdateLoading}
                >
                  <CheckIcon className="h-4 w-4 mr-1" />
                  Resolve
                </Button>
                
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleBulkAction('ignore')}
                  disabled={bulkUpdateLoading}
                >
                  <EyeSlashIcon className="h-4 w-4 mr-1" />
                  Ignore
                </Button>

                <Dropdown
                  trigger={
                    <Button size="sm" variant="outline" disabled={bulkUpdateLoading}>
                      <UserIcon className="h-4 w-4 mr-1" />
                      Assign
                      <ChevronDownIcon className="h-4 w-4 ml-1" />
                    </Button>
                  }
                >
                  <DropdownItem onClick={() => handleBulkAction('assign', undefined)}>
                    Unassign
                  </DropdownItem>
                  {/* TODO: Add team member list */}
                </Dropdown>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Issue cards */}
      <div className="space-y-4">
        {issues.map((issue) => (
          <div key={issue.id} className="flex items-start space-x-3">
            {onIssueSelect && (
              <div className="pt-4">
                <input
                  type="checkbox"
                  checked={selectedIssues.includes(issue.id)}
                  onChange={(e) => onIssueSelect(issue.id, e.target.checked)}
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
              </div>
            )}
            <div className="flex-1">
              <IssueCard
                issue={issue}
                onClick={() => onIssueClick(issue)}
                isSelected={selectedIssues.includes(issue.id)}
              />
            </div>
          </div>
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-6">
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={onPageChange}
            showQuickJumper
          />
        </div>
      )}
    </div>
  )
}

export default IssueList