# Feature Flags & A/B Testing - Quick Reference

## 1-Minute Quick Start

```go
import "github.com/epic1st/rtx/backend/featureflags"

// Initialize
sdk := featureflags.NewSDK()

// Create flag
flag := featureflags.NewFlag("my-feature", "My Feature").Boolean(true).Build()
sdk.CreateFlag(flag, "admin")

// Check if enabled
enabled := sdk.IsEnabled("my-feature", "user123", nil)
```

## Common Patterns

### Boolean Flag
```go
flag := featureflags.NewFlag("new-ui", "New UI").Boolean(true).Build()
sdk.CreateFlag(flag, "admin")
if sdk.IsEnabled("new-ui", userID, nil) {
    showNewUI()
}
```

### Percentage Rollout (25%)
```go
flag := featureflags.NewFlag("beta", "Beta").Percentage(25).Build()
sdk.CreateFlag(flag, "admin")
```

### Targeted Rollout
```go
rules := featureflags.NewSegment().
    Country("US").
    Tier("premium").
    DeviceType("mobile").
    Build()

flag := featureflags.NewFlag("feature", "Feature").
    Boolean(true).
    Rules(rules...).
    Build()
```

### A/B Test
```go
exp := featureflags.NewExperiment("test", "My Test").
    Control("a", "Version A", 50).
    Variant("b", "Version B", 50, nil).
    Build()

sdk.CreateExperiment(exp, "product")
sdk.StartExperiment("test")

variant := sdk.GetExperimentVariant("test", userID, nil)
if variant.ID == "b" {
    showVersionB()
}

// Track conversion
sdk.TrackConversion("test", userID, "purchase", 99.99)
```

### Gradual Rollout
```go
rollout := &featureflags.Rollout{
    ID: "api-v2",
    Strategy: featureflags.StrategyExponential,
    AutoProgress: true,
    RollbackOnError: true,
}
sdk.CreateRollout(rollout)
sdk.StartRollout("api-v2")
```

### Kill Switch
```go
flag := &featureflags.Flag{
    ID: "payments",
    Type: featureflags.FlagTypeKillSwitch,
    Enabled: true,
}
sdk.CreateFlag(flag, "admin")

// Emergency disable
sdk.ToggleFlag("payments", false, "ops")
```

## Cheat Sheet

| Task | Code |
|------|------|
| Create boolean flag | `NewFlag(id, name).Boolean(true).Build()` |
| Create percentage flag | `NewFlag(id, name).Percentage(25).Build()` |
| Add targeting rules | `.Rules(NewSegment().Country("US").Build()...)` |
| Check if enabled | `sdk.IsEnabled(flagID, userID, ctx)` |
| Toggle flag | `sdk.ToggleFlag(flagID, enabled, user)` |
| Create A/B test | `NewExperiment(id, name).Control(...).Variant(...).Build()` |
| Start experiment | `sdk.StartExperiment(experimentID)` |
| Get variant | `sdk.GetExperimentVariant(expID, userID, ctx)` |
| Track conversion | `sdk.TrackConversion(expID, userID, goal, value)` |
| Create rollout | `sdk.CreateRollout(rollout)` |
| Start rollout | `sdk.StartRollout(rolloutID)` |

## Files Reference

| File | Purpose | Key Types |
|------|---------|-----------|
| `flags.go` | Flag management | `FlagManager`, `Flag`, `FlagEvaluation` |
| `experiments.go` | A/B testing | `ExperimentManager`, `Experiment`, `ExperimentResult` |
| `targeting.go` | User targeting | `EvaluationContext`, `TargetingRule`, `SegmentBuilder` |
| `analytics.go` | Analytics | `AnalyticsManager`, `FunnelAnalysis`, `SampleSizeCalculator` |
| `rollout.go` | Gradual rollouts | `RolloutManager`, `Rollout`, `RolloutMetrics` |
| `sdk.go` | Unified API | `SDK`, `FlagBuilder`, `ExperimentBuilder` |

## Performance Tips

1. **Use context caching**: Pass same context object for same user
2. **Enable caching**: SDK caches for 5s by default
3. **Batch operations**: Update multiple flags at once
4. **Pre-warm cache**: Call flags once on startup
5. **Monitor callbacks**: Use callbacks for analytics (async)

## Common Mistakes to Avoid

1. ❌ Don't go 0→100% instantly (use gradual rollout)
2. ❌ Don't skip statistical significance (check p-value < 0.05)
3. ❌ Don't use different user IDs for same user (breaks consistency)
4. ❌ Don't ignore sample size requirements (use calculator)
5. ❌ Don't forget to clean up old flags (maintenance)
6. ❌ Don't disable auto-rollback (safety feature)
7. ❌ Don't skip documentation (use description field)
8. ❌ Don't run tests too short (wait for significance)
9. ❌ Don't forget to track conversions (essential for A/B tests)
10. ❌ Don't hardcode user attributes (use context)

## Performance Targets

| Operation | Target | Actual |
|-----------|--------|--------|
| Flag evaluation (cold) | <1ms | ~500ns |
| Flag evaluation (warm) | <1ms | ~120ns |
| Experiment assignment | <1ms | ~250ns |
| Conversion tracking | <10ms | ~1μs |

## Quick Examples

### Example 1: Simple Feature Toggle
```go
sdk := featureflags.NewSDK()
flag := featureflags.NewFlag("dark-mode", "Dark Mode").Boolean(true).Build()
sdk.CreateFlag(flag, "admin")
if sdk.IsEnabled("dark-mode", userID, nil) { enableDarkMode() }
```

### Example 2: Regional Rollout
```go
rules := featureflags.NewSegment().Country("US", "CA").Build()
flag := featureflags.NewFlag("feature", "Feature").Boolean(true).Rules(rules...).Build()
sdk.CreateFlag(flag, "admin")
```

### Example 3: VIP Feature
```go
rules := featureflags.NewSegment().VIP().Build()
flag := featureflags.NewFlag("vip-lounge", "VIP Lounge").Boolean(true).Rules(rules...).Build()
sdk.CreateFlag(flag, "admin")
```

### Example 4: Simple A/B Test
```go
exp := featureflags.NewExperiment("ui-test", "UI Test").
    Control("old", "Old UI", 50).
    Variant("new", "New UI", 50, nil).
    Build()
sdk.CreateExperiment(exp, "product")
sdk.StartExperiment("ui-test")
```

## Integration Checklist

- [ ] Import package
- [ ] Initialize SDK
- [ ] Create flags
- [ ] Add targeting rules
- [ ] Test locally
- [ ] Create experiments
- [ ] Set up analytics callbacks
- [ ] Configure rollouts
- [ ] Add monitoring
- [ ] Test in staging
- [ ] Deploy to production
- [ ] Monitor metrics
- [ ] Clean up old flags

## Support

For detailed documentation, see `README.md`
For implementation details, see `IMPLEMENTATION_SUMMARY.md`
For examples, see `examples_test.go`
