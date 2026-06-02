// Base types
export interface BaseResponse {
  success: boolean
  message: string
  data?: Record<string, any>
}

export interface ErrorResponse {
  error: string
  message: string
  details?: Record<string, any>
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  total_pages: number
}

// User types
export interface User {
  id: string
  email: string
  name: string
  avatar_url?: string
  is_active: boolean
  email_verified: boolean
  created_at: string
  updated_at: string
}

export interface UserSummary {
  id: string
  email: string
  name: string
  avatar_url?: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface UpdateProfileRequest {
  name?: string
  avatar_url?: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

// Organization types
export type OrganizationRole = 'owner' | 'admin' | 'member'

export interface Organization {
  id: string
  name: string
  slug: string
  description?: string
  role: OrganizationRole
  created_at: string
  updated_at: string
}

export interface CreateOrganizationRequest {
  name: string
  slug: string
  description?: string
}

export interface UpdateOrganizationRequest {
  name?: string
  description?: string
}

export interface OrganizationMember {
  id: string
  organization_id: string
  user_id: string
  role: OrganizationRole
  user: UserSummary
  joined_at: string
}

export interface AddMemberRequest {
  email: string
  role: OrganizationRole
}

export interface UpdateMemberRoleRequest {
  role: OrganizationRole
}

// Project types
export type ProjectPlatform = 'javascript' | 'python' | 'go' | 'java' | 'dotnet' | 'php' | 'ruby'

export interface Project {
  id: string
  organization_id: string
  name: string
  slug: string
  description?: string
  platform: ProjectPlatform
  dsn: string
  public_key: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateProjectRequest {
  name: string
  slug: string
  description?: string
  platform: ProjectPlatform
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
  platform?: ProjectPlatform
}

export interface ProjectConfigurationRequest {
  is_active?: boolean
  platform?: ProjectPlatform
}

export interface ProjectKeyResponse {
  public_key: string
  dsn: string
}

// Issue types
export type IssueStatus = 'unresolved' | 'resolved' | 'ignored'
export type IssueLevel = 'error' | 'warning' | 'info' | 'debug'

export interface IssueFilters {
  status?: IssueStatus[]
  level?: IssueLevel[]
  assigned_to?: string
  date_from?: string
  date_to?: string
  search?: string
  sort?: 'frequency' | 'first_seen' | 'last_seen'
  order?: 'asc' | 'desc'
  page?: number
  limit?: number
  environment?: string
}

export interface Issue {
  id: string
  project_id: string
  fingerprint: string
  title: string
  culprit?: string
  type: string
  level: IssueLevel
  status: IssueStatus
  first_seen: string
  last_seen: string
  times_seen: number
  assignee_id?: string
  created_at: string
  updated_at: string
  assignee?: IssueAssignee
  project?: IssueProject
  latest_event?: IssueEvent
  comment_count?: number
  tags?: Record<string, string>
}

export interface IssueAssignee {
  id: string
  name: string
  email: string
  username: string
}

export interface IssueProject {
  id: string
  name: string
  slug: string
}

export interface IssueEvent {
  id: string
  event_id: string
  timestamp: string
  level: IssueLevel
  message?: string
  exception_type?: string
  exception_value?: string
  environment: string
  release_version?: string
  server_name?: string
  user_context?: Record<string, any>
  tags?: Record<string, any>
}

export interface IssueUpdateRequest {
  status?: IssueStatus
  assignee_id?: string | null
  resolution?: string
}

export interface IssueComment {
  id: string
  issue_id: string
  user_id: string
  content: string
  created_at: string
  updated_at: string
  user: IssueCommentUser
}

export interface IssueCommentUser {
  id: string
  name: string
  email: string
  username: string
}

export interface IssueCommentRequest {
  content: string
}

export interface IssueActivity {
  id: string
  issue_id: string
  user_id?: string
  type: string
  data: Record<string, any>
  created_at: string
  user?: IssueActivityUser
}

export interface IssueActivityUser {
  id: string
  name: string
  email: string
  username: string
}

export interface IssueStats {
  total: number
  unresolved: number
  resolved: number
  ignored: number
  new_today: number
  new_this_week: number
  by_level: Record<string, number>
  by_environment: Record<string, number>
  top_issues: Issue[]
  timeline: IssueTimelineEntry[]
}

export interface IssueTimelineEntry {
  date: string
  count: number
}

export interface BulkUpdateIssuesRequest {
  issue_ids: string[]
  action: 'resolve' | 'ignore' | 'unresolve' | 'assign'
  assignee_id?: string
  resolution?: string
}

export interface BulkUpdateIssuesResponse {
  updated_count: number
  failed_count: number
  errors?: string[]
  updated_ids: string[]
}