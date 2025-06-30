package costcontrol

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/analyzer"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/collector"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
)

// CostController manages cost control and budget enforcement
type CostController struct {
	config     *config.Config
	log        logr.Logger
	costCache  map[string]*TenantCostData
	budgetAlerts map[string]time.Time // Last alert time per tenant
}

// TenantCostData tracks cost information for a tenant
type TenantCostData struct {
	Tenant            string
	DailyCost         float64
	MonthlyCost       float64
	YearlyCost        float64
	ProjectedDaily    float64
	ProjectedMonthly  float64
	ProjectedYearly   float64
	Currency          string
	LastUpdated       time.Time
	BudgetUtilization BudgetUtilization
}

// BudgetUtilization tracks budget usage percentages
type BudgetUtilization struct {
	DailyPercent   float64
	MonthlyPercent float64
	YearlyPercent  float64
}

// CostBreakdown provides detailed cost analysis
type CostBreakdown struct {
	IngestionCost    float64
	StorageCost      float64
	QueryCost        float64
	SeriesCost       float64
	TotalCost        float64
	CostPerSample    float64
	CostPerSeries    float64
	CostPerQuery     float64
}

// CostOptimizationSuggestion provides cost reduction recommendations
type CostOptimizationSuggestion struct {
	Tenant          string
	CurrentCost     float64
	OptimizedCost   float64
	PotentialSaving float64
	Recommendations []string
	Impact          string
}

// NewCostController creates a new cost controller
func NewCostController(cfg *config.Config, log logr.Logger) *CostController {
	return &CostController{
		config:       cfg,
		log:          log,
		costCache:    make(map[string]*TenantCostData),
		budgetAlerts: make(map[string]time.Time),
	}
}

// CalculateCosts calculates costs for all tenants
func (cc *CostController) CalculateCosts(ctx context.Context, tenantMetrics map[string]*collector.TenantMetrics) (map[string]*TenantCostData, error) {
	if !cc.config.CostControl.Enabled {
		return nil, nil
	}

	costs := make(map[string]*TenantCostData)

	for tenant, metrics := range tenantMetrics {
		costData, err := cc.calculateTenantCost(tenant, metrics)
		if err != nil {
			cc.log.Error(err, "failed to calculate cost", "tenant", tenant)
			continue
		}

		costs[tenant] = costData
		cc.costCache[tenant] = costData

		// Update metrics
		cc.updateCostMetrics(tenant, costData)
	}

	cc.log.Info("calculated costs", "tenants", len(costs))
	return costs, nil
}

// EnforceBudgets checks and enforces budget limits
func (cc *CostController) EnforceBudgets(ctx context.Context, costs map[string]*TenantCostData, limits map[string]*analyzer.TenantLimits) (map[string]*analyzer.TenantLimits, error) {
	if !cc.config.CostControl.Enabled {
		return limits, nil
	}

	adjustedLimits := make(map[string]*analyzer.TenantLimits)

	for tenant, costData := range costs {
		tenantLimits := limits[tenant]
		if tenantLimits == nil {
			continue
		}

		// Check budget violations
		budget := cc.getTenantBudget(tenant)
		violation := cc.checkBudgetViolation(costData, budget)

		if violation {
			// Apply budget enforcement
			if budget.EnforceBudget && cc.config.CostControl.AutoLimitReduction {
				adjustedLimits[tenant] = cc.reduceLimitsForBudget(tenantLimits, costData, budget)
				cc.log.Info("reduced limits due to budget violation", "tenant", tenant)
			} else {
				adjustedLimits[tenant] = tenantLimits
			}

			// Send budget alert
			cc.sendBudgetAlert(tenant, costData, budget)
		} else {
			adjustedLimits[tenant] = tenantLimits
		}
	}

	return adjustedLimits, nil
}

// GetCostBreakdown provides detailed cost analysis
func (cc *CostController) GetCostBreakdown(tenant string, metrics *collector.TenantMetrics) (*CostBreakdown, error) {
	if !cc.config.CostControl.Enabled {
		return nil, fmt.Errorf("cost control not enabled")
	}

	breakdown := &CostBreakdown{}
	costPerUnit := cc.config.CostControl.CostPerUnit

	// Calculate ingestion cost (based on samples)
	if ingestionData, ok := metrics.Metrics["cortex_distributor_received_samples_total"]; ok {
		totalSamples := cc.sumMetricValues(ingestionData)
		breakdown.IngestionCost = totalSamples * costPerUnit * 0.4 // 40% of cost
		breakdown.CostPerSample = costPerUnit * 0.4
	}

	// Calculate storage cost (based on series)
	if seriesData, ok := metrics.Metrics["cortex_ingester_memory_series"]; ok {
		totalSeries := cc.sumMetricValues(seriesData)
		breakdown.StorageCost = totalSeries * costPerUnit * 0.3 // 30% of cost
		breakdown.CostPerSeries = costPerUnit * 0.3
	}

	// Calculate query cost (based on queries)
	if queryData, ok := metrics.Metrics["cortex_querier_queries_total"]; ok {
		totalQueries := cc.sumMetricValues(queryData)
		breakdown.QueryCost = totalQueries * costPerUnit * 0.3 // 30% of cost
		breakdown.CostPerQuery = costPerUnit * 0.3
	}

	breakdown.TotalCost = breakdown.IngestionCost + breakdown.StorageCost + breakdown.QueryCost

	return breakdown, nil
}

// GenerateOptimizationSuggestions provides cost optimization recommendations
func (cc *CostController) GenerateOptimizationSuggestions(tenant string, costData *TenantCostData, breakdown *CostBreakdown) *CostOptimizationSuggestion {
	suggestions := &CostOptimizationSuggestion{
		Tenant:          tenant,
		CurrentCost:     costData.DailyCost,
		Recommendations: []string{},
	}

	// Analyze cost patterns and generate suggestions
	if breakdown.IngestionCost > breakdown.TotalCost*0.6 {
		suggestions.Recommendations = append(suggestions.Recommendations,
			"High ingestion cost detected. Consider implementing sampling or data retention policies.")
	}

	if breakdown.StorageCost > breakdown.TotalCost*0.5 {
		suggestions.Recommendations = append(suggestions.Recommendations,
			"High storage cost detected. Consider reducing series cardinality or implementing data lifecycle policies.")
	}

	if breakdown.QueryCost > breakdown.TotalCost*0.4 {
		suggestions.Recommendations = append(suggestions.Recommendations,
			"High query cost detected. Consider optimizing query patterns or implementing caching.")
	}

	// Calculate potential savings (conservative estimate)
	suggestions.OptimizedCost = suggestions.CurrentCost * 0.85 // 15% potential savings
	suggestions.PotentialSaving = suggestions.CurrentCost - suggestions.OptimizedCost

	if suggestions.PotentialSaving > suggestions.CurrentCost*0.1 {
		suggestions.Impact = "High"
	} else if suggestions.PotentialSaving > suggestions.CurrentCost*0.05 {
		suggestions.Impact = "Medium"
	} else {
		suggestions.Impact = "Low"
	}

	return suggestions
}

// GetGlobalCostSummary provides organization-wide cost summary
func (cc *CostController) GetGlobalCostSummary() map[string]interface{} {
	if !cc.config.CostControl.Enabled {
		return nil
	}

	totalDaily := 0.0
	totalMonthly := 0.0
	totalYearly := 0.0
	tenantCount := 0

	for _, costData := range cc.costCache {
		totalDaily += costData.DailyCost
		totalMonthly += costData.MonthlyCost
		totalYearly += costData.YearlyCost
		tenantCount++
	}

	budget := cc.config.CostControl.GlobalBudget

	return map[string]interface{}{
		"total_daily_cost":     totalDaily,
		"total_monthly_cost":   totalMonthly,
		"total_yearly_cost":    totalYearly,
		"currency":             budget.Currency,
		"tenant_count":         tenantCount,
		"daily_budget":         budget.Daily,
		"monthly_budget":       budget.Monthly,
		"yearly_budget":        budget.Annual,
		"daily_utilization":    (totalDaily / budget.Daily) * 100,
		"monthly_utilization":  (totalMonthly / budget.Monthly) * 100,
		"yearly_utilization":   (totalYearly / budget.Annual) * 100,
	}
}

// Helper methods

func (cc *CostController) calculateTenantCost(tenant string, metrics *collector.TenantMetrics) (*TenantCostData, error) {
	costData := &TenantCostData{
		Tenant:      tenant,
		Currency:    cc.config.CostControl.GlobalBudget.Currency,
		LastUpdated: time.Now(),
	}

	breakdown, err := cc.GetCostBreakdown(tenant, metrics)
	if err != nil {
		return nil, err
	}

	// Calculate daily cost based on current usage
	dailyCostRate := breakdown.TotalCost

	// Project costs
	costData.DailyCost = dailyCostRate
	costData.MonthlyCost = dailyCostRate * 30
	costData.YearlyCost = dailyCostRate * 365

	// Calculate projections (with growth factor)
	growthFactor := 1.1 // Assume 10% growth
	costData.ProjectedDaily = dailyCostRate * growthFactor
	costData.ProjectedMonthly = costData.ProjectedDaily * 30
	costData.ProjectedYearly = costData.ProjectedDaily * 365

	// Calculate budget utilization
	budget := cc.getTenantBudget(tenant)
	if budget.Daily > 0 {
		costData.BudgetUtilization.DailyPercent = (costData.DailyCost / budget.Daily) * 100
	}
	if budget.Monthly > 0 {
		costData.BudgetUtilization.MonthlyPercent = (costData.MonthlyCost / budget.Monthly) * 100
	}
	if budget.Annual > 0 {
		costData.BudgetUtilization.YearlyPercent = (costData.YearlyCost / budget.Annual) * 100
	}

	return costData, nil
}

func (cc *CostController) getTenantBudget(tenant string) config.BudgetConfig {
	if budget, exists := cc.config.CostControl.TenantBudgets[tenant]; exists {
		return budget
	}
	return cc.config.CostControl.GlobalBudget
}

func (cc *CostController) checkBudgetViolation(costData *TenantCostData, budget config.BudgetConfig) bool {
	return (budget.Daily > 0 && costData.DailyCost > budget.Daily) ||
		(budget.Monthly > 0 && costData.MonthlyCost > budget.Monthly) ||
		(budget.Annual > 0 && costData.YearlyCost > budget.Annual)
}

func (cc *CostController) reduceLimitsForBudget(limits *analyzer.TenantLimits, costData *TenantCostData, budget config.BudgetConfig) *analyzer.TenantLimits {
	// Calculate reduction factor based on budget violation
	reductionFactor := 0.8 // Reduce by 20%

	if budget.Daily > 0 && costData.DailyCost > budget.Daily {
		reductionFactor = math.Min(reductionFactor, budget.Daily/costData.DailyCost)
	}

	adjustedLimits := &analyzer.TenantLimits{
		Tenant:      limits.Tenant,
		Limits:      make(map[string]interface{}),
		LastUpdated: time.Now(),
		Reason:      "budget-enforcement",
		Source:      "cost-control",
	}

	// Apply reduction to all dynamic limits
	for limitName, limitValue := range limits.Limits {
		adjustedValue := limitValue
		
		switch v := limitValue.(type) {
		case float64:
			adjustedValue = v * reductionFactor
		case int64:
			adjustedValue = int64(float64(v) * reductionFactor)
		}
		
		adjustedLimits.Limits[limitName] = adjustedValue
	}

	return adjustedLimits
}

func (cc *CostController) sendBudgetAlert(tenant string, costData *TenantCostData, budget config.BudgetConfig) {
	// Check if we already sent an alert recently (avoid spam)
	if lastAlert, exists := cc.budgetAlerts[tenant]; exists {
		if time.Since(lastAlert) < time.Hour {
			return
		}
	}

	cc.budgetAlerts[tenant] = time.Now()

	// Log budget violation
	cc.log.Info("budget violation detected",
		"tenant", tenant,
		"daily_cost", costData.DailyCost,
		"daily_budget", budget.Daily,
		"utilization", costData.BudgetUtilization.DailyPercent)

	// TODO: Send alerts to configured channels (Slack, email, etc.)
}

func (cc *CostController) sumMetricValues(data []collector.MetricData) float64 {
	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	return total
}

func (cc *CostController) updateCostMetrics(tenant string, costData *TenantCostData) {
	// Update Prometheus metrics with cost data
	// TODO: Implement metrics updates when metrics package is enhanced
}

// PredictFutureCosts predicts future costs based on trends
func (cc *CostController) PredictFutureCosts(tenant string, days int) (*TenantCostData, error) {
	current, exists := cc.costCache[tenant]
	if !exists {
		return nil, fmt.Errorf("no cost data available for tenant %s", tenant)
	}

	// Simple linear prediction (could be enhanced with ML)
	growthRate := 0.02 // 2% daily growth
	multiplier := math.Pow(1+growthRate, float64(days))

	predicted := &TenantCostData{
		Tenant:          tenant,
		DailyCost:       current.DailyCost * multiplier,
		MonthlyCost:     current.MonthlyCost * multiplier,
		YearlyCost:      current.YearlyCost * multiplier,
		Currency:        current.Currency,
		LastUpdated:     time.Now(),
	}

	return predicted, nil
}

func (cc *CostController) applyBudgetBasedLimits(tenantLimits map[string]*analyzer.TenantLimits, costs map[string]float64, budgets map[string]float64) map[string]*analyzer.TenantLimits {
	adjustedLimits := make(map[string]*analyzer.TenantLimits)

	for tenant, limits := range tenantLimits {
		budget, hasBudget := budgets[tenant]
		cost, hasCost := costs[tenant]

		if !hasBudget || !hasCost {
			// No budget or cost data, use original limits
			adjustedLimits[tenant] = limits
			continue
		}

		// Calculate budget utilization
		utilization := cost / budget
		
		// Copy the original limits
		adjustedLimit := &analyzer.TenantLimits{
			Tenant:      limits.Tenant,
			Limits:      make(map[string]interface{}),
			LastUpdated: time.Now(),
			Reason:      limits.Reason,
			Source:      "cost-control",
		}

		// Copy and potentially adjust all dynamic limits
		for limitName, limitValue := range limits.Limits {
			adjustedValue := limitValue
			
			// Apply budget-based adjustments based on limit type
			if cc.config.CostControl.AutoLimitReduction && utilization > 0.95 {
				// Reduce limits if budget utilization is over 95%
				reductionFactor := 0.8 // Reduce by 20%
				
				switch v := limitValue.(type) {
				case float64:
					adjustedValue = v * reductionFactor
				case int64:
					adjustedValue = int64(float64(v) * reductionFactor)
				}
				
				adjustedLimit.Reason = fmt.Sprintf("budget-exceeded(%.1f%%)", utilization*100)
			} else if cc.config.CostControl.AutoLimitReduction && utilization > 0.75 {
				// Slightly reduce limits if budget utilization is over 75%
				reductionFactor := 0.9 // Reduce by 10%
				
				switch v := limitValue.(type) {
				case float64:
					adjustedValue = v * reductionFactor
				case int64:
					adjustedValue = int64(float64(v) * reductionFactor)
				}
				
				adjustedLimit.Reason = fmt.Sprintf("budget-warning(%.1f%%)", utilization*100)
			}
			
			adjustedLimit.Limits[limitName] = adjustedValue
		}

		adjustedLimits[tenant] = adjustedLimit
	}

	return adjustedLimits
} 