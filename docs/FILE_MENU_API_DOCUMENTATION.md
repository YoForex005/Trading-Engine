# File Menu API Documentation

## Overview
Complete backend API implementation for File menu operations including workspace management, print preferences, user data folder, and session management.

## Architecture

### Components
1. **Models** (`backend/models/workspace.go`)
   - Data structures for workspaces, configurations, and preferences
   - JSONB support for flexible configuration storage

2. **Database Migrations** (`backend/db/migrations/009_create_workspace_tables.go`)
   - Workspace tables with JSONB configuration storage
   - Print preferences table
   - User sessions table for unsaved changes tracking
   - Workspace history for versioning

3. **API Handlers**
   - `backend/api/workspace.go` - Workspace CRUD operations
   - `backend/api/user_preferences.go` - User preferences and session management

## API Endpoints

### 1. Workspace Management

#### Save Workspace
**POST** `/api/workspace/save`

Saves or updates a workspace configuration.

**Request Body:**
```json
{
  "name": "My Trading Setup",
  "description": "Daily trading configuration",
  "config": {
    "charts": [
      {
        "id": "chart-1",
        "symbol": "EURUSD",
        "timeframe": "H1",
        "chartType": "candlestick",
        "indicators": [
          {
            "id": "ema-20",
            "type": "ema",
            "parameters": {
              "period": 20
            },
            "style": {
              "color": "#2196F3",
              "lineWidth": 2,
              "opacity": 1.0
            }
          }
        ],
        "drawings": [
          {
            "id": "trendline-1",
            "type": "trendline",
            "symbol": "EURUSD",
            "timeframe": "H1",
            "points": [
              {"time": 1640000000000, "price": 1.1234},
              {"time": 1640086400000, "price": 1.1345}
            ],
            "properties": {
              "color": "#FF5722",
              "lineWidth": 2
            },
            "createdAt": "2024-01-15T10:30:00Z"
          }
        ]
      }
    ],
    "symbol": "EURUSD",
    "timeframe": "H1",
    "layout": {
      "sidebarOpen": true,
      "activePanels": ["market-watch", "chart", "orders"],
      "panelSizes": {
        "sidebar": 250,
        "marketWatch": 300
      }
    },
    "version": "1.0"
  },
  "isDefault": false,
  "overwrite": false
}
```

**Response (Success - Created):**
```json
{
  "success": true,
  "workspaceId": 123,
  "message": "Workspace saved successfully",
  "action": "created"
}
```

**Response (Conflict - Requires Confirmation):**
```json
{
  "error": "workspace_exists",
  "message": "Workspace with this name already exists",
  "workspaceId": 123,
  "requiresConfirmation": true
}
```

**Response (Success - Updated):**
```json
{
  "success": true,
  "workspaceId": 123,
  "message": "Workspace updated successfully",
  "action": "updated"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation failure
- `401 Unauthorized` - Missing or invalid authentication token
- `409 Conflict` - Workspace name already exists (when overwrite=false)
- `500 Internal Server Error` - Database or server error

---

#### Load Workspace
**GET** `/api/workspace/:id`

Retrieves a saved workspace configuration.

**URL Parameters:**
- `id` (required) - Workspace ID

**Response (Success):**
```json
{
  "id": 123,
  "userId": "demo-user",
  "name": "My Trading Setup",
  "description": "Daily trading configuration",
  "config": {
    "charts": [...],
    "symbol": "EURUSD",
    "timeframe": "H1",
    "layout": {...},
    "version": "1.0"
  },
  "isDefault": false,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T11:45:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid workspace ID
- `401 Unauthorized` - Missing or invalid authentication token
- `404 Not Found` - Workspace not found
- `500 Internal Server Error` - Database or server error

---

#### List Workspaces
**GET** `/api/workspaces`

Lists all workspaces for the authenticated user.

**Response (Success):**
```json
[
  {
    "id": 123,
    "name": "My Trading Setup",
    "description": "Daily trading configuration",
    "isDefault": false,
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T11:45:00Z"
  },
  {
    "id": 124,
    "name": "Default Workspace",
    "description": "Default configuration",
    "isDefault": true,
    "createdAt": "2024-01-10T09:00:00Z",
    "updatedAt": "2024-01-10T09:00:00Z"
  }
]
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Database or server error

---

#### Delete Workspace
**DELETE** `/api/workspace/:id`

Deletes a workspace.

**URL Parameters:**
- `id` (required) - Workspace ID

**Response (Success):**
```json
{
  "success": true,
  "message": "Workspace deleted successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid workspace ID
- `401 Unauthorized` - Missing or invalid authentication token
- `404 Not Found` - Workspace not found
- `500 Internal Server Error` - Database or server error

---

### 2. Print Preferences

#### Save Print Preferences
**POST** `/api/user/print-preferences`

Saves user print preferences.

**Request Body:**
```json
{
  "pageSize": "A4",
  "orientation": "portrait",
  "marginTop": 0.5,
  "marginBottom": 0.5,
  "marginLeft": 0.5,
  "marginRight": 0.5,
  "includeHeader": true,
  "includeFooter": true,
  "headerText": "Trading Report - {{date}}",
  "footerText": "Page {{page}} of {{totalPages}}"
}
```

**Valid Values:**
- `pageSize`: "A4", "Letter", "Legal"
- `orientation`: "portrait", "landscape"
- `marginTop/Bottom/Left/Right`: 0 to 5 (in inches)

**Response (Success):**
```json
{
  "success": true,
  "message": "Print preferences saved successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation failure
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Database or server error

---

#### Get Print Preferences
**GET** `/api/user/print-preferences`

Retrieves user print preferences.

**Response (Success):**
```json
{
  "userId": "demo-user",
  "pageSize": "A4",
  "orientation": "portrait",
  "marginTop": 0.5,
  "marginBottom": 0.5,
  "marginLeft": 0.5,
  "marginRight": 0.5,
  "includeHeader": true,
  "includeFooter": true,
  "headerText": "Trading Report - {{date}}",
  "footerText": "Page {{page}} of {{totalPages}}",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T11:45:00Z"
}
```

**Default Response (No Preferences Saved):**
```json
{
  "userId": "demo-user",
  "pageSize": "A4",
  "orientation": "portrait",
  "marginTop": 0.5,
  "marginBottom": 0.5,
  "marginLeft": 0.5,
  "marginRight": 0.5,
  "includeHeader": true,
  "includeFooter": true
}
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Database or server error

---

### 3. User Data Folder

#### Get User Data Folder
**GET** `/api/user/data-folder?platform=web|desktop`

Returns user data folder path or virtual file system structure.

**Query Parameters:**
- `platform` (optional) - "web" or "desktop" (default: "web")

**Response (Desktop):**
```json
{
  "platform": "desktop",
  "path": "/Users/username/Library/Application Support/TradingEngine/Users/demo-user",
  "os": "darwin"
}
```

**Response (Web):**
```json
{
  "platform": "web",
  "root": {
    "path": "/",
    "type": "directory",
    "modified": "2024-01-15T10:30:00Z",
    "children": [
      {
        "path": "/workspaces",
        "type": "directory",
        "modified": "2024-01-15T10:30:00Z",
        "children": []
      },
      {
        "path": "/templates",
        "type": "directory",
        "modified": "2024-01-15T10:30:00Z",
        "children": []
      },
      {
        "path": "/exports",
        "type": "directory",
        "modified": "2024-01-15T10:30:00Z",
        "children": []
      },
      {
        "path": "/drawings",
        "type": "directory",
        "modified": "2024-01-15T10:30:00Z",
        "children": []
      },
      {
        "path": "/indicators",
        "type": "directory",
        "modified": "2024-01-15T10:30:00Z",
        "children": []
      }
    ]
  }
}
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Server error

---

### 4. Session Management

#### Check Unsaved Changes
**GET** `/api/session/unsaved-changes?sessionId=xxx`

Checks if there are unsaved changes in the session.

**Query Parameters:**
- `sessionId` (required) - Session identifier

**Response (Has Unsaved Data):**
```json
{
  "hasUnsavedData": true,
  "workspace": {
    "name": "Unsaved Workspace",
    "config": {...}
  }
}
```

**Response (No Unsaved Data):**
```json
{
  "hasUnsavedData": false
}
```

**Error Responses:**
- `400 Bad Request` - Missing session ID
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Database or server error

---

#### Terminate Session
**DELETE** `/api/session`

Terminates a user session and cleans up resources.

**Request Body:**
```json
{
  "sessionId": "session-uuid-123"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Session terminated successfully",
  "hadUnsavedData": true,
  "unsavedWorkspaces": ["My Unsaved Workspace"]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Database or server error

---

## Database Schema

### workspaces table
```sql
CREATE TABLE workspaces (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
```

### print_preferences table
```sql
CREATE TABLE print_preferences (
    user_id VARCHAR(255) PRIMARY KEY,
    page_size VARCHAR(50) NOT NULL DEFAULT 'A4',
    orientation VARCHAR(50) NOT NULL DEFAULT 'portrait',
    margin_top DECIMAL(5, 2) NOT NULL DEFAULT 0.5,
    margin_bottom DECIMAL(5, 2) NOT NULL DEFAULT 0.5,
    margin_left DECIMAL(5, 2) NOT NULL DEFAULT 0.5,
    margin_right DECIMAL(5, 2) NOT NULL DEFAULT 0.5,
    include_header BOOLEAN DEFAULT TRUE,
    include_footer BOOLEAN DEFAULT TRUE,
    header_text TEXT,
    footer_text TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### user_sessions table
```sql
CREATE TABLE user_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unsaved_workspace JSONB,
    has_unsaved_data BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### workspace_history table (versioning)
```sql
CREATE TABLE workspace_history (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    config JSONB NOT NULL,
    change_description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

---

## Authentication

All endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

The JWT token should contain the user ID, which is extracted using `getUserIDFromRequest()`.

**TODO:** Integrate with existing `backend/auth/service.go` for proper JWT validation.

---

## Error Handling

All endpoints implement consistent error handling:

1. **Validation Errors** (400 Bad Request)
   - Invalid request body
   - Missing required fields
   - Invalid parameter values

2. **Authentication Errors** (401 Unauthorized)
   - Missing JWT token
   - Invalid or expired JWT token

3. **Authorization Errors** (403 Forbidden)
   - User attempting to access another user's resources

4. **Not Found Errors** (404 Not Found)
   - Workspace not found
   - User not found

5. **Conflict Errors** (409 Conflict)
   - Workspace name already exists (requires overwrite confirmation)

6. **Server Errors** (500 Internal Server Error)
   - Database connection failures
   - JSON marshaling/unmarshaling errors
   - Unexpected server errors

All errors are logged with appropriate context for debugging.

---

## CORS Configuration

All endpoints include CORS headers for cross-origin requests:

```go
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## Integration Guide

### 1. Database Setup

Run the migration:
```bash
cd backend
go run cmd/migrate/main.go up
```

### 2. Initialize Database Connection

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", connectionString)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### 3. Register Routes

In `backend/cmd/server/main.go`:

```go
import (
    "github.com/epic1st/rtx/backend/api"
)

func main() {
    // ... existing code ...

    // Initialize handlers
    workspaceHandler := api.NewWorkspaceHandler(db)
    preferencesHandler := api.NewUserPreferencesHandler(db)

    // Register routes
    workspaceHandler.RegisterRoutes(http.DefaultServeMux)
    preferencesHandler.RegisterRoutes(http.DefaultServeMux)

    // ... start server ...
}
```

### 4. Authentication Integration

Update `getUserIDFromRequest()` in `backend/api/workspace.go`:

```go
func getUserIDFromRequest(r *http.Request) string {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return ""
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    if tokenString == "" {
        return ""
    }

    // Use your auth service
    claims, err := authService.ValidateToken(tokenString)
    if err != nil {
        return ""
    }

    return claims.UserID
}
```

---

## Security Considerations

1. **JWT Validation**: All endpoints must validate JWT tokens before processing requests
2. **User Isolation**: Users can only access their own workspaces and preferences
3. **SQL Injection Prevention**: All queries use parameterized statements
4. **Input Validation**: All user input is validated before processing
5. **CORS**: Configure appropriate CORS policies for production
6. **Rate Limiting**: Consider implementing rate limiting for save operations
7. **Data Sanitization**: Sanitize user input in header/footer text fields

---

## Testing

### Example cURL Commands

**Save Workspace:**
```bash
curl -X POST http://localhost:7999/api/workspace/save \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Test Workspace",
    "description": "Test description",
    "config": {
      "symbol": "EURUSD",
      "timeframe": "H1",
      "charts": [],
      "layout": {
        "sidebarOpen": true,
        "activePanels": []
      },
      "version": "1.0"
    },
    "isDefault": false,
    "overwrite": false
  }'
```

**Load Workspace:**
```bash
curl -X GET http://localhost:7999/api/workspace/123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**List Workspaces:**
```bash
curl -X GET http://localhost:7999/api/workspaces \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Save Print Preferences:**
```bash
curl -X POST http://localhost:7999/api/user/print-preferences \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "pageSize": "A4",
    "orientation": "portrait",
    "marginTop": 0.5,
    "marginBottom": 0.5,
    "marginLeft": 0.5,
    "marginRight": 0.5,
    "includeHeader": true,
    "includeFooter": true
  }'
```

---

## Performance Considerations

1. **JSONB Indexing**: The `config` column is indexed with GIN for efficient JSON queries
2. **Connection Pooling**: Use database connection pooling for production
3. **Caching**: Consider caching frequently accessed workspaces
4. **Pagination**: Implement pagination for large workspace lists
5. **Compression**: Consider compressing large workspace configurations

---

## Future Enhancements

1. **Workspace Sharing**: Allow users to share workspaces with other users
2. **Workspace Templates**: Provide pre-configured workspace templates
3. **Version History**: Track and restore previous workspace versions
4. **Export/Import**: Export workspaces to files and import from files
5. **Cloud Sync**: Synchronize workspaces across devices
6. **Workspace Search**: Full-text search across workspace names and descriptions
7. **Workspace Tags**: Add tagging system for better organization

---

## Troubleshooting

### Common Issues

1. **"Unauthorized" errors**
   - Check JWT token is valid
   - Verify Authorization header format
   - Ensure token hasn't expired

2. **"Workspace not found" errors**
   - Verify workspace ID is correct
   - Check user has permission to access workspace
   - Ensure workspace wasn't deleted

3. **"Invalid request body" errors**
   - Verify JSON format is correct
   - Check required fields are present
   - Validate field types match schema

4. **Database connection errors**
   - Check database is running
   - Verify connection string is correct
   - Ensure migrations have been run

---

## Support

For issues or questions, please contact the development team or create an issue in the project repository.
