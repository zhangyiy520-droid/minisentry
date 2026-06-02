import { ReactNode } from 'react'
import {
  CheckCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  XCircleIcon,
  XMarkIcon
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils'

export interface AlertProps {
  variant?: 'success' | 'error' | 'warning' | 'info'
  title?: string
  description?: string | ReactNode
  children?: ReactNode
  onClose?: () => void
  className?: string
}

const Alert = ({
  variant = 'info',
  title,
  description,
  children,
  onClose,
  className
}: AlertProps) => {
  const variantConfig = {
    success: {
      containerClass: 'bg-success-50 border-success-200',
      iconClass: 'text-success-400',
      titleClass: 'text-success-800',
      descriptionClass: 'text-success-700',
      icon: CheckCircleIcon
    },
    error: {
      containerClass: 'bg-error-50 border-error-200',
      iconClass: 'text-error-400',
      titleClass: 'text-error-800',
      descriptionClass: 'text-error-700',
      icon: XCircleIcon
    },
    warning: {
      containerClass: 'bg-warning-50 border-warning-200',
      iconClass: 'text-warning-400',
      titleClass: 'text-warning-800',
      descriptionClass: 'text-warning-700',
      icon: ExclamationTriangleIcon
    },
    info: {
      containerClass: 'bg-primary-50 border-primary-200',
      iconClass: 'text-primary-400',
      titleClass: 'text-primary-800',
      descriptionClass: 'text-primary-700',
      icon: InformationCircleIcon
    }
  }

  const config = variantConfig[variant]
  const Icon = config.icon

  return (
    <div className={cn(
      'rounded-lg border p-4',
      config.containerClass,
      className
    )}>
      <div className="flex">
        <div className="flex-shrink-0">
          <Icon className={cn('h-5 w-5', config.iconClass)} />
        </div>
        <div className="ml-3 flex-1">
          {title && (
            <h3 className={cn('text-sm font-medium', config.titleClass)}>
              {title}
            </h3>
          )}
          {(description || children) && (
            <div className={cn(
              'text-sm',
              title ? 'mt-2' : '',
              config.descriptionClass
            )}>
              {description || children}
            </div>
          )}
        </div>
        {onClose && (
          <div className="ml-auto pl-3">
            <div className="-mx-1.5 -my-1.5">
              <button
                onClick={onClose}
                className={cn(
                  'inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2',
                  variant === 'success' && 'text-success-500 hover:bg-success-100 focus:ring-success-600',
                  variant === 'error' && 'text-error-500 hover:bg-error-100 focus:ring-error-600',
                  variant === 'warning' && 'text-warning-500 hover:bg-warning-100 focus:ring-warning-600',
                  variant === 'info' && 'text-primary-500 hover:bg-primary-100 focus:ring-primary-600'
                )}
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export { Alert }