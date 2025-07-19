# REST API Reference

This document provides comprehensive REST API documentation for the O-RAN Near-RT RIC platform.

## Base URLs

- **Main Dashboard API**: `http://localhost:8080/api/v1`
- **xApp Dashboard API**: `http://localhost:4200/api/v1`
- **Federated Learning API**: `http://localhost:8080/fl`

## Authentication

### Authentication Endpoints

Based on the authentication handlers in `dashboard-master/dashboard-master/src/app/backend/auth/handler.go`:

#### POST /api/v1/login
Authenticate user with credentials.

**Request Body:**
```json
{
  "username": "string",     // Optional: Username for basic auth
  "password": "string",     // Optional: Password for basic auth  
  "token": "string",        // Optional: Bearer token
  "kubeconfig": "string"    // Optional: Kubeconfig file content
}
```

**Response:** `200 OK`
```json
{
  "name": "string",         // User/subject name if available
  "jweToken": "string",     // Generated JWT token
  "errors": []              // Non-critical errors array
}
```

#### POST /api/v1/token/refresh  
Refresh JWT token.

**Request Body:**
```json
{
  "jweToken": "string"      // Current valid token
}
```

**Response:** `200 OK`
```json
{
  "jweToken": "string",     // New refreshed token
  "errors": []              // Non-critical errors array
}
```

#### GET /api/v1/loginmodes
Get available authentication modes.

**Response:** `200 OK`
```json
{
  "modes": ["token", "basic"]  // Available authentication modes
}
```

#### GET /api/v1/login/status
Get current login status.

**Response:** `200 OK`
```json
{
  "tokenPresent": true,     // Whether valid token is present
  "headerPresent": true,    // Whether auth header is present
  "httpsMode": false        // Whether HTTPS is enabled
}
```

#### GET /api/v1/login/skippable
Check if authentication can be skipped.

**Response:** `200 OK`
```json
{
  "skippable": false        // Whether skip button should be shown
}
```

### Authentication Headers

- **Authorization**: `Bearer <jweToken>` - Required for authenticated requests
- **Content-Type**: `application/json` - Required for all requests with body

## Resource Management APIs

### Pods API

TODO: Extract pod management endpoints from resource handlers.

### Services API  

TODO: Extract service management endpoints.

### Deployments API

TODO: Extract deployment management endpoints.

## Federated Learning APIs

Based on the federated learning implementation in `dashboard-master/dashboard-master/src/app/backend/federatedlearning/api.go`:

### Client Registration

#### POST /fl/clients
Register a new federated learning client (xApp).

**Request Body:**
```json
{
  "xapp_name": "string",           // Required: xApp name
  "xapp_version": "string",        // Optional: xApp version
  "namespace": "string",           // Required: Kubernetes namespace
  "endpoint": "string",            // Required: xApp communication endpoint
  "rrm_tasks": [                   // Required: RRM task capabilities
    "resource_allocation",
    "power_control",
    "spectrum_management",
    "network_slicing",
    "traffic_prediction",
    "qos_optimization",
    "load_balancing",
    "handover_optimization",
    "mobility_prediction",
    "cell_selection",
    "anomaly_detection",
    "interference_management"
  ],
  "model_formats": [               // Supported ML frameworks
    "tensorflow",
    "pytorch", 
    "onnx",
    "sklearn",
    "custom"
  ],
  "compute_resources": {
    "cpu_cores": 4,
    "memory_gb": 8.0,
    "gpu_count": 1,
    "gpu_memory_gb": 8.0,
    "storage_gb": 100.0,
    "network_bandwidth_mbps": 1000,
    "flops_capacity": 1000000,
    "latency_profile": {
      "compute_latency_ms": 50.0,
      "communication_latency_ms": 10.0,
      "e2_latency_ms": 5.0,
      "target_latency_ms": 100.0
    }
  },
  "network_slices": [
    {
      "slice_id": "slice-001",
      "slice_type": "eMBB",
      "priority": 1,
      "bandwidth_mbps": 100,
      "latency_ms": 10.0,
      "reliability": 0.999
    }
  ],
  "metadata": {
    "vendor": "string",
    "version": "string"
  }
}
```

**Response:** `201 Created`
```json
{
  "client_id": "uuid-string",      // Generated unique client ID
  "status": "registered",          // Client status: registered, training, idle, disconnected, failed
  "trust_score": 0.8,             // Initial trust score (0.0-1.0)
  "certificate": "base64-cert",    // Client certificate for secure communication
  "endpoint": "/fl/clients/uuid-string"  // API endpoint for this client
}
```

#### GET /fl/clients
List registered federated learning clients.

**Query Parameters:**
- `rrm_task` (string): Filter by RRM task type
- `status` (string): Filter by client status  
- `trust_score_min` (number): Minimum trust score
- `network_slice` (string): Filter by network slice

#### GET /fl/clients/{client-id}
Get details of a specific client.

#### PUT /fl/clients/{client-id}  
Update client information.

#### DELETE /fl/clients/{client-id}
Unregister a client.

#### POST /fl/clients/{client-id}/heartbeat
Client heartbeat to maintain connection.

### Model Management

#### POST /fl/models
Create a new global federated learning model.

#### GET /fl/models
List global federated learning models.

#### GET /fl/models/{model-id}
Get details of a specific global model.

#### GET /fl/models/{model-id}/download
Download global model parameters for training.

### Training Coordination

#### POST /fl/training
Start a new federated learning training job.

#### GET /fl/training
List federated learning training jobs.

#### GET /fl/training/{job-id}
Get details of a specific training job.

#### POST /fl/training/{job-id}/stop
Stop a running training job.

#### POST /fl/training/{job-id}/updates
Submit local model update from xApp client.

#### GET /fl/training/{job-id}/round
Get current training round information.

## Metrics APIs

### Performance Metrics

TODO: Extract metrics endpoints from handler/metrics.go

### System Health

TODO: Extract health check endpoints.

### Custom Metrics

TODO: Document custom metrics endpoints.

## WebSocket APIs

### Real-time Updates

TODO: Extract WebSocket endpoints for real-time updates.

### Terminal Sessions

TODO: Document terminal session WebSocket endpoints.

### Log Streaming

TODO: Extract log streaming WebSocket endpoints.

## Error Handling

### HTTP Status Codes

The API uses standard HTTP status codes:

- `200 OK` - Successful request
- `201 Created` - Resource created successfully  
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service unavailable

### Error Response Format

TODO: Extract actual error response format from errors package.

### Common Error Scenarios

TODO: Document common error scenarios.

## Rate Limiting

### Request Limits

TODO: Document rate limiting based on actual implementation.

### Authentication Requirements

TODO: Document authentication requirements.

### Usage Guidelines

TODO: Add API usage guidelines.

---

**Status**: This is a documentation stub. Content will be populated based on actual API implementations found in the repository, particularly from the federated learning API handlers.