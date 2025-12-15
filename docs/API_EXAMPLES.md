# API Examples

## 1. Health Check

```bash
curl http://localhost:8080/health
```

## 2. Get Backend Capabilities

```bash
curl -H "Authorization: ApiKey dev-key-12345" \
  http://localhost:8080/api/v1/capabilities
```

## 3. Create a Session

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: ApiKey dev-key-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cortex-m3-test",
    "backend": "qemu",
    "board_config": {
      "processor": {"model": "cortex-m3", "frequency": 72000000},
      "memory": {
        "flash": {"base": 134217728, "size": 131072},
        "ram": {"base": 536870912, "size": 20480}
      }
    }
  }'
```

## 4. Upload and Load Program

```bash
# Upload program
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  -F "file=@firmware.elf" \
  -F "name=my-firmware" \
  -F "format=elf" \
  http://localhost:8080/api/v1/programs

# Load program into session
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  -H "Content-Type: application/json" \
  -d '{"program_id": "PROGRAM_ID"}' \
  http://localhost:8080/api/v1/sessions/SESSION_ID/program
```

## 5. Session Control

```bash
# Power on
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  http://localhost:8080/api/v1/sessions/SESSION_ID/poweron

# Power off
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  http://localhost:8080/api/v1/sessions/SESSION_ID/poweroff

# Reset
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  http://localhost:8080/api/v1/sessions/SESSION_ID/reset
```

## 6. Snapshot Management

```bash
# Create snapshot
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  -H "Content-Type: application/json" \
  -d '{"name": "before-test", "description": "Test checkpoint"}' \
  http://localhost:8080/api/v1/sessions/SESSION_ID/snapshots

# Restore snapshot
curl -X POST \
  -H "Authorization: ApiKey dev-key-12345" \
  -H "Content-Type: application/json" \
  -d '{"snapshot_id": "SNAPSHOT_ID"}' \
  http://localhost:8080/api/v1/sessions/SESSION_ID/restore
```
