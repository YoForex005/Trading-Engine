# Security Incident Response Playbook

## Incident Classification

### Severity Levels

| Level | Description | Response Time | Examples |
|-------|-------------|---------------|----------|
| **P0 - Critical** | Active breach, data exfiltration | Immediate (< 15 min) | Database breach, ransomware |
| **P1 - High** | Potential breach, system compromise | < 1 hour | Unauthorized access, malware detected |
| **P2 - Medium** | Security event, no immediate threat | < 4 hours | Failed login attempts spike, DoS attack |
| **P3 - Low** | Security alert, monitoring required | < 24 hours | Vulnerability discovered, policy violation |

## Incident Response Team

### Roles & Responsibilities

**Incident Commander**
- Overall coordination
- Decision making authority
- External communication

**Security Lead**
- Technical investigation
- Threat analysis
- Remediation planning

**DevOps Lead**
- System access
- Infrastructure changes
- Service restoration

**Compliance Officer**
- Regulatory requirements
- Legal consultation
- Customer notification

**Communications Lead**
- Internal communication
- Customer updates
- PR management

### Contact Information
```
Incident Commander: [Name] - [Phone] - [Email]
Security Lead: [Name] - [Phone] - [Email]
DevOps Lead: [Name] - [Phone] - [Email]
Compliance: [Name] - [Phone] - [Email]
Legal: [Name] - [Phone] - [Email]
```

## Response Procedures

### 1. Detection & Analysis

#### Automated Detection
Monitor for:
- WAF blocked IPs spike (> 100/hour)
- Failed authentication attempts (> 50/hour)
- Unusual data access patterns
- Security test failures
- Compliance check failures
- Abnormal API key usage

#### Manual Detection
Review:
- Audit logs daily
- Security alerts
- User reports
- External notifications

#### Initial Assessment Checklist
- [ ] Confirm incident is real (not false positive)
- [ ] Classify severity level
- [ ] Determine affected systems
- [ ] Identify potential data exposure
- [ ] Estimate user impact
- [ ] Document initial findings

### 2. Containment

#### Immediate Actions (Within 15 Minutes)

**For Active Breach (P0):**
```bash
# 1. Block attacker IPs at WAF
waf.BlockIPPermanent("attacker-ip")

# 2. Invalidate all sessions
sessionManager.DestroyAllSessions()

# 3. Rotate all API keys
keyManager.RotateAllKeys("emergency_breach")

# 4. Enable enhanced logging
auditLogger.SetLevel(AuditLevelSecurity)

# 5. Snapshot system for forensics
# Take database snapshot, preserve logs
```

**For System Compromise (P1):**
```bash
# 1. Isolate affected systems
# Disable network access or restrict to known good IPs

# 2. Block suspicious IPs
waf.BlockIP(suspiciousIP, 24*time.Hour)

# 3. Increase monitoring
# Enable verbose logging

# 4. Preserve evidence
# Copy audit logs to secure location
```

**For DoS/DDoS (P2):**
```bash
# 1. Enable aggressive rate limiting
wafConfig.MaxRequestsPerIP = 10
wafConfig.MaxConcurrentConns = 5000

# 2. Block attacking IPs/ranges
# Use IP ranges if distributed attack

# 3. Enable CDN DDoS protection
# Contact CloudFlare/Akamai

# 4. Notify hosting provider
```

### 3. Eradication

#### Root Cause Analysis
1. Review audit logs around incident time
   ```go
   events, _ := auditLogger.QueryLogs(
       startTime, endTime,
       AuditLevelSecurity,
       "authentication",
   )
   ```

2. Run vulnerability scanner
   ```go
   scanner := security.NewVulnerabilityScanner(auditLogger)
   report, _ := scanner.ScanDirectory("/backend")
   ```

3. Check compliance status
   ```go
   complianceChecker := security.NewComplianceChecker(auditLogger)
   report, _ := complianceChecker.RunAllChecks()
   ```

4. Identify attack vector
   - SQL injection?
   - XSS exploit?
   - Credential theft?
   - Zero-day vulnerability?

#### Remediation Steps

**Close Vulnerability:**
- [ ] Apply security patch
- [ ] Update vulnerable dependencies
- [ ] Fix code vulnerability
- [ ] Run security tests to verify fix

**Strengthen Defenses:**
- [ ] Update WAF rules
- [ ] Add new compliance checks
- [ ] Enhance input validation
- [ ] Add additional logging

**Verify No Backdoors:**
- [ ] Scan for malware
- [ ] Review user accounts
- [ ] Check for unauthorized access
- [ ] Verify system integrity

### 4. Recovery

#### System Restoration Checklist
- [ ] Verify all vulnerabilities patched
- [ ] Run full security test suite
- [ ] Restore from clean backup if needed
- [ ] Rotate all credentials and API keys
- [ ] Re-enable affected services gradually
- [ ] Monitor closely for 24-48 hours

#### Verification Tests
```go
// Run comprehensive security tests
testSuite := security.NewSecurityTestSuite(productionURL)
testSuite.RegisterDefaultTests()
report := testSuite.RunAll()

if report.Failed > 0 {
    // DO NOT proceed with recovery
    log.Fatal("Security tests failing, cannot restore service")
}
```

### 5. Post-Incident Activities

#### Documentation (Within 24 Hours)
- [ ] Incident timeline
- [ ] Attack vector analysis
- [ ] Systems affected
- [ ] Data compromised (if any)
- [ ] Actions taken
- [ ] Lessons learned

#### Communication

**Internal:**
- [ ] Brief incident response team
- [ ] Update engineering team
- [ ] Inform management
- [ ] Document in internal wiki

**External (if required):**
- [ ] Notify affected users (GDPR: 72 hours)
- [ ] Regulatory reporting (varies by jurisdiction)
- [ ] Insurance notification (cyber insurance)
- [ ] Law enforcement (if criminal activity)

#### Improvement Actions
- [ ] Update incident response plan
- [ ] Add new security controls
- [ ] Enhance monitoring
- [ ] Schedule security training
- [ ] Conduct tabletop exercise

## Specific Incident Scenarios

### Scenario 1: SQL Injection Attack

**Detection:**
- WAF blocks SQL injection patterns
- Audit logs show SQL errors

**Response:**
1. Block attacking IPs
2. Review all database queries for injection vulnerabilities
3. Verify parameterized queries everywhere
4. Run database audit trail check
5. Check for data exfiltration

**Prevention:**
- Use parameterized queries exclusively
- Add SQL injection patterns to WAF
- Enable database query logging

### Scenario 2: Credential Stuffing Attack

**Detection:**
- Spike in failed login attempts
- Multiple IPs trying same credentials
- Rate limiter triggering frequently

**Response:**
1. Enable aggressive rate limiting
2. Block attacking IPs
3. Notify affected users
4. Force password reset for compromised accounts
5. Enable 2FA

**Prevention:**
- Implement CAPTCHA on login
- Require strong passwords
- Enable account lockout
- Monitor for credential dumps

### Scenario 3: API Key Compromise

**Detection:**
- API key used from unauthorized IP
- Unusual API call patterns
- External notification

**Response:**
1. Immediately revoke compromised key
   ```go
   keyManager.RevokeKey("service", "key_compromise")
   ```
2. Generate new key
   ```go
   newKey, _ := keyManager.RotateKey("service", "emergency_rotation")
   ```
3. Review audit trail
4. Identify leak source
5. Update key storage

**Prevention:**
- Encrypt keys at rest
- Implement key rotation (90 days)
- Monitor key usage patterns
- Use environment variables only

### Scenario 4: Data Breach

**Detection:**
- Unauthorized data access
- Large data export
- External notification

**Response:**
1. **Immediate (< 15 min):**
   - Isolate affected systems
   - Preserve evidence
   - Stop data exfiltration

2. **Short-term (< 1 hour):**
   - Assess scope of breach
   - Identify compromised data
   - Begin forensic analysis

3. **Communication (< 72 hours for GDPR):**
   - Notify affected users
   - Report to regulators
   - Public disclosure (if required)

**Legal Requirements:**
- GDPR: 72-hour notification
- CCPA: "Without unreasonable delay"
- HIPAA: 60 days
- PCI DSS: Immediate to card brands

### Scenario 5: DDoS Attack

**Detection:**
- Connection limits reached
- Service degradation
- WAF blocking massive traffic

**Response:**
1. Enable DDoS mitigation
   ```go
   wafConfig.MaxConcurrentConns = 5000
   wafConfig.MaxRequestsPerIP = 10
   ```
2. Contact hosting provider
3. Enable CDN DDoS protection
4. Block attacking IP ranges
5. Monitor legitimate user impact

**Prevention:**
- CDN with DDoS protection
- Auto-scaling infrastructure
- Traffic analysis and anomaly detection

## Communication Templates

### User Notification - Data Breach
```
Subject: Important Security Notice

Dear [User],

We are writing to inform you of a security incident that may have affected your account.

What Happened:
On [Date], we discovered unauthorized access to our systems. We immediately took action to secure our platform and investigate the incident.

What Information Was Affected:
[Specific data types - email, encrypted passwords, etc.]

What We're Doing:
- We have secured the vulnerability
- We have rotated all security credentials
- We are conducting a full security audit
- We have notified the appropriate authorities

What You Should Do:
- Change your password immediately
- Enable two-factor authentication
- Monitor your account for suspicious activity
- Be alert for phishing attempts

We take your security very seriously and sincerely apologize for this incident. If you have questions, please contact security@rtx-trading.com.

Sincerely,
RTX Trading Security Team
```

### Internal Alert - Critical Incident
```
SECURITY INCIDENT - P0

Time: [Timestamp]
Severity: Critical
Status: Containment in progress

Summary:
[Brief description of incident]

Affected Systems:
- [System 1]
- [System 2]

Actions Taken:
- [Action 1]
- [Action 2]

Next Steps:
- [Step 1]
- [Step 2]

Incident Commander: [Name]
War Room: [Location/Link]
```

## Incident Log Template

```markdown
## Incident #[NUMBER] - [TITLE]

**Date:** [Date]
**Severity:** [P0/P1/P2/P3]
**Status:** [Active/Contained/Resolved]
**Incident Commander:** [Name]

### Timeline
- HH:MM - Incident detected
- HH:MM - Team notified
- HH:MM - Containment actions started
- HH:MM - Root cause identified
- HH:MM - Fix deployed
- HH:MM - Service restored
- HH:MM - Incident closed

### Summary
[Description of what happened]

### Impact
- Users affected: [Number/Percentage]
- Data compromised: [Yes/No - Details]
- Downtime: [Duration]
- Financial impact: [Estimate]

### Root Cause
[Technical details of vulnerability/attack]

### Resolution
[Steps taken to resolve]

### Lessons Learned
1. [Lesson 1]
2. [Lesson 2]

### Action Items
- [ ] [Action 1] - Owner: [Name] - Due: [Date]
- [ ] [Action 2] - Owner: [Name] - Due: [Date]
```

## Security Incident Metrics

Track and report:
- Mean Time to Detect (MTTD)
- Mean Time to Respond (MTTR)
- Mean Time to Resolve (MTTR)
- Number of incidents by severity
- False positive rate
- User impact (# affected, downtime)

**Target Metrics:**
- MTTD: < 15 minutes
- MTTR (Response): < 30 minutes
- MTTR (Resolve): < 4 hours (P0), < 24 hours (P1)
- False positives: < 10%

## Training & Exercises

### Quarterly Tabletop Exercises
Simulate incidents:
- Database breach
- DDoS attack
- Insider threat
- Ransomware
- API key compromise

### Annual Full-Scale Drill
- Activate incident response team
- Execute full response procedure
- Test communication channels
- Verify backup restoration
- Document lessons learned

---

**Last Updated:** [Date]
**Next Review:** [Date + 6 months]
**Owner:** Security Team Lead
