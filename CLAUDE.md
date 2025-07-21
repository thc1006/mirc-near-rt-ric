# O‑RAN Near‑RT RIC Project Development Guide

## Project Background
This repository provides an **O‑RAN Near Real‑Time RAN Intelligent Controller (Near‑RT RIC)** platform intended to deliver a production‑grade, O‑RAN‑compliant RIC solution.

## Core Functional Requirements

### Mandatory O‑RAN Interfaces
- **E2 Interface** – connects the Near‑RT RIC to E2 nodes (DU, CU, eNB)  
  - Latency SLA: **10 ms – 1 s**  
  - Supports **E2AP** (E2 Application Protocol)  
  - Implements **E2 Service Models (E2SM)**  
  - Provides RIC **subscription, control, and query** capabilities  

- **A1 Interface** – bridges the Non‑RT RIC and Near‑RT RIC  
  - **Policy Management Service**  
  - **ML Model Management Service**  
  - **Enrichment Information Service**  
  - Full **A1 policy** lifecycle management  

- **O1 Interface** – management & configuration plane  
  - Supports **FCAPS** (Fault, Configuration, Accounting, Performance, Security)  
  - Uses **NETCONF/YANG**  
  - Handles **software & file management**

### xApp Development Framework
- xApp **lifecycle management**
- xApp **deployment & configuration**
- **Conflict avoidance** among xApps
- xApp **observability** (monitoring & logging)

### Federated Learning Capabilities
- **Distributed** model training
- Global **model aggregation & synchronization**
- **Privacy‑preserving** mechanisms
- **Model versioning** and rollback

## Technical Architecture

### Backend Tech‑Stack
- **Languages:** Go (primary), Python (ML/AI)
- **Containerization:** Docker, Kubernetes
- **Communication:** gRPC, REST API
- **Datastores:** Time‑series DB (**InfluxDB**), Relational DB (**PostgreSQL**)
- **Message Brokers:** Apache Kafka, Redis

### Front‑End Tech‑Stack
- **Framework:** Angular 15 +
- **UI Library:** Angular Material
- **Charting:** Chart.js, D3.js
- **State Management:** NgRx

### Deployment & DevOps
- **Orchestrator:** Kubernetes
- **Service Mesh:** Istio
- **Monitoring:** Prometheus + Grafana
- **Logging:** ELK Stack (Elasticsearch / Logstash / Kibana)
- **CI/CD:** GitHub Actions

## Coding Guidelines

### Go Guidelines
```go
// Go version: 1.19+
// Always run `gofmt`
// Use `golint` / staticcheck for linting
// All exported members MUST have documentation comments

// Example struct
type E2Interface struct {
    NodeID     string            `json:"node_id"`
    Connection *grpc.ClientConn  `json:"-"`
    Status     ConnectionStatus  `json:"status"`
    Services   []E2ServiceModel  `json:"services"`
}

// Explicit & descriptive error handling
func (e *E2Interface) Connect() error {
    if e.Connection != nil {
        return errors.New("already connected")
    }

    conn, err := grpc.Dial(e.NodeID, grpc.WithInsecure())
    if err != nil {
        return fmt.Errorf("failed to connect to E2 node %s: %w", e.NodeID, err)
    }

    e.Connection = conn
    return nil
}
````

### TypeScript / Angular Guidelines

```typescript
// Strict mode enabled
// Follow the official Angular Style Guide
// TypeScript 4.7+

@Injectable({
  providedIn: 'root'
})
export class XAppService {
  private readonly apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getXAppList(): Observable<XApp[]> {
    return this.http.get<XApp[]>(`${this.apiUrl}/xapps`);
  }

  deployXApp(xapp: XAppDeployment): Observable<XApp> {
    return this.http.post<XApp>(`${this.apiUrl}/xapps`, xapp);
  }
}
```

## Testing Requirements

### Unit Tests

* **Go:** Testify; coverage ≥ **80 %**
* **TypeScript:** Jasmine / Karma; coverage ≥ **80 %**

### Integration Tests

* E2 interface **emulator**
* **End‑to‑end** A1 interface tests
* xApp **deployment** tests

### Performance Tests

* **E2 latency** (< 10 ms)
* **Concurrency** ≥ 100 E2 nodes
* **Federated learning** throughput & convergence

## Security Requirements

### Authentication & Authorization

* **OAuth 2.0 / JWT**
* **RBAC** (Role‑Based Access Control)
* **MFA** support

### Network Security

* **TLS 1.3** encryption
* **Certificate** lifecycle management
* **Firewall** rule hardening

### Data Protection

* Privacy‑preserving **federated learning**
* Data **encryption at rest**
* **Audit logging** (immutability preferred)

## Deployment Guide

### Development Environment

```bash
# Spin‑up dev environment
make dev-setup
docker-compose up -d

# Run tests
make test
make integration-test
```

### Production Deployment

```bash
# Kubernetes deployment
helm install oran-ric ./helm/oran-ric
kubectl apply -f k8s/
```

## Common Commands

### Development

```bash
# Backend
go run cmd/ric/main.go
go test ./...
go mod tidy

# Front‑end
ng serve
ng test
ng e2e
```

### Debugging

```bash
# Inspect E2 connectivity
kubectl logs -f deployment/e2-interface
curl http://localhost:8080/health

# Inspect xApp status
kubectl get pods -l app=xapp
kubectl describe xapp my-xapp
```

## Performance Metrics

### Key Performance Indicators (KPIs)

* **E2 interface latency:** < 10 ms (P99)
* **A1 policy deployment time:** < 1 s
* **xApp deployment time:** < 30 s
* **System availability:** 99.9 %
* **Federated learning convergence:** < 5 min

### Monitoring Metrics

```go
// Prometheus metric examples
var (
    e2MessageCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "e2_messages_total",
            Help: "Total E2 messages processed",
        },
        []string{"node_id", "message_type", "status"},
    )

    a1PolicySuccessRate = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "a1_policy_success_rate",
            Help: "Success rate of A1 policy deployments",
        },
    )
)
```

## Directory Layout

```
near-rt-ric/
├── cmd/                    # Entrypoints
│   ├── ric/               # RIC main binary
│   └── xapp-manager/      # xApp manager
├── pkg/                   # Shared libraries
│   ├── e2/               # E2 implementation
│   ├── a1/               # A1 implementation
│   ├── o1/               # O1 implementation
│   ├── xapp/             # xApp framework
│   └── federation/       # Federated Learning
├── internal/             # Private packages
│   ├── config/          # Configuration mgmt
│   ├── database/        # Persistence layer
│   └── metrics/         # Custom Prom metrics
├── web/                 # Front‑end
│   ├── src/            # Angular sources
│   └── dist/           # Build artifacts
├── helm/               # Helm charts
├── k8s/                # Kubernetes manifests
├── docker/             # Container artifacts
└── docs/               # Documentation
```

## Important Notes

### ⚠️ Critical Reminders

1. **No mock data** – all O‑RAN functionality must be genuinely implemented.
2. **Tight performance budgets** – the E2 interface MUST satisfy the 10 ms – 1 s latency requirement.
3. **Standards compliance** – the project MUST conform to O‑RAN Alliance specifications.
4. **Security first** – absolutely **no** hard‑coded secrets or insecure defaults.
5. **Test‑driven development** – every new feature MUST ship with tests.

### 🔧 Technical‑Debt Management

* Prioritize performance‑critical code
* **Incremental refactoring** – avoid big‑bang rewrites
* Maintain **backward compatibility**
* **Regularly** update dependencies

### 📚 Learning Resources

* [O‑RAN Alliance Specifications](https://www.o-ran.org/specifications)
* [E2 Interface Spec (ETSI TS 104 038)](https://www.etsi.org/deliver/etsi_ts/104000_104099/104038/)
* [A1 Interface Spec (ETSI TS 103 983)](https://www.etsi.org/deliver/etsi_ts/103900_103999/103983/)
* [O‑RAN Software Community](https://o-ran-sc.org/)