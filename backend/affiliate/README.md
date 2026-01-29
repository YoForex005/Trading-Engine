# Affiliate & Referral Program System

A comprehensive affiliate marketing and referral program system designed for trading platforms. Scalable to 10,000+ affiliates with automated fraud detection, multi-tier commissions, and real-time analytics.

## Features

### 1. Affiliate Program Management
- **Multi-Tier Commission Structure**
  - CPA (Cost Per Acquisition): One-time payment per signup
  - RevShare (Revenue Share): Percentage of trading fees for lifetime
  - Hybrid: CPA + reduced RevShare
  - Sub-affiliate support (MLM structure) with configurable percentages

- **Flexible Commission Tiers**
  - 5-tier system with automatic multipliers
  - Custom commission rates per affiliate
  - Performance-based tier upgrades

### 2. Advanced Tracking System
- **Click Tracking**
  - Real-time click recording with geo-location
  - Unique click detection (IP + UserAgent + 24hr window)
  - Device, browser, and OS detection
  - Referrer and UTM parameter tracking
  - 30-day cookie tracking (configurable)

- **Conversion Tracking**
  - Multiple conversion types: SIGNUP, DEPOSIT, FIRST_TRADE
  - Attribution models: First-click, Last-click, Linear
  - Cross-device tracking support
  - Conversion funnel analysis

- **Fraud Detection**
  - IP velocity checking (max clicks per hour)
  - Private IP range blocking
  - Bot detection via user agent analysis
  - Click fraud prevention
  - Cookie stuffing detection
  - Duplicate account detection
  - Fraud scoring system (0-100)

### 3. Commission Management
- **Automated Calculation**
  - CPA commissions on signup
  - RevShare calculated monthly on trading fees
  - Sub-affiliate commissions (percentage of parent earnings)
  - Automatic approval after configurable days

- **Commission States**
  - PENDING: Awaiting approval
  - APPROVED: Ready for payout
  - PAID: Included in payout
  - REVERSED: Charged back

### 4. Payout System
- **Multiple Payment Methods**
  - Bank transfer
  - PayPal
  - Cryptocurrency
  - Wire transfer

- **Payout Schedules**
  - Monthly (default)
  - Bi-weekly
  - On-demand (minimum balance required)

- **Automated Processing**
  - Minimum payout thresholds
  - Bulk payout processing
  - Transaction tracking
  - Tax form generation (1099-MISC)

### 5. User Referral Program
- **Referral Codes**
  - Unique codes per user
  - Both parties receive bonuses
  - Social sharing integration
  - Maximum uses limits
  - Expiration dates

- **Gamification**
  - Achievement badges (First Referral, Rising Star, Influencer, etc.)
  - Leaderboard rankings
  - Referral contests with prizes
  - Progress tracking

- **Social Sharing**
  - Pre-built share links for:
    - Facebook
    - Twitter
    - WhatsApp
    - Email
  - Tracking pixel for email campaigns

### 6. Marketing Tools
- **Materials Library**
  - Banner ads (multiple sizes)
  - Email templates
  - Landing pages
  - Video content
  - Social media posts
  - SEO-optimized content

- **Link Builder**
  - Custom tracking links
  - Campaign tagging (UTM parameters)
  - Landing page selection
  - QR code generation

### 7. Analytics Dashboard
- **Real-Time Statistics**
  - Today/Week/Month/Lifetime metrics
  - Clicks, Signups, Deposits, Earnings
  - Conversion rates
  - Pending/Available balance

- **Performance Analysis**
  - Conversion funnel visualization
  - Traffic source breakdown
  - Geographic performance
  - Time-series data (hourly/daily/monthly)
  - Link performance comparison

- **Advanced Reports**
  - Custom date ranges
  - Period comparison
  - Export to CSV/PDF
  - Scheduled email reports

### 8. Admin Features
- **Affiliate Management**
  - Application approval/rejection
  - Status management (Active, Suspended, Banned)
  - Custom commission rates
  - Tier assignments
  - Fraud investigation tools

- **Commission Review**
  - Bulk approval
  - Manual reversals
  - Fraud flagging
  - Adjustment capabilities

- **Payout Processing**
  - Batch processing
  - Payment verification
  - Transaction tracking
  - Failure handling

- **Compliance**
  - Tax ID collection
  - KYC verification
  - Audit trails
  - Regulatory reporting

## Architecture

### Core Components

1. **ProgramManager** (`program.go`)
   - Affiliate registration and approval
   - Commission tier management
   - Sub-affiliate relationships
   - Program configuration

2. **TrackingManager** (`tracking.go`)
   - Click recording and cookie management
   - Conversion tracking
   - Attribution logic
   - Fraud detection

3. **CommissionManager** (`commissions.go`)
   - Commission calculation (CPA, RevShare, Hybrid)
   - Approval workflow
   - Payout processing
   - Chargeback handling

4. **ReferralManager** (`referrals.go`)
   - User referral code generation
   - Reward distribution
   - Gamification features
   - Contest management

5. **AnalyticsEngine** (`analytics.go`)
   - Real-time dashboard statistics
   - Conversion funnel analysis
   - Traffic source breakdown
   - Performance reporting

## API Endpoints

### Public Endpoints (No Auth)
```
GET  /track/click?ref=CODE          - Track affiliate click
GET  /track/pixel.gif?ref=CODE      - Tracking pixel (1x1 GIF)
```

### Affiliate Endpoints
```
POST /api/affiliate/register         - Register new affiliate
POST /api/affiliate/login            - Affiliate login
GET  /api/affiliate/profile          - Get profile
PUT  /api/affiliate/update           - Update profile

GET  /api/affiliate/dashboard        - Dashboard statistics
GET  /api/affiliate/stats            - Detailed statistics
GET  /api/affiliate/funnel           - Conversion funnel data
GET  /api/affiliate/traffic          - Traffic source analysis
GET  /api/affiliate/performance      - Performance report

GET  /api/affiliate/links            - List tracking links
POST /api/affiliate/links/create     - Create tracking link

GET  /api/affiliate/commissions      - Commission history
GET  /api/affiliate/payouts          - Payout history
POST /api/affiliate/payout/request   - Request payout

GET  /api/affiliate/materials        - Marketing materials
```

### Referral Endpoints
```
GET  /api/referral/code              - Get user's referral code
POST /api/referral/code              - Create referral code
POST /api/referral/apply             - Apply referral code
GET  /api/referral/stats             - Referral statistics
GET  /api/referral/leaderboard       - Top referrers
```

### Admin Endpoints
```
GET  /admin/affiliate/list           - List all affiliates
POST /admin/affiliate/approve        - Approve affiliate
POST /admin/affiliate/suspend        - Suspend affiliate
POST /admin/affiliate/ban            - Ban affiliate

GET  /admin/affiliate/commissions    - All commissions
POST /admin/affiliate/commissions/approve - Approve commission
POST /admin/affiliate/commissions/reverse - Reverse commission

GET  /admin/affiliate/payouts        - All payouts
POST /admin/affiliate/payouts/process - Process payout
POST /admin/affiliate/payouts/complete - Mark payout complete
POST /admin/affiliate/payouts/fail   - Mark payout failed

GET  /admin/affiliate/fraud          - Fraud incidents
POST /admin/affiliate/fraud/resolve  - Resolve fraud case
```

## Database Schema

See `schema.sql` for the complete database structure including:
- affiliate_programs
- affiliates
- affiliate_links
- affiliate_clicks
- affiliate_conversions
- affiliate_commissions
- affiliate_payouts
- referral_codes
- referral_rewards
- marketing_materials
- fraud_detection_rules
- fraud_incidents

## Usage Example

### 1. Initialize the System
```go
import "github.com/epic1st/rtx/backend/affiliate"

// Create API handler
affiliateAPI := affiliate.NewAffiliateAPI()

// Register routes
affiliateAPI.RegisterRoutes()
```

### 2. Register an Affiliate
```bash
curl -X POST http://localhost:7999/api/affiliate/register \
  -H "Content-Type: application/json" \
  -d '{
    "companyName": "Marketing Pro Inc",
    "contactName": "John Doe",
    "email": "john@marketingpro.com",
    "phone": "+1-555-1234",
    "country": "USA",
    "website": "https://marketingpro.com"
  }'
```

### 3. Create Tracking Link
```bash
curl -X POST http://localhost:7999/api/affiliate/links/create \
  -H "Content-Type: application/json" \
  -H "X-Affiliate-ID: 1" \
  -d '{
    "campaign": "summer-promo",
    "source": "facebook",
    "medium": "cpc",
    "content": "banner-728x90"
  }'
```

### 4. Track Click
```
https://yourbroker.com/track/click?ref=ABC123XY&utm_campaign=summer-promo
```

### 5. Record Conversion
```go
// After user signs up
cookie, _ := trackingManager.GetClickByCookie(clickID)
conversion, _ := trackingManager.TrackConversion(
    clickID,
    userID,
    accountID,
    "SIGNUP",
    0,
)

// Calculate CPA commission
commission, _ := commissionManager.CalculateCPACommission(
    cookie.AffiliateID,
    conversion.ID,
    accountID,
)
```

### 6. Process Monthly RevShare
```go
// Calculate monthly revenue share
accountData := map[int64]struct{
    Volume float64
    Fees   float64
}{
    12345: {Volume: 1000000, Fees: 2500},
    67890: {Volume: 500000, Fees: 1200},
}

commissionManager.CalculateMonthlyRevShare("2024-01", accountData)
```

## Fraud Prevention

The system includes multiple layers of fraud detection:

1. **IP-based Detection**
   - Rate limiting (max clicks per hour)
   - Private IP blocking
   - Geo-location verification

2. **Behavioral Analysis**
   - Click patterns
   - Conversion timing
   - Device fingerprinting

3. **Bot Detection**
   - User agent analysis
   - JavaScript challenges
   - CAPTCHA integration points

4. **Account Verification**
   - Duplicate account detection
   - KYC requirements
   - Email/phone verification

## Scalability

The system is designed to scale to 10,000+ affiliates:

- In-memory caching with sync.RWMutex
- Database indexing on critical fields
- TimescaleDB support for click history
- Batch processing for payouts
- Background workers for analytics
- Rate limiting on API endpoints

## Integration Points

### With Trading Platform
```go
// On user signup
if clickID := r.Cookie("aff_click"); clickID != nil {
    affiliateAPI.trackingManager.TrackConversion(
        clickID.Value,
        user.ID,
        account.ID,
        "SIGNUP",
        0,
    )
}

// On first deposit
affiliateAPI.trackingManager.TrackConversion(
    clickID,
    user.ID,
    account.ID,
    "DEPOSIT",
    depositAmount,
)

// Monthly revenue share calculation
go affiliateAPI.commissionMgr.CalculateMonthlyRevShare(
    period,
    accountTradingData,
)
```

## Security Considerations

1. **Authentication**: Use JWT tokens for all affiliate endpoints
2. **Rate Limiting**: Implement rate limits on tracking endpoints
3. **HTTPS**: All tracking links must use HTTPS
4. **Cookie Security**: HttpOnly, Secure, SameSite flags
5. **Input Validation**: Sanitize all user inputs
6. **SQL Injection**: Use parameterized queries
7. **XSS Protection**: Encode all outputs

## Performance Metrics

- Click tracking: <10ms latency
- Conversion recording: <50ms latency
- Dashboard load: <500ms
- Fraud detection: Real-time (synchronous)
- Monthly RevShare calculation: <5 minutes for 10,000 affiliates

## Future Enhancements

- [ ] Machine learning fraud detection
- [ ] A/B testing for landing pages
- [ ] Email automation for affiliates
- [ ] Mobile app for affiliates
- [ ] Advanced geo-targeting rules
- [ ] Multi-currency support
- [ ] API rate limiting per affiliate
- [ ] Webhook notifications
- [ ] White-label affiliate portal
- [ ] Integration with ad networks

## Support

For questions or issues, contact the development team.
