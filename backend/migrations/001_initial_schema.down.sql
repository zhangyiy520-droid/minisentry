-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS releases;
DROP TABLE IF EXISTS issue_activities;
DROP TABLE IF EXISTS issue_comments;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "pgcrypto";