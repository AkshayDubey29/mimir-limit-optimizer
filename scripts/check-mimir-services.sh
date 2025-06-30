#!/bin/bash

echo "🔍 Mimir Service Discovery Diagnostic Script"
echo "============================================"

NAMESPACE="mimir"

echo "📋 Services in $NAMESPACE namespace:"
kubectl get services -n $NAMESPACE --no-headers | awk '{print "  - " $1 " (" $2 ")"}'

echo -e "\n🏷️  Service Labels Analysis:"
echo "Checking common labels across your Mimir services..."

# Check for common label patterns
echo -e "\n1. Checking for 'app.kubernetes.io/name' labels:"
kubectl get services -n $NAMESPACE -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.metadata.labels.app\.kubernetes\.io/name}{"\n"}{end}' | grep -v "^.*  $" || echo "  No services with this label"

echo -e "\n2. Checking for 'app.kubernetes.io/part-of' labels:"
kubectl get services -n $NAMESPACE -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.metadata.labels.app\.kubernetes\.io/part-of}{"\n"}{end}' | grep -v "^.*  $" || echo "  No services with this label"

echo -e "\n3. Checking for 'app' labels:"
kubectl get services -n $NAMESPACE -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.metadata.labels.app}{"\n"}{end}' | grep -v "^.*  $" || echo "  No services with this label"

echo -e "\n4. Checking for 'component' labels:"
kubectl get services -n $NAMESPACE -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.metadata.labels.component}{"\n"}{end}' | grep -v "^.*  $" || echo "  No services with this label"

echo -e "\n🔍 Detailed service information for metrics discovery:"
echo "Checking ports and endpoints..."

for service in distributor ingester-zone-a querier query-frontend compactor store-gateway-multi-zone alertmanager; do
    if kubectl get service $service -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "\n📊 Service: $service"
        kubectl get service $service -n $NAMESPACE -o jsonpath='{.spec.ports[*].name}{"  ports: "}{.spec.ports[*].port}{"\n"}' | sed 's/^/  /'
        kubectl get service $service -n $NAMESPACE -o jsonpath='{"  labels: "}{.metadata.labels}{"\n"}'
    fi
done

echo -e "\n💡 Recommended Configuration:"
echo "Based on your services, try these configurations in your values.yaml:"

echo -e "\nOption 1 - If your services have app.kubernetes.io labels:"
echo "  serviceLabelSelector: \"app.kubernetes.io/part-of=mimir\""

echo -e "\nOption 2 - If your services have app labels:"
echo "  serviceLabelSelector: \"app=mimir\""

echo -e "\nOption 3 - If your services have component labels:"
echo "  serviceLabelSelector: \"component in (distributor,ingester,querier,query-frontend,compactor,store-gateway,alertmanager)\""

echo -e "\nOption 4 - Manual service list (fallback):"
echo "  Use the serviceNames list I provided in values.yaml"

echo -e "\n🚀 Test the configuration:"
echo "After updating values.yaml, deploy and check the logs:"
echo "  kubectl logs -n <your-namespace> deployment/mimir-limit-optimizer -f" 