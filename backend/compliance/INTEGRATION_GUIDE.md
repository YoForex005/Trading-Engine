# Compliance System Integration Guide

## Quick Start

### 1. Add Compliance to Main Server

Edit `/Users/epic1st/Documents/trading engine/backend/cmd/server/main.go`:

```go
import (
    "github.com/epic1st/rtx/backend/compliance"
    // ... other imports
)

func main() {
    // ... existing initialization

    // Initialize compliance system
    complianceSystem := compliance.NewComplianceSystem()

    // Add compliance API routes
    http.HandleFunc("/compliance/reports/pending", complianceSystem.Handler.HandleGetPendingReports)
    http.HandleFunc("/compliance/reports/submit", complianceSystem.Handler.HandleSubmitReport)
    http.HandleFunc("/compliance/reports/daily", complianceSystem.Handler.HandleDailyReport)

    http.HandleFunc("/compliance/kyc/create", complianceSystem.Handler.HandleCreateKYC)
    http.HandleFunc("/compliance/kyc/screen-pep", complianceSystem.Handler.HandleScreenPEP)
    http.HandleFunc("/compliance/kyc/screen-sanctions", complianceSystem.Handler.HandleScreenSanctions)
    http.HandleFunc("/compliance/aml/file-sar", complianceSystem.Handler.HandleFileSAR)

    http.HandleFunc("/compliance/audit/history", complianceSystem.Handler.HandleGetAuditHistory)
    http.HandleFunc("/compliance/audit/verify", complianceSystem.Handler.HandleVerifyAuditIntegrity)
    http.HandleFunc("/compliance/audit/export", complianceSystem.Handler.HandleExportAuditTrail)

    http.HandleFunc("/compliance/execution/rts27", complianceSystem.Handler.HandleGenerateRTS27)
    http.HandleFunc("/compliance/execution/quality", complianceSystem.Handler.HandleGetExecutionQuality)

    http.HandleFunc("/compliance/leverage/validate", complianceSystem.Handler.HandleValidateLeverage)
    http.HandleFunc("/compliance/leverage/esma-limits", complianceSystem.Handler.HandleGetESMALimits)

    // ... rest of server initialization
}
```

### 2. Hook Into Order Flow

Edit your order execution logic to integrate compliance:

```go
// Before placing order - validate leverage
valid, maxLeverage, warning, err := complianceSystem.ValidateLeverageCompliance(
    "EU",               // jurisdiction
    "RETAIL",           // client classification
    symbol,
    instrumentClass,
    requestedLeverage,
)

if !valid {
    // Return error with warning
    return fmt.Errorf("leverage exceeds limit: %s", warning)
}

// Place order
order := placeOrder(...)

// Log order placement
complianceSystem.OnOrderPlaced(
    userID,
    clientID,
    order.ID,
    symbol,
    orderData,
    request.RemoteAddr,  // IP address
    request.UserAgent(), // User agent
)

// Execute trade
trade := executeTrade(order)

// Log trade execution and create regulatory reports
complianceSystem.OnTradeExecuted(
    userID,
    clientID,
    order.ID,
    trade.ID,
    symbol,
    tradeData,
    executedPrice,
    quantity,
    lpName,
    "EU",        // jurisdiction
    "RETAIL",    // client class
    request.RemoteAddr,
    request.UserAgent(),
)
```

### 3. Hook Into Position Management

```go
// Close position
closedPosition := closePosition(tradeID)

// Log position closure
complianceSystem.OnPositionClosed(
    userID,
    clientID,
    tradeID,
    symbol,
    positionData,
    request.RemoteAddr,
    request.UserAgent(),
)
```

### 4. Hook Into Fund Movements

```go
// Withdrawal
complianceSystem.OnWithdrawal(
    userID,
    clientID,
    amount,
    request.RemoteAddr,
    request.UserAgent(),
)

// Deposit
complianceSystem.OnDeposit(
    userID,
    clientID,
    amount,
    request.RemoteAddr,
    request.UserAgent(),
)
```

## API Usage Examples

### Transaction Reporting

**Get Pending Reports:**
```bash
curl http://localhost:8080/compliance/reports/pending
```

**Submit Report:**
```bash
curl -X POST http://localhost:8080/compliance/reports/submit \
  -H "Content-Type: application/json" \
  -d '{"reportId": "report-123"}'
```

**Daily Summary:**
```bash
curl http://localhost:8080/compliance/reports/daily
```

### KYC/AML

**Create KYC Record:**
```bash
curl -X POST http://localhost:8080/compliance/kyc/create \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "client-123",
    "fullName": "John Doe",
    "dob": "1990-01-15"
  }'
```

**Screen for PEP:**
```bash
curl -X POST http://localhost:8080/compliance/kyc/screen-pep \
  -H "Content-Type: application/json" \
  -d '{
    "kycId": "kyc-123",
    "fullName": "John Doe"
  }'
```

**Screen Sanctions:**
```bash
curl -X POST http://localhost:8080/compliance/kyc/screen-sanctions \
  -H "Content-Type: application/json" \
  -d '{
    "kycId": "kyc-123",
    "fullName": "John Doe",
    "nationality": "US"
  }'
```

**File SAR:**
```bash
curl -X POST http://localhost:8080/compliance/aml/file-sar \
  -H "Content-Type: application/json" \
  -d '{
    "alertId": "alert-123",
    "narrative": "Suspicious pattern detected in trading activity..."
  }'
```

### Audit Trail

**Get Audit History:**
```bash
curl "http://localhost:8080/compliance/audit/history?clientId=client-123&start=2026-01-01&end=2026-12-31"
```

**Verify Integrity:**
```bash
curl "http://localhost:8080/compliance/audit/verify?start=2026-01-01&end=2026-01-31"
```

**Export Audit Trail:**
```bash
curl "http://localhost:8080/compliance/audit/export?start=2026-01-01&end=2026-12-31&format=JSON"
```

### Best Execution

**Generate RTS 27 Report:**
```bash
curl -X POST http://localhost:8080/compliance/execution/rts27 \
  -H "Content-Type: application/json" \
  -d '{
    "year": 2026,
    "quarter": "1",
    "instrumentClass": "FOREX_MAJOR"
  }'
```

**Get Execution Quality:**
```bash
curl "http://localhost:8080/compliance/execution/quality?lp=OANDA&symbol=EURUSD&hours=24"
```

### Leverage Limits

**Validate Leverage:**
```bash
curl -X POST http://localhost:8080/compliance/leverage/validate \
  -H "Content-Type: application/json" \
  -d '{
    "jurisdiction": "EU",
    "clientClass": "RETAIL",
    "symbol": "EURUSD",
    "instrumentClass": "MAJOR_PAIRS",
    "requestedLeverage": 50
  }'
```

**Get ESMA Limits:**
```bash
curl "http://localhost:8080/compliance/leverage/esma-limits?clientClass=RETAIL"
```

## Scheduled Tasks Setup

Create a cron job or scheduled task runner:

```go
// Daily tasks (run at midnight)
func runDailyTasks(cs *compliance.ComplianceSystem) {
    // Submit pending transaction reports
    pendingReports, _ := cs.TransactionService.GetPendingReports()
    reportIDs := make([]string, len(pendingReports))
    for i, r := range pendingReports {
        reportIDs[i] = r.ID
    }
    cs.TransactionService.BatchSubmit(reportIDs)

    // Generate daily summary
    cs.TransactionService.GenerateDailyReport()

    // Reconcile segregated accounts
    // ... implementation
}

// Weekly tasks (run on Monday)
func runWeeklyTasks(cs *compliance.ComplianceSystem) {
    // Run ongoing monitoring
    cs.KYCAMLService.OngoingMonitoring()

    // Verify audit trail integrity
    startDate := time.Now().AddDate(0, 0, -7)
    endDate := time.Now()
    cs.AuditService.VerifyIntegrity(startDate, endDate)
}

// Quarterly tasks
func runQuarterlyTasks(cs *compliance.ComplianceSystem) {
    year := time.Now().Year()
    quarter := getQuarter(time.Now())

    // Generate RTS 27 reports for all instrument classes
    cs.BestExecService.GenerateRTS27Report(year, quarter, "FOREX_MAJOR")
    cs.BestExecService.GenerateRTS27Report(year, quarter, "FOREX_MINOR")
    cs.BestExecService.GenerateRTS27Report(year, quarter, "GOLD")
    cs.BestExecService.GenerateRTS27Report(year, quarter, "INDICES")
    cs.BestExecService.GenerateRTS27Report(year, quarter, "COMMODITIES")
    cs.BestExecService.GenerateRTS27Report(year, quarter, "CRYPTO")
}

// Annual tasks
func runAnnualTasks(cs *compliance.ComplianceSystem) {
    year := time.Now().Year() - 1 // Previous year

    // Generate RTS 28 reports
    cs.BestExecService.GenerateRTS28Report(year, "FOREX_MAJOR")
    cs.BestExecService.GenerateRTS28Report(year, "FOREX_MINOR")
    // ... other instrument classes
}
```

## Database Migration (Production)

When moving to production, migrate from in-memory to PostgreSQL:

```sql
-- Transaction Reports
CREATE TABLE transaction_reports (
    id VARCHAR(255) PRIMARY KEY,
    report_type VARCHAR(50) NOT NULL,
    jurisdiction VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    order_id VARCHAR(255) NOT NULL,
    execution_id VARCHAR(255),
    client_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    isin VARCHAR(50),
    side VARCHAR(10) NOT NULL,
    quantity DECIMAL(18,8) NOT NULL,
    price DECIMAL(18,8) NOT NULL,
    execution_timestamp TIMESTAMP NOT NULL,
    trading_venue VARCHAR(100),
    liquidity_provider VARCHAR(100),
    currency VARCHAR(10),
    client_classification VARCHAR(50),
    investment_decision_maker VARCHAR(255),
    executing_trader VARCHAR(255),
    transmission_of_order BOOLEAN,
    buyer_identification VARCHAR(255),
    seller_identification VARCHAR(255),
    short_selling_indicator BOOLEAN,
    waiver_indicator VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'PENDING',
    INDEX idx_client_id (client_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- KYC Records
CREATE TABLE kyc_records (
    id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    nationality VARCHAR(10),
    residence_country VARCHAR(10),
    address TEXT,
    document_type VARCHAR(50),
    document_number VARCHAR(100),
    document_expiry DATE,
    document_verified BOOLEAN DEFAULT FALSE,
    verification_provider VARCHAR(100),
    proof_of_address VARCHAR(255),
    address_verified BOOLEAN DEFAULT FALSE,
    pep_status VARCHAR(50) DEFAULT 'NOT_PEP',
    sanctions_match BOOLEAN DEFAULT FALSE,
    sanctions_lists JSON,
    risk_rating VARCHAR(50) DEFAULT 'MEDIUM',
    ongoing_monitoring BOOLEAN DEFAULT TRUE,
    last_screening TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_client_id (client_id),
    INDEX idx_risk_rating (risk_rating),
    INDEX idx_last_screening (last_screening)
);

-- Audit Trail
CREATE TABLE audit_trail (
    id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    user_id VARCHAR(255),
    user_role VARCHAR(50),
    client_id VARCHAR(255),
    order_id VARCHAR(255),
    trade_id VARCHAR(255),
    symbol VARCHAR(50),
    action VARCHAR(100),
    before_state JSON,
    after_state JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL,
    hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64),
    INDEX idx_client_id (client_id),
    INDEX idx_event_type (event_type),
    INDEX idx_timestamp (timestamp),
    INDEX idx_hash (hash)
);

-- Best Execution Reports
CREATE TABLE best_execution_reports (
    id VARCHAR(255) PRIMARY KEY,
    report_type VARCHAR(50) NOT NULL,
    period VARCHAR(50) NOT NULL,
    instrument_class VARCHAR(100),
    venue VARCHAR(100),
    lp_name VARCHAR(100),
    price_improvement DECIMAL(10,4),
    fill_rate DECIMAL(5,4),
    average_execution_time DECIMAL(10,2),
    slippage_rate DECIMAL(10,4),
    rejection_rate DECIMAL(5,4),
    execution_volume DECIMAL(18,2),
    number_of_orders INT,
    passive_orders INT,
    aggressive_orders INT,
    directed_orders INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    INDEX idx_report_type (report_type),
    INDEX idx_period (period),
    INDEX idx_instrument_class (instrument_class)
);

-- AML Alerts
CREATE TABLE aml_alerts (
    id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    description TEXT,
    severity VARCHAR(50) NOT NULL,
    transaction_ids JSON,
    detected_at TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    assigned_to VARCHAR(255),
    sar_filed BOOLEAN DEFAULT FALSE,
    sar_filed_at TIMESTAMP,
    resolution TEXT,
    resolved_at TIMESTAMP,
    INDEX idx_client_id (client_id),
    INDEX idx_status (status),
    INDEX idx_severity (severity),
    INDEX idx_detected_at (detected_at)
);
```

## Environment Configuration

Create `.env` file:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_compliance
DB_USER=compliance_user
DB_PASSWORD=secure_password

# KYC Providers
ONFIDO_API_KEY=your_onfido_api_key
JUMIO_API_TOKEN=your_jumio_token
JUMIO_API_SECRET=your_jumio_secret
TRULIOO_API_KEY=your_trulioo_key

# Regulatory Endpoints
MIFID_SUBMISSION_URL=https://regulatory.example.com/mifid
EMIR_SUBMISSION_URL=https://regulatory.example.com/emir
CAT_SUBMISSION_URL=https://regulatory.example.com/cat

# Sanctions Lists (update URLs)
OFAC_LIST_URL=https://www.treasury.gov/ofac/downloads/sdnlist.txt
UN_LIST_URL=https://www.un.org/securitycouncil/sanctions/1267/aq_sanctions_list
EU_LIST_URL=https://webgate.ec.europa.eu/fsd/fsf

# Application
COMPLIANCE_LOG_LEVEL=info
AUDIT_RETENTION_YEARS=7
KYC_SCREENING_INTERVAL_DAYS=90
```

## Monitoring Setup

Add monitoring for compliance system:

```go
// Prometheus metrics
var (
    transactionReportsSubmitted = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "compliance_transaction_reports_submitted_total",
            Help: "Total number of transaction reports submitted",
        },
    )

    amlAlertsCreated = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "compliance_aml_alerts_created_total",
            Help: "Total number of AML alerts created",
        },
    )

    auditTrailEntriesLogged = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "compliance_audit_trail_entries_total",
            Help: "Total number of audit trail entries logged",
        },
    )
)
```

## Testing

Run compliance system tests:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend/compliance
go test ./... -v
```

## Support

For compliance integration support:
- Chief Compliance Officer: compliance@yourcompany.com
- Technical Support: techsupport@yourcompany.com
- Regulatory Affairs: regulatory@yourcompany.com
