-- Performance indexes
CREATE INDEX idx_events_issue_id ON events(issue_id);
CREATE INDEX idx_events_project_timestamp ON events(project_id, timestamp DESC);
CREATE INDEX idx_events_fingerprint ON events(fingerprint);
CREATE INDEX idx_events_level ON events(level);
CREATE INDEX idx_events_environment ON events(environment);
CREATE INDEX idx_issues_project_status ON issues(project_id, status);
CREATE INDEX idx_issues_last_seen ON issues(last_seen DESC);
CREATE INDEX idx_issues_times_seen ON issues(times_seen DESC);

-- Full-text search indexes
CREATE INDEX idx_events_message_fts ON events USING GIN(to_tsvector('english', message));
CREATE INDEX idx_issues_title_fts ON issues USING GIN(to_tsvector('english', title));

-- JSONB indexes for fast queries
CREATE INDEX idx_events_tags ON events USING GIN(tags);
CREATE INDEX idx_events_user_context ON events USING GIN(user_context);
CREATE INDEX idx_events_stack_trace ON events USING GIN(stack_trace);

-- Organization and project indexes
CREATE INDEX idx_organization_members_org_id ON organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX idx_projects_organization_id ON projects(organization_id);
CREATE INDEX idx_projects_dsn ON projects(dsn);

-- User indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);

-- Issue activity indexes
CREATE INDEX idx_issue_activities_issue_id ON issue_activities(issue_id);
CREATE INDEX idx_issue_activities_created_at ON issue_activities(created_at DESC);

-- Issue comments indexes
CREATE INDEX idx_issue_comments_issue_id ON issue_comments(issue_id);