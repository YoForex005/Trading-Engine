# Drawing Template System - Quick Start Guide

## 📋 Overview

The Drawing Template System allows traders to save, load, and manage their chart drawing configurations as reusable templates. This saves time by enabling quick application of favorite drawing setups across different charts.

---

## 🚀 Quick Start (5 Minutes)

### For End Users

#### 1. Save Your First Template
1. Draw some objects on your chart (trendlines, levels, shapes, etc.)
2. Click the **folder icon** 📁 in the toolbar (next to the trash icon)
3. Click **"Save Current Drawings"**
4. Enter a name (e.g., "Support & Resistance Setup")
5. Optionally add a description
6. Click **"Save Template"**

#### 2. Load a Template
1. Click the **folder icon** 📁 in the toolbar
2. Click on any template in the list
3. Confirm replacement of current drawings
4. Your drawings will appear on the chart!

#### 3. Delete a Template
1. Click the **folder icon** 📁 in the toolbar
2. Hover over a template
3. Click the **trash icon** 🗑️
4. Confirm deletion

---

## 📁 File Locations

### Frontend Files (Already Implemented ✅)

```
clients/desktop/src/
├── services/
│   └── drawingTemplateManager.ts       ← Template service
├── components/
│   └── layout/
│       ├── DrawingTemplatesDropdown.tsx ← UI component
│       └── TopToolbar.tsx              ← Integration point
└── config/
    └── api.ts                          ← API endpoints

docs/
├── DRAWING_TEMPLATE_API_SPEC.md        ← Backend API spec
├── DRAWING_TEMPLATE_IMPLEMENTATION_SUMMARY.md
├── DRAWING_TEMPLATE_VERIFICATION.md
└── DRAWING_TEMPLATE_QUICK_START.md     ← This file
```

---

## 🔧 Backend Implementation (Required)

### Step 1: Create Database Table

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

    INDEX idx_templates_type_symbol_account (type, symbol, account_id),
    INDEX idx_templates_account (account_id),
    INDEX idx_templates_created (created_at DESC)
);
```

### Step 2: Implement API Endpoints

#### POST /api/workspace/templates
```go
// Create template
func CreateTemplate(c *gin.Context) {
    var template Template
    if err := c.ShouldBindJSON(&template); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Validate
    if template.Type != "drawing" {
        c.JSON(400, gin.H{"error": "Invalid template type"})
        return
    }

    // Save to database
    result := db.Create(&template)
    if result.Error != nil {
        c.JSON(500, gin.H{"error": result.Error.Error()})
        return
    }

    c.JSON(200, template)
}
```

#### GET /api/workspace/templates
```go
// List templates
func ListTemplates(c *gin.Context) {
    templateType := c.Query("type")
    symbol := c.Query("symbol")
    accountId := c.Query("accountId")

    var templates []Template
    result := db.Where("type = ? AND symbol = ? AND account_id = ?",
        templateType, symbol, accountId).Find(&templates)

    if result.Error != nil {
        c.JSON(500, gin.H{"error": result.Error.Error()})
        return
    }

    c.JSON(200, templates)
}
```

#### DELETE /api/workspace/templates/:id
```go
// Delete template
func DeleteTemplate(c *gin.Context) {
    id := c.Param("id")
    symbol := c.Query("symbol")
    accountId := c.Query("accountId")

    result := db.Where("id = ? AND symbol = ? AND account_id = ?",
        id, symbol, accountId).Delete(&Template{})

    if result.Error != nil {
        c.JSON(500, gin.H{"error": result.Error.Error()})
        return
    }

    c.JSON(200, gin.H{"success": true, "id": id})
}
```

### Step 3: Register Routes

```go
// In your router setup
api.POST("/workspace/templates", CreateTemplate)
api.GET("/workspace/templates", ListTemplates)
api.DELETE("/workspace/templates/:id", DeleteTemplate)
```

---

## 🧪 Testing

### Manual Test Plan

1. **Save Template**
   - Draw 3-5 objects on chart
   - Save as template "Test Template 1"
   - Verify template appears in list
   - Check browser localStorage has data

2. **Load Template**
   - Clear all drawings
   - Load "Test Template 1"
   - Verify all drawings appear correctly
   - Check positions match original

3. **Delete Template**
   - Delete "Test Template 1"
   - Verify removed from list
   - Check localStorage cleared

4. **Export/Import**
   - Save 2-3 templates
   - Export to JSON
   - Clear all templates
   - Import from JSON
   - Verify templates restored

5. **Backend Integration** (once available)
   - Save template
   - Refresh page
   - Verify template persists (loaded from backend)
   - Test on different device
   - Verify template available

### API Test (with curl)

```bash
# Create template
curl -X POST http://localhost:7999/workspace/templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Template",
    "description": "Test description",
    "type": "drawing",
    "symbol": "EURUSD",
    "accountId": 1,
    "data": {
      "drawings": []
    }
  }'

# List templates
curl "http://localhost:7999/workspace/templates?type=drawing&symbol=EURUSD&accountId=1"

# Delete template
curl -X DELETE "http://localhost:7999/workspace/templates/1?symbol=EURUSD&accountId=1"
```

---

## 📊 Usage Statistics

Track these metrics:
- Templates created per user
- Most loaded templates
- Average drawings per template
- Export/import frequency
- Template deletion rate

---

## 🐛 Common Issues

### Issue: Template not saving
**Cause:** Backend API not implemented
**Solution:** Check browser console, verify localStorage fallback working
**Workaround:** Export templates as JSON backup

### Issue: Template not loading
**Cause:** Symbol mismatch or corrupt data
**Solution:** Check template symbol matches chart symbol
**Workaround:** Delete and recreate template

### Issue: Drawings disappear after load
**Cause:** Chart not initialized or DrawingManager error
**Solution:** Check console for errors, ensure chart is loaded
**Workaround:** Refresh page and try again

---

## 💡 Pro Tips

1. **Naming Convention:** Use descriptive names like "EURUSD Daily S/R" instead of "Template 1"
2. **Descriptions:** Add notes about when/why to use each template
3. **Regular Exports:** Export templates weekly as backup
4. **Template Library:** Build a collection of templates for different strategies
5. **Cleanup:** Delete unused templates monthly to keep list manageable

---

## 🔮 Future Enhancements

Planned features:
- Template preview thumbnails
- Template categories/tags
- Template sharing between accounts
- Template marketplace
- Auto-save on symbol change
- Template versioning
- Favorite templates
- Search and filter

---

## 📚 Additional Resources

- **Full API Spec:** `docs/DRAWING_TEMPLATE_API_SPEC.md`
- **Implementation Details:** `docs/DRAWING_TEMPLATE_IMPLEMENTATION_SUMMARY.md`
- **Verification Checklist:** `docs/DRAWING_TEMPLATE_VERIFICATION.md`

---

## 🆘 Support

For issues or questions:
1. Check this quick start guide
2. Review the verification checklist
3. Check browser console for errors
4. Contact development team

---

## ✅ Checklist for Go-Live

Frontend:
- [x] Service implemented
- [x] UI component implemented
- [x] Integration complete
- [x] localStorage working
- [x] Documentation complete

Backend:
- [ ] Database table created
- [ ] API endpoints implemented
- [ ] Authorization added
- [ ] Testing complete
- [ ] Deployed to production

---

**Status:** Frontend ✅ Complete | Backend ⏳ Pending

**Last Updated:** 2026-02-13
