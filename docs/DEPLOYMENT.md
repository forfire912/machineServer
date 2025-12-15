# Deployment Guide

## Prerequisites

### System Requirements
- Linux or macOS (Windows via WSL2)
- Go 1.21 or later (for building from source)
- Docker and Docker Compose (for containerized deployment)
- Kubernetes cluster (for K8s deployment)

### Backend Requirements
- QEMU (if using QEMU backend)
- Renode (if using Renode backend)

## Local Development

### 1. Build from Source

```bash
# Clone repository
git clone https://github.com/forfire912/machineServer.git
cd machineServer

# Install dependencies
go mod download

# Build
make build

# The binary will be in bin/machineserver
```

### 2. Configuration

Edit `configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # Use "debug" for development

auth:
  enabled: false  # Disable auth for local development
  
backends:
  qemu:
    enabled: true
    binary: "/usr/bin/qemu-system-arm"  # Adjust path
  renode:
    enabled: true
    binary: "/usr/bin/renode"  # Adjust path
```

### 3. Run Locally

```bash
# Option 1: Using make
make run

# Option 2: Direct execution
./bin/machineserver -config=configs/config.yaml

# Option 3: Using go run
go run cmd/server/main.go -config=configs/config.yaml
```

### 4. Test the Server

```bash
# Health check
curl http://localhost:8080/health

# Get capabilities (no auth in dev mode)
curl http://localhost:8080/api/v1/capabilities
```

## Docker Deployment

### 1. Build Docker Image

```bash
# Build image
docker build -t machineserver:latest .

# Verify image
docker images | grep machineserver
```

### 2. Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f machineserver

# Stop services
docker-compose down
```

### 3. Docker Compose Services

The stack includes:
- **machineserver**: Main application
- **redis**: Job queue backend
- **prometheus**: Metrics collection

### 4. Accessing Services

- API: http://localhost:8080
- Prometheus: http://localhost:9091
- Metrics: http://localhost:8080/metrics

### 5. Data Persistence

Volumes are mounted for:
- `./data`: Database and storage
- `./logs`: Application logs
- `./configs`: Configuration files

## Kubernetes Deployment

### 1. Prerequisites

```bash
# Ensure kubectl is configured
kubectl cluster-info

# Ensure you have appropriate permissions
kubectl auth can-i create deployments --all-namespaces
```

### 2. Create Namespace (Optional)

```bash
kubectl create namespace machineserver
kubectl config set-context --current --namespace=machineserver
```

### 3. Apply Configurations

```bash
# Deploy Redis
kubectl apply -f deployments/kubernetes/redis.yaml

# Deploy MachineServer
kubectl apply -f deployments/kubernetes/deployment.yaml

# Verify deployments
kubectl get deployments
kubectl get pods
kubectl get services
```

### 4. Configuration Management

The ConfigMap in `deployment.yaml` contains the configuration. To update:

```bash
# Edit the ConfigMap in the YAML file
vim deployments/kubernetes/deployment.yaml

# Apply changes
kubectl apply -f deployments/kubernetes/deployment.yaml

# Restart pods to pick up changes
kubectl rollout restart deployment machineserver
```

### 5. Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment machineserver --replicas=5

# Check status
kubectl get pods -l app=machineserver

# Auto-scaling (optional)
kubectl autoscale deployment machineserver \
  --cpu-percent=80 --min=3 --max=10
```

### 6. Access the Service

```bash
# Get service details
kubectl get service machineserver

# For LoadBalancer (cloud providers)
# External IP will be assigned automatically

# For NodePort (on-premise)
kubectl get service machineserver -o wide

# For Port-forwarding (testing)
kubectl port-forward service/machineserver 8080:80
```

### 7. Monitoring

```bash
# Check logs
kubectl logs -f deployment/machineserver

# Check pod status
kubectl describe pod <pod-name>

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

### 8. Persistent Storage

The deployment uses a PersistentVolumeClaim for data storage:

```bash
# Check PVC status
kubectl get pvc machineserver-data

# Check PV status
kubectl get pv
```

For production, configure storage class:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: machineserver-data
spec:
  storageClassName: fast-ssd  # Your storage class
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```

## Production Deployment

### Security Best Practices

1. **Enable Authentication**
```yaml
auth:
  enabled: true
  jwt_secret: "<strong-random-secret>"
  api_keys:
    - "<secure-api-key-1>"
    - "<secure-api-key-2>"
```

2. **Use TLS/HTTPS**
```bash
# Add TLS certificate to Kubernetes secret
kubectl create secret tls machineserver-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key

# Update service to use HTTPS
# Add ingress with TLS configuration
```

3. **Network Policies**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: machineserver-policy
spec:
  podSelector:
    matchLabels:
      app: machineserver
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 8080
```

4. **Resource Limits**
```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "2000m"
```

5. **Security Context**
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  capabilities:
    drop:
      - ALL
```

### High Availability

1. **Multiple Replicas**
```bash
kubectl scale deployment machineserver --replicas=3
```

2. **Pod Disruption Budget**
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: machineserver-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: machineserver
```

3. **Health Probes**
Already configured in deployment:
- Liveness probe: Restarts unhealthy pods
- Readiness probe: Routes traffic only to ready pods

### Monitoring Setup

1. **Prometheus Integration**
```yaml
# Add ServiceMonitor for Prometheus Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: machineserver
spec:
  selector:
    matchLabels:
      app: machineserver
  endpoints:
  - port: metrics
    interval: 30s
```

2. **Grafana Dashboards**
Create dashboards for:
- Request rate and latency
- Active sessions
- Error rates
- Resource utilization

### Backup and Recovery

1. **Database Backup**
```bash
# Backup SQLite database
kubectl exec deployment/machineserver -- \
  sqlite3 /root/data/machineserver.db ".backup /root/data/backup.db"

# Copy backup locally
kubectl cp machineserver-pod:/root/data/backup.db ./backup.db
```

2. **Volume Snapshots**
```bash
# For cloud providers with snapshot support
kubectl create volumesnapshot machineserver-snapshot \
  --volume-snapshot-class=csi-snapshot-class \
  --source=machineserver-data
```

### Rolling Updates

```bash
# Update image
kubectl set image deployment/machineserver \
  machineserver=machineserver:v2.0.0

# Check rollout status
kubectl rollout status deployment/machineserver

# Rollback if needed
kubectl rollout undo deployment/machineserver
```

## Troubleshooting

### Common Issues

**1. Pod Not Starting**
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

**2. Service Unreachable**
```bash
kubectl get endpoints machineserver
kubectl get service machineserver
```

**3. Database Connection Issues**
```bash
# Check PVC
kubectl get pvc
kubectl describe pvc machineserver-data

# Check if database file exists
kubectl exec -it <pod-name> -- ls -la /root/data/
```

**4. Backend Not Found**
```bash
# Check if QEMU/Renode is available in container
kubectl exec -it <pod-name> -- which qemu-system-arm
kubectl exec -it <pod-name> -- which renode
```

### Performance Tuning

1. **Database Optimization**
   - Use separate database server for production
   - Configure connection pooling
   - Regular VACUUM operations

2. **Redis Configuration**
   - Increase max memory if needed
   - Configure persistence (RDB/AOF)
   - Enable clustering for high load

3. **Resource Allocation**
   - Monitor CPU/memory usage
   - Adjust pod limits accordingly
   - Use vertical pod autoscaler

## Maintenance

### Regular Tasks

1. **Log Rotation**
```bash
# Logs are handled by Kubernetes
# Configure retention in logging backend
```

2. **Database Maintenance**
```bash
kubectl exec -it <pod-name> -- \
  sqlite3 /root/data/machineserver.db "VACUUM;"
```

3. **Cleanup Old Data**
```bash
# Implement cleanup jobs
# Delete old sessions, snapshots, etc.
```

### Upgrades

1. **Minor Updates**
```bash
# Pull latest config changes
git pull

# Apply updates
kubectl apply -f deployments/kubernetes/
```

2. **Major Version Upgrades**
- Review changelog
- Test in staging environment
- Backup data
- Perform rolling update
- Monitor for issues
- Rollback if needed
