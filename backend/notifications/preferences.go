package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PreferencesStore interface for storing user preferences
type PreferencesStore interface {
	Get(ctx context.Context, userID string) (*UserPreferences, error)
	Save(ctx context.Context, prefs *UserPreferences) error
	Delete(ctx context.Context, userID string) error
}

// InMemoryPreferencesStore implements an in-memory preferences store
type InMemoryPreferencesStore struct {
	preferences map[string]*UserPreferences
	mu          sync.RWMutex
}

// NewInMemoryPreferencesStore creates a new in-memory preferences store
func NewInMemoryPreferencesStore() *InMemoryPreferencesStore {
	return &InMemoryPreferencesStore{
		preferences: make(map[string]*UserPreferences),
	}
}

// Get retrieves user preferences
func (s *InMemoryPreferencesStore) Get(ctx context.Context, userID string) (*UserPreferences, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefs, exists := s.preferences[userID]
	if !exists {
		return s.getDefaultPreferences(userID), nil
	}

	return prefs, nil
}

// Save saves user preferences
func (s *InMemoryPreferencesStore) Save(ctx context.Context, prefs *UserPreferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefs.UpdatedAt = time.Now()
	s.preferences[prefs.UserID] = prefs
	return nil
}

// Delete removes user preferences
func (s *InMemoryPreferencesStore) Delete(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.preferences, userID)
	return nil
}

// getDefaultPreferences returns default notification preferences
func (s *InMemoryPreferencesStore) getDefaultPreferences(userID string) *UserPreferences {
	prefs := &UserPreferences{
		UserID:         userID,
		Preferences:    make(map[NotificationType]ChannelPreference),
		Locale:         "en",
		Timezone:       "UTC",
		UnsubscribeAll: false,
		UpdatedAt:      time.Now(),
	}

	// Default preferences for each notification type
	defaultPrefs := map[NotificationType]ChannelPreference{
		// Critical notifications - all channels enabled
		NotifMarginCallWarning: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelEmail, ChannelSMS, ChannelPush, ChannelInApp},
			MinimumPriority: PriorityCritical,
		},
		NotifStopOut: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelEmail, ChannelSMS, ChannelPush, ChannelInApp},
			MinimumPriority: PriorityCritical,
		},
		NotifSecurityAlert: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelEmail, ChannelPush, ChannelInApp},
			MinimumPriority: PriorityHigh,
		},

		// High priority notifications
		NotifOrderExecuted: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelPush, ChannelInApp, ChannelEmail},
			MinimumPriority: PriorityHigh,
		},
		NotifPositionClosed: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelPush, ChannelInApp, ChannelEmail},
			MinimumPriority: PriorityHigh,
		},
		NotifLoginNewDevice: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelEmail, ChannelPush},
			MinimumPriority: PriorityHigh,
		},

		// Normal priority notifications
		NotifBalanceChange: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelInApp, ChannelEmail},
			MinimumPriority: PriorityNormal,
		},
		NotifPriceMovement: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelPush, ChannelInApp},
			MinimumPriority: PriorityNormal,
		},
		NotifNewsAlert: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelInApp},
			MinimumPriority: PriorityNormal,
		},

		// Low priority notifications
		NotifTradingHoursChange: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelInApp, ChannelEmail},
			MinimumPriority: PriorityLow,
		},
		NotifSystemMaintenance: {
			Enabled:         true,
			Channels:        []NotificationChannel{ChannelInApp, ChannelEmail},
			MinimumPriority: PriorityLow,
		},
	}

	prefs.Preferences = defaultPrefs
	return prefs
}

// PreferencesManager manages notification preferences
type PreferencesManager struct {
	store PreferencesStore
}

// NewPreferencesManager creates a new preferences manager
func NewPreferencesManager(store PreferencesStore) *PreferencesManager {
	return &PreferencesManager{
		store: store,
	}
}

// GetPreferences retrieves user preferences
func (m *PreferencesManager) GetPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	return m.store.Get(ctx, userID)
}

// UpdatePreferences updates user preferences
func (m *PreferencesManager) UpdatePreferences(ctx context.Context, prefs *UserPreferences) error {
	return m.store.Save(ctx, prefs)
}

// SetChannelPreference sets preference for a specific notification type and channel
func (m *PreferencesManager) SetChannelPreference(ctx context.Context, userID string, notifType NotificationType, channels []NotificationChannel, enabled bool) error {
	prefs, err := m.store.Get(ctx, userID)
	if err != nil {
		return err
	}

	pref, exists := prefs.Preferences[notifType]
	if !exists {
		pref = ChannelPreference{
			MinimumPriority: PriorityNormal,
		}
	}

	pref.Enabled = enabled
	pref.Channels = channels
	prefs.Preferences[notifType] = pref

	return m.store.Save(ctx, prefs)
}

// SetQuietHours sets quiet hours for a user
func (m *PreferencesManager) SetQuietHours(ctx context.Context, userID string, enabled bool, startTime, endTime, timezone string) error {
	prefs, err := m.store.Get(ctx, userID)
	if err != nil {
		return err
	}

	prefs.QuietHours = &QuietHours{
		Enabled:   enabled,
		StartTime: startTime,
		EndTime:   endTime,
		Timezone:  timezone,
	}

	return m.store.Save(ctx, prefs)
}

// UnsubscribeAll unsubscribes a user from all notifications
func (m *PreferencesManager) UnsubscribeAll(ctx context.Context, userID string) error {
	prefs, err := m.store.Get(ctx, userID)
	if err != nil {
		return err
	}

	prefs.UnsubscribeAll = true
	return m.store.Save(ctx, prefs)
}

// ResubscribeAll re-subscribes a user to notifications
func (m *PreferencesManager) ResubscribeAll(ctx context.Context, userID string) error {
	prefs, err := m.store.Get(ctx, userID)
	if err != nil {
		return err
	}

	prefs.UnsubscribeAll = false
	return m.store.Save(ctx, prefs)
}

// ShouldSend checks if a notification should be sent based on user preferences
func (m *PreferencesManager) ShouldSend(ctx context.Context, userID string, notif *Notification, channel NotificationChannel) (bool, error) {
	prefs, err := m.store.Get(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if user unsubscribed from all notifications
	if prefs.UnsubscribeAll {
		// Still allow critical notifications
		if notif.Priority != PriorityCritical {
			return false, nil
		}
	}

	// Check quiet hours for non-critical notifications
	if notif.Priority != PriorityCritical && prefs.QuietHours != nil && prefs.QuietHours.Enabled {
		if m.isInQuietHours(prefs.QuietHours) {
			return false, nil
		}
	}

	// Check notification type preferences
	pref, exists := prefs.Preferences[notif.Type]
	if !exists {
		// If no preference exists, allow high and critical priority notifications
		return notif.Priority == PriorityHigh || notif.Priority == PriorityCritical, nil
	}

	// Check if notification type is enabled
	if !pref.Enabled {
		return false, nil
	}

	// Check if priority meets minimum requirement
	if !m.meetsPriorityRequirement(notif.Priority, pref.MinimumPriority) {
		return false, nil
	}

	// Check if channel is enabled for this notification type
	for _, enabledChannel := range pref.Channels {
		if enabledChannel == channel {
			return true, nil
		}
	}

	return false, nil
}

// isInQuietHours checks if current time is within quiet hours
func (m *PreferencesManager) isInQuietHours(quietHours *QuietHours) bool {
	// Load timezone
	loc, err := time.LoadLocation(quietHours.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentTime := now.Format("15:04")

	// Simple string comparison (HH:MM format)
	// This handles cases where quiet hours cross midnight
	if quietHours.StartTime <= quietHours.EndTime {
		return currentTime >= quietHours.StartTime && currentTime < quietHours.EndTime
	}
	return currentTime >= quietHours.StartTime || currentTime < quietHours.EndTime
}

// meetsPriorityRequirement checks if notification priority meets the minimum requirement
func (m *PreferencesManager) meetsPriorityRequirement(notifPriority, minPriority Priority) bool {
	priorityLevels := map[Priority]int{
		PriorityLow:      0,
		PriorityNormal:   1,
		PriorityHigh:     2,
		PriorityCritical: 3,
	}

	notifLevel, _ := priorityLevels[notifPriority]
	minLevel, _ := priorityLevels[minPriority]

	return notifLevel >= minLevel
}

// ExportPreferences exports user preferences as JSON
func (m *PreferencesManager) ExportPreferences(ctx context.Context, userID string) ([]byte, error) {
	prefs, err := m.store.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(prefs, "", "  ")
}

// ImportPreferences imports user preferences from JSON
func (m *PreferencesManager) ImportPreferences(ctx context.Context, userID string, data []byte) error {
	var prefs UserPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return fmt.Errorf("failed to parse preferences: %w", err)
	}

	prefs.UserID = userID
	prefs.UpdatedAt = time.Now()

	return m.store.Save(ctx, &prefs)
}

// GetChannelsForNotification returns the channels to use for a notification
func (m *PreferencesManager) GetChannelsForNotification(ctx context.Context, userID string, notif *Notification) ([]NotificationChannel, error) {
	_, err := m.store.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	var channels []NotificationChannel

	for _, channel := range notif.Channels {
		shouldSend, err := m.ShouldSend(ctx, userID, notif, channel)
		if err != nil {
			continue
		}
		if shouldSend {
			channels = append(channels, channel)
		}
	}

	return channels, nil
}
