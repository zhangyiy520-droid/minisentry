import { IssueStats as IssueStatsType } from '@/types/api'
import { Card, CardContent, CardHeader, CardTitle, Loading } from '@/components/ui'
import { 
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeSlashIcon,
  ChartBarIcon
} from '@heroicons/react/24/outline'

interface IssueStatsProps {
  stats?: IssueStatsType
  isLoading?: boolean
  className?: string
}

export const IssueStats = ({ stats, isLoading, className }: IssueStatsProps) => {
  if (isLoading) {
    return (
      <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 ${className}`}>
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardContent className="p-6">
              <Loading />
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  if (!stats) return null

  const statCards = [
    {
      title: 'Total Issues',
      value: stats.total,
      icon: ChartBarIcon,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50'
    },
    {
      title: 'Unresolved',
      value: stats.unresolved,
      icon: ExclamationTriangleIcon,
      color: 'text-red-600',
      bgColor: 'bg-red-50'
    },
    {
      title: 'Resolved',
      value: stats.resolved,
      icon: CheckCircleIcon,
      color: 'text-green-600',
      bgColor: 'bg-green-50'
    },
    {
      title: 'Ignored',
      value: stats.ignored,
      icon: EyeSlashIcon,
      color: 'text-gray-600',
      bgColor: 'bg-gray-50'
    }
  ]

  return (
    <div className={className}>
      {/* Main stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {statCards.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.title}>
              <CardContent className="p-6">
                <div className="flex items-center">
                  <div className={`p-2 rounded-lg ${stat.bgColor}`}>
                    <Icon className={`h-6 w-6 ${stat.color}`} />
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-600">{stat.title}</p>
                    <p className="text-2xl font-semibold text-gray-900">{stat.value.toLocaleString()}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Additional stats */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">New today</span>
                <span className="font-semibold">{stats.new_today}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">New this week</span>
                <span className="font-semibold">{stats.new_this_week}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* By level */}
        <Card>
          <CardHeader>
            <CardTitle>By Level</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {Object.entries(stats.by_level).map(([level, count]) => (
                <div key={level} className="flex justify-between items-center">
                  <div className="flex items-center">
                    <div className={`w-3 h-3 rounded-full mr-2 ${
                      level === 'error' ? 'bg-red-500' :
                      level === 'warning' ? 'bg-yellow-500' :
                      level === 'info' ? 'bg-blue-500' :
                      'bg-gray-500'
                    }`} />
                    <span className="text-sm text-gray-600 capitalize">{level}</span>
                  </div>
                  <span className="font-semibold">{count}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* By environment */}
        {Object.keys(stats.by_environment).length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>By Environment</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {Object.entries(stats.by_environment).map(([env, count]) => (
                  <div key={env} className="flex justify-between items-center">
                    <span className="text-sm text-gray-600">{env}</span>
                    <span className="font-semibold">{count}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Timeline chart placeholder */}
        {stats.timeline && stats.timeline.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Timeline</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-32 flex items-end justify-between space-x-1">
                {stats.timeline.slice(-14).map((entry) => (
                  <div
                    key={entry.date}
                    className="bg-primary-500 rounded-t"
                    style={{
                      height: `${Math.max(4, (entry.count / Math.max(...stats.timeline.map(e => e.count))) * 100)}%`,
                      width: `${100 / 14}%`
                    }}
                    title={`${entry.date}: ${entry.count} issues`}
                  />
                ))}
              </div>
              <p className="text-xs text-gray-500 mt-2">Last 14 days</p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

export default IssueStats