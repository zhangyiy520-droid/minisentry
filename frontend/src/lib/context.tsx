import { createContext, useContext, ReactNode, useState, useEffect } from 'react'
import { Organization, Project } from '@/types/api'

interface AppContextType {
  selectedOrganization: Organization | null
  selectedProject: Project | null
  setSelectedOrganization: (org: Organization | null) => void
  setSelectedProject: (project: Project | null) => void
}

const AppContext = createContext<AppContextType | undefined>(undefined)

interface AppProviderProps {
  children: ReactNode
}

export const AppProvider = ({ children }: AppProviderProps) => {
  const [selectedOrganization, setSelectedOrganization] = useState<Organization | null>(null)
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)

  // Clear project when organization changes
  useEffect(() => {
    if (selectedOrganization && selectedProject?.organization_id !== selectedOrganization.id) {
      setSelectedProject(null)
    }
  }, [selectedOrganization, selectedProject])

  const value = {
    selectedOrganization,
    selectedProject,
    setSelectedOrganization,
    setSelectedProject
  }

  return (
    <AppContext.Provider value={value}>
      {children}
    </AppContext.Provider>
  )
}

export const useAppContext = () => {
  const context = useContext(AppContext)
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider')
  }
  return context
}

export default AppProvider