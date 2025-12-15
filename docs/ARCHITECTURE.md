# Architecture Documentation

## Overview

Machine Server is a unified simulation microservices platform designed to provide a consistent RESTful API for embedded system simulation using multiple backends (QEMU and Renode).

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Applications                     │
│   (Web UI, CLI Tools, IDEs, Testing Frameworks)             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    HTTP/WebSocket API                        │
│  ┌──────────┬──────────┬──────────┬───────────┬──────────┐  │
│  │ Sessions │ Programs │ Control  │ Snapshots │ Streams  │  │
│  └──────────┴──────────┴──────────┴───────────┴──────────┘  │
│         ┌──────────────┬──────────────┬──────────────┐       │
│         │ Auth         │ Audit        │ Metrics      │       │
│         │ Middleware   │ Middleware   │ Middleware   │       │
│         └──────────────┴──────────────┴──────────────┘       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Session Management │ Program Management │ Snapshots │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Coverage Analysis  │ Job Queue  │ Audit Logging     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Adapter Layer                           │
│  ┌──────────────────────┬─────────────────────────────┐     │
│  │   QEMU Adapter       │    Renode Adapter           │     │
│  │   - QMP Protocol     │    - Monitor Protocol       │     │
│  │   - Process Mgmt     │    - Process Mgmt           │     │
│  │   - GDB Integration  │    - GDB Integration        │     │
│  └──────────────────────┴─────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Backend Simulation Engines                      │
│  ┌──────────────────────┬─────────────────────────────┐     │
│  │   QEMU               │    Renode                   │     │
│  │   (ARM, RISC-V, etc) │    (Multi-architecture)     │     │
│  └──────────────────────┴─────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. API Layer (`internal/api`)

**Responsibilities:**
- HTTP request handling
- Request validation
- Response formatting
- WebSocket management
- Authentication/Authorization
- Audit logging
- Metrics collection

**Key Files:**
- `handler.go`: HTTP endpoint handlers
- `router.go`: Route definitions
- `middleware.go`: Authentication, CORS, audit
- `websocket.go`: WebSocket streaming
- `metrics.go`: Prometheus metrics

### 2. Service Layer (`internal/service`)

**Responsibilities:**
- Business logic implementation
- Data persistence
- Session lifecycle management
- Resource allocation
- Error handling

**Key Operations:**
- Session creation and management
- Program upload and loading
- Snapshot creation and restoration
- Coverage data collection
- Job queue management

### 3. Adapter Layer (`internal/adapter`)

**Responsibilities:**
- Backend abstraction
- Process management
- Protocol communication
- Capability discovery

**Adapters:**
- **QEMU Adapter**: Manages QEMU processes, QMP protocol
- **Renode Adapter**: Manages Renode processes, Monitor protocol

**Interface:**
```go
type Adapter interface {
    GetCapabilities() (*model.Capability, error)
    StartSession(ctx, *Session, *BoardConfig) error
    StopSession(ctx, sessionID) error
    ResetSession(ctx, sessionID) error
    LoadProgram(ctx, sessionID, programPath) error
    ExecuteProgram(ctx, sessionID) error
    PauseExecution(ctx, sessionID) error
    CreateSnapshot(ctx, sessionID, path) error
    RestoreSnapshot(ctx, sessionID, path) error
}
```

### 4. Model Layer (`internal/model`)

**Data Models:**
- `Session`: Simulation session state
- `BoardConfig`: Hardware configuration
- `Program`: Uploaded program metadata
- `Snapshot`: Snapshot information
- `Job`: Async job tracking
- `AuditLog`: Operation audit trail
- `Capability`: Backend capabilities

### 5. Package Layer (`pkg`)

**Coverage (`pkg/coverage`):**
- Code coverage collection
- LCOV format export
- HTML report generation
- Line, function, and branch coverage

**GDB (`pkg/gdb`):**
- GDB Remote Serial Protocol server
- Debug session management
- Breakpoint handling

**Queue (`pkg/queue`):**
- Redis-based job queue
- Worker management
- Job status tracking

## Data Flow

### Session Creation Flow

```
Client → API Handler → Service Layer → Adapter → Backend
   ↓         ↓             ↓              ↓          ↓
Request   Validate     Allocate      Start      Launch
   ↓      Session      Resources    Session     Process
   ↓         ↓             ↓              ↓          ↓
Response ← Handler ← Save to DB ← Get Ports ← Return PID
```

### Program Execution Flow

```
1. Upload Program
   Client → API → Service → Storage
   
2. Load into Session
   API → Service → Adapter → Backend
   
3. Execute
   API → Service → Adapter → Backend (GDB/Monitor)
   
4. Stream Output
   Backend → Adapter → WebSocket Hub → Client
```

## Security Architecture

### Authentication Flow

```
Client Request → Auth Middleware
                      ↓
               Check Auth Header
                      ↓
          ┌───────────┴───────────┐
          ↓                       ↓
    API Key Validation      JWT Validation
          ↓                       ↓
    Validate against          Verify Signature
    configured keys           and expiration
          ↓                       ↓
          └───────────┬───────────┘
                      ↓
               Set User Context
                      ↓
               Next Handler
```

### Audit Trail

All API operations are logged with:
- User ID
- Action type
- Resource affected
- Timestamp
- IP address
- Operation result

## Monitoring & Observability

### Prometheus Metrics

- `http_requests_total`: Total HTTP requests
- `http_request_duration_seconds`: Request duration histogram
- `simulation_active_sessions`: Active session count
- `simulation_programs_uploaded_total`: Programs uploaded
- `simulation_jobs_queued`: Jobs in queue by type

### Health Checks

- `/health`: Service health status
- Database connectivity
- Redis connectivity (if enabled)

## Deployment Architecture

### Docker Deployment

```
┌─────────────────────────────────────────────────┐
│  Docker Compose Stack                           │
│  ┌───────────────┐  ┌──────────────────────┐   │
│  │ MachineServer │  │     Redis            │   │
│  │ Container     │──│     Container        │   │
│  └───────────────┘  └──────────────────────┘   │
│          │                                      │
│  ┌───────────────────────────────────────┐     │
│  │     Prometheus Container              │     │
│  └───────────────────────────────────────┘     │
└─────────────────────────────────────────────────┘
```

### Kubernetes Deployment

```
┌─────────────────────────────────────────────────┐
│  Kubernetes Cluster                             │
│  ┌───────────────────────────────────────┐     │
│  │ MachineServer Deployment (3 replicas) │     │
│  │  ┌──────┐  ┌──────┐  ┌──────┐        │     │
│  │  │ Pod  │  │ Pod  │  │ Pod  │        │     │
│  │  └──────┘  └──────┘  └──────┘        │     │
│  └────────────────┬──────────────────────┘     │
│                   │                            │
│  ┌────────────────┴──────────────────────┐     │
│  │        LoadBalancer Service           │     │
│  └────────────────┬──────────────────────┘     │
│                   │                            │
│  ┌────────────────┴──────────┬───────────┐     │
│  │  Redis Deployment         │  PVC      │     │
│  └───────────────────────────┴───────────┘     │
└─────────────────────────────────────────────────┘
```

## Scalability Considerations

### Horizontal Scaling
- Stateless API design allows multiple replicas
- Session affinity not required
- Database handles concurrency

### Resource Management
- Configurable session limits
- Memory and CPU quotas per session
- Automatic timeout for idle sessions

### Performance Optimization
- Connection pooling for database
- Redis for distributed caching
- Async job processing for long-running tasks

## Extension Points

### Adding New Backends
1. Implement `Adapter` interface
2. Add backend configuration
3. Register adapter in service initialization

### Custom Peripherals
1. Define peripheral configuration in BoardConfig
2. Implement adapter-specific peripheral setup
3. Document peripheral capabilities

### Additional Metrics
1. Define new Prometheus metrics
2. Update middleware or service layer
3. Configure Prometheus scraping

## Configuration

Key configuration areas:
- Server settings (host, port, mode)
- Database configuration
- Redis configuration
- Authentication settings
- Backend binaries and options
- Resource limits
- Monitoring settings
- Storage paths
