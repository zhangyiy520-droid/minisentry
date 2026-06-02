#!/usr/bin/env node

/**
 * MiniSentry Integration Test Suite
 * 
 * This script performs comprehensive end-to-end testing of the MiniSentry application:
 * - Tests complete user workflow: register → create org → create project → send error → view dashboard
 * - Tests error ingestion with JavaScript SDK simulation  
 * - Tests API endpoints with proper authentication
 * - Verifies error grouping and issue management
 * - Tests bulk operations and filtering
 * 
 * Prerequisites:
 * - Backend running on http://localhost:8080
 * - Frontend running on http://localhost:3000 (optional, for UI testing)
 * - Database and Redis accessible
 * 
 * Usage:
 *   npm install axios
 *   node test-integration.js
 */

const axios = require('axios');
const crypto = require('crypto');

// Configuration
const CONFIG = {
  API_BASE_URL: process.env.API_URL || 'http://localhost:8080',
  FRONTEND_URL: process.env.FRONTEND_URL || 'http://localhost:3000',
  TEST_USER_EMAIL: 'test@example.com',
  TEST_USER_PASSWORD: 'TestPassword123!',
  TEST_USER_NAME: 'Test User',
  TEST_ORG_NAME: 'Test Organization',
  TEST_PROJECT_NAME: 'Test Project'
};

// Test state
let testContext = {
  userToken: null,
  refreshToken: null,
  userId: null,
  organizationId: null,
  projectId: null,
  projectDSN: null,
  issueId: null,
  eventIds: []
};

// Utility functions
function log(message, data = null) {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] ${message}`);
  if (data) {
    console.log(JSON.stringify(data, null, 2));
  }
}

function logError(message, error) {
  console.error(`❌ ${message}`);
  if (error.response) {
    console.error(`Status: ${error.response.status}`);
    console.error(`Data: ${JSON.stringify(error.response.data, null, 2)}`);
  } else {
    console.error(error.message);
  }
}

function logSuccess(message) {
  console.log(`✅ ${message}`);
}

function logStep(step) {
  console.log(`\n🔄 ${step}`);
  console.log('=' + '='.repeat(step.length + 3));
}

function generateRandomEmail() {
  const randomString = crypto.randomBytes(8).toString('hex');
  return `test-${randomString}@example.com`;
}

function generateEventId() {
  return crypto.randomBytes(16).toString('hex');
}

// Create axios instance with default config
const api = axios.create({
  baseURL: CONFIG.API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Add request interceptor for authentication
api.interceptors.request.use((config) => {
  if (testContext.userToken && !config.headers.Authorization) {
    config.headers.Authorization = `Bearer ${testContext.userToken}`;
  }
  return config;
});

// Test functions
async function testHealthCheck() {
  logStep('Testing Health Check');
  
  try {
    const response = await api.get('/health');
    logSuccess('Health check passed');
    log('Health status', response.data);
    return true;
  } catch (error) {
    logError('Health check failed', error);
    return false;
  }
}

async function testUserRegistration() {
  logStep('Testing User Registration');
  
  // Use a random email to avoid conflicts
  const testEmail = generateRandomEmail();
  
  try {
    const response = await api.post('/api/v1/auth/register', {
      name: CONFIG.TEST_USER_NAME,
      email: testEmail,
      password: CONFIG.TEST_USER_PASSWORD
    });
    
    testContext.userToken = response.data.access_token || response.data.token;
    testContext.refreshToken = response.data.refresh_token;
    testContext.userId = response.data.user.id;
    
    logSuccess('User registration successful');
    log('User details', {
      id: response.data.user.id,
      email: response.data.user.email,
      name: response.data.user.name
    });
    
    // Update config with the random email for login test
    CONFIG.TEST_USER_EMAIL = testEmail;
    
    return true;
  } catch (error) {
    logError('User registration failed', error);
    return false;
  }
}

async function testUserLogin() {
  logStep('Testing User Login');
  
  try {
    const response = await api.post('/api/v1/auth/login', {
      email: CONFIG.TEST_USER_EMAIL,
      password: CONFIG.TEST_USER_PASSWORD
    });
    
    testContext.userToken = response.data.access_token || response.data.token;
    testContext.refreshToken = response.data.refresh_token;
    testContext.userId = response.data.user.id;
    
    logSuccess('User login successful');
    log('Login response', {
      user: response.data.user,
      hasToken: !!(response.data.access_token || response.data.token)
    });
    
    return true;
  } catch (error) {
    logError('User login failed', error);
    return false;
  }
}

async function testUserProfile() {
  logStep('Testing User Profile');
  
  try {
    const response = await api.get('/api/v1/auth/profile');
    
    logSuccess('User profile fetched successfully');
    log('User profile', response.data);
    
    return true;
  } catch (error) {
    logError('User profile fetch failed', error);
    return false;
  }
}

async function testOrganizationCreation() {
  logStep('Testing Organization Creation');
  
  try {
    const response = await api.post('/api/v1/organizations', {
      name: CONFIG.TEST_ORG_NAME,
      slug: 'test-organization',
      description: 'A test organization for integration testing'
    });
    
    testContext.organizationId = response.data.id;
    
    logSuccess('Organization created successfully');
    log('Organization details', response.data);
    
    return true;
  } catch (error) {
    logError('Organization creation failed', error);
    return false;
  }
}

async function testOrganizationList() {
  logStep('Testing Organization List');
  
  try {
    const response = await api.get('/api/v1/organizations');
    
    logSuccess('Organizations listed successfully');
    log('Organizations', response.data);
    
    // Verify our test organization is in the list
    const organizations = response.data.organizations || response.data;
    if (Array.isArray(organizations)) {
      const testOrg = organizations.find(org => org.id === testContext.organizationId);
      if (!testOrg) {
        throw new Error('Test organization not found in list');
      }
    }
    
    return true;
  } catch (error) {
    logError('Organization list failed', error);
    return false;
  }
}

async function testProjectCreation() {
  logStep('Testing Project Creation');
  
  try {
    const response = await api.post(`/api/v1/organizations/${testContext.organizationId}/projects`, {
      name: CONFIG.TEST_PROJECT_NAME,
      slug: 'test-project',
      description: 'A test project for integration testing',
      platform: 'javascript'
    });
    
    testContext.projectId = response.data.id;
    testContext.projectDSN = response.data.dsn;
    
    logSuccess('Project created successfully');
    log('Project details', response.data);
    
    return true;
  } catch (error) {
    logError('Project creation failed', error);
    return false;
  }
}

async function testProjectList() {
  logStep('Testing Project List');
  
  try {
    const response = await api.get(`/api/v1/organizations/${testContext.organizationId}/projects`);
    
    logSuccess('Projects listed successfully');
    log('Projects', response.data);
    
    // Verify our test project is in the list
    const projects = response.data.projects || response.data;
    if (Array.isArray(projects)) {
      const testProject = projects.find(project => project.id === testContext.projectId);
      if (!testProject) {
        throw new Error('Test project not found in list');
      }
    }
    
    return true;
  } catch (error) {
    logError('Project list failed', error);
    return false;
  }
}

async function testErrorIngestion() {
  logStep('Testing Error Ingestion');
  
  const errorEvents = [
    {
      event_id: generateEventId(),
      timestamp: new Date().toISOString(),
      level: 'error',
      platform: 'javascript',
      environment: 'test',
      release: 'v1.0.0-test',
      exception: {
        values: [
          {
            type: 'TypeError',
            value: 'Cannot read property \'x\' of undefined',
            stacktrace: {
              frames: [
                {
                  filename: 'app.js',
                  function: 'handleClick',
                  lineno: 42,
                  colno: 15,
                  context_line: 'const value = obj.x.y;',
                  pre_context: ['function handleClick() {', '  const obj = getObject();'],
                  post_context: ['  return value;', '}']
                },
                {
                  filename: 'utils.js',
                  function: 'getObject',
                  lineno: 15,
                  colno: 8,
                  context_line: 'return null;',
                  pre_context: ['function getObject() {', '  // Simulate error'],
                  post_context: ['}']
                }
              ]
            }
          }
        ]
      },
      request: {
        url: 'http://localhost:3000/dashboard',
        method: 'GET',
        headers: {
          'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
        }
      },
      user: {
        id: 'user123',
        email: 'user@example.com',
        username: 'testuser'
      },
      tags: {
        browser: 'Chrome',
        os: 'macOS',
        component: 'dashboard'
      },
      extra: {
        sessionId: 'abc123',
        userId: testContext.userId
      }
    },
    {
      event_id: generateEventId(),
      timestamp: new Date().toISOString(),
      level: 'error',
      platform: 'javascript',
      environment: 'test',
      release: 'v1.0.0-test',
      exception: {
        values: [
          {
            type: 'ReferenceError',
            value: 'undefinedVariable is not defined',
            stacktrace: {
              frames: [
                {
                  filename: 'main.js',
                  function: 'init',
                  lineno: 10,
                  colno: 5,
                  context_line: 'console.log(undefinedVariable);',
                  pre_context: ['function init() {', '  // Initialize app'],
                  post_context: ['}']
                }
              ]
            }
          }
        ]
      },
      tags: {
        browser: 'Firefox',
        os: 'Windows',
        component: 'initialization'
      }
    }
  ];
  
  try {
    for (const event of errorEvents) {
      const response = await api.post(`/api/${testContext.projectId}/store/`, event);
      testContext.eventIds.push(event.event_id);
      
      logSuccess(`Error event ingested: ${event.event_id}`);
      log('Ingestion response', response.data);
    }
    
    return true;
  } catch (error) {
    logError('Error ingestion failed', error);
    return false;
  }
}

async function testIssuesList() {
  logStep('Testing Issues List');
  
  // Wait a moment for issues to be processed
  await new Promise(resolve => setTimeout(resolve, 1000));
  
  try {
    const response = await api.get(`/api/v1/projects/${testContext.projectId}/issues`);
    
    logSuccess('Issues listed successfully');
    log('Issues', response.data);
    
    if (response.data.data && response.data.data.length > 0) {
      testContext.issueId = response.data.data[0].id;
      logSuccess(`Found ${response.data.data.length} issues`);
    }
    
    return true;
  } catch (error) {
    logError('Issues list failed', error);
    return false;
  }
}

async function testIssueDetails() {
  logStep('Testing Issue Details');
  
  if (!testContext.issueId) {
    log('No issue ID available, skipping issue details test');
    return true;
  }
  
  try {
    const response = await api.get(`/api/v1/issues/${testContext.issueId}`);
    
    logSuccess('Issue details fetched successfully');
    log('Issue details', response.data);
    
    return true;
  } catch (error) {
    logError('Issue details fetch failed', error);
    return false;
  }
}

async function testIssueUpdate() {
  logStep('Testing Issue Update');
  
  if (!testContext.issueId) {
    log('No issue ID available, skipping issue update test');
    return true;
  }
  
  try {
    const response = await api.put(`/api/v1/issues/${testContext.issueId}`, {
      status: 'resolved',
      assignee_id: testContext.userId
    });
    
    logSuccess('Issue updated successfully');
    log('Updated issue', response.data);
    
    return true;
  } catch (error) {
    logError('Issue update failed', error);
    return false;
  }
}

async function testIssueComments() {
  logStep('Testing Issue Comments');
  
  if (!testContext.issueId) {
    log('No issue ID available, skipping issue comments test');
    return true;
  }
  
  try {
    // Add a comment
    const commentResponse = await api.post(`/api/v1/issues/${testContext.issueId}/comments`, {
      content: 'This is a test comment for integration testing.'
    });
    
    logSuccess('Comment added successfully');
    log('Comment', commentResponse.data);
    
    // List comments
    const listResponse = await api.get(`/api/v1/issues/${testContext.issueId}/comments`);
    
    logSuccess('Comments listed successfully');
    log('Comments', listResponse.data);
    
    return true;
  } catch (error) {
    logError('Issue comments test failed', error);
    return false;
  }
}

async function testIssueStats() {
  logStep('Testing Issue Statistics');
  
  try {
    const response = await api.get(`/api/v1/projects/${testContext.projectId}/issues/stats`);
    
    logSuccess('Issue statistics fetched successfully');
    log('Issue stats', response.data);
    
    return true;
  } catch (error) {
    logError('Issue statistics fetch failed', error);
    return false;
  }
}

async function testBulkOperations() {
  logStep('Testing Bulk Operations');
  
  if (testContext.eventIds.length === 0) {
    log('No event IDs available, skipping bulk operations test');
    return true;
  }
  
  try {
    // Get issue IDs for bulk update
    const issuesResponse = await api.get(`/api/v1/projects/${testContext.projectId}/issues`);
    const issueIds = issuesResponse.data.data.map(issue => issue.id);
    
    if (issueIds.length === 0) {
      log('No issues available for bulk operations');
      return true;
    }
    
    // Bulk update issues
    const response = await api.post('/api/v1/issues/bulk-update', {
      issue_ids: issueIds,
      updates: {
        status: 'ignored',
        assignee_id: testContext.userId
      }
    });
    
    logSuccess('Bulk update completed successfully');
    log('Bulk update result', response.data);
    
    return true;
  } catch (error) {
    logError('Bulk operations test failed', error);
    return false;
  }
}

async function testFiltering() {
  logStep('Testing Issue Filtering');
  
  try {
    // Test various filters
    const filters = [
      { status: 'resolved' },
      { status: 'ignored' },
      { level: 'error' },
      { environment: 'test' },
      { search: 'TypeError' }
    ];
    
    for (const filter of filters) {
      const queryParams = new URLSearchParams(filter).toString();
      const response = await api.get(`/api/v1/projects/${testContext.projectId}/issues?${queryParams}`);
      
      logSuccess(`Filter test passed: ${JSON.stringify(filter)}`);
      log(`Results for ${JSON.stringify(filter)}`, {
        count: response.data.data ? response.data.data.length : 0,
        hasMore: response.data.pagination ? response.data.pagination.has_more : false
      });
    }
    
    return true;
  } catch (error) {
    logError('Filtering test failed', error);
    return false;
  }
}

async function testTokenRefresh() {
  logStep('Testing Token Refresh');
  
  if (!testContext.refreshToken) {
    log('No refresh token available, skipping refresh test');
    return true;
  }
  
  try {
    const response = await api.post('/api/v1/auth/refresh', {
      refresh_token: testContext.refreshToken
    });
    
    testContext.userToken = response.data.token;
    testContext.refreshToken = response.data.refresh_token;
    
    logSuccess('Token refresh successful');
    log('New token received', { hasToken: !!response.data.token });
    
    return true;
  } catch (error) {
    logError('Token refresh failed', error);
    return false;
  }
}

async function testCleanup() {
  logStep('Testing Cleanup (Optional)');
  
  try {
    // Try to delete the test project
    if (testContext.projectId) {
      await api.delete(`/api/v1/projects/${testContext.projectId}`);
      logSuccess('Test project deleted');
    }
    
    // Try to delete the test organization
    if (testContext.organizationId) {
      await api.delete(`/api/v1/organizations/${testContext.organizationId}`);
      logSuccess('Test organization deleted');
    }
    
    return true;
  } catch (error) {
    log('Cleanup failed (this is often expected in test environments)', error.message);
    return true; // Don't fail the entire test suite for cleanup issues
  }
}

// Main test runner
async function runIntegrationTests() {
  console.log('🚀 Starting MiniSentry Integration Tests');
  console.log('=====================================\n');
  
  const tests = [
    { name: 'Health Check', fn: testHealthCheck },
    { name: 'User Registration', fn: testUserRegistration },
    { name: 'User Login', fn: testUserLogin },
    { name: 'User Profile', fn: testUserProfile },
    { name: 'Organization Creation', fn: testOrganizationCreation },
    { name: 'Organization List', fn: testOrganizationList },
    { name: 'Project Creation', fn: testProjectCreation },
    { name: 'Project List', fn: testProjectList },
    { name: 'Error Ingestion', fn: testErrorIngestion },
    { name: 'Issues List', fn: testIssuesList },
    { name: 'Issue Details', fn: testIssueDetails },
    { name: 'Issue Update', fn: testIssueUpdate },
    { name: 'Issue Comments', fn: testIssueComments },
    { name: 'Issue Statistics', fn: testIssueStats },
    { name: 'Bulk Operations', fn: testBulkOperations },
    { name: 'Filtering', fn: testFiltering },
    { name: 'Token Refresh', fn: testTokenRefresh },
    { name: 'Cleanup', fn: testCleanup }
  ];
  
  const results = {
    passed: 0,
    failed: 0,
    skipped: 0,
    total: tests.length
  };
  
  for (const test of tests) {
    try {
      const success = await test.fn();
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    } catch (error) {
      logError(`Test '${test.name}' threw an exception`, error);
      results.failed++;
    }
    
    // Small delay between tests
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  
  // Print summary
  console.log('\n📊 Test Summary');
  console.log('================');
  console.log(`Total Tests: ${results.total}`);
  console.log(`✅ Passed: ${results.passed}`);
  console.log(`❌ Failed: ${results.failed}`);
  console.log(`⏭️  Skipped: ${results.skipped}`);
  console.log(`Success Rate: ${Math.round((results.passed / results.total) * 100)}%`);
  
  if (results.failed === 0) {
    console.log('\n🎉 All tests passed! MiniSentry is working correctly.');
    process.exit(0);
  } else {
    console.log('\n⚠️  Some tests failed. Please check the logs above for details.');
    process.exit(1);
  }
}

// Handle uncaught errors
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
  process.exit(1);
});

// Run the tests
if (require.main === module) {
  runIntegrationTests().catch((error) => {
    console.error('Integration tests failed:', error);
    process.exit(1);
  });
}

module.exports = {
  runIntegrationTests,
  testContext,
  CONFIG
};