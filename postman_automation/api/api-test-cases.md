# Task Manager API - Test Cases Document

## Document Information
| Field | Value |
|-------|-------|
| **API Version** | 1.0.0 |
| **Base URL** | http://localhost:8080 |
| **Authentication** | JWT Bearer Token (24h expiry) |
| **Created Date** | 2026-01-14 |
| **Source** | swagger.yaml |

---

## Table of Contents
1. [Health Check Endpoint](#1-health-check-endpoint)
2. [Authentication Endpoints](#2-authentication-endpoints)
   - [User Registration](#21-user-registration)
   - [User Login](#22-user-login)
3. [Tasks Endpoints](#3-tasks-endpoints)
   - [Get All Tasks](#31-get-all-tasks)
   - [Create Task](#32-create-task)
   - [Get Task by ID](#33-get-task-by-id)
   - [Update Task](#34-update-task)
   - [Delete Task](#35-delete-task)
   - [Update Task Status](#36-update-task-status)

---

## 1. Health Check Endpoint

### Endpoint: `GET /api/v1/health`

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| HC-001 | Happy Path | Verify health check returns OK | Service is running | 1. Send GET request to `/api/v1/health` | Status: 200, Body: `{"status": "ok"}` | High |
| HC-002 | Negative | Verify health check when service is down | Service is stopped | 1. Stop the service<br>2. Send GET request to `/api/v1/health` | Connection refused or timeout | Medium |
| HC-003 | Edge Case | Verify health check with invalid HTTP method | Service is running | 1. Send POST request to `/api/v1/health` | Status: 405 Method Not Allowed | Low |
| HC-004 | Edge Case | Verify health check with query parameters | Service is running | 1. Send GET request to `/api/v1/health?foo=bar` | Status: 200, Body: `{"status": "ok"}` (should ignore params) | Low |

---

## 2. Authentication Endpoints

### 2.1 User Registration

### Endpoint: `POST /api/v1/auth/register`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| REG-001 | Happy Path | Register new user with valid credentials | Email not already registered | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "newuser@example.com", "password": "securePass123"}` | Status: 201<br>Body contains: `user_id` (UUID), `email` | High |
| REG-002 | Happy Path | Register user with minimum password length | Email not already registered | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "minpass@example.com", "password": "a"}` | Status: 201<br>Body contains: `user_id`, `email` | Medium |
| REG-003 | Happy Path | Register user with complex email | Email not already registered | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "user.name+tag@sub.domain.com", "password": "pass123"}` | Status: 201<br>Body contains: `user_id`, `email` | Medium |

#### Validation Error Scenarios (400)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| REG-004 | Validation | Register with missing email | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"password": "securePass123"}` | Status: 400<br>Error: `validation_error`<br>Message: "Email is required" | High |
| REG-005 | Validation | Register with missing password | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "user@example.com"}` | Status: 400<br>Error: `validation_error`<br>Message: "Password is required" | High |
| REG-006 | Validation | Register with empty body | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{}` | Status: 400<br>Error: `validation_error` | High |
| REG-007 | Validation | Register with invalid JSON | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{invalid json}` | Status: 400<br>Error: `invalid_request`<br>Message: "Invalid request body" | High |
| REG-008 | Validation | Register with empty email | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "", "password": "pass123"}` | Status: 400<br>Error: `validation_error` | High |
| REG-009 | Validation | Register with empty password | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "user@example.com", "password": ""}` | Status: 400<br>Error: `validation_error` | High |
| REG-010 | Validation | Register with invalid email format | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "invalid-email", "password": "pass123"}` | Status: 400<br>Error: `validation_error` | High |
| REG-011 | Validation | Register with email without domain | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": "user@", "password": "pass123"}` | Status: 400<br>Error: `validation_error` | Medium |
| REG-012 | Validation | Register with null values | None | 1. Send POST to `/api/v1/auth/register`<br>2. Body: `{"email": null, "password": null}` | Status: 400<br>Error: `validation_error` | Medium |

#### Conflict Scenarios (409)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| REG-013 | Conflict | Register with already existing email | User with email exists | 1. Register user with email<br>2. Send POST to `/api/v1/auth/register` with same email | Status: 409<br>Error: `user_exists`<br>Message: "User with this email already exists" | High |
| REG-014 | Conflict | Register with existing email (different case) | User with email exists | 1. Register with "User@Example.com"<br>2. Register with "user@example.com" | Status: 409 (if case-insensitive) or 201 (if case-sensitive) | Medium |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| REG-015 | Edge Case | Register with very long email | None | 1. Send POST with email of 255+ characters | Status: 400 or 201 (depending on limit) | Low |
| REG-016 | Edge Case | Register with very long password | None | 1. Send POST with password of 1000+ characters | Status: 400 or 201 (depending on limit) | Low |
| REG-017 | Edge Case | Register with special characters in password | None | 1. Send POST with password containing `!@#$%^&*()` | Status: 201 | Medium |
| REG-018 | Edge Case | Register with unicode characters in password | None | 1. Send POST with password containing unicode: `пароль123` | Status: 201 or 400 (depending on support) | Low |
| REG-019 | Edge Case | Register with SQL injection in email | None | 1. Send POST with email: `'; DROP TABLE users;--@test.com` | Status: 400 (invalid email format) | High |
| REG-020 | Edge Case | Register with XSS in email | None | 1. Send POST with email: `<script>alert('xss')</script>@test.com` | Status: 400 or sanitized | High |
| REG-021 | Edge Case | Register with whitespace email | None | 1. Send POST with email: `"  user@example.com  "` | Status: 400 or 201 (if trimmed) | Medium |
| REG-022 | Edge Case | Extra fields in request body | None | 1. Send POST with extra fields: `{"email": "...", "password": "...", "role": "admin"}` | Status: 201 (extra fields ignored) | Medium |

---

### 2.2 User Login

### Endpoint: `POST /api/v1/auth/login`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| LOG-001 | Happy Path | Login with valid credentials | User is registered | 1. Send POST to `/api/v1/auth/login`<br>2. Body: `{"email": "user@example.com", "password": "correctPassword"}` | Status: 200<br>Body contains: `token` (JWT), `user_id` (UUID) | High |
| LOG-002 | Happy Path | Login and verify token is valid JWT | User is registered | 1. Login successfully<br>2. Decode JWT token<br>3. Verify structure | Token has header, payload, signature<br>Payload contains `user_id`, `exp`, `iat` | High |
| LOG-003 | Happy Path | Login multiple times with same credentials | User is registered | 1. Login first time<br>2. Login second time | Both return 200 with valid tokens | Medium |

#### Validation Error Scenarios (400)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| LOG-004 | Validation | Login with missing email | None | 1. Send POST to `/api/v1/auth/login`<br>2. Body: `{"password": "pass123"}` | Status: 400<br>Error: `validation_error`<br>Message: "Email is required" | High |
| LOG-005 | Validation | Login with missing password | None | 1. Send POST to `/api/v1/auth/login`<br>2. Body: `{"email": "user@example.com"}` | Status: 400<br>Error: `validation_error`<br>Message: "Password is required" | High |
| LOG-006 | Validation | Login with empty body | None | 1. Send POST to `/api/v1/auth/login`<br>2. Body: `{}` | Status: 400<br>Error: `validation_error` | High |
| LOG-007 | Validation | Login with invalid JSON | None | 1. Send POST to `/api/v1/auth/login`<br>2. Body: malformed JSON | Status: 400<br>Error: `invalid_request` | High |
| LOG-008 | Validation | Login with empty email and password | None | 1. Send POST with: `{"email": "", "password": ""}` | Status: 400<br>Error: `validation_error` | High |

#### Authentication Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| LOG-009 | Auth Error | Login with non-existent email | User not registered | 1. Send POST with unregistered email | Status: 401<br>Error: `invalid_credentials`<br>Message: "Invalid email or password" | High |
| LOG-010 | Auth Error | Login with wrong password | User is registered | 1. Send POST with correct email, wrong password | Status: 401<br>Error: `invalid_credentials`<br>Message: "Invalid email or password" | High |
| LOG-011 | Auth Error | Login with correct password, wrong email | User is registered | 1. Send POST with wrong email, correct password | Status: 401<br>Error: `invalid_credentials` | High |
| LOG-012 | Auth Error | Login with email case mismatch | User registered with "User@Example.com" | 1. Login with "user@example.com" | Status: 200 (if case-insensitive) or 401 (if case-sensitive) | Medium |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| LOG-013 | Edge Case | Login with SQL injection in credentials | None | 1. Send POST with email: `' OR '1'='1` | Status: 401 (not exploited) | High |
| LOG-014 | Edge Case | Brute force protection test | User registered | 1. Attempt 10+ logins with wrong password | May return 429 Too Many Requests (if implemented) | Medium |
| LOG-015 | Edge Case | Login with token in request | User registered | 1. Send POST with Authorization header | Should ignore header, process login normally | Low |

---

## 3. Tasks Endpoints

### 3.1 Get All Tasks

### Endpoint: `GET /api/v1/tasks`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| GT-001 | Happy Path | Get all tasks for authenticated user | User logged in, has tasks | 1. Send GET to `/api/v1/tasks`<br>2. Include valid Bearer token | Status: 200<br>Body: Array of tasks belonging to user | High |
| GT-002 | Happy Path | Get tasks when user has no tasks | User logged in, no tasks | 1. Send GET to `/api/v1/tasks`<br>2. Include valid Bearer token | Status: 200<br>Body: Empty array `[]` | High |
| GT-003 | Happy Path | Get tasks - verify response structure | User logged in, has tasks | 1. Get tasks<br>2. Verify each task has required fields | Each task has: `id`, `user_id`, `title`, `description`, `status`, `priority`, `created_at`, `updated_at` | High |
| GT-004 | Happy Path | User only sees own tasks | Two users with tasks | 1. Login as User A<br>2. Get tasks<br>3. Verify all tasks have User A's user_id | All returned tasks belong to authenticated user | High |

#### Authentication Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| GT-005 | Auth Error | Get tasks without Authorization header | None | 1. Send GET to `/api/v1/tasks` without header | Status: 401<br>Error: `missing_authorization`<br>Message: "Authorization header is required" | High |
| GT-006 | Auth Error | Get tasks with empty Authorization header | None | 1. Send GET with `Authorization: ` | Status: 401<br>Error: `invalid_authorization` | High |
| GT-007 | Auth Error | Get tasks with invalid token format | None | 1. Send GET with `Authorization: Bearer` (no token) | Status: 401<br>Error: `invalid_authorization`<br>Message: "Authorization header must be Bearer token" | High |
| GT-008 | Auth Error | Get tasks with malformed JWT | None | 1. Send GET with `Authorization: Bearer invalid.token.here` | Status: 401<br>Error: `invalid_token`<br>Message: "Invalid or expired token" | High |
| GT-009 | Auth Error | Get tasks with expired token | Token expired (>24h) | 1. Wait for token to expire<br>2. Send GET with expired token | Status: 401<br>Error: `invalid_token`<br>Message: "Invalid or expired token" | High |
| GT-010 | Auth Error | Get tasks with Basic auth instead of Bearer | None | 1. Send GET with `Authorization: Basic base64credentials` | Status: 401<br>Error: `invalid_authorization` | Medium |
| GT-011 | Auth Error | Get tasks with modified JWT payload | Valid token modified | 1. Modify JWT payload<br>2. Send GET | Status: 401<br>Error: `invalid_token` | High |
| GT-012 | Auth Error | Get tasks with token from different user | User A's token, User B logged in | 1. Get token for User A<br>2. Delete User A<br>3. Use token | Status: 401<br>Error: `invalid_token` | Medium |

---

### 3.2 Create Task

### Endpoint: `POST /api/v1/tasks`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| CT-001 | Happy Path | Create task with title only | User authenticated | 1. POST to `/api/v1/tasks`<br>2. Body: `{"title": "My task"}` | Status: 201<br>Task created with default priority `medium`, status `pending` | High |
| CT-002 | Happy Path | Create task with all fields | User authenticated | 1. POST with: `{"title": "Task", "description": "Desc", "priority": "high"}` | Status: 201<br>All fields set as provided | High |
| CT-003 | Happy Path | Create task with low priority | User authenticated | 1. POST with priority: "low" | Status: 201<br>Priority: "low" | Medium |
| CT-004 | Happy Path | Create task with medium priority | User authenticated | 1. POST with priority: "medium" | Status: 201<br>Priority: "medium" | Medium |
| CT-005 | Happy Path | Create task with high priority | User authenticated | 1. POST with priority: "high" | Status: 201<br>Priority: "high" | Medium |
| CT-006 | Happy Path | Verify created_at and updated_at are set | User authenticated | 1. Create task<br>2. Check timestamps | Both timestamps are set and equal (on creation) | High |
| CT-007 | Happy Path | Verify task ID is UUID format | User authenticated | 1. Create task<br>2. Validate ID format | ID matches UUID pattern | High |
| CT-008 | Happy Path | Create multiple tasks | User authenticated | 1. Create 5 tasks<br>2. Get all tasks | All 5 tasks returned in list | Medium |

#### Validation Error Scenarios (400)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| CT-009 | Validation | Create task without title | User authenticated | 1. POST with: `{"description": "No title"}` | Status: 400<br>Error: `validation_error`<br>Message: "Title is required" | High |
| CT-010 | Validation | Create task with empty title | User authenticated | 1. POST with: `{"title": ""}` | Status: 400<br>Error: `validation_error`<br>Message: "Title is required" | High |
| CT-011 | Validation | Create task with invalid priority | User authenticated | 1. POST with: `{"title": "Task", "priority": "urgent"}` | Status: 400<br>Error: `validation_error`<br>Message: "Priority must be low, medium, or high" | High |
| CT-012 | Validation | Create task with invalid JSON | User authenticated | 1. POST with malformed JSON | Status: 400<br>Error: `invalid_request`<br>Message: "Invalid request body" | High |
| CT-013 | Validation | Create task with empty body | User authenticated | 1. POST with: `{}` | Status: 400<br>Error: `validation_error` | High |
| CT-014 | Validation | Create task with null title | User authenticated | 1. POST with: `{"title": null}` | Status: 400<br>Error: `validation_error` | High |
| CT-015 | Validation | Create task with numeric title | User authenticated | 1. POST with: `{"title": 12345}` | Status: 400 or 201 (if coerced to string) | Medium |
| CT-016 | Validation | Create task with priority as integer | User authenticated | 1. POST with: `{"title": "Task", "priority": 1}` | Status: 400<br>Error: `validation_error` | Medium |

#### Authentication Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| CT-017 | Auth Error | Create task without token | None | 1. POST to `/api/v1/tasks` without Authorization | Status: 401<br>Error: `missing_authorization` | High |
| CT-018 | Auth Error | Create task with invalid token | None | 1. POST with invalid Bearer token | Status: 401<br>Error: `invalid_token` | High |
| CT-019 | Auth Error | Create task with expired token | Token expired | 1. POST with expired token | Status: 401<br>Error: `invalid_token` | High |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| CT-020 | Edge Case | Create task with very long title | User authenticated | 1. POST with title of 10000+ characters | Status: 201 or 400 (if length limit exists) | Low |
| CT-021 | Edge Case | Create task with very long description | User authenticated | 1. POST with description of 100000+ characters | Status: 201 or 400 (if length limit exists) | Low |
| CT-022 | Edge Case | Create task with HTML in title | User authenticated | 1. POST with title: `<b>Bold Task</b>` | Status: 201 (HTML should be stored/escaped safely) | High |
| CT-023 | Edge Case | Create task with SQL injection in title | User authenticated | 1. POST with title: `'; DROP TABLE tasks;--` | Status: 201 (SQL injection prevented) | High |
| CT-024 | Edge Case | Create task with unicode title | User authenticated | 1. POST with title: `Задача на русском 中文任务` | Status: 201 | Medium |
| CT-025 | Edge Case | Create task with only whitespace title | User authenticated | 1. POST with title: `"   "` | Status: 400 (should be treated as empty) | Medium |
| CT-026 | Edge Case | Create task with extra fields | User authenticated | 1. POST with: `{"title": "T", "status": "done", "id": "custom"}` | Status: 201 (extra fields ignored, status defaults to pending) | Medium |

---

### 3.3 Get Task by ID

### Endpoint: `GET /api/v1/tasks/{id}`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| GID-001 | Happy Path | Get existing task by ID | User owns task | 1. GET `/api/v1/tasks/{valid-uuid}` | Status: 200<br>Body: Complete task object | High |
| GID-002 | Happy Path | Verify all response fields present | User owns task | 1. GET task<br>2. Verify response structure | All required fields present: id, user_id, title, description, status, priority, created_at, updated_at | High |

#### Not Found Scenarios (404)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| GID-003 | Not Found | Get task with non-existent UUID | User authenticated | 1. GET `/api/v1/tasks/00000000-0000-0000-0000-000000000000` | Status: 404<br>Error: `task_not_found`<br>Message: "Task not found" | High |
| GID-004 | Not Found | Get deleted task | Task was deleted | 1. Create and delete task<br>2. GET the deleted task | Status: 404<br>Error: `task_not_found` | High |

#### Authorization Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| GID-005 | Auth Error | Get task without token | None | 1. GET `/api/v1/tasks/{id}` without Authorization | Status: 401<br>Error: `missing_authorization` | High |
| GID-006 | Auth Error | Get task with invalid token | None | 1. GET with invalid Bearer token | Status: 401<br>Error: `invalid_token` | High |
| GID-007 | Auth Error | Get task owned by another user | Task belongs to User B | 1. Login as User A<br>2. GET User B's task ID | Status: 401<br>Error: `unauthorized`<br>Message: "You don't have access to this task" | High |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| GID-008 | Edge Case | Get task with invalid UUID format | User authenticated | 1. GET `/api/v1/tasks/not-a-uuid` | Status: 400 or 404 | Medium |
| GID-009 | Edge Case | Get task with empty ID | User authenticated | 1. GET `/api/v1/tasks/` | Status: 404 or redirects to GET all tasks | Low |
| GID-010 | Edge Case | Get task with SQL injection in ID | User authenticated | 1. GET `/api/v1/tasks/'; DROP TABLE tasks;--` | Status: 400 or 404 (injection prevented) | High |

---

### 3.4 Update Task

### Endpoint: `PUT /api/v1/tasks/{id}`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| UT-001 | Happy Path | Update task title | User owns task | 1. PUT to `/api/v1/tasks/{id}`<br>2. Body: `{"title": "Updated Title"}` | Status: 200<br>Title updated, other fields unchanged | High |
| UT-002 | Happy Path | Update all task fields | User owns task | 1. PUT with: `{"title": "New", "description": "New desc", "priority": "high"}` | Status: 200<br>All specified fields updated | High |
| UT-003 | Happy Path | Update task priority to low | User owns task | 1. PUT with: `{"title": "T", "priority": "low"}` | Status: 200<br>Priority: "low" | Medium |
| UT-004 | Happy Path | Verify updated_at changes | User owns task | 1. Get task (note updated_at)<br>2. Update task<br>3. Verify updated_at changed | updated_at is newer than before | High |
| UT-005 | Happy Path | Update task with same values | User owns task | 1. PUT with same values as existing | Status: 200<br>updated_at still changes | Low |

#### Validation Error Scenarios (400)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| UT-006 | Validation | Update task without title | User owns task | 1. PUT with: `{"description": "No title"}` | Status: 400<br>Error: `validation_error`<br>Message: "Title is required" | High |
| UT-007 | Validation | Update task with empty title | User owns task | 1. PUT with: `{"title": ""}` | Status: 400<br>Error: `validation_error` | High |
| UT-008 | Validation | Update task with invalid priority | User owns task | 1. PUT with: `{"title": "T", "priority": "critical"}` | Status: 400<br>Error: `validation_error`<br>Message: "Priority must be low, medium, or high" | High |
| UT-009 | Validation | Update task with invalid JSON | User owns task | 1. PUT with malformed JSON | Status: 400<br>Error: `invalid_request` | High |

#### Authorization Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| UT-010 | Auth Error | Update task without token | None | 1. PUT without Authorization header | Status: 401<br>Error: `missing_authorization` | High |
| UT-011 | Auth Error | Update task with invalid token | None | 1. PUT with invalid Bearer token | Status: 401<br>Error: `invalid_token` | High |
| UT-012 | Auth Error | Update task owned by another user | Task belongs to User B | 1. Login as User A<br>2. PUT to User B's task | Status: 401<br>Error: `unauthorized`<br>Message: "You don't have access to this task" | High |

#### Not Found Scenarios (404)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| UT-013 | Not Found | Update non-existent task | User authenticated | 1. PUT to `/api/v1/tasks/{non-existent-uuid}` | Status: 404<br>Error: `task_not_found` | High |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| UT-014 | Edge Case | Update task - try to change status via PUT | User owns task | 1. PUT with: `{"title": "T", "status": "done"}` | Status: 200 (status field ignored in PUT) | Medium |
| UT-015 | Edge Case | Update task - try to change ID | User owns task | 1. PUT with: `{"title": "T", "id": "new-id"}` | Status: 200 (id unchanged) | Medium |
| UT-016 | Edge Case | Update task - try to change user_id | User owns task | 1. PUT with: `{"title": "T", "user_id": "other-user"}` | Status: 200 (user_id unchanged) | High |

---

### 3.5 Delete Task

### Endpoint: `DELETE /api/v1/tasks/{id}`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| DT-001 | Happy Path | Delete existing task | User owns task | 1. DELETE `/api/v1/tasks/{id}` | Status: 204 (No Content) | High |
| DT-002 | Happy Path | Verify task is deleted | User owns task | 1. Delete task<br>2. GET the same task | GET returns 404 | High |
| DT-003 | Happy Path | Delete task and verify not in list | User owns task | 1. Delete task<br>2. GET all tasks | Deleted task not in list | High |

#### Authorization Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| DT-004 | Auth Error | Delete task without token | None | 1. DELETE without Authorization | Status: 401<br>Error: `missing_authorization` | High |
| DT-005 | Auth Error | Delete task with invalid token | None | 1. DELETE with invalid Bearer token | Status: 401<br>Error: `invalid_token` | High |
| DT-006 | Auth Error | Delete task owned by another user | Task belongs to User B | 1. Login as User A<br>2. DELETE User B's task | Status: 401<br>Error: `unauthorized`<br>Message: "You don't have access to this task" | High |

#### Not Found Scenarios (404)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| DT-007 | Not Found | Delete non-existent task | User authenticated | 1. DELETE `/api/v1/tasks/{non-existent-uuid}` | Status: 404<br>Error: `task_not_found` | High |
| DT-008 | Not Found | Delete already deleted task | Task was deleted | 1. Delete task<br>2. Delete same task again | Status: 404<br>Error: `task_not_found` | Medium |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| DT-009 | Edge Case | Delete with invalid UUID format | User authenticated | 1. DELETE `/api/v1/tasks/invalid-uuid` | Status: 400 or 404 | Medium |
| DT-010 | Edge Case | Delete with request body (should be ignored) | User owns task | 1. DELETE with body: `{"confirm": true}` | Status: 204 (body ignored) | Low |

---

### 3.6 Update Task Status

### Endpoint: `PATCH /api/v1/tasks/{id}/status`

#### Happy Path Scenarios

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| US-001 | Happy Path | Update status to in_progress | User owns task, status=pending | 1. PATCH to `/api/v1/tasks/{id}/status`<br>2. Body: `{"status": "in_progress"}` | Status: 200<br>Task status: "in_progress" | High |
| US-002 | Happy Path | Update status to done | User owns task | 1. PATCH with: `{"status": "done"}` | Status: 200<br>Task status: "done" | High |
| US-003 | Happy Path | Update status to pending | User owns task, status=done | 1. PATCH with: `{"status": "pending"}` | Status: 200<br>Task status: "pending" | High |
| US-004 | Happy Path | Verify updated_at changes | User owns task | 1. Note current updated_at<br>2. Update status<br>3. Check updated_at | updated_at is newer than before | High |
| US-005 | Happy Path | Verify other fields unchanged | User owns task | 1. Update status<br>2. Verify title, description, priority unchanged | Only status and updated_at changed | High |

#### Validation Error Scenarios (400)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| US-006 | Validation | Update status without status field | User owns task | 1. PATCH with: `{}` | Status: 400<br>Error: `validation_error`<br>Message: "Status is required" | High |
| US-007 | Validation | Update status with invalid value | User owns task | 1. PATCH with: `{"status": "completed"}` | Status: 400<br>Error: `validation_error`<br>Message: "Status must be pending, in_progress, or done" | High |
| US-008 | Validation | Update status with empty value | User owns task | 1. PATCH with: `{"status": ""}` | Status: 400<br>Error: `validation_error` | High |
| US-009 | Validation | Update status with null | User owns task | 1. PATCH with: `{"status": null}` | Status: 400<br>Error: `validation_error` | High |
| US-010 | Validation | Update status with invalid JSON | User owns task | 1. PATCH with malformed JSON | Status: 400<br>Error: `invalid_request` | High |
| US-011 | Validation | Update status with integer | User owns task | 1. PATCH with: `{"status": 1}` | Status: 400<br>Error: `validation_error` | Medium |
| US-012 | Validation | Update status with boolean | User owns task | 1. PATCH with: `{"status": true}` | Status: 400<br>Error: `validation_error` | Medium |

#### Authorization Error Scenarios (401)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| US-013 | Auth Error | Update status without token | None | 1. PATCH without Authorization | Status: 401<br>Error: `missing_authorization` | High |
| US-014 | Auth Error | Update status with invalid token | None | 1. PATCH with invalid Bearer token | Status: 401<br>Error: `invalid_token` | High |
| US-015 | Auth Error | Update status of another user's task | Task belongs to User B | 1. Login as User A<br>2. PATCH User B's task status | Status: 401<br>Error: `unauthorized`<br>Message: "You don't have access to this task" | High |

#### Not Found Scenarios (404)

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| US-016 | Not Found | Update status of non-existent task | User authenticated | 1. PATCH `/api/v1/tasks/{non-existent-uuid}/status` | Status: 404<br>Error: `task_not_found` | High |

#### Edge Cases

| TC ID | Category | Test Case | Preconditions | Test Steps | Expected Result | Priority |
|-------|----------|-----------|---------------|------------|-----------------|----------|
| US-017 | Edge Case | Update status to same value | Task status=pending | 1. PATCH with: `{"status": "pending"}` | Status: 200 (updated_at still changes) | Low |
| US-018 | Edge Case | Status with extra whitespace | User owns task | 1. PATCH with: `{"status": " done "}` | Status: 400 or 200 (if trimmed) | Medium |
| US-019 | Edge Case | Status with different case | User owns task | 1. PATCH with: `{"status": "DONE"}` | Status: 400 (case-sensitive) or 200 (if case-insensitive) | Medium |
| US-020 | Edge Case | Extra fields in request | User owns task | 1. PATCH with: `{"status": "done", "title": "New"}` | Status: 200 (extra fields ignored, only status changes) | Medium |

---

## 4. Cross-Endpoint Test Scenarios

### 4.1 End-to-End Workflows

| TC ID | Category | Test Case | Test Steps | Expected Result | Priority |
|-------|----------|-----------|------------|-----------------|----------|
| E2E-001 | Workflow | Complete user journey | 1. Register user<br>2. Login<br>3. Create task<br>4. Get all tasks<br>5. Update task<br>6. Update status to done<br>7. Delete task | All operations succeed | High |
| E2E-002 | Workflow | Multi-user isolation | 1. Register User A and B<br>2. Each creates tasks<br>3. Verify A only sees A's tasks<br>4. Verify B only sees B's tasks | Complete data isolation | High |


## 5. Test Data Requirements

### Test Users
| User | Email | Password | Purpose |
|------|-------|----------|---------|
| Primary | test.user@example.com | TestPass123! | Main test account |
| Secondary | secondary@example.com | SecondPass456! | Multi-user testing |
| Edge Case | edge.case+test@sub.domain.co.uk | P@$$w0rd!#% | Special characters |

### Test Tasks
| Task | Title | Description | Priority | Status |
|------|-------|-------------|----------|--------|
| Standard | Test Task | Standard test task | medium | pending |
| High Priority | Urgent Task | High priority test | high | pending |
| Completed | Done Task | Completed task | low | done |

---

## Summary Statistics

| Category | Count |
|----------|-------|
| Health Check Tests | 4 |
| Registration Tests | 22 |
| Login Tests | 15 |
| Get All Tasks Tests | 12 |
| Create Task Tests | 26 |
| Get Task by ID Tests | 10 |
| Update Task Tests | 16 |
| Delete Task Tests | 10 |
| Update Status Tests | 20 |
| E2E Tests | 2 |
| Concurrency Tests | 3 |
| Performance Tests | 3 |
| Security Tests | 8 |
| **Total Test Cases** | **151** |

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-14 | QA Expert Agent | Initial creation based on swagger.yaml analysis |