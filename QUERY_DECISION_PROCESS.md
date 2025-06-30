# 📊 Query & Decision Process: How Mimir Limit Optimizer Analyzes Metrics

## 🔍 **Core Query Mechanism**

### **1. Data Collection Strategy**

The Mimir Limit Optimizer uses a **hybrid approach** for collecting and analyzing metrics:

```yaml
# Configuration (config.yaml)
trendAnalysis:
  analysisWindow: 48h          # ← YOUR 48 HOURS OF HISTORICAL DATA
  percentile: 95.0             # Use 95th percentile for limit calculation
  useMovingAverage: true       # Include moving averages in analysis
  includePeaks: true          # Consider peak usage patterns
  
updateInterval: 5m             # How often we collect new data points
```

### **2. Historical Data Queries**

**NEW: PromQL Query Capability**
```go
// Example: Get 48 hours of ingestion rate data for tenant "webapp"
query := `rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])`
startTime := time.Now().Add(-48 * time.Hour)
endTime := time.Now()
step := 30 * time.Minute  // 96 data points over 48 hours

data := collector.QueryHistoricalData(ctx, query, startTime, endTime, step)
```

**Actual PromQL Queries Used:**
```promql
# Ingestion Rate Analysis
rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])

# Series Growth Analysis  
cortex_ingester_memory_series{tenant="webapp"}

# Query Performance Analysis
histogram_quantile(0.95, cortex_query_frontend_query_duration_seconds_bucket{tenant="webapp"})

# Peak Detection Query
max_over_time(cortex_distributor_received_samples_total{tenant="webapp"}[48h])

# Spike Detection Query (Compare current vs baseline)
(
  rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m]) /
  rate(cortex_distributor_received_samples_total{tenant="webapp"}[48h:24h] offset 24h)
) > 2.0
```

## 📈 **Decision-Making Algorithm**

### **3. Step-by-Step Analysis Process**

**Step 1: Collect 48 Hours of Data**
```
Timeline: [Now-48h] ──────────── [Now-24h] ──────────── [Now]
            │                        │                    │
        Baseline Period          Recent Period         Current
        (for spike detection)    (for trends)          (for decisions)
```

**Step 2: Calculate Key Statistics**
```go
// From analyzer.go - Real implementation
func analyzeMetric(tenant, metricName string, data []MetricData) *AnalysisResult {
    // Get 48 hours of historical data
    historical := getHistoricalData(tenant, metricName)
    
    // Filter to analysis window
    cutoff := time.Now().Add(-48 * time.Hour)
    var windowData []MetricData
    for _, d := range historical {
        if d.Timestamp.After(cutoff) {
            windowData = append(windowData, d)
        }
    }
    
    values := extractValues(windowData)
    
    return &AnalysisResult{
        CurrentValue:    values[len(values)-1],           // Latest value
        Percentile:     calculatePercentile(values, 95.0), // 95th percentile
        MovingAverage:  calculateMovingAverage(values),    // Smooth trend
        Peak:          calculatePeak(values),              // Maximum observed
        Trend:         calculateTrend(values),             // Growth rate
        SpikeDetected: detectSpike(values),               // Event detection
    }
}
```

**Step 3: Apply Decision Logic**
```go
func calculateRecommendedLimit(result *AnalysisResult) float64 {
    // START: Use 95th percentile as base
    base := result.Percentile
    
    // ENHANCE: Use higher of percentile or moving average
    if result.MovingAverage > base {
        base = result.MovingAverage
    }
    
    // PEAKS: Consider 80% of peak value (safety margin)
    if result.Peak > 0 {
        peakBased := result.Peak * 0.8
        if peakBased > base {
            base = peakBased
        }
    }
    
    // SPIKES: Apply event-driven multiplier (2x to 5x)
    if result.SpikeDetected {
        base *= result.SpikeMultiplier  // Dynamic based on spike severity
    }
    
    // BUFFER: Add 20% safety buffer
    finalLimit := base * 1.20
    
    return finalLimit
}
```

## 🎯 **Real-World Example**

### **4. Complete Analysis Flow for `ingestion_rate`**

**Input Data (Last 48 Hours):**
```
Tenant: webapp
Metric: cortex_distributor_received_samples_total{tenant="webapp"}
Query: rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])

Raw Data Points (samples/sec):
Time Range: 2024-06-28 15:00 → 2024-06-30 15:00

Business Hours (9AM-6PM):  40K-80K samples/sec
Peak Hours (6PM-9PM):      100K-150K samples/sec  
Night Hours (9PM-9AM):     15K-35K samples/sec
Weekend (Sat-Sun):         25K-60K samples/sec
```

**Statistical Analysis:**
```go
// Calculated from 48h of data (96 data points)
statistics := AnalysisResult{
    CurrentValue:    75000,     // Current rate: 75K samples/sec
    Percentile:     145000,     // 95th percentile: 145K samples/sec
    MovingAverage:  82000,      // 12-hour moving average: 82K
    Peak:          148000,      // Maximum observed: 148K
    Trend:         0.15,        // Growing at 15% over 48h
    SpikeDetected: false,       // No unusual spikes detected
}
```

**Decision Calculation:**
```go
// Step 1: Choose base value
base := max(145000, 82000) = 145000  // Use percentile (higher)

// Step 2: Consider peaks  
peakBased := 148000 * 0.8 = 118400
base := max(145000, 118400) = 145000  // Percentile still higher

// Step 3: Apply spike multiplier
spikeMultiplier := 1.0  // No spike detected
base := 145000 * 1.0 = 145000

// Step 4: Add safety buffer
recommendedLimit := 145000 * 1.20 = 174000

// FINAL DECISION: Set ingestion_rate = 174,000 samples/sec
```

**Configuration Applied:**
```yaml
# Applied to Mimir runtime overrides ConfigMap
overrides:
  webapp:
    ingestion_rate: 174000     # ← CALCULATED FROM 48H ANALYSIS
    ingestion_burst_size: 348000  # 2x ingestion_rate
```

### **5. Spike Detection Example**

**Scenario: Sudden Traffic Spike**
```
Normal Baseline (24-48h ago): 50K samples/sec average
Current Rate (last 5min):     120K samples/sec

Spike Detection:
ratio = 120000 / 50000 = 2.4
threshold = 2.0
spike_detected = true (2.4 > 2.0)

Response:
spikeMultiplier = min(2.4, 5.0) = 2.4
tempLimit = baseLimit * 2.4
cooldownPeriod = 30 minutes
```

## 🚀 **Advanced Query Patterns**

### **6. Sophisticated PromQL Queries Used**

**Rate Calculations:**
```promql
# 5-minute rate for ingestion
rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])

# Peak rate over 48 hours
max_over_time(rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])[48h:5m])
```

**Percentile Analysis:**
```promql
# 95th percentile query duration
histogram_quantile(0.95, 
  rate(cortex_query_frontend_query_duration_seconds_bucket{tenant="webapp"}[5m])
)
```

**Growth Trend Detection:**
```promql
# Series growth rate over 48 hours
(
  cortex_ingester_memory_series{tenant="webapp"} - 
  cortex_ingester_memory_series{tenant="webapp"} offset 48h
) / cortex_ingester_memory_series{tenant="webapp"} offset 48h
```

**Anomaly Detection:**
```promql
# Detect when current rate exceeds 2x the 48h average
rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m]) >
2 * avg_over_time(rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])[48h:5m])
```

## ⚙️ **Configuration Options**

### **7. Tunable Parameters**

```yaml
trendAnalysis:
  analysisWindow: 48h           # Historical window size
  percentile: 95.0              # Percentile for limit calculation
  useMovingAverage: true        # Include moving averages
  includePeaks: true           # Consider peak usage
  timeOfDayBuffers:            # Time-specific adjustments
    "business_hours": 1.5       # 50% higher during business hours
    "weekend": 0.8              # 20% lower on weekends

eventSpike:
  enabled: true                 # Enable spike detection
  threshold: 2.0                # 2x baseline triggers spike
  detectionWindow: 5m           # Detection window
  cooldownPeriod: 30m          # How long to maintain higher limits
  maxSpikeMultiplier: 5.0      # Maximum spike adjustment

limits:
  bufferPercentage: 20.0        # Default safety buffer
  minLimits:                   # Minimum allowed values
    ingestion_rate: 1000
  maxLimits:                   # Maximum allowed values  
    ingestion_rate: 1000000
```

## 📋 **Query Execution Summary**

### **8. What Actually Happens Every 5 Minutes**

1. **Execute PromQL Queries:**
   ```bash
   GET /api/v1/query_range?query=rate(cortex_distributor_received_samples_total{tenant="webapp"}[5m])&start=1719742800&end=1719916800&step=1800
   ```

2. **Process Historical Data:**
   - Extract 96 data points over 48 hours
   - Calculate percentiles, averages, peaks
   - Detect spikes and anomalies

3. **Make Limit Decisions:**
   - Apply decision algorithm
   - Consider tenant tiers and overrides
   - Apply safety buffers

4. **Update ConfigMap:**
   ```yaml
   overrides:
     webapp:
       ingestion_rate: 174000        # ← CALCULATED FROM ANALYSIS
       max_global_series_per_user: 2500000
       query_timeout: 45s
   ```

5. **Mimir Applies Changes:**
   - Runtime overrides loaded automatically
   - No component restarts required
   - New limits active within seconds

This is **exactly** how we analyze your "last 48 hours" of metrics and make intelligent limit decisions! 🎯 