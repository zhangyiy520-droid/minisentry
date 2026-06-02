import { Organization, OrganizationRole } from '@/types/api'
import { Card, CardContent, CardHeader, CardTitle, Badge } from '@/components/ui'
import { formatRelativeTime } from '@/lib/utils'
import { 
  BuildingOfficeIcon,
  UsersIcon,
  CogIcon,
  ShieldCheckIcon
} from '@heroicons/react/24/outline'

interface OrganizationCardProps {
  organization: Organization
  onClick?: () => void
  showStats?: boolean
  className?: string
}

const roleColors: Record<OrganizationRole, string> = {
  owner: 'bg-purple-100 text-purple-800',
  admin: 'bg-blue-100 text-blue-800',
  member: 'bg-green-100 text-green-800'
}

const roleIcons: Record<OrganizationRole, React.ElementType> = {
  owner: ShieldCheckIcon,
  admin: CogIcon,
  member: UsersIcon
}

export const OrganizationCard = ({ 
  organization, 
  onClick, 
  showStats = true, 
  className 
}: OrganizationCardProps) => {
  const RoleIcon = roleIcons[organization.role]

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
                <BuildingOfficeIcon className="h-6 w-6 text-gray-600" />
              </div>
            </div>
            <div className="min-w-0 flex-1">
              <CardTitle className="text-lg truncate">{organization.name}</CardTitle>
              <p className="text-sm text-gray-500 truncate">{organization.slug}</p>
            </div>
          </div>
          <Badge 
            variant="default" 
            className={roleColors[organization.role]}
          >
            <RoleIcon className="h-3 w-3 mr-1" />
            {organization.role}
          </Badge>
        </div>
      </CardHeader>
      
      <CardContent>
        {organization.description && (
          <p className="text-sm text-gray-600 mb-3 line-clamp-2">
            {organization.description}
          </p>
        )}
        
        <div className="flex items-center justify-between text-xs text-gray-500">
          <span>Created {formatRelativeTime(organization.created_at)}</span>
          <span>Updated {formatRelativeTime(organization.updated_at)}</span>
        </div>

        {showStats && (
          <div className="mt-4 pt-4 border-t border-gray-100">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <div className="text-lg font-semibold text-gray-900">-</div>
                <div className="text-xs text-gray-500">Projects</div>
              </div>
              <div>
                <div className="text-lg font-semibold text-gray-900">-</div>
                <div className="text-xs text-gray-500">Members</div>
              </div>
              <div>
                <div className="text-lg font-semibold text-gray-900">-</div>
                <div className="text-xs text-gray-500">Issues</div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default OrganizationCard