# Production Deployment Checklist

## Pre-Deployment

### Infrastructure
- [ ] Redis Cluster set up (3+ masters, 3+ replicas)
- [ ] Load balancer configured (HAProxy/Nginx/AWS ALB)
- [ ] SSL/TLS certificates installed
- [ ] DNS configured
- [ ] Firewall rules configured
- [ ] File descriptor limits increased (`ulimit -n 200000`)
- [ ] TCP tuning applied (buffer sizes, connection tracking)

### Configuration
- [ ] `MaxConnections` set per node (default: 10,000)
- [ ] `HeartbeatInterval` configured (default: 5s)
- [ ] `NodeTimeout` configured (default: 15s)
- [ ] `ScaleUpThreshold` set (default: 0.8)
- [ ] `ScaleDownThreshold` set (default: 0.3)
- [ ] `MinNodes` and `MaxNodes` defined
- [ ] Load balancing algorithm selected
- [ ] Failover strategy chosen
- [ ] Batch configuration set
- [ ] Compression enabled (if needed)

### Security
- [ ] Authentication implemented (JWT/OAuth2)
- [ ] Rate limiting configured per user/IP
- [ ] CORS policies set
- [ ] Input validation enabled
- [ ] DDoS protection configured
- [ ] Redis AUTH enabled
- [ ] TLS for Redis connections
- [ ] Firewall rules tested
- [ ] Security audit completed

### Monitoring
- [ ] Prometheus configured and running
- [ ] Grafana dashboards created
- [ ] Alert rules defined
- [ ] Error tracking setup (Sentry/etc)
- [ ] Logging aggregation configured
- [ ] Health check endpoints tested
- [ ] Metrics endpoints verified
- [ ] PagerDuty/alerting integrated

## Deployment

### Initial Deployment
- [ ] Deploy Redis cluster first
- [ ] Verify Redis connectivity
- [ ] Deploy first WebSocket node
- [ ] Test single node operation
- [ ] Deploy additional nodes (minimum 3)
- [ ] Verify cluster formation
- [ ] Configure load balancer
- [ ] Test load balancer routing

### Testing
- [ ] Run health checks
- [ ] Test WebSocket connections
- [ ] Verify sticky sessions work
- [ ] Test message broadcasting
- [ ] Simulate node failure (failover test)
- [ ] Run load test (1k connections)
- [ ] Run load test (10k connections)
- [ ] Run load test (100k connections if possible)
- [ ] Verify metrics collection
- [ ] Test auto-scaling triggers

### Performance Validation
- [ ] Connection latency < 50ms (P99)
- [ ] Message latency < 10ms (P99)
- [ ] Failover time < 5s
- [ ] CPU usage < 80% under load
- [ ] Memory usage < 80% under load
- [ ] No memory leaks detected
- [ ] Redis commands < 10ms
- [ ] Load balanced evenly across nodes

## Post-Deployment

### Monitoring & Alerts
- [ ] Set up 24/7 monitoring
- [ ] Configure critical alerts:
  - [ ] Node down
  - [ ] High error rate (>5%)
  - [ ] High latency (P99 > 100ms)
  - [ ] High CPU (>90%)
  - [ ] High memory (>90%)
  - [ ] Redis unavailable
  - [ ] Disk space low
- [ ] Test alert delivery
- [ ] Define on-call rotation

### Documentation
- [ ] Document deployment architecture
- [ ] Create runbook for common issues
- [ ] Document scaling procedures
- [ ] Create incident response plan
- [ ] Document rollback procedure
- [ ] Update API documentation

### Operational
- [ ] Set up automated backups (Redis)
- [ ] Configure log rotation
- [ ] Set up automated health checks
- [ ] Configure auto-restart on crash
- [ ] Set up blue/green deployment
- [ ] Test disaster recovery plan

## Load Testing Checklist

### Test Scenarios
- [ ] **Test 1: Connection Stress**
  - [ ] 50,000 connections at 2000/sec
  - [ ] Success rate > 99%
  - [ ] Connection latency P99 < 50ms

- [ ] **Test 2: Message Throughput**
  - [ ] 10,000 connections
  - [ ] 1000 messages per connection
  - [ ] Message latency P99 < 10ms
  - [ ] No message loss

- [ ] **Test 3: Sustained Load**
  - [ ] 20,000 connections
  - [ ] 10 minute duration
  - [ ] Stable memory usage
  - [ ] No connection drops

- [ ] **Test 4: Failover**
  - [ ] Kill random node
  - [ ] Verify reconnection < 5s
  - [ ] No data loss
  - [ ] Load redistributed

- [ ] **Test 5: Auto-Scaling**
  - [ ] Trigger scale-up (>80% load)
  - [ ] Verify new node joins
  - [ ] Load redistributed
  - [ ] Trigger scale-down (<30% load)

## Scaling Checklist

### Vertical Scaling (Per Node)
- [ ] Increase `MaxConnections`
- [ ] Add more CPU cores
- [ ] Add more RAM
- [ ] Increase file descriptors
- [ ] Tune TCP parameters
- [ ] Enable compression
- [ ] Optimize batch settings

### Horizontal Scaling (More Nodes)
- [ ] Deploy new node
- [ ] Verify node joins cluster
- [ ] Verify load balancer includes node
- [ ] Monitor load distribution
- [ ] Verify failover works with new topology

### Auto-Scaling
- [ ] Configure scaling policies
- [ ] Set up infrastructure automation (Terraform/CloudFormation)
- [ ] Configure container orchestration (K8s/ECS)
- [ ] Test scale-up triggers
- [ ] Test scale-down triggers
- [ ] Verify graceful node shutdown

## Troubleshooting Checklist

### Connection Issues
- [ ] Check load balancer health
- [ ] Verify DNS resolution
- [ ] Check SSL certificate validity
- [ ] Verify firewall rules
- [ ] Check Redis connectivity
- [ ] Review WebSocket upgrade logs

### Performance Issues
- [ ] Check CPU usage per node
- [ ] Check memory usage per node
- [ ] Review Redis latency
- [ ] Check network bandwidth
- [ ] Review goroutine count
- [ ] Check for memory leaks
- [ ] Review message batch sizes

### Failover Issues
- [ ] Verify heartbeat frequency
- [ ] Check node timeout settings
- [ ] Review Redis session data
- [ ] Verify failover strategy
- [ ] Check session migration logs
- [ ] Review client reconnection logic

### Load Balancing Issues
- [ ] Verify algorithm selection
- [ ] Check node health scores
- [ ] Review sticky session config
- [ ] Check session affinity cache
- [ ] Review load distribution metrics
- [ ] Verify weighted node settings

## Rollback Plan

### Prepare for Rollback
- [ ] Keep previous version available
- [ ] Document rollback steps
- [ ] Test rollback procedure
- [ ] Have database backups ready
- [ ] Communicate rollback plan to team

### Rollback Triggers
- [ ] Error rate > 10%
- [ ] Connection success rate < 90%
- [ ] Critical bug discovered
- [ ] Performance degradation > 50%
- [ ] Data corruption detected

### Rollback Steps
1. [ ] Stop accepting new connections
2. [ ] Drain existing connections (graceful shutdown)
3. [ ] Deploy previous version
4. [ ] Verify health checks pass
5. [ ] Gradually increase traffic
6. [ ] Monitor for 30 minutes
7. [ ] Document issues for next deployment

## Maintenance Checklist

### Daily
- [ ] Review error logs
- [ ] Check alert status
- [ ] Review connection metrics
- [ ] Check disk space

### Weekly
- [ ] Review performance trends
- [ ] Check for memory leaks
- [ ] Review security logs
- [ ] Update dependencies (security patches)

### Monthly
- [ ] Review and update documentation
- [ ] Conduct disaster recovery drill
- [ ] Review and optimize costs
- [ ] Update monitoring dashboards
- [ ] Review capacity planning

### Quarterly
- [ ] Conduct security audit
- [ ] Review architecture
- [ ] Load test at scale
- [ ] Review and update runbooks
- [ ] Team training on new features

## Emergency Contacts

- [ ] On-call engineer: __________________
- [ ] DevOps lead: __________________
- [ ] Database admin: __________________
- [ ] Security team: __________________
- [ ] Infrastructure provider support: __________________

## Go-Live Checklist

### Final Checks
- [ ] All tests passing
- [ ] Code review completed
- [ ] Security review completed
- [ ] Performance benchmarks met
- [ ] Documentation updated
- [ ] Monitoring configured
- [ ] Alerts tested
- [ ] Team trained
- [ ] Rollback plan ready
- [ ] On-call schedule set

### Communication
- [ ] Notify stakeholders of deployment
- [ ] Update status page
- [ ] Prepare incident communication template
- [ ] Schedule post-deployment review

### Launch
- [ ] Deploy to staging
- [ ] Run full test suite in staging
- [ ] Deploy to production (off-peak hours)
- [ ] Monitor for 2 hours post-deployment
- [ ] Gradually increase traffic
- [ ] Document any issues
- [ ] Send success notification

---

## Success Criteria

✅ **Uptime**: 99.9%+
✅ **Connection Success Rate**: 99.5%+
✅ **Message Latency (P99)**: < 10ms
✅ **Failover Time**: < 5s
✅ **Error Rate**: < 1%
✅ **Concurrent Connections**: 100,000+

---

**Last Updated**: [Date]
**Deployment Lead**: [Name]
**Next Review**: [Date]
