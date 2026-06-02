import { useState } from 'react'
import { IssueFilters as IssueFiltersType, IssueStatus, IssueLevel } from '@/types/api'
import { Button, Input, Select, Badge } from '@/components/ui'
import { MagnifyingGlassIcon, FunnelIcon, XMarkIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils'

interface IssueFiltersProps {
  filters: IssueFiltersType
  onFiltersChange: (filters: IssueFiltersType) => void
  className?: string
}

const statusOptions: { value: IssueStatus; label: string }[] = [
  { value: 'unresolved', label: 'Unresolved' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'ignored', label: 'Ignored' }
]

const levelOptions: { value: IssueLevel; label: string }[] = [
  { value: 'error', label: 'Error' },
  { value: 'warning', label: 'Warning' },
  { value: 'info', label: 'Info' },
  { value: 'debug', label: 'Debug' }
]

const sortOptions = [
  { value: 'frequency', label: 'Frequency' },
  { value: 'first_seen', label: 'First Seen' },
  { value: 'last_seen', label: 'Last Seen' }
]

export const IssueFilters = ({ filters, onFiltersChange, className }: IssueFiltersProps) => {
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [searchValue, setSearchValue] = useState(filters.search || '')

  const updateFilter = (key: keyof IssueFiltersType, value: any) => {
    onFiltersChange({ ...filters, [key]: value, page: 1 })
  }

  const toggleStatus = (status: IssueStatus) => {
    const currentStatuses = filters.status || []
    const newStatuses = currentStatuses.includes(status)
      ? currentStatuses.filter(s => s !== status)
      : [...currentStatuses, status]
    updateFilter('status', newStatuses.length > 0 ? newStatuses : undefined)
  }

  const toggleLevel = (level: IssueLevel) => {
    const currentLevels = filters.level || []
    const newLevels = currentLevels.includes(level)
      ? currentLevels.filter(l => l !== level)
      : [...currentLevels, level]
    updateFilter('level', newLevels.length > 0 ? newLevels : undefined)
  }

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateFilter('search', searchValue.trim() || undefined)
  }

  const clearFilters = () => {
    setSearchValue('')
    onFiltersChange({ page: 1, limit: filters.limit })
  }

  const hasActiveFilters = filters.status?.length || filters.level?.length || filters.search || 
    filters.assigned_to || filters.date_from || filters.date_to

  return (
    <div className={cn('space-y-4', className)}>
      {/* Search and basic filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <form onSubmit={handleSearchSubmit} className="flex-1">
          <div className="relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              type="text"
              placeholder="Search issues..."
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              className="pl-10"
            />
          </div>
        </form>
        
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setShowAdvanced(!showAdvanced)}
          >
            <FunnelIcon className="h-4 w-4 mr-1" />
            Filters
          </Button>
          
          {hasActiveFilters && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={clearFilters}
            >
              <XMarkIcon className="h-4 w-4 mr-1" />
              Clear
            </Button>
          )}
        </div>
      </div>

      {/* Active filter badges */}
      {hasActiveFilters && (
        <div className="flex flex-wrap gap-2">
          {filters.status?.map(status => (
            <Badge key={status} variant="info" className="cursor-pointer" onClick={() => toggleStatus(status)}>
              Status: {status} ×
            </Badge>
          ))}
          {filters.level?.map(level => (
            <Badge key={level} variant="warning" className="cursor-pointer" onClick={() => toggleLevel(level)}>
              Level: {level} ×
            </Badge>
          ))}
          {filters.search && (
            <Badge variant="default" className="cursor-pointer" onClick={() => updateFilter('search', undefined)}>
              Search: "{filters.search}" ×
            </Badge>
          )}
        </div>
      )}

      {/* Advanced filters */}
      {showAdvanced && (
        <div className="bg-gray-50 p-4 rounded-lg space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Status filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Status</label>
              <div className="space-y-1">
                {statusOptions.map(option => (
                  <label key={option.value} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={filters.status?.includes(option.value) || false}
                      onChange={() => toggleStatus(option.value)}
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                    <span className="ml-2 text-sm text-gray-700">{option.label}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* Level filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Level</label>
              <div className="space-y-1">
                {levelOptions.map(option => (
                  <label key={option.value} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={filters.level?.includes(option.value) || false}
                      onChange={() => toggleLevel(option.value)}
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                    <span className="ml-2 text-sm text-gray-700">{option.label}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* Sort */}
            <div>
              <Select
                label="Sort by"
                value={filters.sort || 'frequency'}
                onChange={(e) => updateFilter('sort', e.target.value)}
              >
                {sortOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Select>
            </div>

            {/* Order */}
            <div>
              <Select
                label="Order"
                value={filters.order || 'desc'}
                onChange={(e) => updateFilter('order', e.target.value)}
              >
                <option value="desc">Descending</option>
                <option value="asc">Ascending</option>
              </Select>
            </div>
          </div>

          {/* Date range */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              type="date"
              label="From date"
              value={filters.date_from || ''}
              onChange={(e) => updateFilter('date_from', e.target.value || undefined)}
            />
            <Input
              type="date"
              label="To date"
              value={filters.date_to || ''}
              onChange={(e) => updateFilter('date_to', e.target.value || undefined)}
            />
          </div>
        </div>
      )}
    </div>
  )
}

export default IssueFilters