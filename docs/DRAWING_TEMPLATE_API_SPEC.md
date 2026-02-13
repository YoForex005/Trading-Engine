# Drawing Template API Specification

## Overview

The Drawing Template API allows users to save, load, and manage chart drawing configurations as reusable templates. This enables traders to quickly apply their favorite drawing setups across different charts and sessions.

## Base URL

```
/api/workspace/templates
```

## Authentication

All endpoints require authentication. The `accountId` parameter is used to scope templates to specific trading accounts.

---

## Endpoints

### 1. Create Template

**Endpoint:** `POST /api/workspace/templates`

**Description:** Save a new drawing template.

**Request Body:**

```json
{
  "id": "drawing_template_1707960453879_a8k2j9",
  "name": "Support & Resistance Setup",
  "description": "My favorite S/R levels with trendlines",
  "type": "drawing",
  "symbol": "EURUSD",
  "accountId": 1,
  "data": {
    "drawings": [
      {
        "id": "drawing-1707960123456",
        "type": "hline",
        "points": [
          {
            "time": 1707960000,
            "price": 1.0850
          }
        ],
        "color": "#3b82f6",
        "lineWidth": 2,
        "lineStyle": "solid",
        "selected": false,
        "locked": false
      },
      {
        "id": "drawing-1707960123457",
        "type": "trendline",
        "points": [
          {
            "time": 1707950000,
            "price": 1.0820
          },
          {
            "time": 1707960000,
            "price": 1.0860
          }
        ],
        "color": "#ef4444",
        "lineWidth": 2,
        "lineStyle": "solid"
      }
    ]
  },
  "createdAt": 1707960453879,
  "updatedAt": 1707960453879
}
```

**Response (200 OK):**

```json
{
  "id": "12345",
  "name": "Support & Resistance Setup",
  "description": "My favorite S/R levels with trendlines",
  "type": "drawing",
  "symbol": "EURUSD",
  "accountId": 1,
  "data": {
    "drawings": [...]
  },
  "createdAt": 1707960453879,
  "updatedAt": 1707960453879
}
```

**Error Responses:**

- `400 Bad Request` - Invalid request body or missing required fields
- `401 Unauthorized` - Authentication required
- `500 Internal Server Error` - Server error

---

### 2. List Templates

**Endpoint:** `GET /api/workspace/templates`

**Description:** Retrieve all templates for a specific type, symbol, and account.

**Query Parameters:**

| Parameter   | Type   | Required | Description                              |
|-------------|--------|----------|------------------------------------------|
| type        | string | Yes      | Template type (must be "drawing")        |
| symbol      | string | Yes      | Trading symbol (e.g., "EURUSD")          |
| accountId   | number | Yes      | Account ID to scope templates            |

**Example Request:**

```
GET /api/workspace/templates?type=drawing&symbol=EURUSD&accountId=1
```

**Response (200 OK):**

```json
[
  {
    "id": "12345",
    "name": "Support & Resistance Setup",
    "description": "My favorite S/R levels with trendlines",
    "type": "drawing",
    "symbol": "EURUSD",
    "accountId": 1,
    "data": {
      "drawings": [...]
    },
    "createdAt": 1707960453879,
    "updatedAt": 1707960453879
  },
  {
    "id": "12346",
    "name": "Fibonacci Levels",
    "description": "Standard fib retracement setup",
    "type": "drawing",
    "symbol": "EURUSD",
    "accountId": 1,
    "data": {
      "drawings": [...]
    },
    "createdAt": 1707960500000,
    "updatedAt": 1707960500000
  }
]
```

**Error Responses:**

- `400 Bad Request` - Missing or invalid query parameters
- `401 Unauthorized` - Authentication required
- `500 Internal Server Error` - Server error

---

### 3. Delete Template

**Endpoint:** `DELETE /api/workspace/templates/:id`

**Description:** Delete a specific template by ID.

**URL Parameters:**

| Parameter | Type   | Description         |
|-----------|--------|---------------------|
| id        | string | Template ID         |

**Query Parameters:**

| Parameter   | Type   | Required | Description                      |
|-------------|--------|----------|----------------------------------|
| symbol      | string | Yes      | Trading symbol                   |
| accountId   | number | Yes      | Account ID for authorization     |

**Example Request:**

```
DELETE /api/workspace/templates/12345?symbol=EURUSD&accountId=1
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Template deleted successfully",
  "id": "12345"
}
```

**Error Responses:**

- `400 Bad Request` - Missing query parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Template belongs to different account
- `404 Not Found` - Template not found
- `500 Internal Server Error` - Server error

---

## Database Schema

### Table: `workspace_templates`

```sql
CREATE TABLE workspace_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    account_id INTEGER NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Indexes
    INDEX idx_templates_type_symbol_account (type, symbol, account_id),
    INDEX idx_templates_account (account_id),
    INDEX idx_templates_created (created_at DESC)
);
```

### Column Descriptions:

- `id`: Auto-incrementing primary key
- `name`: User-defined template name (required)
- `description`: Optional description of the template
- `type`: Template type (e.g., "drawing", "chart", "workspace")
- `symbol`: Trading symbol this template is associated with
- `account_id`: Foreign key to the account that owns this template
- `data`: JSONB field containing the template data (drawings, indicators, etc.)
- `created_at`: Timestamp when template was created
- `updated_at`: Timestamp when template was last modified

---

## Drawing Data Structure

### Drawing Object

Each drawing in the `data.drawings` array should follow this structure:

```typescript
interface Drawing {
  id: string;                    // Unique identifier
  type: DrawingType;             // Drawing type (see below)
  subtype?: string;              // Optional subtype for shapes
  points: DrawingPoint[];        // Array of coordinate points
  text?: string;                 // Text content (for text drawings)
  color?: string;                // Line/shape color (hex)
  lineWidth?: number;            // Line width in pixels
  lineStyle?: 'solid' | 'dashed' | 'dotted';
  selected?: boolean;            // Selection state
  locked?: boolean;              // Lock state (prevent editing)
}

interface DrawingPoint {
  time: number;                  // Unix timestamp
  price: number;                 // Price level
}

type DrawingType =
  | 'trendline'
  | 'hline'
  | 'vline'
  | 'text'
  | 'channel'
  | 'fibonacci'
  | 'shapes'
  | 'rectangle'
  | 'ellipse'
  | 'arrow'
  | 'pitchfork';
```

---

## Implementation Notes

### Backend Requirements

1. **Authorization**: Ensure users can only access templates for accounts they own
2. **Validation**: Validate that `type` is "drawing" for drawing templates
3. **JSONB Indexing**: Consider adding GIN indexes on `data` for faster queries if needed
4. **Soft Deletes**: Consider implementing soft deletes instead of hard deletes
5. **Audit Trail**: Log template creation/modification/deletion for compliance

### Frontend Integration

The frontend uses these endpoints through the `drawingTemplateManager` service:

- **Save**: Calls `POST /api/workspace/templates` with template data
- **Load**: Calls `GET /api/workspace/templates?type=drawing&symbol=X&accountId=Y`
- **Delete**: Calls `DELETE /api/workspace/templates/:id`
- **Fallback**: Uses localStorage if backend is unavailable

### Error Handling

The frontend gracefully handles backend errors by:

1. Displaying user-friendly error messages
2. Falling back to localStorage for persistence
3. Retrying failed requests on network errors
4. Logging errors for debugging

---

## Testing Recommendations

### Unit Tests

- Validate request body schema
- Test CRUD operations
- Test authorization logic
- Test error handling

### Integration Tests

- Test full save/load/delete flow
- Test cross-account isolation
- Test concurrent modifications
- Test large template data

### Performance Tests

- Load test with 100+ templates per account
- Test JSONB query performance
- Test pagination for large result sets

---

## Future Enhancements

1. **Template Sharing**: Allow users to share templates with other accounts
2. **Template Categories**: Add category/tag support for organization
3. **Version History**: Track template versions and allow rollback
4. **Import/Export**: Bulk import/export across accounts
5. **Template Marketplace**: Public template library
6. **Favorites**: Mark frequently used templates as favorites
7. **Search**: Full-text search across template names and descriptions
8. **Pagination**: Implement pagination for large template lists

---

## Contact

For questions or issues with this API, please contact the backend development team.
