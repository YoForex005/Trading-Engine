# Security Deployment Checklist

## Pre-Deployment Security Verification

### 1. Code Security Audit
- [ ] Run vulnerability scanner on entire codebase
  ```bash
  go run scripts/security_scan.go
  ```
- [ ] No CRITICAL vulnerabilities remaining
- [ ] All HIGH vulnerabilities addressed or documented
- [ ] Secret scanning completed (no hardcoded credentials)
- [ ] Code review completed by security team

### 2. Compliance Verification
- [ ] Run compliance checker
  ```bash
  go run scripts/compliance_check.go
  ```
- [ ] OWASP Top 10 compliance: PASS
- [ ] All critical compliance rules passing
- [ ] Financial compliance requirements met (PCI DSS, KYC/AML)
- [ ] Compliance report generated and reviewed

### 3. Security Testing
- [ ] Automated security tests passing (100%)
  ```bash
  go test ./security/... -v
  ```
- [ ] Penetration testing completed
- [ ] SQL injection tests: PASS
- [ ] XSS tests: PASS
- [ ] CSRF protection verified
- [ ] Authentication/authorization tests: PASS
- [ ] Rate limiting verified
- [ ] Session management tested

### 4. Configuration Hardening
- [ ] Master encryption key set (NOT default)
- [ ] JWT secret rotated from development
- [ ] API keys moved to environment variables
- [ ] Database credentials secured
- [ ] CORS configured with specific origins (not wildcard)
- [ ] WAF configured with production limits
- [ ] Session timeout appropriate for production (30 min)
- [ ] Admin IP whitelist configured

### 5. HTTPS/TLS Configuration
- [ ] Valid SSL/TLS certificate installed
- [ ] TLS 1.2+ enforced (TLS 1.0/1.1 disabled)
- [ ] Strong cipher suites configured
- [ ] HSTS header configured (max-age=31536000)
- [ ] HTTP to HTTPS redirect enabled
- [ ] Certificate auto-renewal configured

### 6. Security Headers
Verify all headers are set:
- [ ] Strict-Transport-Security
- [ ] Content-Security-Policy
- [ ] X-Frame-Options: DENY
- [ ] X-Content-Type-Options: nosniff
- [ ] X-XSS-Protection: 1; mode=block
- [ ] Referrer-Policy
- [ ] Permissions-Policy

### 7. Audit Logging
- [ ] Audit logger initialized
- [ ] Log directory permissions: 0700
- [ ] Log rotation configured (100MB)
- [ ] Log retention set (90 days)
- [ ] Remote log shipping configured
- [ ] Log monitoring/alerting enabled
- [ ] Test log writes successful

### 8. Database Security
- [ ] Database uses parameterized queries only
- [ ] Database user has minimum required privileges
- [ ] Database connections encrypted
- [ ] Database backups encrypted
- [ ] Database passwords rotated
- [ ] No default database credentials

### 9. API Key Management
- [ ] All API keys rotated from development
- [ ] API key rotation schedule configured (90 days)
- [ ] API keys encrypted at rest
- [ ] API key usage logged
- [ ] Emergency key rotation procedure documented

### 10. Dependency Security
- [ ] All dependencies updated to latest secure versions
  ```bash
  go list -m -u all
  ```
- [ ] No known CVEs in dependencies
- [ ] Dependency scanning automated (CI/CD)
- [ ] Private dependencies secured

## Deployment Steps

### 1. Infrastructure Security
- [ ] Firewall rules configured
- [ ] Only required ports open (443, optionally 80 for redirect)
- [ ] SSH key-based authentication only
- [ ] fail2ban or equivalent configured
- [ ] OS security updates applied
- [ ] Intrusion detection system (IDS) configured

### 2. Application Deployment
- [ ] Deploy behind reverse proxy (nginx/Apache)
- [ ] Enable WAF at reverse proxy level
- [ ] Configure rate limiting at proxy
- [ ] Set appropriate file permissions (0600 for configs)
- [ ] Environment variables set securely
- [ ] Application user (non-root) created
- [ ] systemd service configured with security options

### 3. Monitoring Setup
- [ ] Application performance monitoring (APM)
- [ ] Security event monitoring
- [ ] Log aggregation (ELK, Splunk, etc.)
- [ ] Uptime monitoring
- [ ] SSL certificate expiry monitoring
- [ ] Disk space monitoring (audit logs)

### 4. Backup & Recovery
- [ ] Automated backups configured
- [ ] Backup encryption enabled
- [ ] Backup retention policy set
- [ ] Disaster recovery plan documented
- [ ] Backup restoration tested
- [ ] Incident response plan ready

## Post-Deployment Verification

### Immediate (Within 1 Hour)
- [ ] Application accessible via HTTPS
- [ ] HTTP redirects to HTTPS working
- [ ] Login functionality working
- [ ] CSRF protection active (test state-changing operation)
- [ ] Rate limiting working (test multiple requests)
- [ ] Audit logs being written
- [ ] Security headers present (use securityheaders.com)

### First 24 Hours
- [ ] Monitor error rates
- [ ] Check audit logs for anomalies
- [ ] Verify session management working
- [ ] Test API key rotation
- [ ] Monitor blocked IPs
- [ ] Check WAF statistics
- [ ] Review security alerts

### First Week
- [ ] Run full security test suite
- [ ] Generate compliance report
- [ ] Review all audit logs
- [ ] Test incident response procedure
- [ ] Verify backup restoration
- [ ] Performance testing under load

## Ongoing Security Maintenance

### Daily
- [ ] Monitor security alerts
- [ ] Review critical audit log events
- [ ] Check WAF blocked IPs

### Weekly
- [ ] Review audit logs
- [ ] Check security test results
- [ ] Monitor API key rotation status
- [ ] Review blocked/suspicious IPs

### Monthly
- [ ] Run vulnerability scanner
- [ ] Update dependencies
- [ ] Review and update firewall rules
- [ ] Test backup restoration
- [ ] Security training for team

### Quarterly
- [ ] Full compliance audit
- [ ] Penetration testing
- [ ] Disaster recovery drill
- [ ] Review and update security policies
- [ ] API key rotation verification

### Annually
- [ ] Third-party security audit
- [ ] Review and update incident response plan
- [ ] Security architecture review
- [ ] Insurance policy review (cyber insurance)

## Emergency Procedures

### Security Breach Detected
1. Activate incident response team
2. Isolate affected systems
3. Rotate all API keys and secrets
4. Invalidate all sessions
5. Review audit logs
6. Notify affected users (if required by law)
7. Document incident
8. Implement fixes
9. Post-mortem analysis

### DDoS Attack
1. Enable additional rate limiting
2. Contact hosting provider
3. Enable DDoS mitigation service
4. Block attacking IPs at firewall
5. Monitor legitimate user impact
6. Document attack patterns

### Vulnerability Disclosed
1. Verify vulnerability
2. Develop and test fix
3. Apply fix in controlled manner
4. Run security tests
5. Monitor for exploitation attempts
6. Document and communicate

## Compliance Requirements

### PCI DSS (if processing card payments)
- [ ] Cardholder data encrypted
- [ ] Access control implemented
- [ ] Audit trails maintained
- [ ] Regular security testing
- [ ] Secure development practices

### GDPR (if handling EU user data)
- [ ] Privacy policy updated
- [ ] Data encryption at rest and in transit
- [ ] Right to erasure implemented
- [ ] Data breach notification procedure
- [ ] Data protection impact assessment

### SOC 2 (if applicable)
- [ ] Security controls documented
- [ ] Access controls implemented
- [ ] Change management process
- [ ] Incident response plan
- [ ] Regular audits scheduled

## Sign-Off

**Deployment approved by:**

Security Team Lead: ___________________ Date: ___________

DevOps Lead: ___________________ Date: ___________

Engineering Manager: ___________________ Date: ___________

Compliance Officer: ___________________ Date: ___________

**Production deployment authorized on: ___________**

---

*This checklist must be completed and signed before production deployment.*
