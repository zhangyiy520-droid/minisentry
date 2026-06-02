import { Project, ProjectPlatform } from '@/types/api'
import { Card, CardContent, CardHeader, CardTitle, Badge } from '@/components/ui'
import { formatRelativeTime } from '@/lib/utils'
import { 
  CodeBracketIcon,
  GlobeAltIcon,
} from '@heroicons/react/24/outline'

interface ProjectCardProps {
  project: Project
  onClick?: () => void
  showStats?: boolean
  className?: string
}

const platformIcons: Record<ProjectPlatform, React.ElementType> = {
  javascript: CodeBracketIcon,
  python: CodeBracketIcon,
  go: CodeBracketIcon,
  java: CodeBracketIcon,
  dotnet: CodeBracketIcon,
  php: CodeBracketIcon,
  ruby: CodeBracketIcon
}

const platformColors: Record<ProjectPlatform, string> = {
  javascript: 'bg-yellow-100 text-yellow-800',
  python: 'bg-blue-100 text-blue-800',
  go: 'bg-cyan-100 text-cyan-800',
  java: 'bg-red-100 text-red-800',
  dotnet: 'bg-purple-100 text-purple-800',
  php: 'bg-indigo-100 text-indigo-800',
  ruby: 'bg-red-100 text-red-800'
}

export const ProjectCard = ({ 
  project, 
  onClick, 
  showStats = true, 
  className 
}: ProjectCardProps) => {
  const PlatformIcon = platformIcons[project.platform] || CodeBracketIcon

  return (
    <Card 
      className={`cursor-pointer transition-all hover:shadow-md ${className}`}
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-3">
            <div className="flex-shrink-0">
              <div className="p-2 bg-gray-100 rounded-lg">
                <PlatformIcon className="h-6 w-6 text-gray-600" />
              </div>
            </div>
            <div className="min-w-0 flex-1">
              <CardTitle className="text-lg truncate">{project.name}</CardTitle>
              <p className="text-sm text-gray-500 truncate">{project.slug}</p>
            </div>
          </div>
          <div className="flex flex-col items-end space-y-1">
            <Badge 
              variant="default" 
              className={platformColors[project.platform]}
            >
              {project.platform}
            </Badge>
            <Badge variant={project.is_active ? 'success' : 'default'}>
              {project.is_active ? 'Active' : 'Inactive'}
            </Badge>
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        {project.description && (
          <p className="text-sm text-gray-600 mb-3 line-clamp-2">
            {project.description}
          </p>
        )}
        
        <div className="space-y-2">
          <div className="flex items-center text-xs text-gray-500">
            <GlobeAltIcon className="h-4 w-4 mr-1" />
            <span>DSN: {project.dsn.slice(0, 20)}...</span>
          </div>
          
          <div className="flex items-center justify-between text-xs text-gray-500">
            <span>Created {formatRelativeTime(project.created_at)}</span>
            <span>Updated {formatRelativeTime(project.updated_at)}</span>
          </div>
        </div>

        {showStats && (
          <div className="mt-4 pt-4 border-t border-gray-100">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <div className="text-lg font-semibold text-gray-900">-</div>
                <div className="text-xs text-gray-500">Issues</div>
              </div>
              <div>
                <div className="text-lg font-semibold text-gray-900">-</div>
                <div className="text-xs text-gray-500">Events</div>
              </div>
              <div>
                <div className="text-lg font-semibold text-gray-900">-</div>
                <div className="text-xs text-gray-500">Errors</div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default ProjectCard