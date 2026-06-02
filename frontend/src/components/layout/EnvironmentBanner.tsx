import { useState } from 'react'
import { XMarkIcon } from '@heroicons/react/24/outline'

interface EnvironmentBannerProps {
  env?: string
}

export const EnvironmentBanner = ({ env }: EnvironmentBannerProps) => {
  const [dismissed, setDismissed] = useState(false)

  if (!env || env === 'production' || dismissed) return null

  const colors: Record<string, string> = {
    development: 'bg-yellow-500 text-yellow-900',
    staging: 'bg-orange-500 text-orange-900',
    testing: 'bg-blue-500 text-blue-900',
  }

  return (
    <div className={`flex items-center justify-center px-4 py-1.5 text-xs font-medium ${colors[env] || 'bg-purple-500 text-purple-900'}`}>
      <span>⚠ {env.toUpperCase()} ENVIRONMENT — data may be reset at any time</span>
      <button onClick={() => setDismissed(true)} className="ml-3 opacity-60 hover:opacity-100">
        <XMarkIcon className="h-3.5 w-3.5" />
      </button>
    </div>
  )
}

export default EnvironmentBanner
