import { Issue } from '@/types/api'
import { Badge, Card, CardContent } from '@/components/ui'
import { formatRelativeTime } from '@/lib/utils'
import { 
  ExclamationTriangleIcon, 
  InformationCircleIcon,
  ExclamationCircleIcon,
  BugAntIcon 
} from '@heroicons/react/24/outline'

interface IssueCardProps {
  issue: Issue
  onClick?: () => void
  isSelected?: boolean
}

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

export const IssueCard = ({ issue, onClick, isSelected }: IssueCardProps) => {
  const LevelIcon = levelIcons[issue.level]
  
  return (
    <Card 
      className={`cursor-pointer transition-all hover:shadow-md ${isSelected ? 'ring-2 ring-primary-500' : ''}`}
      onClick={onClick}
    >
      <CardContent className="p-4">
        <div className="flex items-start justify-between">
          <div className="flex items-start space-x-3 flex-1 min-w-0">
            <div className="flex-shrink-0">
              <LevelIcon className={`h-5 w-5 ${
                issue.level === 'error' ? 'text-red-500' :
                issue.level === 'warning' ? 'text-yellow-500' :
                issue.level === 'info' ? 'text-blue-500' :
                'text-gray-500'
              }`} />
            </div>
            
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-2">
                <Badge variant={levelColors[issue.level]}>
                  {issue.level}
                </Badge>
                <Badge variant={statusColors[issue.status]}>
                  {issue.status}
                </Badge>
                {issue.assignee && (
                  <Badge variant="info">
                    assigned to {issue.assignee.name}
                  </Badge>
                )}
              </div>
              
              <h3 className="text-sm font-medium text-gray-900 truncate mb-1">
                {issue.title}
              </h3>
              
              {issue.culprit && (
                <p className="text-xs text-gray-500 truncate mb-2">
                  {issue.culprit}
                </p>
              )}
              
              <div className="flex items-center justify-between text-xs text-gray-500">
                <div className="flex items-center space-x-4">
                  <span>First seen {formatRelativeTime(issue.first_seen)}</span>
                  <span>Last seen {formatRelativeTime(issue.last_seen)}</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="font-medium">{issue.times_seen} events</span>
                  {issue.comment_count && issue.comment_count > 0 && (
                    <span>{issue.comment_count} comments</span>
                  )}
                </div>
              </div>
              
              {issue.latest_event && (
                <div className="mt-2 pt-2 border-t border-gray-100">
                  <p className="text-xs text-gray-600 truncate">
                    Latest: {issue.latest_event.message || issue.latest_event.exception_value}
                  </p>
                  <div className="flex items-center gap-2 mt-1">
                    {issue.latest_event.environment && (
                      <Badge variant="default" size="sm">
                        {issue.latest_event.environment}
                      </Badge>
                    )}
                    {issue.latest_event.release_version && (
                      <Badge variant="default" size="sm">
                        {issue.latest_event.release_version}
                      </Badge>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default IssueCard