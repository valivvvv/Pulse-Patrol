# Pulse-Patrol

## Run it locally

```bash
go mod init pulse-patrol-document-service
go run .
```

---

## Quick test with curl

### Create document (patient)

```bash
curl -i -X POST "http://localhost:18081/documents" \
  -H "Content-Type: application/json" \
  -H "X-Role: PATIENT" \
  -H "X-Patient-Id: patient-123" \
  -d '{"hospitalId":"hospital-1","title":"Blood test results","category":"LAB_RESULTS","notes":"Uploaded from home"}'
```

### List patient documents (patient)

```bash
curl -i "http://localhost:18081/patients/patient-123/documents" \
  -H "X-Role: PATIENT" \
  -H "X-Patient-Id: patient-123"
```

### Review document (staff)

```bash
curl -i -X PATCH "http://localhost:18081/documents/doc-1/review" \
  -H "Content-Type: application/json" \
  -H "X-Role: STAFF" \
  -H "X-Hospital-Id: hospital-1" \
  -d '{"status":"APPROVED","reviewNote":"Looks valid."}'
```

If you want, next we can add one small “nice” improvement for learning: a `/health` endpoint and a tiny request logger middleware (still readable).
