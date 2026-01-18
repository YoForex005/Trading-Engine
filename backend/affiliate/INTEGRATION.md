# Affiliate System Integration Guide

## Quick Start

### 1. Import the Package
Add to your `main.go`:
```go
import "github.com/epic1st/rtx/backend/affiliate"
```

### 2. Initialize the System
```go
// Create affiliate API
affiliateAPI := affiliate.NewAffiliateAPI()

// Register all affiliate routes
affiliateAPI.RegisterRoutes()
```

### 3. Run Database Migrations
```bash
psql -U postgres -d trading_engine -f backend/affiliate/schema.sql
```

## Integration Points

### A. User Signup Integration

When a new user signs up, check for affiliate tracking:

```go
func handleUserSignup(w http.ResponseWriter, r *http.Request) {
    // ... user creation logic ...

    // Check for affiliate tracking cookie
    cookie, err := r.Cookie("aff_click")
    if err == nil && cookie.Value != "" {
        // Track conversion
        go func(clickID, userID string, accountID int64) {
            // Get tracking manager from affiliate API
            conversion, err := affiliateAPI.TrackingManager().TrackConversion(
                clickID,
                userID,
                accountID,
                "SIGNUP",
                0,
            )

            if err != nil {
                log.Printf("Failed to track conversion: %v", err)
                return
            }

            // Calculate CPA commission
            if conversion != nil {
                cookie, ok := affiliateAPI.TrackingManager().GetClickByCookie(clickID)
                if ok {
                    commission, err := affiliateAPI.CommissionManager().CalculateCPACommission(
                        cookie.AffiliateID,
                        conversion.ID,
                        accountID,
                    )

                    if err == nil {
                        log.Printf("Created CPA commission: $%.2f for affiliate %d",
                            commission.Amount, cookie.AffiliateID)
                    }
                }
            }
        }(cookie.Value, newUser.ID, newAccount.ID)
    }
}
```

### B. Deposit Integration

Track deposit conversions for commission calculation:

```go
func handleDeposit(w http.ResponseWriter, r *http.Request) {
    // ... deposit processing logic ...

    // Track deposit conversion
    go func(userID string, accountID int64, amount float64) {
        // Get click ID from user metadata (stored during signup)
        clickID := getUserAffiliateClickID(userID)
        if clickID != "" {
            affiliateAPI.TrackingManager().TrackConversion(
                clickID,
                userID,
                accountID,
                "DEPOSIT",
                amount,
            )
        }
    }(user.ID, account.ID, depositAmount)
}
```

### C. Trading Fee Integration

Calculate monthly revenue share based on trading fees:

```go
// Run this monthly (e.g., via cron job)
func calculateMonthlyRevShare() {
    // Get all accounts with affiliate attribution
    accounts := getAccountsWithAffiliates()

    period := time.Now().Format("2006-01")

    // Group by affiliate
    affiliateData := make(map[int64]map[int64]struct{
        Volume float64
        Fees   float64
    })

    for _, account := range accounts {
        affiliateID := account.AffiliateID
        if affiliateID == 0 {
            continue
        }

        if _, ok := affiliateData[affiliateID]; !ok {
            affiliateData[affiliateID] = make(map[int64]struct{
                Volume float64
                Fees   float64
            })
        }

        // Get account trading stats for the month
        stats := getAccountTradingStats(account.ID, period)

        affiliateData[affiliateID][account.ID] = struct{
            Volume float64
            Fees   float64
        }{
            Volume: stats.TotalVolume,
            Fees:   stats.TotalFees,
        }
    }

    // Calculate commissions for each affiliate
    for affiliateID, accounts := range affiliateData {
        for accountID, data := range accounts {
            _, err := affiliateAPI.CommissionManager().CalculateRevShareCommission(
                affiliateID,
                accountID,
                data.Volume,
                data.Fees,
                period,
            )

            if err != nil {
                log.Printf("Failed to calculate RevShare for affiliate %d: %v", affiliateID, err)
            }
        }
    }

    log.Printf("Monthly RevShare calculation completed for period: %s", period)
}
```

### D. Referral Program Integration

Add referral code input to signup form:

```go
func handleSignupWithReferral(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email        string `json:"email"`
        Password     string `json:"password"`
        ReferralCode string `json:"referralCode"`
    }

    json.NewDecoder(r.Body).Decode(&req)

    // ... create user and account ...

    // Apply referral code if provided
    if req.ReferralCode != "" {
        reward, err := affiliateAPI.ReferralManager().ApplyReferralCode(
            req.ReferralCode,
            newUser.ID,
        )

        if err == nil {
            // Credit both parties
            affiliateAPI.ReferralManager().CreditReferralReward(reward.ID)

            // Add bonus to both accounts
            addBonusToAccount(reward.ReferrerUserID, reward.ReferrerReward)
            addBonusToAccount(reward.RefereeUserID, reward.RefereeReward)

            log.Printf("Referral reward credited: Referrer=$%.2f, Referee=$%.2f",
                reward.ReferrerReward, reward.RefereeReward)
        }
    }
}
```

### E. Admin Panel Integration

Add affiliate management to admin panel:

```go
// Admin routes
http.HandleFunc("/admin/affiliates", affiliateAPI.HandleAdminListAffiliates)
http.HandleFunc("/admin/affiliates/approve", affiliateAPI.HandleAdminApproveAffiliate)
http.HandleFunc("/admin/affiliates/suspend", affiliateAPI.HandleAdminSuspendAffiliate)
http.HandleFunc("/admin/commissions", affiliateAPI.HandleAdminGetCommissions)
http.HandleFunc("/admin/payouts/process", affiliateAPI.HandleAdminProcessPayout)
```

## Automated Tasks

### 1. Auto-Approve Old Commissions

Run daily to approve commissions older than 30 days:

```go
func autoApproveCommissions() {
    approved := affiliateAPI.CommissionManager().AutoApprovePendingCommissions(30)
    log.Printf("Auto-approved %d commissions", approved)
}
```

### 2. Process Scheduled Payouts

Run monthly for automatic payouts:

```go
func processScheduledPayouts() {
    // Get all active affiliates with minimum balance
    affiliates := affiliateAPI.ProgramManager().ListAffiliates("ACTIVE")
    program := affiliateAPI.ProgramManager().GetActiveProgram()

    for _, aff := range affiliates {
        if aff.PendingBalance >= program.MinPayout {
            // Create payout
            payout, err := affiliateAPI.CommissionManager().ProcessPayout(
                aff.ID,
                aff.PendingBalance,
                aff.PayoutMethod,
            )

            if err != nil {
                log.Printf("Failed to create payout for affiliate %d: %v", aff.ID, err)
                continue
            }

            // Process payment (integrate with payment gateway)
            err = processPayment(payout)
            if err != nil {
                affiliateAPI.CommissionManager().FailPayout(payout.ID, err.Error())
                log.Printf("Failed to process payout %d: %v", payout.ID, err)
                continue
            }

            // Mark as completed
            affiliateAPI.CommissionManager().CompletePayout(payout.ID, "TXN-"+generateID())
            log.Printf("Payout completed for affiliate %d: $%.2f", aff.ID, payout.Amount)
        }
    }
}
```

### 3. Fraud Detection Scan

Run hourly to detect fraud patterns:

```go
func scanForFraud() {
    // This is handled automatically in TrackClick
    // But you can add additional checks here

    // Example: Flag affiliates with high fraud scores
    affiliates := affiliateAPI.ProgramManager().ListAffiliates("ACTIVE")
    for _, aff := range affiliates {
        if aff.FraudScore > 70 {
            // Suspend affiliate
            affiliateAPI.ProgramManager().SuspendAffiliate(aff.ID, "High fraud score")

            // Notify admin
            sendAdminNotification(fmt.Sprintf(
                "Affiliate %s (ID=%d) suspended due to high fraud score: %.2f",
                aff.ContactName, aff.ID, aff.FraudScore,
            ))
        }
    }
}
```

## Environment Configuration

Add to your `.env` or config file:

```bash
# Affiliate System Configuration
AFFILIATE_COOKIE_DURATION=30          # Days
AFFILIATE_MIN_PAYOUT=100.00          # USD
AFFILIATE_CPA_AMOUNT=100.00          # USD
AFFILIATE_REVSHARE_PERCENT=25.00     # %
AFFILIATE_PAYOUT_SCHEDULE=MONTHLY    # MONTHLY, BIWEEKLY, ON_DEMAND
AFFILIATE_FRAUD_THRESHOLD=70         # Score 0-100
AFFILIATE_BASE_URL=https://yourbroker.com
```

## Frontend Integration

### Landing Page with Tracking

```html
<!DOCTYPE html>
<html>
<head>
    <title>Sign Up - Your Broker</title>
</head>
<body>
    <!-- Tracking pixel for email campaigns -->
    <img src="/track/pixel.gif?ref=ABC123XY" width="1" height="1" style="display:none" />

    <form action="/signup" method="POST">
        <input type="email" name="email" placeholder="Email" required />
        <input type="password" name="password" placeholder="Password" required />

        <!-- Hidden field for referral code (from URL) -->
        <input type="hidden" name="referralCode" id="referralCode" />

        <button type="submit">Sign Up</button>
    </form>

    <script>
        // Extract referral code from URL
        const urlParams = new URLSearchParams(window.location.search);
        const ref = urlParams.get('ref');
        if (ref) {
            document.getElementById('referralCode').value = ref;
        }
    </script>
</body>
</html>
```

### Affiliate Dashboard (React Example)

```jsx
import React, { useEffect, useState } from 'react';

function AffiliateDashboard() {
    const [stats, setStats] = useState(null);

    useEffect(() => {
        fetch('/api/affiliate/dashboard', {
            headers: {
                'X-Affiliate-ID': localStorage.getItem('affiliateId')
            }
        })
        .then(res => res.json())
        .then(data => setStats(data));
    }, []);

    if (!stats) return <div>Loading...</div>;

    return (
        <div className="dashboard">
            <h1>Affiliate Dashboard</h1>

            <div className="stats-grid">
                <div className="stat-card">
                    <h3>Today's Clicks</h3>
                    <p className="stat-value">{stats.todayClicks}</p>
                </div>

                <div className="stat-card">
                    <h3>Today's Signups</h3>
                    <p className="stat-value">{stats.todaySignups}</p>
                </div>

                <div className="stat-card">
                    <h3>Today's Earnings</h3>
                    <p className="stat-value">${stats.todayEarnings.toFixed(2)}</p>
                </div>

                <div className="stat-card">
                    <h3>Pending Balance</h3>
                    <p className="stat-value">${stats.pendingBalance.toFixed(2)}</p>
                </div>
            </div>

            <div className="conversion-rate">
                <h3>Conversion Rate</h3>
                <p>{stats.conversionRate.toFixed(2)}%</p>
            </div>

            <div className="links">
                <h3>Your Tracking Links</h3>
                <button onClick={() => createLink()}>Create New Link</button>
            </div>
        </div>
    );
}
```

## Testing

Run the test suite:

```bash
cd backend/affiliate
go test -v
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## Monitoring

### Key Metrics to Track

1. **Click Tracking Performance**
   - Target: <10ms per click
   - Monitor: Request latency

2. **Fraud Detection Rate**
   - Target: <5% false positives
   - Monitor: Fraud score distribution

3. **Conversion Rate**
   - Target: >5% click-to-signup
   - Monitor: Funnel drop-off points

4. **Commission Accuracy**
   - Target: 100% accuracy
   - Monitor: Manual audit vs. automated

5. **Payout Success Rate**
   - Target: >95% success
   - Monitor: Failed payouts

### Logging

All key operations are logged:

```go
log.Printf("[Affiliate] Registered new affiliate: %s (Code=%s)", affiliate.ContactName, affiliate.AffiliateCode)
log.Printf("[Tracking] Click recorded: LinkCode=%s, ClickID=%s, Fraud=%v", linkCode, clickID, isFraudulent)
log.Printf("[Commission] CPA commission created: Affiliate=%d, Amount=%.2f", affiliateID, cpa)
log.Printf("[Payout] Completed payout: ID=%d, Affiliate=%d, Amount=%.2f", payoutID, affiliateID, amount)
```

## Support & Maintenance

### Database Maintenance

```sql
-- Clean up old clicks (older than 90 days)
DELETE FROM affiliate_clicks WHERE created_at < NOW() - INTERVAL '90 days';

-- Optimize tables
VACUUM ANALYZE affiliate_clicks;
VACUUM ANALYZE affiliate_conversions;
VACUUM ANALYZE affiliate_commissions;
```

### Troubleshooting

**Issue: Clicks not tracking**
- Check cookie settings (HttpOnly, Secure, SameSite)
- Verify tracking pixel is loaded
- Check for ad blockers

**Issue: Commissions not calculating**
- Verify conversion was approved
- Check affiliate commission rates
- Review fraud detection flags

**Issue: Payouts failing**
- Verify payment method details
- Check minimum payout threshold
- Review transaction logs

## Security Checklist

- [ ] Use HTTPS for all tracking links
- [ ] Implement rate limiting on tracking endpoints
- [ ] Validate all inputs
- [ ] Use prepared statements for database queries
- [ ] Encrypt sensitive data (bank details, crypto addresses)
- [ ] Implement CSRF protection
- [ ] Regular security audits
- [ ] Monitor for suspicious activity
- [ ] Keep fraud detection rules updated

## Next Steps

1. Configure environment variables
2. Run database migrations
3. Initialize affiliate API in main.go
4. Integrate tracking in signup flow
5. Test with sample affiliate
6. Configure automated tasks (cron jobs)
7. Set up monitoring and alerts
8. Train admin team on affiliate management

For questions or support, refer to the main README.md or contact the development team.
