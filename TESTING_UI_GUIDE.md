# Testing UI with Kubernetes Port Forwarding

## 🚀 **Step-by-Step UI Testing Guide**

### **1. Deploy to Kubernetes**

```bash
# Deploy with UI enabled (recommended for testing)
helm upgrade --install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set image.repository=artifactory.coupang.net/ghcr-remote/akshaydubey29/mimir-limit-optimizer \
  --set image.tag=v3.0.0 \
  --set controller.mode=dry-run \
  --set mimir.namespace=mimir \
  --set metricsDiscovery.enabled=false \
  --set ui.enabled=true \
  --set ui.port=8082 \
  --set ui.service.type=ClusterIP \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **2. Verify Deployment**

```bash
# Check if pods are running
kubectl get pods -n mimir-limit-optimizer

# Check services
kubectl get svc -n mimir-limit-optimizer

# Expected output:
# NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
# mimir-limit-optimizer   ClusterIP   10.96.xxx.xxx   <none>        8080/TCP,8082/TCP   1m
```

### **3. Access UI via Port Forward**

#### **Option A: Forward Main Service - All Ports (Recommended)**
```bash
# Forward both metrics (8080) and UI (8082) ports from main service
kubectl port-forward -n mimir-limit-optimizer service/mimir-limit-optimizer 8080:8080 8082:8082

# Access:
# - UI: http://localhost:8082
# - Metrics: http://localhost:8080/metrics
```

#### **Option B: Forward UI Port Only**
```bash
# Forward only the UI port (8082) to local port 8082
kubectl port-forward -n mimir-limit-optimizer service/mimir-limit-optimizer 8082:8082

# Access the UI at: http://localhost:8082
```

#### **Option C: Forward UI Service (Alternative)**
```bash
# Use the dedicated UI service (if you prefer separation)
kubectl port-forward -n mimir-limit-optimizer service/mimir-limit-optimizer-ui 8082:8082

# Access the UI at: http://localhost:8082
```

#### **Option D: Forward Pod Directly**
```bash
# Get pod name
POD_NAME=$(kubectl get pods -n mimir-limit-optimizer -l app.kubernetes.io/name=mimir-limit-optimizer -o jsonpath='{.items[0].metadata.name}')

# Forward from pod (both ports)
kubectl port-forward -n mimir-limit-optimizer pod/$POD_NAME 8080:8080 8082:8082

# Access:
# - UI: http://localhost:8082
# - Metrics: http://localhost:8080/metrics
```

### **4. View Logs for Debugging**

#### **Real-time Logs**
```bash
# Follow logs from the deployment
kubectl logs -n mimir-limit-optimizer deployment/mimir-limit-optimizer -f

# Filter for UI-related logs
kubectl logs -n mimir-limit-optimizer deployment/mimir-limit-optimizer -f | grep -i "ui\|web\|8082"
```

#### **Historical Logs**
```bash
# Get recent logs
kubectl logs -n mimir-limit-optimizer deployment/mimir-limit-optimizer --tail=100

# Get logs from specific pod
POD_NAME=$(kubectl get pods -n mimir-limit-optimizer -l app.kubernetes.io/name=mimir-limit-optimizer -o jsonpath='{.items[0].metadata.name}')
kubectl logs -n mimir-limit-optimizer pod/$POD_NAME --tail=100
```

### **5. Health Checks**

```bash
# Check if UI is responding (after port-forward is active)
curl -s http://localhost:8082/ | head -20

# Check API endpoints
curl -s http://localhost:8082/api/health
curl -s http://localhost:8082/api/tenants

# Check metrics endpoint
curl -s http://localhost:8080/metrics | grep mimir_limit_optimizer
```

### **6. Ingress Verification**

If ingress is enabled, verify it's working correctly:

```bash
# Check ingress status
kubectl get ingress -n mimir-limit-optimizer

# Describe ingress for details
kubectl describe ingress -n mimir-limit-optimizer

# Verify ingress endpoints
kubectl get endpoints -n mimir-limit-optimizer

# Test ingress access (if DNS is configured)
curl -I https://your-domain.com/

# Test with local hosts file (for testing)
echo "INGRESS_IP your-domain.com" >> /etc/hosts  # Replace INGRESS_IP
curl -I https://your-domain.com/
```

### **7. Expected Log Messages**

When UI is **enabled**, you should see:
```json
{"level":"info","ts":"2025-07-01T...","logger":"setup","msg":"Web UI enabled","port":8082}
{"level":"info","ts":"2025-07-01T...","logger":"api","msg":"Starting UI server","port":8082}
```

When UI is **disabled**:
```json
{"level":"info","ts":"2025-07-01T...","logger":"setup","msg":"Web UI disabled"}
```

### **8. Troubleshooting Common Issues**

#### **Issue: Port Forward Connection Refused**
```bash
# Check if service exists
kubectl get svc -n mimir-limit-optimizer

# Check if pods are ready
kubectl get pods -n mimir-limit-optimizer -o wide

# Check pod logs for errors
kubectl logs -n mimir-limit-optimizer deployment/mimir-limit-optimizer --tail=50
```

#### **Issue: UI Not Loading**
```bash
# Verify UI is enabled in config
kubectl get configmap -n mimir-limit-optimizer mimir-limit-optimizer -o yaml | grep -A5 -B5 "ui:"

# Check if UI port is exposed
kubectl describe svc -n mimir-limit-optimizer mimir-limit-optimizer
```

#### **Issue: Permission Errors**
```bash
# Check RBAC permissions
kubectl get clusterrole mimir-limit-optimizer -o yaml
kubectl get clusterrolebinding mimir-limit-optimizer -o yaml
```

### **9. Advanced Testing Scenarios**

#### **Test with Different Configurations**
```bash
# Test with UI disabled
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=false \
  --reuse-values

# Test with different UI port
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.port=9090 \
  --reuse-values
```

#### **Test External Access (NodePort)**
```bash
# Deploy with NodePort for external testing
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.service.type=NodePort \
  --reuse-values

# Get NodePort
kubectl get svc -n mimir-limit-optimizer mimir-limit-optimizer
```

#### **Test with Ingress**
```bash
# Deploy with Ingress enabled (basic)
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.ingress.enabled=true \
  --set ui.ingress.hosts[0].host=mimir-optimizer.local \
  --set ui.ingress.hosts[0].paths[0].path=/ \
  --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
  --reuse-values

# Deploy with NGINX ingress and SSL
helm upgrade mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.ingress.enabled=true \
  --set ui.ingress.className=nginx \
  --set ui.ingress.hosts[0].host=mimir-optimizer.your-domain.com \
  --set ui.ingress.hosts[0].paths[0].path=/ \
  --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
  --set ui.ingress.annotations."cert-manager\.io/cluster-issuer"=letsencrypt-prod \
  --reuse-values

# Test ingress (after DNS is configured)
curl -I https://mimir-optimizer.your-domain.com/
```

### **10. Cleanup**

```bash
# Stop port forwarding (Ctrl+C)
# Then remove deployment
helm uninstall mimir-limit-optimizer -n mimir-limit-optimizer

# Remove namespace if no longer needed
kubectl delete namespace mimir-limit-optimizer
```

## 🎯 **Quick Test Script**

```bash
#!/bin/bash
# quick-ui-test.sh

echo "🚀 Deploying Mimir Limit Optimizer with UI..."
helm upgrade --install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set controller.mode=dry-run \
  --namespace mimir-limit-optimizer \
  --create-namespace

echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mimir-limit-optimizer -n mimir-limit-optimizer --timeout=120s

echo "📊 Checking deployment status..."
kubectl get pods,svc -n mimir-limit-optimizer

echo "🌐 Starting port forward to UI (port 8082)..."
echo "Access UI at: http://localhost:8082"
kubectl port-forward -n mimir-limit-optimizer service/mimir-limit-optimizer-ui 8082:8082
```

Run with: `chmod +x quick-ui-test.sh && ./quick-ui-test.sh` 