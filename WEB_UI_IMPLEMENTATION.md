# Mimir Limit Optimizer - Web UI Dashboard Implementation

## Overview

The Mimir Limit Optimizer now includes a comprehensive web-based UI dashboard that provides real-time monitoring, visualization, and control capabilities for the limit optimization system. The UI is served directly from the controller binary on port 8081 and embedded using Go's `embed` package.

## Architecture

### Backend (Go)

- **API Server**: RESTful API server using `gorilla/mux` router
- **Embedded Assets**: React build artifacts embedded using `//go:embed`
- **Port**: Serves UI and APIs on port 8081
- **Integration**: Direct integration with controller components

### Frontend (React + TypeScript)

- **Framework**: React 18 with TypeScript
- **Styling**: TailwindCSS with dark/light theme support
- **Charts**: Recharts for data visualization
- **Routing**: React Router for SPA navigation
- **Animations**: Framer Motion for smooth transitions

## Features Implemented

### 1. System Overview Dashboard

**Location**: `/dashboard`

**Capabilities**:
- Real-time system status display
- Controller mode indicator (Dry-Run/Production)
- Tenant count and health metrics
- Circuit breaker status
- Reconciliation activity charts
- Component health monitoring

**API Endpoints**:
- `GET /api/status` - System status and metrics
- `GET /api/tenants` - Tenant summary data

### 2. Tenant Management

**Location**: `/tenants`

**Capabilities**:
- Searchable tenant list
- Tenant health indicators
- Ingestion rate and series count
- Applied vs suggested limits comparison
- Spike detection status
- Individual tenant detail views

**API Endpoints**:
- `GET /api/tenants` - List all tenants
- `GET /api/tenants/{id}` - Detailed tenant information

### 3. Runtime Configuration Editor

**Location**: `/config`

**Capabilities**:
- Live configuration editing
- Form controls for all major settings:
  - Operating mode (Dry-Run/Production)
  - Buffer percentage slider
  - Spike detection threshold
  - Update interval dropdown
  - Circuit breaker toggle
  - Auto-discovery settings
  - Tenant include/exclude lists
- Real-time validation
- ConfigMap updates in production mode

**API Endpoints**:
- `GET /api/config` - Current configuration
- `POST /api/config` - Update configuration

### 4. Audit Log Viewer

**Location**: `/audit`

**Capabilities**:
- Filterable audit log entries
- Search by tenant, action type, date range
- Paginated results
- Detailed view of configuration changes
- Export capabilities

**API Endpoints**:
- `GET /api/audit` - Audit log entries with filters

### 5. Dry-Run vs Production Diff Viewer

**Location**: `/diff`

**Capabilities**:
- Side-by-side comparison of limits
- Color-coded differences:
  - 🟢 Green: Identical values
  - 🔴 Red: Mismatched values
  - 🟡 Yellow: Dry-run only (not applied)
- Delta calculations
- "Apply Dry-Run Limits" button (production mode)
- Export diff reports

**API Endpoints**:
- `GET /api/diff` - Limit differences analysis

### 6. System Metrics Viewer

**Location**: `/metrics`

**Capabilities**:
- Controller performance metrics
- Reconciliation statistics
- Per-tenant aggregated metrics
- Prometheus metrics preview
- Export to CSV/JSON

**API Endpoints**:
- `GET /api/metrics` - Redirect to Prometheus endpoint
- `GET /metrics` - Prometheus metrics

### 7. Synthetic Test Tools

**Location**: `/test-tools`

**Capabilities**:
- Trigger synthetic ingestion spikes
- Test alert webhooks (Slack/PagerDuty)
- Manual reconciliation trigger
- Configuration validation
- Health check endpoints

**API Endpoints**:
- `POST /api/test/spike` - Trigger test spike
- `POST /api/test/alert` - Send test alert
- `POST /api/test/reconcile` - Manual reconciliation

## Build Process

### Development Workflow

1. **Start Backend API**:
   ```bash
   make run-dev  # Runs without building UI
   ```

2. **Start Frontend Dev Server**:
   ```bash
   make ui-dev   # Starts React dev server on :3000
   ```

3. **Development URLs**:
   - Frontend: `http://localhost:3000` (with API proxy)
   - Backend API: `http://localhost:8081/api/*`

### Production Build

1. **Build UI and Binary**:
   ```bash
   make build    # Builds UI first, then Go binary with embedded assets
   ```

2. **Docker Build**:
   ```bash
   make docker-build  # Includes UI build step
   ```

3. **Manual UI Build**:
   ```bash
   make build-ui  # Builds just the React app
   ```

## Integration Points

### Controller Integration

The API server integrates directly with the `MimirLimitController`:

```go
// In main.go
mimirController := &controller.MimirLimitController{...}
apiServer := api.NewServer(mimirController, cfg, logger)
```

### Embedded Assets

UI assets are embedded at build time:

```go
//go:embed ui/build/*
var uiAssets embed.FS
```

### API Routes

The API server provides these route groups:

- `/api/status` - System information
- `/api/config` - Configuration management  
- `/api/tenants` - Tenant operations
- `/api/diff` - Limit comparison
- `/api/audit` - Audit log access
- `/api/test/*` - Testing utilities
- `/health` - Health checks
- `/metrics` - Prometheus metrics
- `/` - Static UI files

## Security Considerations

### Current Implementation

- CORS enabled for development
- No authentication (suitable for internal cluster access)
- API validation for configuration updates

### Future Enhancements

- Kubernetes RBAC integration
- JWT token authentication
- Rate limiting
- Request logging and monitoring

## Configuration

### Runtime Configuration

The web UI can modify these runtime settings:

```yaml
# Modifiable via UI
mode: "dry-run"           # or "prod"
bufferPercentage: 20.0    # Buffer %
updateInterval: "5m"      # Update frequency
circuitBreaker:
  enabled: true           # Circuit breaker toggle
eventSpike:
  enabled: true           # Spike detection toggle
  threshold: 2.0          # Spike threshold multiplier
tenantScoping:
  skipList: []            # Tenant exclusions
  includeList: []         # Tenant inclusions
```

### Build Configuration

```bash
# Environment variables for build
GENERATE_SOURCEMAP=false  # Smaller build size
NODE_ENV=production       # Production optimizations
```

## Performance

### Frontend Optimizations

- Code splitting with React.lazy()
- Production build minification
- Asset compression
- Efficient chart rendering with Recharts
- Responsive images and icons

### Backend Optimizations

- Embedded assets (no file system access)
- Gzip compression for API responses
- Efficient JSON marshaling
- Request timeout handling

## Development Guidelines

### Adding New API Endpoints

1. **Define handler in `pkg/api/handlers.go`**:
   ```go
   func (s *Server) handleNewEndpoint(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

2. **Add route in `pkg/api/server.go`**:
   ```go
   api.HandleFunc("/new-endpoint", s.handleNewEndpoint).Methods("GET")
   ```

3. **Update frontend API context**:
   ```typescript
   const newEndpoint = useCallback(() => 
     apiCall(() => axios.get(`${API_BASE}/new-endpoint`).then(res => res.data)), 
     [apiCall]
   );
   ```

### Adding New UI Pages

1. **Create component in `ui/src/pages/`**
2. **Add route in `ui/src/App.tsx`**
3. **Add navigation item in `ui/src/components/layout/Sidebar.tsx`**
4. **Update TypeScript types in `ui/src/context/ApiContext.tsx`**

## Deployment

### Helm Chart Configuration

The UI feature is configurable in the Helm chart and **enabled by default**. You can control it using the following values:

```yaml
# values.yaml
ui:
  # Enable/disable the web dashboard (default: true)
  enabled: true
  
  # Port for the web UI and API server (default: 8081)
  port: 8081
  
  # Service configuration for the UI
  service:
    # Service type (ClusterIP, NodePort, LoadBalancer)
    type: ClusterIP
    
    # Port for the service
    port: 8081
    
    # Target port on the pod
    targetPort: 8081
    
    # Annotations for the service
    annotations: {}
  
  # Ingress configuration for external access
  ingress:
    # Enable ingress for external access (default: false)
    enabled: false
    
    # Ingress class name
    className: ""
    
    # Annotations for the ingress
    annotations: {}
      # kubernetes.io/ingress.class: nginx
      # cert-manager.io/cluster-issuer: letsencrypt-prod
    
    # Hosts configuration
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix
    
    # TLS configuration
    tls: []
      # - secretName: mimir-optimizer-tls
      #   hosts:
      #     - mimir-optimizer.example.com
```

#### Deployment Examples

**1. Default Deployment (UI Enabled)**
```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir \
  --create-namespace
```

**2. Disable UI Feature**
```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir \
  --create-namespace \
  --set ui.enabled=false
```

**3. Enable External Access via Ingress**
```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir \
  --create-namespace \
  --set ui.ingress.enabled=true \
  --set ui.ingress.hosts[0].host=mimir-optimizer.example.com \
  --set ui.ingress.hosts[0].paths[0].path=/ \
  --set ui.ingress.hosts[0].paths[0].pathType=Prefix
```

**4. Custom UI Port and Service Type**
```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir \
  --create-namespace \
  --set ui.port=9090 \
  --set ui.service.type=NodePort \
  --set ui.service.port=9090
```

#### Values File Example

Create a custom values file for production deployment:
```yaml
# custom-values.yaml
# Enable UI with external access
ui:
  enabled: true
  port: 8081
  service:
    type: ClusterIP
    port: 8081
    annotations:
      prometheus.io/scrape: "false"
  ingress:
    enabled: true
    className: "nginx"
    annotations:
      kubernetes.io/ingress.class: nginx
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
    hosts:
      - host: mimir-optimizer.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-optimizer-tls
        hosts:
          - mimir-optimizer.example.com

# Configure controller settings
controller:
  mode: "prod"
  bufferPercentage: 25

# Configure Mimir integration
mimir:
  namespace: "mimir"
  configMapName: "mimir-runtime-overrides"
```

Then deploy:
```bash
helm install mimir-limit-optimizer ./helm/mimir-limit-optimizer \
  --namespace mimir \
  --create-namespace \
  --values custom-values.yaml
```

### Kubernetes Deployment

The web UI is included in the standard controller deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mimir-limit-optimizer
spec:
  template:
    spec:
      containers:
      - name: controller
        image: your-registry/mimir-limit-optimizer:v2.4.0
        ports:
        - containerPort: 8080  # Metrics
          name: metrics
        - containerPort: 8081  # Web UI + API
          name: web-ui
```

### Service Configuration

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mimir-limit-optimizer-ui
spec:
  ports:
  - port: 8081
    targetPort: 8081
    name: web-ui
  selector:
    app: mimir-limit-optimizer
```

### Ingress Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mimir-limit-optimizer-ui
spec:
  rules:
  - host: mimir-optimizer.your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: mimir-limit-optimizer-ui
            port:
              number: 8081
```

## Monitoring and Logging

### API Request Logging

All API requests are logged with:
- Method and path
- Duration
- Response status
- Client IP address

### Error Handling

- Structured error responses
- Client-side error boundaries
- Automatic retry for failed requests
- User-friendly error messages

### Health Checks

- `/health` endpoint for basic health
- Component health status in dashboard
- Real-time status indicators

## Testing

### Frontend Testing

```bash
cd ui
npm test  # Jest + React Testing Library
```

### Backend Testing

```bash
go test ./pkg/api/...  # API handler tests
```

### Integration Testing

```bash
make test  # Full test suite
```

## Future Enhancements

### Planned Features

1. **Real-time WebSocket updates**
2. **Advanced alerting configuration**
3. **Tenant tier management UI**
4. **Cost analysis dashboards**
5. **Multi-cluster support**
6. **Custom dashboard widgets**
7. **Grafana dashboard integration**
8. **Export/import configurations**

### Architecture Improvements

1. **Micro-frontend architecture**
2. **Progressive Web App (PWA)**
3. **Server-side rendering (SSR)**
4. **Advanced caching strategies**

## Troubleshooting

### Common Issues

1. **UI not loading**: Check if `ui/build` directory exists
2. **API errors**: Verify controller is running and healthy
3. **Build failures**: Ensure Node.js 16+ and npm are installed
4. **Theme issues**: Check localStorage for theme preference

### Debug Mode

```bash
# Start with debug logging
go run main.go --log-level=debug

# Check API endpoints
curl http://localhost:8081/api/status
```

### Log Analysis

```bash
# Controller logs
kubectl logs -f deployment/mimir-limit-optimizer -c controller

# API server logs  
kubectl logs -f deployment/mimir-limit-optimizer -c controller | grep "api"
```

## Conclusion

The web UI dashboard transforms the Mimir Limit Optimizer from a command-line tool into a comprehensive, user-friendly platform for monitoring and managing Mimir tenant limits. The embedded architecture ensures zero additional deployment complexity while providing enterprise-grade observability and control capabilities. 