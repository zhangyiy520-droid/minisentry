import { createRouter, createRoute, createRootRoute, Outlet } from '@tanstack/react-router'
import { QueryClient } from '@tanstack/react-query'
import { AuthProvider } from '@/lib/auth'
import { AppProvider } from '@/lib/context'
import ProtectedRoute from '@/components/layout/ProtectedRoute'
import AppLayout from '@/components/layout/AppLayout'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import DashboardPage from '@/pages/DashboardPage'
import OrganizationsPage from '@/pages/OrganizationsPage'
import CreateProjectPage from '@/pages/CreateProjectPage'
import IssueDetailPage from '@/pages/IssueDetailPage'
import ProjectIssuesPage from '@/pages/ProjectIssuesPage'

// Create a query client instance
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      retry: (failureCount, error: any) => {
        // Don't retry on 401/403 errors
        if (error?.status === 401 || error?.status === 403) {
          return false
        }
        return failureCount < 3
      }
    }
  }
})

// Root route
const rootRoute = createRootRoute({
  component: () => (
    <AuthProvider>
      <AppProvider>
        <Outlet />
      </AppProvider>
    </AuthProvider>
  )
})

// Protected layout route (requires authentication)
const protectedLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'protected',
  component: () => (
    <ProtectedRoute>
      <AppLayout>
        <Outlet />
      </AppLayout>
    </ProtectedRoute>
  )
})

// Public layout route (no authentication required)
const publicLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'public',
  component: () => (
    <ProtectedRoute requireAuth={false}>
      <Outlet />
    </ProtectedRoute>
  )
})

// Dashboard route (protected)
const dashboardRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/',
  component: DashboardPage
})

// Organizations route (protected)
const organizationsRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations',
  component: OrganizationsPage
})

// Organization detail route
const organizationDetailRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations/$orgSlug',
  component: () => (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Organization Details</h1>
      <p className="mt-2 text-gray-600">Organization management page.</p>
    </div>
  )
})

// Projects route (protected)
const projectsRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/projects',
  component: () => (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Projects</h1>
      <p className="mt-2 text-gray-600">Manage your projects here.</p>
    </div>
  )
})

// Organization projects route
const orgProjectsRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations/$orgSlug/projects',
  component: () => (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Projects</h1>
      <p className="mt-2 text-gray-600">Manage organization projects.</p>
    </div>
  )
})

// Create project route
const createProjectRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations/$orgSlug/projects/new',
  component: CreateProjectPage
})

// Project detail route
const projectDetailRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations/$orgSlug/projects/$projectSlug',
  component: () => (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Project Details</h1>
      <p className="mt-2 text-gray-600">Project management page.</p>
    </div>
  )
})

// Issues route (protected)
const issuesRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/issues',
  component: () => (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Issues</h1>
      <p className="mt-2 text-gray-600">Monitor your issues here.</p>
    </div>
  )
})

// Project issues route
const projectIssuesRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations/$orgSlug/projects/$projectSlug/issues',
  component: ProjectIssuesPage
})

// Issue detail route
const issueDetailRoute = createRoute({
  getParentRoute: () => protectedLayoutRoute,
  path: '/organizations/$orgSlug/projects/$projectSlug/issues/$issueId',
  component: IssueDetailPage
})

// Login route (public)
const loginRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/login',
  component: LoginPage,
  validateSearch: (search: Record<string, unknown>) => {
    return {
      return: typeof search.return === 'string' ? search.return : undefined
    }
  }
})

// Register route (public)
const registerRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/register',
  component: RegisterPage
})

// Create the route tree
const routeTree = rootRoute.addChildren([
  protectedLayoutRoute.addChildren([
    dashboardRoute,
    organizationsRoute,
    organizationDetailRoute,
    projectsRoute,
    orgProjectsRoute,
    createProjectRoute,
    projectDetailRoute,
    issuesRoute,
    projectIssuesRoute,
    issueDetailRoute
  ]),
  publicLayoutRoute.addChildren([
    loginRoute,
    registerRoute
  ])
])

// Create the router
export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0
})

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

export default router