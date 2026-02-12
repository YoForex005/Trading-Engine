package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Types
// ============================================

type BackupType string

const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
	BackupTypeConfigOnly   BackupType = "config_only"
	BackupTypeDatabaseOnly BackupType = "database_only"
)

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
	BackupStatusExpired    BackupStatus = "expired"
)

type BackupFrequency string

const (
	FrequencyHourly  BackupFrequency = "hourly"
	FrequencyDaily   BackupFrequency = "daily"
	FrequencyWeekly  BackupFrequency = "weekly"
	FrequencyMonthly BackupFrequency = "monthly"
)

type RestoreStatus string

const (
	RestoreStatusPending    RestoreStatus = "pending"
	RestoreStatusInProgress RestoreStatus = "in_progress"
	RestoreStatusCompleted  RestoreStatus = "completed"
	RestoreStatusFailed     RestoreStatus = "failed"
)

// ============================================
// Data Structures
// ============================================

type Backup struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Type        BackupType          `json:"type"`
	Status      BackupStatus        `json:"status"`
	Size        int64               `json:"size"` // bytes
	CreatedAt   time.Time           `json:"createdAt"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
	CreatedBy   string              `json:"createdBy"`
	StoragePath string              `json:"storagePath"`
	Retention   int                 `json:"retention"` // days
	Components  []string            `json:"components"`
	Error       string              `json:"error,omitempty"`
	Duration    int                 `json:"duration"` // seconds
}

type BackupSchedule struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Frequency     BackupFrequency `json:"frequency"`
	NextRun       time.Time       `json:"nextRun"`
	LastRun       *time.Time      `json:"lastRun,omitempty"`
	Type          BackupType      `json:"type"`
	RetentionDays int             `json:"retentionDays"`
	IsActive      bool            `json:"isActive"`
	Components    []string        `json:"components"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type RestorePoint struct {
	ID          string        `json:"id"`
	BackupID    string        `json:"backupId"`
	Timestamp   time.Time     `json:"timestamp"`
	Description string        `json:"description"`
	Status      RestoreStatus `json:"status"`
	RestoredBy  string        `json:"restoredBy"`
	RestoredAt  time.Time     `json:"restoredAt"`
	Duration    int           `json:"duration"` // seconds
	Error       string        `json:"error,omitempty"`
}

type BackupComponent struct {
	Name       string     `json:"name"`
	Size       int64      `json:"size"` // bytes
	ItemCount  int        `json:"itemCount"`
	LastBackup *time.Time `json:"lastBackup,omitempty"`
	Enabled    bool       `json:"enabled"`
}

type BackupStats struct {
	TotalBackups     int       `json:"totalBackups"`
	StorageUsed      int64     `json:"storageUsed"` // bytes
	LastSuccessful   time.Time `json:"lastSuccessful"`
	NextScheduled    time.Time `json:"nextScheduled"`
	FailedLast24h    int       `json:"failedLast24h"`
	CompletedLast24h int       `json:"completedLast24h"`
	AverageSize      int64     `json:"averageSize"`
	AverageDuration  int       `json:"averageDuration"` // seconds
}

// ============================================
// Store
// ============================================

type BackupStore struct {
	mu            sync.RWMutex
	backups       map[string]*Backup
	schedules     map[string]*BackupSchedule
	restorePoints map[string]*RestorePoint
	components    map[string]*BackupComponent
}

func NewBackupStore() *BackupStore {
	store := &BackupStore{
		backups:       make(map[string]*Backup),
		schedules:     make(map[string]*BackupSchedule),
		restorePoints: make(map[string]*RestorePoint),
		components:    make(map[string]*BackupComponent),
	}

	// Generate mock data
	store.generateMockData()

	return store
}

func (s *BackupStore) generateMockData() {
	now := time.Now()

	// System components
	components := []struct {
		name      string
		size      int64
		itemCount int
	}{
		{"database", 2147483648, 150000},      // 2GB
		{"config", 52428800, 350},             // 50MB
		{"ticks", 10737418240, 5000000},       // 10GB
		{"logs", 1073741824, 250000},          // 1GB
		{"uploads", 5368709120, 8500},         // 5GB
		{"users", 104857600, 2000},            // 100MB
		{"reports", 536870912, 1200},          // 500MB
		{"cache", 268435456, 50000},           // 250MB
	}

	for _, comp := range components {
		lastBackup := now.Add(-time.Duration(rand.Intn(48)) * time.Hour)
		s.components[comp.name] = &BackupComponent{
			Name:       comp.name,
			Size:       comp.size,
			ItemCount:  comp.itemCount,
			LastBackup: &lastBackup,
			Enabled:    true,
		}
	}

	// 30 backup records (last 30 days)
	backupTypes := []BackupType{
		BackupTypeFull, BackupTypeIncremental, BackupTypeDifferential,
		BackupTypeConfigOnly, BackupTypeDatabaseOnly,
	}
	statuses := []BackupStatus{
		BackupStatusCompleted, BackupStatusCompleted, BackupStatusCompleted,
		BackupStatusCompleted, BackupStatusFailed, BackupStatusExpired,
	}
	admins := []string{"admin", "system", "backup-service", "admin2"}

	for i := 1; i <= 30; i++ {
		backupType := backupTypes[rand.Intn(len(backupTypes))]
		status := statuses[rand.Intn(len(statuses))]
		createdAt := now.Add(-time.Duration(30-i) * 24 * time.Hour).Add(-time.Duration(rand.Intn(12)) * time.Hour)

		// Determine components based on type
		var selectedComponents []string
		switch backupType {
		case BackupTypeFull:
			selectedComponents = []string{"database", "config", "ticks", "logs", "uploads", "users", "reports"}
		case BackupTypeIncremental:
			selectedComponents = []string{"database", "ticks", "logs"}
		case BackupTypeDifferential:
			selectedComponents = []string{"database", "config", "ticks"}
		case BackupTypeConfigOnly:
			selectedComponents = []string{"config"}
		case BackupTypeDatabaseOnly:
			selectedComponents = []string{"database"}
		}

		// Calculate size based on components
		var totalSize int64
		for _, compName := range selectedComponents {
			if comp, ok := s.components[compName]; ok {
				if backupType == BackupTypeIncremental {
					totalSize += comp.Size / 10 // Incremental is ~10% of full
				} else if backupType == BackupTypeDifferential {
					totalSize += comp.Size / 4 // Differential is ~25% of full
				} else {
					totalSize += comp.Size
				}
			}
		}

		duration := 60 + rand.Intn(600) // 1-10 minutes
		var completedAt *time.Time
		var errorMsg string

		if status == BackupStatusCompleted {
			completed := createdAt.Add(time.Duration(duration) * time.Second)
			completedAt = &completed
		} else if status == BackupStatusFailed {
			completed := createdAt.Add(time.Duration(30+rand.Intn(60)) * time.Second)
			completedAt = &completed
			errorMsg = []string{
				"Disk space insufficient",
				"Network connection lost",
				"Database locked",
				"Permission denied",
			}[rand.Intn(4)]
		} else if status == BackupStatusExpired {
			completed := createdAt.Add(time.Duration(duration) * time.Second)
			completedAt = &completed
		}

		backup := &Backup{
			ID:          fmt.Sprintf("backup-%d", i),
			Name:        fmt.Sprintf("%s-backup-%s", backupType, createdAt.Format("2006-01-02")),
			Type:        backupType,
			Status:      status,
			Size:        totalSize,
			CreatedAt:   createdAt,
			CompletedAt: completedAt,
			CreatedBy:   admins[rand.Intn(len(admins))],
			StoragePath: fmt.Sprintf("/backups/%d/%s/%s", createdAt.Year(), createdAt.Format("01"), fmt.Sprintf("backup-%d.tar.gz", i)),
			Retention:   []int{7, 14, 30, 90, 180}[rand.Intn(5)],
			Components:  selectedComponents,
			Error:       errorMsg,
			Duration:    duration,
		}

		s.backups[backup.ID] = backup
	}

	// 5 backup schedules
	schedules := []struct {
		name      string
		frequency BackupFrequency
		type_     BackupType
		retention int
		active    bool
		components []string
	}{
		{"Daily Full Backup", FrequencyDaily, BackupTypeFull, 30, true, []string{"database", "config", "ticks", "logs", "uploads", "users", "reports"}},
		{"Hourly Incremental", FrequencyHourly, BackupTypeIncremental, 7, true, []string{"database", "ticks", "logs"}},
		{"Weekly Differential", FrequencyWeekly, BackupTypeDifferential, 90, true, []string{"database", "config", "ticks"}},
		{"Monthly Archive", FrequencyMonthly, BackupTypeFull, 365, true, []string{"database", "config", "ticks", "logs", "uploads", "users", "reports"}},
		{"Config Backup", FrequencyDaily, BackupTypeConfigOnly, 14, false, []string{"config"}},
	}

	for i, sched := range schedules {
		lastRun := now.Add(-time.Duration(rand.Intn(24)) * time.Hour)
		var nextRun time.Time

		switch sched.frequency {
		case FrequencyHourly:
			nextRun = now.Add(time.Hour)
		case FrequencyDaily:
			nextRun = now.Add(24 * time.Hour)
		case FrequencyWeekly:
			nextRun = now.Add(7 * 24 * time.Hour)
		case FrequencyMonthly:
			nextRun = now.Add(30 * 24 * time.Hour)
		}

		schedule := &BackupSchedule{
			ID:            fmt.Sprintf("schedule-%d", i+1),
			Name:          sched.name,
			Frequency:     sched.frequency,
			NextRun:       nextRun,
			LastRun:       &lastRun,
			Type:          sched.type_,
			RetentionDays: sched.retention,
			IsActive:      sched.active,
			Components:    sched.components,
			CreatedAt:     now.Add(-90 * 24 * time.Hour),
			UpdatedAt:     now.Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		}

		s.schedules[schedule.ID] = schedule
	}

	// 10 restore points
	completedBackups := []string{}
	for id, backup := range s.backups {
		if backup.Status == BackupStatusCompleted {
			completedBackups = append(completedBackups, id)
		}
	}

	for i := 1; i <= 10; i++ {
		if len(completedBackups) == 0 {
			break
		}
		backupID := completedBackups[rand.Intn(len(completedBackups))]
		backup := s.backups[backupID]

		restoredAt := now.Add(-time.Duration(rand.Intn(180)) * 24 * time.Hour)
		status := RestoreStatusCompleted
		var errorMsg string

		if i > 8 { // Last 2 failed
			status = RestoreStatusFailed
			errorMsg = []string{
				"Restore target not available",
				"Backup file corrupted",
			}[rand.Intn(2)]
		}

		restorePoint := &RestorePoint{
			ID:          fmt.Sprintf("restore-%d", i),
			BackupID:    backupID,
			Timestamp:   restoredAt,
			Description: fmt.Sprintf("Restore from %s backup", backup.Type),
			Status:      status,
			RestoredBy:  admins[rand.Intn(len(admins))],
			RestoredAt:  restoredAt,
			Duration:    300 + rand.Intn(900), // 5-20 minutes
			Error:       errorMsg,
		}

		s.restorePoints[restorePoint.ID] = restorePoint
	}

	log.Printf("[BackupStore] Mock data generated: %d backups, %d schedules, %d components, %d restore points",
		len(s.backups), len(s.schedules), len(s.components), len(s.restorePoints))
}

// ============================================
// Handler
// ============================================

type BackupHandler struct {
	store       *BackupStore
	authService *auth.Service
}

func NewBackupHandler(store *BackupStore, authService *auth.Service) *BackupHandler {
	return &BackupHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// 1. GET /admin/backups — List all backups
// ============================================

func (h *BackupHandler) HandleListBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JWT validation (admin only)
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	// Query params for filtering
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")

	backups := []*Backup{}
	for _, backup := range h.store.backups {
		if typeFilter != "" && string(backup.Type) != typeFilter {
			continue
		}
		if statusFilter != "" && string(backup.Status) != statusFilter {
			continue
		}
		backups = append(backups, backup)
	}

	// Sort by created date (newest first)
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].CreatedAt.Before(backups[j].CreatedAt) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
		"total":   len(backups),
	})
}

// ============================================
// 2. POST /admin/backups/create — Trigger manual backup
// ============================================

func (h *BackupHandler) HandleCreateBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JWT validation
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Type       BackupType `json:"type"`
		Components []string   `json:"components"`
		Name       string     `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	now := time.Now()
	backupID := fmt.Sprintf("backup-%d", len(h.store.backups)+1)

	// Calculate size
	var totalSize int64
	for _, compName := range req.Components {
		if comp, ok := h.store.components[compName]; ok {
			totalSize += comp.Size
		}
	}

	backup := &Backup{
		ID:          backupID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      BackupStatusPending,
		Size:        totalSize,
		CreatedAt:   now,
		CreatedBy:   "admin",
		StoragePath: fmt.Sprintf("/backups/%d/%s/%s.tar.gz", now.Year(), now.Format("01"), backupID),
		Retention:   30,
		Components:  req.Components,
		Duration:    0,
	}

	h.store.backups[backupID] = backup

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backup)
}

// ============================================
// 3. GET /admin/backups/:id — Backup details
// ============================================

func (h *BackupHandler) HandleGetBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/backups/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Backup ID required", http.StatusBadRequest)
		return
	}
	backupID := parts[0]

	h.store.mu.RLock()
	backup, exists := h.store.backups[backupID]
	h.store.mu.RUnlock()

	if !exists {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	// Build component breakdown
	h.store.mu.RLock()
	componentDetails := []map[string]interface{}{}
	for _, compName := range backup.Components {
		if comp, ok := h.store.components[compName]; ok {
			componentDetails = append(componentDetails, map[string]interface{}{
				"name":      comp.Name,
				"size":      comp.Size,
				"itemCount": comp.ItemCount,
			})
		}
	}
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backup":     backup,
		"components": componentDetails,
	})
}

// ============================================
// 4. DELETE /admin/backups/:id — Delete backup
// ============================================

func (h *BackupHandler) HandleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/backups/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Backup ID required", http.StatusBadRequest)
		return
	}
	backupID := parts[0]

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	if _, exists := h.store.backups[backupID]; !exists {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	delete(h.store.backups, backupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Backup deleted successfully",
	})
}

// ============================================
// 5. GET /admin/backups/schedules — List schedules
// ============================================

func (h *BackupHandler) HandleListSchedules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	schedules := []*BackupSchedule{}
	for _, schedule := range h.store.schedules {
		schedules = append(schedules, schedule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schedules": schedules,
		"total":     len(schedules),
	})
}

// ============================================
// 6. PUT /admin/backups/schedules/:id — Update schedule
// ============================================

func (h *BackupHandler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/backups/schedules/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Schedule ID required", http.StatusBadRequest)
		return
	}
	scheduleID := parts[0]

	var req struct {
		Frequency     *BackupFrequency `json:"frequency,omitempty"`
		RetentionDays *int             `json:"retentionDays,omitempty"`
		IsActive      *bool            `json:"isActive,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	schedule, exists := h.store.schedules[scheduleID]
	if !exists {
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Frequency != nil {
		schedule.Frequency = *req.Frequency
	}
	if req.RetentionDays != nil {
		schedule.RetentionDays = *req.RetentionDays
	}
	if req.IsActive != nil {
		schedule.IsActive = *req.IsActive
	}
	schedule.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// ============================================
// 7. GET /admin/backups/components — System components
// ============================================

func (h *BackupHandler) HandleGetComponents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	components := []*BackupComponent{}
	for _, component := range h.store.components {
		components = append(components, component)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"components": components,
		"total":      len(components),
	})
}

// ============================================
// 8. POST /admin/backups/:id/restore — Initiate restore
// ============================================

func (h *BackupHandler) HandleRestoreBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID
	pathParts := strings.Split(r.URL.Path, "/")
	var backupID string
	for i, part := range pathParts {
		if part == "backups" && i+1 < len(pathParts) {
			backupID = pathParts[i+1]
			break
		}
	}

	if backupID == "" || backupID == "create" || backupID == "schedules" || backupID == "components" || backupID == "stats" {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	backup, exists := h.store.backups[backupID]
	if !exists {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	if backup.Status != BackupStatusCompleted {
		http.Error(w, "Cannot restore from incomplete backup", http.StatusBadRequest)
		return
	}

	// Create restore point
	now := time.Now()
	restoreID := fmt.Sprintf("restore-%d", len(h.store.restorePoints)+1)

	restorePoint := &RestorePoint{
		ID:          restoreID,
		BackupID:    backupID,
		Timestamp:   now,
		Description: fmt.Sprintf("Restore from %s backup", backup.Type),
		Status:      RestoreStatusPending,
		RestoredBy:  "admin",
		RestoredAt:  now,
		Duration:    0,
	}

	h.store.restorePoints[restoreID] = restorePoint

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restorePoint)
}

// ============================================
// 9. GET /admin/backups/stats — Backup statistics
// ============================================

func (h *BackupHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	var totalBackups int
	var storageUsed int64
	var lastSuccessful time.Time
	var failedLast24h int
	var completedLast24h int
	var totalDuration int
	var completedCount int

	for _, backup := range h.store.backups {
		totalBackups++
		storageUsed += backup.Size

		if backup.Status == BackupStatusCompleted {
			if backup.CompletedAt.After(lastSuccessful) {
				lastSuccessful = *backup.CompletedAt
			}
			totalDuration += backup.Duration
			completedCount++

			if backup.CompletedAt.After(yesterday) {
				completedLast24h++
			}
		}

		if backup.Status == BackupStatusFailed && backup.CreatedAt.After(yesterday) {
			failedLast24h++
		}
	}

	// Find next scheduled backup
	var nextScheduled time.Time
	for _, schedule := range h.store.schedules {
		if schedule.IsActive {
			if nextScheduled.IsZero() || schedule.NextRun.Before(nextScheduled) {
				nextScheduled = schedule.NextRun
			}
		}
	}

	averageSize := int64(0)
	if totalBackups > 0 {
		averageSize = storageUsed / int64(totalBackups)
	}

	averageDuration := 0
	if completedCount > 0 {
		averageDuration = totalDuration / completedCount
	}

	stats := BackupStats{
		TotalBackups:     totalBackups,
		StorageUsed:      storageUsed,
		LastSuccessful:   lastSuccessful,
		NextScheduled:    nextScheduled,
		FailedLast24h:    failedLast24h,
		CompletedLast24h: completedLast24h,
		AverageSize:      averageSize,
		AverageDuration:  averageDuration,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
