package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

type ComplianceReport struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"` // "MiFID_II", "EMIR", "ASIC", "FCA", "CySEC"
	Regulator    string    `json:"regulator"`
	Period       string    `json:"period"` // "Q1", "Q2", "Q3", "Q4", "Monthly", "Annual"
	Format       string    `json:"format"` // "PDF", "CSV", "XML"
	Status       string    `json:"status"` // "pending", "generating", "completed", "failed"
	FilePath     string    `json:"filePath"`
	FileSize     int64     `json:"fileSize"` // bytes
	GeneratedAt  time.Time `json:"generatedAt"`
	GeneratedBy  string    `json:"generatedBy"`
	DataSummary  ReportSummary `json:"dataSummary"`
	SubmittedTo  string    `json:"submittedTo,omitempty"`
	SubmittedAt  *time.Time `json:"submittedAt,omitempty"`
}

type ReportSummary struct {
	TotalTrades      int     `json:"totalTrades"`
	TotalVolume      float64 `json:"totalVolume"`
	TotalClients     int     `json:"totalClients"`
	TotalTransactions int    `json:"totalTransactions"`
	ComplianceScore  float64 `json:"complianceScore"`
	IssuesFound      int     `json:"issuesFound"`
}

type ReportTemplate struct {
	ID           int64    `json:"id"`
	Regulator    string   `json:"regulator"`
	ReportType   string   `json:"reportType"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RequiredData []string `json:"requiredData"`
	Frequency    string   `json:"frequency"` // "Daily", "Weekly", "Monthly", "Quarterly", "Annual"
	Deadline     string   `json:"deadline"`  // e.g., "T+1", "Month end + 5 days"
	Format       []string `json:"format"`    // Supported formats
}

type ComplianceCheck struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pass", "fail", "warning"
	Score       float64   `json:"score"`  // 0-100
	Details     string    `json:"details"`
	LastChecked time.Time `json:"lastChecked"`
	Priority    string    `json:"priority"` // "critical", "high", "medium", "low"
}

type ComplianceCheckResult struct {
	CheckID     int64     `json:"checkId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Score       float64   `json:"score"`
	Details     string    `json:"details"`
	CheckedAt   time.Time `json:"checkedAt"`
}

type ComplianceBreach struct {
	ID          int64      `json:"id"`
	Date        time.Time  `json:"date"`
	Type        string     `json:"type"` // "KYC", "AML", "Best Execution", "Margin", "Reporting"
	Severity    string     `json:"severity"` // "critical", "high", "medium", "low"
	Description string     `json:"description"`
	Status      string     `json:"status"` // "open", "investigating", "resolved", "closed"
	Resolution  string     `json:"resolution,omitempty"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy  string     `json:"resolvedBy,omitempty"`
	ClientID    int64      `json:"clientId,omitempty"`
	TradeID     int64      `json:"tradeId,omitempty"`
}

type ComplianceStats struct {
	OverallScore      float64             `json:"overallScore"` // 0-100
	ChecksPassed      int                 `json:"checksPassed"`
	ChecksFailed      int                 `json:"checksFailed"`
	ChecksWarning     int                 `json:"checksWarning"`
	TotalChecks       int                 `json:"totalChecks"`
	ReportsGenerated  int                 `json:"reportsGenerated"`
	ReportsThisQuarter int                `json:"reportsThisQuarter"`
	OpenBreaches      int                 `json:"openBreaches"`
	ResolvedBreaches  int                 `json:"resolvedBreaches"`
	UpcomingDeadlines []RegulatoryDeadline `json:"upcomingDeadlines"`
	LastAuditDate     time.Time           `json:"lastAuditDate"`
}

type RegulatoryDeadline struct {
	Regulator   string    `json:"regulator"`
	ReportType  string    `json:"reportType"`
	DueDate     time.Time `json:"dueDate"`
	DaysUntil   int       `json:"daysUntil"`
	Status      string    `json:"status"` // "upcoming", "due_soon", "overdue", "submitted"
}

// ============================================
// Service
// ============================================

type ComplianceReportService struct {
	reports   map[int64]*ComplianceReport
	templates map[int64]*ReportTemplate
	checks    map[int64]*ComplianceCheck
	breaches  map[int64]*ComplianceBreach
	nextReportID  int64
	nextTemplateID int64
	nextCheckID   int64
	nextBreachID  int64
	mu        sync.RWMutex
}

func NewComplianceReportService() *ComplianceReportService {
	service := &ComplianceReportService{
		reports:   make(map[int64]*ComplianceReport),
		templates: make(map[int64]*ReportTemplate),
		checks:    make(map[int64]*ComplianceCheck),
		breaches:  make(map[int64]*ComplianceBreach),
		nextReportID:  1,
		nextTemplateID: 1,
		nextCheckID:   1,
		nextBreachID:  1,
	}
	service.initializeMockData()
	return service
}

func (s *ComplianceReportService) initializeMockData() {
	now := time.Now()

	// ============================================
	// Initialize Report Templates (5 regulators)
	// ============================================
	templates := []ReportTemplate{
		{
			ID:          1,
			Regulator:   "MiFID II",
			ReportType:  "Transaction Reporting",
			Name:        "RTS 22 Transaction Reports",
			Description: "Detailed transaction reporting under MiFID II",
			RequiredData: []string{
				"Trade timestamp", "Client ID", "Instrument ISIN", "Buy/Sell indicator",
				"Quantity", "Price", "Venue", "Counterparty", "Client categorization",
			},
			Frequency: "Daily",
			Deadline:  "T+1",
			Format:    []string{"XML", "CSV"},
		},
		{
			ID:          2,
			Regulator:   "MiFID II",
			ReportType:  "Best Execution",
			Name:        "RTS 27 Best Execution Report",
			Description: "Annual best execution quality report",
			RequiredData: []string{
				"Top 5 execution venues", "Execution quality metrics", "Price improvement",
				"Speed of execution", "Likelihood of execution", "Payment for order flow",
			},
			Frequency: "Annual",
			Deadline:  "April 30",
			Format:    []string{"PDF", "XML"},
		},
		{
			ID:          3,
			Regulator:   "EMIR",
			ReportType:  "Trade Repository Reporting",
			Name:        "Derivatives Transaction Reporting",
			Description: "Reporting of derivative transactions to trade repositories",
			RequiredData: []string{
				"Counterparty details", "Derivative type", "Notional amount",
				"Maturity date", "Valuation", "Collateral", "Clearing status",
			},
			Frequency: "Daily",
			Deadline:  "T+1",
			Format:    []string{"XML"},
		},
		{
			ID:          4,
			Regulator:   "FCA",
			ReportType:  "Client Asset Report",
			Name:        "CASS Resolution Pack",
			Description: "Client asset segregation and reconciliation",
			RequiredData: []string{
				"Client money holdings", "Custody assets", "Reconciliation records",
				"Segregated accounts", "Client disclosures", "Acknowledgement letters",
			},
			Frequency: "Monthly",
			Deadline:  "Month end + 5 days",
			Format:    []string{"PDF", "CSV"},
		},
		{
			ID:          5,
			Regulator:   "CySEC",
			ReportType:  "Client Categorization",
			Name:        "Client Classification Report",
			Description: "Annual client categorization review",
			RequiredData: []string{
				"Retail clients count", "Professional clients count", "Eligible counterparties",
				"Opt-up requests", "Opt-down requests", "Appropriateness assessments",
			},
			Frequency: "Annual",
			Deadline:  "January 31",
			Format:    []string{"PDF", "CSV"},
		},
		{
			ID:          6,
			Regulator:   "ASIC",
			ReportType:  "AFS Licensee Report",
			Name:        "ASIC Financial Report",
			Description: "Australian financial services annual report",
			RequiredData: []string{
				"Breach reporting", "Complaints summary", "Financial resources",
				"Audit report", "Professional indemnity insurance", "Key person changes",
			},
			Frequency: "Annual",
			Deadline:  "45 days after year end",
			Format:    []string{"PDF"},
		},
		{
			ID:          7,
			Regulator:   "FCA",
			ReportType:  "Product Governance",
			Name:        "Product Intervention Report",
			Description: "Product governance and intervention measures",
			RequiredData: []string{
				"Target market definition", "Distribution strategy", "Product reviews",
				"Conflicts of interest", "Inducements disclosure", "Value assessment",
			},
			Frequency: "Quarterly",
			Deadline:  "Quarter end + 30 days",
			Format:    []string{"PDF", "CSV"},
		},
		{
			ID:          8,
			Regulator:   "MiFID II",
			ReportType:  "Costs and Charges",
			Name:        "Ex-ante and Ex-post Costs Disclosure",
			Description: "Disclosure of costs and charges to clients",
			RequiredData: []string{
				"Transaction costs", "Ongoing charges", "Performance fees",
				"Third-party payments", "Ancillary services", "Currency conversion costs",
			},
			Frequency: "Annual",
			Deadline:  "March 31",
			Format:    []string{"PDF"},
		},
	}

	for i := range templates {
		s.templates[templates[i].ID] = &templates[i]
	}
	s.nextTemplateID = 9

	// ============================================
	// Initialize 50 Compliance Reports
	// ============================================
	reportTypes := []string{"MiFID_II", "EMIR", "ASIC", "FCA", "CySEC"}
	regulators := []string{"MiFID II", "EMIR", "ASIC", "FCA", "CySEC"}
	periods := []string{"Q1_2026", "Q2_2026", "Q3_2026", "Q4_2025", "Monthly_Jan_2026", "Monthly_Dec_2025", "Annual_2025"}
	formats := []string{"PDF", "CSV", "XML"}
	statuses := []string{"completed", "completed", "completed", "completed", "generating", "pending"}

	for i := 0; i < 50; i++ {
		daysAgo := i * 5
		report := &ComplianceReport{
			ID:          int64(i + 1),
			Type:        reportTypes[i%len(reportTypes)],
			Regulator:   regulators[i%len(regulators)],
			Period:      periods[i%len(periods)],
			Format:      formats[i%len(formats)],
			Status:      statuses[i%len(statuses)],
			FilePath:    fmt.Sprintf("/reports/compliance_%d.%s", i+1, strings.ToLower(formats[i%len(formats)])),
			FileSize:    int64(500000 + i*10000),
			GeneratedAt: now.AddDate(0, 0, -daysAgo),
			GeneratedBy: fmt.Sprintf("admin%d@rtx5.com", (i%5)+1),
			DataSummary: ReportSummary{
				TotalTrades:      1000 + i*50,
				TotalVolume:      float64(5000000 + i*100000),
				TotalClients:     50 + i*2,
				TotalTransactions: 200 + i*10,
				ComplianceScore:  85.0 + float64(i%15),
				IssuesFound:      i % 5,
			},
		}

		if report.Status == "completed" && i%3 == 0 {
			submittedAt := now.AddDate(0, 0, -daysAgo+1)
			report.SubmittedTo = report.Regulator
			report.SubmittedAt = &submittedAt
		}

		s.reports[report.ID] = report
	}
	s.nextReportID = 51

	// ============================================
	// Initialize 20 Compliance Checks
	// ============================================
	checks := []ComplianceCheck{
		{
			ID:          1,
			Name:        "KYC Completion Rate",
			Category:    "Client Onboarding",
			Description: "Percentage of clients with complete KYC documentation",
			Status:      "pass",
			Score:       98.5,
			Details:     "197 out of 200 clients have complete KYC. 3 pending document uploads.",
			LastChecked: now.Add(-2 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          2,
			Name:        "AML Transaction Monitoring",
			Category:    "Anti-Money Laundering",
			Description: "Real-time monitoring of suspicious transaction patterns",
			Status:      "pass",
			Score:       95.0,
			Details:     "15 alerts triggered in last 24h. All reviewed. 2 escalated to compliance team.",
			LastChecked: now.Add(-1 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          3,
			Name:        "Best Execution Policy",
			Category:    "Trading",
			Description: "Adherence to best execution obligations",
			Status:      "pass",
			Score:       92.0,
			Details:     "All trades executed within policy parameters. Average price improvement: 0.3 pips.",
			LastChecked: now.Add(-3 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          4,
			Name:        "Client Categorization",
			Category:    "Client Protection",
			Description: "Accurate client categorization (Retail/Professional/ECP)",
			Status:      "pass",
			Score:       100.0,
			Details:     "All 200 clients correctly categorized. 5 opt-up requests processed.",
			LastChecked: now.Add(-24 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          5,
			Name:        "Negative Balance Protection",
			Category:    "Client Protection",
			Description: "Ensure retail clients cannot lose more than deposited",
			Status:      "pass",
			Score:       100.0,
			Details:     "NBP active for all 150 retail clients. No negative balances detected.",
			LastChecked: now.Add(-30 * time.Minute),
			Priority:    "critical",
		},
		{
			ID:          6,
			Name:        "Segregated Client Accounts",
			Category:    "Client Money",
			Description: "Client funds held in segregated accounts",
			Status:      "pass",
			Score:       100.0,
			Details:     "All client funds segregated. Daily reconciliation completed.",
			LastChecked: now.Add(-5 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          7,
			Name:        "Leverage Limits Compliance",
			Category:    "Trading",
			Description: "Adherence to regulatory leverage limits",
			Status:      "pass",
			Score:       97.0,
			Details:     "Retail: 30:1 max. Professional: 500:1 max. 6 warnings issued.",
			LastChecked: now.Add(-1 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          8,
			Name:        "Trade Reporting (T+1)",
			Category:    "Regulatory Reporting",
			Description: "Timely submission of trade reports to regulators",
			Status:      "pass",
			Score:       99.0,
			Details:     "99.2% of trades reported within T+1. 8 out of 1000 late (technical issues).",
			LastChecked: now.Add(-2 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          9,
			Name:        "Product Governance",
			Category:    "Product Compliance",
			Description: "Product approval and target market assessment",
			Status:      "pass",
			Score:       88.0,
			Details:     "50 products approved. 3 under review. All have target market defined.",
			LastChecked: now.Add(-48 * time.Hour),
			Priority:    "medium",
		},
		{
			ID:          10,
			Name:        "Conflicts of Interest",
			Category:    "Organizational",
			Description: "Management and disclosure of conflicts of interest",
			Status:      "pass",
			Score:       90.0,
			Details:     "Conflicts register updated. 12 conflicts identified and mitigated.",
			LastChecked: now.Add(-72 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          11,
			Name:        "Market Abuse Monitoring",
			Category:    "Market Integrity",
			Description: "Detection of market manipulation and insider trading",
			Status:      "warning",
			Score:       75.0,
			Details:     "5 potential cases flagged for investigation. 2 referred to regulator.",
			LastChecked: now.Add(-6 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          12,
			Name:        "Record Keeping (5 years)",
			Category:    "Organizational",
			Description: "Retention of client and transaction records",
			Status:      "pass",
			Score:       100.0,
			Details:     "All records archived. Automated deletion after 7 years.",
			LastChecked: now.Add(-24 * time.Hour),
			Priority:    "medium",
		},
		{
			ID:          13,
			Name:        "Client Suitability Assessment",
			Category:    "Client Protection",
			Description: "Appropriateness and suitability tests",
			Status:      "pass",
			Score:       94.0,
			Details:     "188 out of 200 clients assessed. 12 pending responses to questionnaire.",
			LastChecked: now.Add(-12 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          14,
			Name:        "Risk Warnings Disclosure",
			Category:    "Client Protection",
			Description: "Adequate risk warnings provided to clients",
			Status:      "pass",
			Score:       100.0,
			Details:     "All clients acknowledged risk warnings. Refreshed quarterly.",
			LastChecked: now.Add(-24 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          15,
			Name:        "Professional Indemnity Insurance",
			Category:    "Financial Resources",
			Description: "Adequate PI insurance coverage",
			Status:      "pass",
			Score:       100.0,
			Details:     "£5M coverage. Policy expires 30 Jun 2026. Renewal in progress.",
			LastChecked: now.Add(-168 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          16,
			Name:        "Capital Adequacy",
			Category:    "Financial Resources",
			Description: "Minimum capital requirements met",
			Status:      "pass",
			Score:       100.0,
			Details:     "Current capital: £2.5M. Required: £750K. Surplus: £1.75M.",
			LastChecked: now.Add(-24 * time.Hour),
			Priority:    "critical",
		},
		{
			ID:          17,
			Name:        "Complaints Handling",
			Category:    "Client Protection",
			Description: "Timely and fair handling of client complaints",
			Status:      "pass",
			Score:       85.0,
			Details:     "15 complaints this quarter. 12 resolved within 5 days. 3 pending.",
			LastChecked: now.Add(-48 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          18,
			Name:        "Data Protection (GDPR)",
			Category:    "Data Privacy",
			Description: "Compliance with GDPR and data protection laws",
			Status:      "pass",
			Score:       92.0,
			Details:     "Privacy policy updated. 20 DSAR requests processed. 1 breach reported.",
			LastChecked: now.Add(-72 * time.Hour),
			Priority:    "high",
		},
		{
			ID:          19,
			Name:        "Outsourcing Oversight",
			Category:    "Organizational",
			Description: "Due diligence on third-party service providers",
			Status:      "pass",
			Score:       88.0,
			Details:     "8 critical suppliers. All have contracts. Annual reviews completed.",
			LastChecked: now.Add(-168 * time.Hour),
			Priority:    "medium",
		},
		{
			ID:          20,
			Name:        "Staff Training and Competence",
			Category:    "Organizational",
			Description: "Regulatory and compliance training for staff",
			Status:      "warning",
			Score:       70.0,
			Details:     "18 out of 25 staff completed annual training. 7 pending.",
			LastChecked: now.Add(-24 * time.Hour),
			Priority:    "medium",
		},
	}

	for i := range checks {
		s.checks[checks[i].ID] = &checks[i]
	}
	s.nextCheckID = 21

	// ============================================
	// Initialize 30 Compliance Breaches
	// ============================================
	breachTypes := []string{"KYC", "AML", "Best Execution", "Margin", "Reporting", "Client Money", "Data Protection", "Market Abuse"}
	severities := []string{"critical", "high", "high", "medium", "medium", "low"}
	breachStatuses := []string{"resolved", "resolved", "resolved", "investigating", "open"}

	for i := 0; i < 30; i++ {
		daysAgo := i * 3
		status := breachStatuses[i%len(breachStatuses)]

		breach := &ComplianceBreach{
			ID:          int64(i + 1),
			Date:        now.AddDate(0, 0, -daysAgo),
			Type:        breachTypes[i%len(breachTypes)],
			Severity:    severities[i%len(severities)],
			Status:      status,
			ClientID:    int64((i % 50) + 1),
			TradeID:     int64((i % 100) + 1000),
		}

		switch breach.Type {
		case "KYC":
			breach.Description = fmt.Sprintf("Incomplete KYC documentation for client %d. Missing proof of address.", breach.ClientID)
		case "AML":
			breach.Description = fmt.Sprintf("Suspicious transaction pattern detected for client %d. Value: $%.0f.", breach.ClientID, float64(50000+i*5000))
		case "Best Execution":
			breach.Description = fmt.Sprintf("Trade %d executed outside best execution policy (slippage: %.2f pips).", breach.TradeID, 5.0+float64(i%10)*0.5)
		case "Margin":
			breach.Description = fmt.Sprintf("Client %d exceeded leverage limit (used: %d:1, max: 30:1).", breach.ClientID, 35+i%20)
		case "Reporting":
			breach.Description = fmt.Sprintf("Trade report for trade %d submitted late (T+%d instead of T+1).", breach.TradeID, 2+i%3)
		case "Client Money":
			breach.Description = fmt.Sprintf("Client funds reconciliation discrepancy: $%.2f.", float64(100+i*50))
		case "Data Protection":
			breach.Description = fmt.Sprintf("Unauthorized access to client %d data. Internal investigation initiated.", breach.ClientID)
		case "Market Abuse":
			breach.Description = fmt.Sprintf("Potential market manipulation detected on trade %d. Referred to compliance.", breach.TradeID)
		}

		if status == "resolved" || status == "closed" {
			resolvedAt := now.AddDate(0, 0, -daysAgo+i%5+1)
			breach.ResolvedAt = &resolvedAt
			breach.ResolvedBy = fmt.Sprintf("compliance_officer_%d@rtx5.com", (i%3)+1)

			switch breach.Type {
			case "KYC":
				breach.Resolution = "Client provided missing documentation. KYC now complete."
			case "AML":
				breach.Resolution = "Transaction verified as legitimate business activity. No further action."
			case "Best Execution":
				breach.Resolution = "Execution policy updated. Additional liquidity provider added."
			case "Margin":
				breach.Resolution = "Client account leverage reduced. Warning issued to client."
			case "Reporting":
				breach.Resolution = "Technical issue resolved. Automated reporting system updated."
			case "Client Money":
				breach.Resolution = "Reconciliation error corrected. Bank statement reconciled."
			case "Data Protection":
				breach.Resolution = "Access revoked. Staff member disciplined. Systems access reviewed."
			case "Market Abuse":
				breach.Resolution = "Investigation concluded. No manipulation found. Normal trading activity."
			}

			if i%10 == 0 {
				breach.Status = "closed"
			}
		}

		s.breaches[breach.ID] = breach
	}
	s.nextBreachID = 31

	log.Printf("[ComplianceReportService] Initialized with %d reports, %d templates, %d checks, %d breaches",
		len(s.reports), len(s.templates), len(s.checks), len(s.breaches))
}

// ============================================
// Service Methods
// ============================================

func (s *ComplianceReportService) GetReports(filterType, filterRegulator string) []*ComplianceReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reports := make([]*ComplianceReport, 0)
	for _, report := range s.reports {
		if filterType != "" && report.Type != filterType {
			continue
		}
		if filterRegulator != "" && report.Regulator != filterRegulator {
			continue
		}
		reports = append(reports, report)
	}
	return reports
}

func (s *ComplianceReportService) GetReportByID(id int64) *ComplianceReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.reports[id]
}

func (s *ComplianceReportService) GenerateReport(reportType, regulator, period, format string) *ComplianceReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	report := &ComplianceReport{
		ID:          s.nextReportID,
		Type:        reportType,
		Regulator:   regulator,
		Period:      period,
		Format:      format,
		Status:      "generating",
		FilePath:    fmt.Sprintf("/reports/compliance_%d.%s", s.nextReportID, strings.ToLower(format)),
		FileSize:    0,
		GeneratedAt: time.Now(),
		GeneratedBy: "admin@rtx5.com",
		DataSummary: ReportSummary{
			TotalTrades:      1500,
			TotalVolume:      7500000,
			TotalClients:     200,
			TotalTransactions: 450,
			ComplianceScore:  92.5,
			IssuesFound:      3,
		},
	}

	s.reports[s.nextReportID] = report
	s.nextReportID++

	// Simulate async generation
	go func() {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		report.Status = "completed"
		report.FileSize = 856000
		s.mu.Unlock()
	}()

	return report
}

func (s *ComplianceReportService) GetTemplates() []*ReportTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*ReportTemplate, 0, len(s.templates))
	for _, template := range s.templates {
		templates = append(templates, template)
	}
	return templates
}

func (s *ComplianceReportService) GetChecks() []*ComplianceCheck {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checks := make([]*ComplianceCheck, 0, len(s.checks))
	for _, check := range s.checks {
		checks = append(checks, check)
	}
	return checks
}

func (s *ComplianceReportService) RunComplianceChecks() []ComplianceCheckResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]ComplianceCheckResult, 0, len(s.checks))
	now := time.Now()

	for _, check := range s.checks {
		result := ComplianceCheckResult{
			CheckID:   check.ID,
			Name:      check.Name,
			Status:    check.Status,
			Score:     check.Score,
			Details:   check.Details,
			CheckedAt: now,
		}
		results = append(results, result)

		// Update last checked time
		check.LastChecked = now
	}

	return results
}

func (s *ComplianceReportService) GetBreaches(filterStatus, filterSeverity string) []*ComplianceBreach {
	s.mu.RLock()
	defer s.mu.RUnlock()

	breaches := make([]*ComplianceBreach, 0)
	for _, breach := range s.breaches {
		if filterStatus != "" && breach.Status != filterStatus {
			continue
		}
		if filterSeverity != "" && breach.Severity != filterSeverity {
			continue
		}
		breaches = append(breaches, breach)
	}
	return breaches
}

func (s *ComplianceReportService) GetStats() ComplianceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := ComplianceStats{
		TotalChecks:       len(s.checks),
		ReportsGenerated:  len(s.reports),
		LastAuditDate:     time.Now().AddDate(0, -1, 0),
	}

	// Count check statuses
	totalScore := 0.0
	for _, check := range s.checks {
		switch check.Status {
		case "pass":
			stats.ChecksPassed++
		case "fail":
			stats.ChecksFailed++
		case "warning":
			stats.ChecksWarning++
		}
		totalScore += check.Score
	}

	if stats.TotalChecks > 0 {
		stats.OverallScore = totalScore / float64(stats.TotalChecks)
	}

	// Count breaches
	for _, breach := range s.breaches {
		if breach.Status == "open" || breach.Status == "investigating" {
			stats.OpenBreaches++
		} else {
			stats.ResolvedBreaches++
		}
	}

	// Count reports this quarter
	quarterStart := time.Date(time.Now().Year(), ((time.Now().Month()-1)/3)*3+1, 1, 0, 0, 0, 0, time.UTC)
	for _, report := range s.reports {
		if report.GeneratedAt.After(quarterStart) {
			stats.ReportsThisQuarter++
		}
	}

	// Generate upcoming deadlines
	now := time.Now()
	stats.UpcomingDeadlines = []RegulatoryDeadline{
		{
			Regulator:  "MiFID II",
			ReportType: "Transaction Reporting",
			DueDate:    now.AddDate(0, 0, 1),
			DaysUntil:  1,
			Status:     "due_soon",
		},
		{
			Regulator:  "EMIR",
			ReportType: "Derivatives Reporting",
			DueDate:    now.AddDate(0, 0, 1),
			DaysUntil:  1,
			Status:     "due_soon",
		},
		{
			Regulator:  "FCA",
			ReportType: "CASS Resolution Pack",
			DueDate:    now.AddDate(0, 0, 8),
			DaysUntil:  8,
			Status:     "upcoming",
		},
		{
			Regulator:  "MiFID II",
			ReportType: "Best Execution Report",
			DueDate:    time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			DaysUntil:  int(time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC).Sub(now).Hours() / 24),
			Status:     "upcoming",
		},
		{
			Regulator:  "CySEC",
			ReportType: "Client Classification",
			DueDate:    time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC),
			DaysUntil:  int(time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC).Sub(now).Hours() / 24),
			Status:     "upcoming",
		},
	}

	return stats
}

// ============================================
// HTTP Handlers
// ============================================

type ComplianceReportHandler struct {
	service     *ComplianceReportService
	authService *auth.Service
}

func NewComplianceReportHandler(service *ComplianceReportService, authService *auth.Service) *ComplianceReportHandler {
	return &ComplianceReportHandler{
		service:     service,
		authService: authService,
	}
}

// 1. GET /admin/compliance/reports - List all compliance reports
func (h *ComplianceReportHandler) HandleListReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filterType := r.URL.Query().Get("type")
	filterRegulator := r.URL.Query().Get("regulator")

	reports := h.service.GetReports(filterType, filterRegulator)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
	})
}

// 2. POST /admin/compliance/reports/generate - Generate new compliance report
func (h *ComplianceReportHandler) HandleGenerateReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Type      string `json:"type"`
		Regulator string `json:"regulator"`
		Period    string `json:"period"`
		Format    string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report := h.service.GenerateReport(req.Type, req.Regulator, req.Period, req.Format)
	json.NewEncoder(w).Encode(report)
}

// 3. GET /admin/compliance/reports/:id - Get report details
func (h *ComplianceReportHandler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/compliance/reports/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	report := h.service.GetReportByID(id)
	if report == nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(report)
}

// 4. GET /admin/compliance/reports/templates - Get report templates
func (h *ComplianceReportHandler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates := h.service.GetTemplates()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// 5. GET /admin/compliance/checks - Get compliance checklist
func (h *ComplianceReportHandler) HandleListChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	checks := h.service.GetChecks()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"checks": checks,
		"count":  len(checks),
	})
}

// 6. POST /admin/compliance/checks/run - Run compliance checks
func (h *ComplianceReportHandler) HandleRunChecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	results := h.service.RunComplianceChecks()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"count":   len(results),
		"timestamp": time.Now(),
	})
}

// 7. GET /admin/compliance/breaches - Get compliance breach log
func (h *ComplianceReportHandler) HandleListBreaches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filterStatus := r.URL.Query().Get("status")
	filterSeverity := r.URL.Query().Get("severity")

	breaches := h.service.GetBreaches(filterStatus, filterSeverity)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"breaches": breaches,
		"count":    len(breaches),
	})
}

// 8. GET /admin/compliance/stats - Get compliance statistics
func (h *ComplianceReportHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}
