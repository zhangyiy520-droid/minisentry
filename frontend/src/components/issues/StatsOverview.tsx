import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'
import { Card, Loading } from '@/components/ui'
import {
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeSlashIcon,
  ClockIcon,
  FireIcon
} from '@heroicons/react/24/outline'

const statCard = (label: string, value: number, icon: React.ReactNode, color: string) => (
  <Card className="flex items-center gap-4 p-4">
    <div className={`p-2.5 rounded-lg ${color}`}>{icon}</div>
    <div>
      <p className="text-2xl font-bold text-gray-900">{value.toLocaleString()}</p>
      <p className="text-xs text-gray-500">{label}</p>
    </div>
  </Card>
)

export const StatsOverview = () => {
  const { data, isLoading } = useQuery({
    queryKey: ['stats-overview'],
    queryFn: () => apiClient.getStatsOverview(),
    refetchInterval: 30_000,
  })

  if (isLoading) return <Loading />
  if (!data) return null

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-gray-900">System Health</h2>
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {statCard('Total Issues', data.total_issues, <ExclamationTriangleIcon className="h-5 w-5 text-red-600" />, 'bg-red-50')}
        {statCard('Unresolved', data.unresolved_count, <FireIcon className="h-5 w-5 text-orange-600" />, 'bg-orange-50')}
        {statCard('Resolved', data.resolved_count, <CheckCircleIcon className="h-5 w-5 text-green-600" />, 'bg-green-50')}
        {statCard('Ignored', data.ignored_count, <EyeSlashIcon className="h-5 w-5 text-gray-600" />, 'bg-gray-50')}
        {statCard('Last 24h', data.recent_24h, <ClockIcon className="h-5 w-5 text-blue-600" />, 'bg-blue-50')}
      </div>
      {data.top_projects.length > 0 && (
        <div className="text-xs text-gray-400">
          Top projects: {data.top_projects.map(p => `${p.project_name} (${p.issue_count})`).join(' · ')}
        </div>
      )}
    </div>
  )
}

export default StatsOverview
