# 🚀 Production Deployment Guide

This guide provides comprehensive instructions for deploying Mimir Limit Optimizer in production environments.

## 📋 Quick Start Options

### Option 1: Using Values File (Recommended)
```bash
# 1. Customize the values file
cp prod-values.yaml my-prod-values.yaml
# Edit my-prod-values.yaml with your specific configuration

# 2. Deploy using the script
./deploy-production.sh prod your-domain.com

# OR deploy manually
helm install mimir-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir-optimizer \
  --create-namespace \
  --values my-prod-values.yaml \
  --wait
```

### Option 2: Command Line Deployment
```bash
helm install mimir-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir-optimizer \
  --create-namespace \
  --set controller.mode=prod \
  --set image.tag=v3.0.2 \
  --set ui.enabled=true \
  --set ui.ingress.enabled=true \
  --set ui.ingress.hosts[0].host=your-domain.com \
  --set mimir.namespace=mimir \
  --set serviceMonitor.enabled=true \
  --wait
```

## 🔧 Essential Customizations

### 1. Update Domain and TLS
```yaml
ui:
  ingress:
    enabled: true
    hosts:
      - host: mimir-optimizer.your-domain.com  # CHANGE THIS
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.your-domain.com  # CHANGE THIS
```

### 2. Configure Mimir Environment
```yaml
mimir:
  namespace: "mimir"  # Your Mimir namespace
  configMapName: "mimir-runtime-overrides"  # Your ConfigMap name

metricsDiscovery:
  namespace: "mimir"  # Same as above
  serviceLabelSelector: "app.kubernetes.io/part-of=mimir"  # Update for your labels
  serviceNames:  # Update for your Mimir services
    - "distributor"
    - "ingester"
    - "querier"
    - "query-frontend"
    # Add your specific services
```

### 3. Resource Sizing
```yaml
# Small environment
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Large environment
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

### 4. High Availability
```yaml
replicaCount: 3  # For HA

affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app.kubernetes.io/name
          operator: In
          values:
          - mimir-limit-optimizer
      topologyKey: kubernetes.io/hostname
```

## 🌐 Ingress Examples

### NGINX Ingress
```yaml
ui:
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
```

### Traefik Ingress
```yaml
ui:
  ingress:
    enabled: true
    className: "traefik"
    annotations:
      traefik.ingress.kubernetes.io/router.tls: "true"
      traefik.ingress.kubernetes.io/router.tls.certresolver: "letsencrypt"
```

### AWS ALB
```yaml
ui:
  ingress:
    enabled: true
    className: "alb"
    annotations:
      kubernetes.io/ingress.class: alb
      alb.ingress.kubernetes.io/scheme: internet-facing
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/healthcheck-path: /metrics
      alb.ingress.kubernetes.io/healthcheck-port: "8080"
      alb.ingress.kubernetes.io/success-codes: "200"
```

## 📊 Monitoring Integration

### Prometheus ServiceMonitor
```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
  labels:
    team: platform
    environment: production
  metricRelabelings:
    - sourceLabels: [__name__]
      regex: 'go_.*'
      action: drop
```

### Grafana Dashboard
The optimizer exposes metrics at `/metrics`. Key metrics to monitor:
- `mimir_limit_optimizer_reconcile_total`
- `mimir_limit_optimizer_configmap_updates_total`
- `mimir_limit_optimizer_tenant_limits_applied_total`
- `mimir_limit_optimizer_recommendations_total`

## 🛡️ Security Configuration

### Pod Security Standards
```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 65534
```

### Network Policies
```yaml
# Example NetworkPolicy (create separately)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mimir-optimizer-netpol
  namespace: mimir-optimizer
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: mimir-limit-optimizer
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: mimir
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: mimir
    ports:
    - protocol: TCP
      port: 8080
```

## 🔍 Troubleshooting

### Common Issues

1. **Pods not starting**
   ```bash
   kubectl describe pods -n mimir-optimizer
   kubectl logs -n mimir-optimizer -l app.kubernetes.io/name=mimir-limit-optimizer
   ```

2. **Unable to connect to Mimir**
   ```bash
   # Check if services are discoverable
   kubectl get svc -n mimir -l app.kubernetes.io/part-of=mimir
   
   # Test connectivity
   kubectl exec -n mimir-optimizer deployment/mimir-optimizer -- \
     curl -s http://SERVICE_NAME.mimir.svc.cluster.local:8080/metrics
   ```

3. **UI not accessible**
   ```bash
   # Check ingress
   kubectl get ingress -n mimir-optimizer
   kubectl describe ingress -n mimir-optimizer
   
   # Port forward for testing
   kubectl port-forward -n mimir-optimizer svc/mimir-optimizer-ui 8080:8082
   ```

### Debug Commands
```bash
# Check all resources
kubectl get all -n mimir-optimizer

# Check events
kubectl get events -n mimir-optimizer --sort-by='.firstTimestamp'

# Check logs with follow
kubectl logs -f -n mimir-optimizer -l app.kubernetes.io/name=mimir-limit-optimizer

# Check configuration
kubectl get configmap -n mimir-optimizer -o yaml

# Test health endpoints
kubectl exec -n mimir-optimizer deployment/mimir-optimizer -- \
  curl -s http://localhost:8081/healthz
```

## 🔄 Upgrade Process

### Upgrade to New Version
```bash
# Update image tag in values file or use --set
helm upgrade mimir-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir-optimizer \
  --values prod-values.yaml \
  --set image.tag=v3.0.2 \
  --wait

# Or use the deployment script
./deploy-production.sh prod your-domain.com
```

### Rollback
```bash
# List releases
helm history mimir-optimizer -n mimir-optimizer

# Rollback to previous version
helm rollback mimir-optimizer -n mimir-optimizer
```

## 📈 Performance Tuning

### High Volume Environments
```yaml
resources:
  requests:
    cpu: 1000m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 2Gi

performance:
  cache:
    enabled: true
    sizeMB: 512
  batchProcessing:
    enabled: true
    size: 200
    maxConcurrent: 20

auditLog:
  retention:
    cleanupInterval: "15m"
    cleanupBatchSize: 500
```

### Low Resource Environments
```yaml
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi

performance:
  cache:
    enabled: true
    sizeMB: 64
  batchProcessing:
    enabled: true
    size: 50
    maxConcurrent: 5
```

## 🚨 Production Checklist

- [ ] Set `controller.mode: prod`
- [ ] Configure proper resource limits
- [ ] Enable monitoring (`serviceMonitor.enabled: true`)
- [ ] Configure ingress with TLS
- [ ] Set up proper RBAC
- [ ] Configure security contexts
- [ ] Set up audit logging retention
- [ ] Configure tenant scoping
- [ ] Test UI accessibility
- [ ] Verify metrics collection
- [ ] Test ConfigMap updates
- [ ] Set up alerting
- [ ] Document runbooks
- [ ] Test backup/restore procedures

## 📞 Support

For issues and questions:
1. Check the logs: `kubectl logs -n mimir-optimizer -l app.kubernetes.io/name=mimir-limit-optimizer`
2. Review the troubleshooting section above
3. Check GitHub issues: https://github.com/AkshayDubey29/mimir-limit-optimizer/issues
4. Review the comprehensive guides in the repository 