# 🏥 Health Check Configuration Guide

This guide covers all aspects of configuring health checks for Mimir Limit Optimizer.

## 📍 **Port Configuration Overview**

### Default Port Layout
```
8080 - Metrics endpoint (/metrics)
8081 - Health checks (/healthz, /readyz)  [CONFIGURABLE]
8082 - Web UI (/ and /api/*)              [CONFIGURABLE]
```

### Health Check Endpoints
```
GET /healthz  - Liveness probe  (port 8081)
GET /readyz   - Readiness probe (port 8081)
GET /metrics  - Metrics export  (port 8080)
```

## 🔧 **Configuration Options**

### 1. Command Line Configuration
```bash
# Default configuration
./mimir-limit-optimizer \
  --health-probe-bind-address=:8081 \
  --metrics-bind-address=:8080

# Custom health probe port
./mimir-limit-optimizer \
  --health-probe-bind-address=:9000 \
  --metrics-bind-address=:8080

# Bind to all interfaces (Kubernetes style)
./mimir-limit-optimizer \
  --health-probe-bind-address=0.0.0.0:8081 \
  --metrics-bind-address=0.0.0.0:8080
```

### 2. Helm Chart Configuration

#### Basic Configuration
```yaml
# values.yaml
healthProbes:
  port: 8081  # Health check port (NEW - now configurable)
  
  liveness:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 30
    timeoutSeconds: 5
    failureThreshold: 3
    
  readiness:
    enabled: true
    initialDelaySeconds: 5
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3

metrics:
  port: 8080  # Metrics port

ui:
  port: 8082  # UI port
```

#### Deployment Command
```bash
helm install mimir-optimizer ./helm/mimir-limit-optimizer \
  --set healthProbes.port=8081 \
  --set metrics.port=8080 \
  --set ui.port=8082
```

## 🚀 **Production Configuration Examples**

### Standard Production Setup
```yaml
healthProbes:
  port: 8081
  liveness:
    enabled: true
    initialDelaySeconds: 60    # Longer delay for production
    periodSeconds: 30
    timeoutSeconds: 10         # Longer timeout
    failureThreshold: 3
  readiness:
    enabled: true
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
```

### High-Volume Environment
```yaml
healthProbes:
  port: 8081
  liveness:
    enabled: true
    initialDelaySeconds: 120   # Even longer delay
    periodSeconds: 60          # Less frequent checks
    timeoutSeconds: 15
    failureThreshold: 5        # More tolerance
  readiness:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 5
```

### Security-Constrained Environment
```yaml
healthProbes:
  port: 9999  # Non-standard port
  liveness:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 30
    timeoutSeconds: 5
    failureThreshold: 3
  readiness:
    enabled: true
    initialDelaySeconds: 5
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
```

## 🔍 **Common Issues and Solutions**

### Issue 1: Port Already in Use
```
Error: failed to start UI server: listen tcp :8082: bind: address already in use
```

**Solutions:**
```bash
# 1. Check what's using the port
lsof -i :8082
netstat -tulpn | grep 8082

# 2. Kill the process
kill -9 <PID>

# 3. Use different ports
helm install mimir-optimizer ./helm/mimir-limit-optimizer \
  --set ui.port=9082 \
  --set healthProbes.port=9081 \
  --set metrics.port=9080
```

### Issue 2: Health Checks Failing
```
Liveness probe failed: Get "http://10.244.0.5:8081/healthz": dial tcp 10.244.0.5:8081: connect: connection refused
```

**Solutions:**
```bash
# 1. Check if health server is running
kubectl exec -n mimir-optimizer deployment/mimir-optimizer -- \
  curl -f http://localhost:8081/healthz

# 2. Check logs for binding issues
kubectl logs -n mimir-optimizer deployment/mimir-optimizer

# 3. Verify port configuration
kubectl describe pod -n mimir-optimizer

# 4. Test with port forward
kubectl port-forward -n mimir-optimizer deployment/mimir-optimizer 8081:8081
curl http://localhost:8081/healthz
```

### Issue 3: Wrong Health Check URL in Load Balancers
```
ALB health check failing with 404 on /
```

**Solution for AWS ALB:**
```yaml
ui:
  ingress:
    enabled: true
    className: "alb"
    annotations:
      alb.ingress.kubernetes.io/healthcheck-path: /healthz
      alb.ingress.kubernetes.io/healthcheck-port: "8081"
      alb.ingress.kubernetes.io/success-codes: "200"
      # OR use metrics endpoint
      # alb.ingress.kubernetes.io/healthcheck-path: /metrics
      # alb.ingress.kubernetes.io/healthcheck-port: "8080"
      # alb.ingress.kubernetes.io/success-codes: "200"
```

**Solution for NGINX Ingress:**
```yaml
ui:
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
      # Health checks handled by Kubernetes, not ingress
```

## 🧪 **Testing Health Checks**

### Local Testing
```bash
# Test health endpoints locally
curl -f http://localhost:8081/healthz
curl -f http://localhost:8081/readyz
curl -s http://localhost:8080/metrics | head -10

# Test with the application
./mimir-limit-optimizer --health-probe-bind-address=:8081 &
sleep 5
curl -f http://localhost:8081/healthz && echo "✅ Health check OK"
```

### Kubernetes Testing
```bash
# Test from within cluster
kubectl exec -n mimir-optimizer deployment/mimir-optimizer -- \
  curl -f http://localhost:8081/healthz

# Test via service
kubectl exec -n mimir-optimizer deployment/mimir-optimizer -- \
  curl -f http://mimir-optimizer:8081/healthz

# Port forward and test
kubectl port-forward -n mimir-optimizer svc/mimir-optimizer 8081:8081 &
curl -f http://localhost:8081/healthz
```

### Monitoring Health Checks
```bash
# Watch health check status
kubectl get pods -n mimir-optimizer -w

# Check health check events
kubectl describe pod -n mimir-optimizer | grep -A 10 -B 10 "health"

# View health check logs
kubectl logs -n mimir-optimizer deployment/mimir-optimizer | grep health
```

## 🔧 **Advanced Configuration**

### Different Ports for Different Environments
```yaml
# Development
healthProbes:
  port: 8081

# Staging  
healthProbes:
  port: 18081

# Production
healthProbes:
  port: 28081
```

### Custom Health Check Script
```bash
#!/bin/bash
# health-check.sh

HEALTH_URL="http://localhost:8081/healthz"
METRICS_URL="http://localhost:8080/metrics"

# Check health endpoint
if curl -f -s "$HEALTH_URL" > /dev/null; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

# Check if metrics are being served
if curl -f -s "$METRICS_URL" | grep -q "mimir_limit_optimizer"; then
    echo "✅ Metrics check passed"
else
    echo "❌ Metrics check failed"
    exit 1
fi

echo "🎉 All health checks passed"
```

## 📊 **Health Check Monitoring**

### Prometheus Queries
```promql
# Health check success rate
rate(prometheus_http_requests_total{handler="/healthz",code="200"}[5m])

# Health check response time
histogram_quantile(0.95, rate(prometheus_http_request_duration_seconds_bucket{handler="/healthz"}[5m]))

# Failed health checks
increase(prometheus_http_requests_total{handler="/healthz",code!="200"}[5m])
```

### Grafana Dashboard Panels
- Health check success rate over time
- Health check response time percentiles
- Pod restart count due to health check failures
- Health check availability by deployment

## 🚨 **Troubleshooting Checklist**

- [ ] Verify health probe port is not conflicting with other services
- [ ] Check that health endpoints return 200 status
- [ ] Ensure security context allows binding to the port
- [ ] Verify firewall/network policies allow health check traffic
- [ ] Check resource limits aren't causing health check timeouts
- [ ] Validate health check configuration in Helm values
- [ ] Test health checks manually with curl/wget
- [ ] Monitor health check metrics in Prometheus
- [ ] Review pod events for health check failures
- [ ] Check application logs for health check related errors

## 📞 **Quick Reference**

### Default Ports
| Service | Port | Endpoint | Purpose |
|---------|------|----------|---------|
| Health | 8081 | /healthz, /readyz | Kubernetes health checks |
| Metrics | 8080 | /metrics | Prometheus scraping |
| UI | 8082 | /, /api/* | Web dashboard |

### Test Commands
```bash
# Health check
curl -f http://localhost:8081/healthz

# Readiness check  
curl -f http://localhost:8081/readyz

# Metrics check
curl -s http://localhost:8080/metrics | grep mimir_limit_optimizer

# UI check
curl -I http://localhost:8082/
``` 