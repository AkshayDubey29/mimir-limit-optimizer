# Ingress Configuration Guide for Mimir Limit Optimizer

## 🌐 **Overview**

The Mimir Limit Optimizer supports external access through Kubernetes Ingress. You can configure ingress to use either:
- **UI Service** (dedicated service for UI only)
- **Main Service** (includes both metrics and UI ports)

## 🚀 **Quick Setup Examples**

### **1. Basic Ingress with Default Settings** ⭐

```bash
# Deploy with ingress enabled using default UI service
helm upgrade --install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set ui.ingress.enabled=true \
  --set ui.ingress.hosts[0].host=mimir-optimizer.your-domain.com \
  --set ui.ingress.hosts[0].paths[0].path=/ \
  --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **2. Ingress with NGINX Ingress Controller**

```bash
# Deploy with NGINX ingress controller and annotations
helm upgrade --install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set ui.ingress.enabled=true \
  --set ui.ingress.className=nginx \
  --set ui.ingress.hosts[0].host=mimir-optimizer.your-domain.com \
  --set ui.ingress.hosts[0].paths[0].path=/ \
  --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
  --set ui.ingress.annotations."nginx\.ingress\.kubernetes\.io/ssl-redirect"=true \
  --set ui.ingress.annotations."cert-manager\.io/cluster-issuer"=letsencrypt-prod \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

### **3. Ingress with Main Service (Both Metrics + UI)**

```bash
# Use main service that includes both metrics (8080) and UI (8082)
helm upgrade --install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --set ui.enabled=true \
  --set ui.ingress.enabled=true \
  --set ui.ingress.serviceName=mimir-limit-optimizer \
  --set ui.ingress.servicePort=8082 \
  --set ui.ingress.hosts[0].host=mimir-optimizer.your-domain.com \
  --set ui.ingress.hosts[0].paths[0].path=/ \
  --set ui.ingress.hosts[0].paths[0].pathType=Prefix \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

## 📋 **Values.yaml Configuration**

### **Complete Ingress Configuration**

```yaml
ui:
  enabled: true
  port: 8082
  
  service:
    type: ClusterIP
    port: 8082
    targetPort: 8082
  
  ingress:
    # Enable ingress for external access
    enabled: true
    
    # Ingress class name (e.g., nginx, traefik, alb)
    className: "nginx"
    
    # Service configuration for ingress
    serviceName: ""  # Default: uses UI service
    servicePort: ""  # Default: uses ui.service.port
    
    # Alternative: Use main service (includes both metrics and UI)
    # serviceName: "mimir-limit-optimizer"  # Use main service
    # servicePort: 8082                     # UI port on main service
    
    # Annotations for the ingress
    annotations:
      kubernetes.io/ingress.class: nginx
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/rewrite-target: /
    
    # Hosts configuration
    hosts:
      - host: mimir-optimizer.your-domain.com
        paths:
          - path: /
            pathType: Prefix
    
    # TLS configuration
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.your-domain.com
```

## 🔧 **Service Options**

### **Option 1: UI Service (Default)**
- **Service Name**: `<release-name>-mimir-limit-optimizer-ui`
- **Port**: `8082`
- **Purpose**: Dedicated UI service
- **Use Case**: When you want separated services

```yaml
ui:
  ingress:
    # Uses UI service by default
    serviceName: ""  # Auto-resolves to UI service
    servicePort: ""  # Auto-resolves to ui.service.port
```

### **Option 2: Main Service**
- **Service Name**: `<release-name>-mimir-limit-optimizer`
- **Ports**: `8080` (metrics) + `8082` (UI)
- **Purpose**: Combined service for both metrics and UI
- **Use Case**: When you want single service with multiple ports

```yaml
ui:
  ingress:
    # Use main service explicitly
    serviceName: "mimir-limit-optimizer"  # Main service
    servicePort: 8082                     # UI port
```

## 🌐 **Ingress Controller Specific Examples**

### **NGINX Ingress Controller**

```yaml
ui:
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/rewrite-target: /
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.example.com
```

### **Traefik Ingress Controller**

```yaml
ui:
  ingress:
    enabled: true
    className: "traefik"
    annotations:
      traefik.ingress.kubernetes.io/router.entrypoints: websecure
      traefik.ingress.kubernetes.io/router.tls: "true"
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.example.com
```

### **AWS ALB Ingress Controller**

```yaml
ui:
  ingress:
    enabled: true
    className: "alb"
    annotations:
      kubernetes.io/ingress.class: alb
      alb.ingress.kubernetes.io/scheme: internet-facing
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:region:account:certificate/cert-id
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix
```

## 🔍 **Verification Steps**

### **1. Check Ingress Status**
```bash
# Check if ingress is created
kubectl get ingress -n mimir-limit-optimizer

# Describe ingress for details
kubectl describe ingress -n mimir-limit-optimizer
```

### **2. Verify Service Endpoints**
```bash
# Check services
kubectl get svc -n mimir-limit-optimizer

# Check endpoints
kubectl get endpoints -n mimir-limit-optimizer
```

### **3. Test Access**
```bash
# Test UI access (replace with your actual domain)
curl -I https://mimir-optimizer.your-domain.com/

# Check if UI is loading
curl -s https://mimir-optimizer.your-domain.com/ | grep -i "mimir"
```

## 🛠️ **Troubleshooting**

### **Issue: Ingress Not Created**
```bash
# Check if UI is enabled
helm get values mimir-limit-optimizer -n mimir-limit-optimizer | grep -A10 "ui:"

# Verify ingress is enabled in values
kubectl get configmap -n mimir-limit-optimizer -o yaml | grep ingress
```

### **Issue: 404 Not Found**
```bash
# Check service endpoints
kubectl get endpoints -n mimir-limit-optimizer

# Verify pod is running and ready
kubectl get pods -n mimir-limit-optimizer

# Check ingress backend service
kubectl describe ingress -n mimir-limit-optimizer | grep -A5 "Backend"
```

### **Issue: SSL/TLS Errors**
```bash
# Check certificate
kubectl get certificates -n mimir-limit-optimizer

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager
```

### **Issue: Wrong Service**
```bash
# Check which service ingress is pointing to
kubectl get ingress -n mimir-limit-optimizer -o yaml | grep -A3 "service:"

# List all services to verify names
kubectl get svc -n mimir-limit-optimizer
```

## 📖 **Example Values Files**

### **Basic Ingress (values-ingress-basic.yaml)**
```yaml
ui:
  enabled: true
  ingress:
    enabled: true
    hosts:
      - host: mimir-optimizer.local
        paths:
          - path: /
            pathType: Prefix
```

### **Production Ingress (values-ingress-prod.yaml)**
```yaml
ui:
  enabled: true
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    hosts:
      - host: mimir-optimizer.company.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.company.com
```

### **Development Ingress (values-ingress-dev.yaml)**
```yaml
ui:
  enabled: true
  ingress:
    enabled: true
    serviceName: "mimir-limit-optimizer"  # Use main service
    servicePort: 8082
    hosts:
      - host: mimir-optimizer-dev.local
        paths:
          - path: /
            pathType: Prefix
```

Deploy with custom values:
```bash
helm upgrade --install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  -f values-ingress-prod.yaml \
  --namespace mimir-limit-optimizer \
  --create-namespace
```

## ✅ **Production Checklist**

- [ ] **Domain/DNS**: Configure DNS to point to your ingress controller
- [ ] **SSL Certificate**: Set up cert-manager or provide manual certificates
- [ ] **Ingress Controller**: Ensure ingress controller is installed and running
- [ ] **Network Policies**: Configure appropriate network policies if needed
- [ ] **Authentication**: Consider adding authentication proxy if required
- [ ] **Monitoring**: Monitor ingress metrics and SSL certificate expiry
- [ ] **Backup**: Backup ingress configurations and certificates

## 🔗 **Related Documentation**

- [Testing UI Guide](TESTING_UI_GUIDE.md)
- [Helm Chart Values](helm/mimir-limit-optimizer/values.yaml)
- [Service Configuration](helm/mimir-limit-optimizer/templates/service.yaml)
- [UI Service Template](helm/mimir-limit-optimizer/templates/ui-service.yaml) 