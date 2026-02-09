package models

import (
	"encoding/json"
	"time"
)

// Workspace represents a saved workspace configuration
type Workspace struct {
	ID          int64           `json:"id" db:"id"`
	UserID      string          `json:"userId" db:"user_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description,omitempty" db:"description"`
	Config      WorkspaceConfig `json:"config" db:"config"`
	IsDefault   bool            `json:"isDefault" db:"is_default"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
}

// WorkspaceConfig holds the complete workspace state
type WorkspaceConfig struct {
	// Chart Configuration
	Charts []ChartConfig `json:"charts"`

	// Active Symbol and Timeframe
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`

	// Layout Configuration
	Layout LayoutConfig `json:"layout"`

	// Drawings and Indicators
	Drawings   []DrawingConfig   `json:"drawings,omitempty"`
	Indicators []IndicatorConfig `json:"indicators,omitempty"`

	// Version for migration compatibility
	Version string `json:"version"`
}

// ChartConfig represents a single chart configuration
type ChartConfig struct {
	ID         string            `json:"id"`
	Symbol     string            `json:"symbol"`
	Timeframe  string            `json:"timeframe"`
	ChartType  string            `json:"chartType"` // candlestick, line, bar, etc.
	Indicators []IndicatorConfig `json:"indicators,omitempty"`
	Drawings   []DrawingConfig   `json:"drawings,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// LayoutConfig represents panel layout configuration
type LayoutConfig struct {
	SidebarOpen     bool               `json:"sidebarOpen"`
	ActivePanels    []string           `json:"activePanels"`
	PanelSizes      map[string]float64 `json:"panelSizes,omitempty"`
	WindowPositions map[string]Position `json:"windowPositions,omitempty"`
}

// Position represents a window or panel position
type Position struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DrawingConfig represents a chart drawing (trendline, fibonacci, etc.)
type DrawingConfig struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // trendline, fibonacci, horizontal, etc.
	Symbol     string                 `json:"symbol"`
	Timeframe  string                 `json:"timeframe"`
	Points     []Point                `json:"points"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// Point represents a coordinate on the chart
type Point struct {
	Time  int64   `json:"time"`
	Price float64 `json:"price"`
}

// IndicatorConfig represents a technical indicator configuration
type IndicatorConfig struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // ema, sma, rsi, macd, etc.
	Parameters map[string]interface{} `json:"parameters"`
	Style      IndicatorStyle         `json:"style"`
}

// IndicatorStyle represents visual styling for an indicator
type IndicatorStyle struct {
	Color     string  `json:"color"`
	LineWidth int     `json:"lineWidth"`
	Opacity   float64 `json:"opacity"`
}

// PrintPreferences represents user print settings
type PrintPreferences struct {
	UserID        string    `json:"userId" db:"user_id"`
	PageSize      string    `json:"pageSize" db:"page_size"`           // A4, Letter, Legal
	Orientation   string    `json:"orientation" db:"orientation"`       // portrait, landscape
	MarginTop     float64   `json:"marginTop" db:"margin_top"`          // in inches
	MarginBottom  float64   `json:"marginBottom" db:"margin_bottom"`
	MarginLeft    float64   `json:"marginLeft" db:"margin_left"`
	MarginRight   float64   `json:"marginRight" db:"margin_right"`
	IncludeHeader bool      `json:"includeHeader" db:"include_header"`
	IncludeFooter bool      `json:"includeFooter" db:"include_footer"`
	HeaderText    string    `json:"headerText,omitempty" db:"header_text"`
	FooterText    string    `json:"footerText,omitempty" db:"footer_text"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// UserDataFolder represents the virtual file system structure for web
type UserDataFolder struct {
	Path     string           `json:"path"`
	Type     string           `json:"type"` // file, directory
	Size     int64            `json:"size,omitempty"`
	Modified time.Time        `json:"modified,omitempty"`
	Children []UserDataFolder `json:"children,omitempty"`
}

// Session represents user session state
type Session struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"userId" db:"user_id"`
	LastActiveAt     time.Time `json:"lastActiveAt" db:"last_active_at"`
	UnsavedWorkspace *Workspace `json:"unsavedWorkspace,omitempty" db:"unsaved_workspace"`
	HasUnsavedData   bool      `json:"hasUnsavedData" db:"has_unsaved_data"`
}

// Custom JSON marshaling for WorkspaceConfig to handle JSONB storage
func (w *WorkspaceConfig) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	bytes, ok := src.([]byte)
	if !ok {
		return json.Unmarshal([]byte(src.(string)), w)
	}
	return json.Unmarshal(bytes, w)
}

func (w WorkspaceConfig) Value() (interface{}, error) {
	return json.Marshal(w)
}
