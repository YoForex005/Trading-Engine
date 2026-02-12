package notifications

import (
	"sort"
	"sync"
)

// Store manages in-memory notification storage
type Store struct {
	notifications map[string]*Notification // key: notification ID
	userIndex     map[string][]string      // key: userID, value: notification IDs
	mu            sync.RWMutex
}

// NewStore creates a new notification store
func NewStore() *Store {
	return &Store{
		notifications: make(map[string]*Notification),
		userIndex:     make(map[string][]string),
	}
}

// Add stores a new notification
func (s *Store) Add(notification *Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.notifications[notification.ID] = notification
	
	// Add to user index
	if _, exists := s.userIndex[notification.UserID]; !exists {
		s.userIndex[notification.UserID] = []string{}
	}
	s.userIndex[notification.UserID] = append(s.userIndex[notification.UserID], notification.ID)
}

// GetUnread returns all unread notifications for a user
func (s *Store) GetUnread(userID string) []*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	notifIDs, exists := s.userIndex[userID]
	if !exists {
		return []*Notification{}
	}
	
	var unread []*Notification
	for _, id := range notifIDs {
		notif, exists := s.notifications[id]
		if exists && !notif.Read {
			unread = append(unread, notif)
		}
	}
	
	// Sort by created time (newest first)
	sort.Slice(unread, func(i, j int) bool {
		return unread[i].CreatedAt > unread[j].CreatedAt
	})
	
	return unread
}

// GetHistory returns notification history for a user with pagination
func (s *Store) GetHistory(userID string, limit, offset int) []*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	notifIDs, exists := s.userIndex[userID]
	if !exists {
		return []*Notification{}
	}
	
	var all []*Notification
	for _, id := range notifIDs {
		notif, exists := s.notifications[id]
		if exists {
			all = append(all, notif)
		}
	}
	
	// Sort by created time (newest first)
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt > all[j].CreatedAt
	})
	
	// Apply pagination
	if offset >= len(all) {
		return []*Notification{}
	}
	
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	
	return all[offset:end]
}

// MarkAsRead marks a single notification as read
func (s *Store) MarkAsRead(userID, notificationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	notif, exists := s.notifications[notificationID]
	if !exists {
		return ErrNotificationNotFound
	}
	
	// Verify ownership
	if notif.UserID != userID {
		return ErrUnauthorized
	}
	
	notif.Read = true
	return nil
}

// MarkAllRead marks all notifications as read for a user
func (s *Store) MarkAllRead(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	notifIDs, exists := s.userIndex[userID]
	if !exists {
		return nil
	}
	
	for _, id := range notifIDs {
		notif, exists := s.notifications[id]
		if exists {
			notif.Read = true
		}
	}
	
	return nil
}

// Delete removes a notification
func (s *Store) Delete(userID, notificationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	notif, exists := s.notifications[notificationID]
	if !exists {
		return ErrNotificationNotFound
	}
	
	// Verify ownership
	if notif.UserID != userID {
		return ErrUnauthorized
	}
	
	// Remove from main storage
	delete(s.notifications, notificationID)
	
	// Remove from user index
	notifIDs := s.userIndex[userID]
	for i, id := range notifIDs {
		if id == notificationID {
			s.userIndex[userID] = append(notifIDs[:i], notifIDs[i+1:]...)
			break
		}
	}
	
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *Store) GetUnreadCount(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	notifIDs, exists := s.userIndex[userID]
	if !exists {
		return 0
	}
	
	count := 0
	for _, id := range notifIDs {
		notif, exists := s.notifications[id]
		if exists && !notif.Read {
			count++
		}
	}
	
	return count
}

// Get retrieves a single notification by ID
func (s *Store) Get(notificationID string) (*Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	notif, exists := s.notifications[notificationID]
	if !exists {
		return nil, ErrNotificationNotFound
	}
	
	return notif, nil
}

// Common errors
var (
	ErrNotificationNotFound = &StoreError{Message: "notification not found"}
	ErrUnauthorized         = &StoreError{Message: "unauthorized access"}
)

// StoreError represents a store error
type StoreError struct {
	Message string
}

func (e *StoreError) Error() string {
	return e.Message
}
