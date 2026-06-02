import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useLocation } from '@tanstack/react-router'
import { useAppContext } from '@/lib/context'
import { apiClient } from '@/lib/api'
import { Badge } from '@/components/ui'
import { OrganizationSelector } from '@/components/organizations'
import { ProjectSelector } from '@/components/projects'
import {
  HomeIcon,
  BuildingOfficeIcon,
  FolderIcon,
  ExclamationTriangleIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  Bars3Icon,
  XMarkIcon
} from '@heroicons/react/24/outline'

const navigation = [
  { name: 'Dashboard', href: '/', icon: HomeIcon },
  { name: 'Issues', href: '/issues', icon: ExclamationTriangleIcon },
  { name: 'Projects', href: '/projects', icon: FolderIcon },
  { name: 'Organizations', href: '/organizations', icon: BuildingOfficeIcon },
]

export const Sidebar = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [isProjectsExpanded, setIsProjectsExpanded] = useState(true)
  const location = useLocation()
  const { selectedOrganization, selectedProject, setSelectedOrganization, setSelectedProject } = useAppContext()

  // Load issue stats for badge counts
  const { data: issueStats } = useQuery({
    queryKey: ['issue-stats', selectedOrganization?.slug, selectedProject?.slug],
    queryFn: () => apiClient.getIssueStats(selectedOrganization!.slug, selectedProject!.slug),
    enabled: !!selectedOrganization && !!selectedProject,
    refetchInterval: 30000 // Refresh every 30 seconds
  })

  // Load projects for the sidebar
  const { data: projects = [] } = useQuery({
    queryKey: ['projects', selectedOrganization?.slug],
    queryFn: () => apiClient.getProjects(selectedOrganization!.slug),
    enabled: !!selectedOrganization
  })

  const isActive = (href: string) => {
    if (href === '/') {
      return location.pathname === '/'
    }
    return location.pathname.startsWith(href)
  }

  const getProjectIssuesHref = (projectSlug: string) => {
    if (selectedOrganization) {
      return `/organizations/${selectedOrganization.slug}/projects/${projectSlug}/issues`
    }
    return '#'
  }

  const SidebarContent = () => (
    <div className="flex h-full flex-col">
      {/* Context selectors */}
      <div className="p-4 border-b border-gray-200">
        <div className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">
              Organization
            </label>
            <OrganizationSelector
              selectedOrgSlug={selectedOrganization?.slug}
              onOrganizationChange={setSelectedOrganization}
              placeholder="Select organization"
              className="w-full"
            />
          </div>
          
          {selectedOrganization && (
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">
                Project
              </label>
              <ProjectSelector
                organizationId={selectedOrganization.id}
                selectedProjectSlug={selectedProject?.slug}
                onProjectChange={setSelectedProject}
                placeholder="Select project"
                className="w-full"
              />
            </div>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {navigation.map((item) => {
          const Icon = item.icon
          const active = isActive(item.href)
          const showBadge = item.name === 'Issues' && issueStats?.unresolved

          return (
            <Link
              key={item.name}
              to={item.href}
              className={`group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                active
                  ? 'bg-primary-100 text-primary-700'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              }`}
            >
              <Icon
                className={`mr-3 h-5 w-5 ${
                  active ? 'text-primary-500' : 'text-gray-400 group-hover:text-gray-500'
                }`}
              />
              {item.name}
              {showBadge && (
                <Badge variant="error" size="sm" className="ml-auto">
                  {issueStats.unresolved}
                </Badge>
              )}
            </Link>
          )
        })}

        {/* Project-specific navigation */}
        {selectedOrganization && projects.length > 0 && (
          <div className="pt-4">
            <button
              onClick={() => setIsProjectsExpanded(!isProjectsExpanded)}
              className="flex items-center w-full px-3 py-2 text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
            >
              {isProjectsExpanded ? (
                <ChevronDownIcon className="mr-2 h-4 w-4" />
              ) : (
                <ChevronRightIcon className="mr-2 h-4 w-4" />
              )}
              Projects ({projects.length})
            </button>
            
            {isProjectsExpanded && (
              <div className="ml-6 space-y-1">
                {projects.slice(0, 10).map((project) => (
                  <Link
                    key={project.id}
                    to={getProjectIssuesHref(project.slug)}
                    className={`block px-3 py-1 text-sm rounded-md transition-colors truncate ${
                      selectedProject?.id === project.id
                        ? 'bg-primary-50 text-primary-700'
                        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                    }`}
                    title={project.name}
                  >
                    {project.name}
                  </Link>
                ))}
                {projects.length > 10 && (
                  <Link
                    to="/projects"
                    className="block px-3 py-1 text-xs text-gray-500 hover:text-gray-700"
                  >
                    View all projects...
                  </Link>
                )}
              </div>
            )}
          </div>
        )}
      </nav>

      {/* Footer */}
      {selectedProject && (
        <div className="p-4 border-t border-gray-200">
          <div className="text-xs text-gray-500">
            <p className="font-medium truncate">{selectedProject.name}</p>
            <p className="truncate">{selectedProject.platform}</p>
          </div>
        </div>
      )}
    </div>
  )

  return (
    <>
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-4 left-4 z-50">
        <button
          onClick={() => setIsMobileMenuOpen(true)}
          className="p-2 bg-white rounded-md shadow-lg border border-gray-200"
        >
          <Bars3Icon className="h-6 w-6 text-gray-600" />
        </button>
      </div>

      {/* Mobile sidebar */}
      {isMobileMenuOpen && (
        <div className="lg:hidden">
          <div className="fixed inset-0 z-40 flex">
            <div
              className="fixed inset-0 bg-gray-600 bg-opacity-75"
              onClick={() => setIsMobileMenuOpen(false)}
            />
            <div className="relative flex w-full max-w-xs flex-1 flex-col bg-white">
              <div className="absolute top-0 right-0 -mr-12 pt-2">
                <button
                  onClick={() => setIsMobileMenuOpen(false)}
                  className="ml-1 flex h-10 w-10 items-center justify-center rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                >
                  <XMarkIcon className="h-6 w-6 text-white" />
                </button>
              </div>
              <SidebarContent />
            </div>
          </div>
        </div>
      )}

      {/* Desktop sidebar */}
      <div className="hidden lg:flex lg:w-64 lg:flex-col lg:fixed lg:inset-y-0">
        <div className="flex flex-col flex-grow bg-white border-r border-gray-200">
          <SidebarContent />
        </div>
      </div>
    </>
  )
}

export default Sidebar