import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = "http://localhost:18081";

// --------------- scenario stages by TEST_TYPE ---------------
const stagesByTestType = {
  load: [
    { duration: "30s", target: 20 },
    { duration: "1m", target: 20 },
    { duration: "10s", target: 0 },
  ],
  stress: [
    { duration: "30s", target: 50 },
    { duration: "30s", target: 100 },
    { duration: "1m", target: 100 },
    { duration: "10s", target: 0 },
  ],
  spike: [
    { duration: "30s", target: 5 },
    { duration: "5s", target: 100 },
    { duration: "15s", target: 100 },
    { duration: "5s", target: 5 },
    { duration: "30s", target: 5 },
  ],
  volume: [
    { duration: "2m", target: 10 },
  ],
};

const testType = __ENV.TEST_TYPE || "load";
export const options = {
  stages: stagesByTestType[testType],
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
  },
};

// --------------- request headers ---------------
// Volume test: all VUs share one patient so documents pile up and the list
// response grows over time. Other tests: each VU gets its own patient to
// isolate concurrency as the only variable.
const patientId = testType === "volume" ? "patient-1" : `patient-${__VU}`;

function buildPatientHeaders() {
  return {
    "Content-Type": "application/json",
    "X-Role": "PATIENT",
    "X-Patient-Id": patientId,
  };
}

const staffHeaders = {
  "Content-Type": "application/json",
  "X-Role": "STAFF",
  "X-Hospital-Id": "hospital-1",
};

// --------------- main workflow ---------------
export default function () {
  // 1. Patient creates a document
  const createResponse = http.post(
    `${BASE_URL}/documents`,
    JSON.stringify({
      hospitalId: "hospital-1",
      title: "Blood Test Results",
      category: "LAB_RESULTS",
      notes: "Routine checkup",
    }),
    { headers: buildPatientHeaders(), tags: { endpoint: "POST /documents" } }
  );
  check(createResponse, {
    "create: status 201": (response) => response.status === 201,
  });

  const document = createResponse.json();
  const documentId = document.documentId;

  // 2. Patient lists their documents
  const listResponse = http.get(
    `${BASE_URL}/patients/${patientId}/documents`,
    {
      headers: buildPatientHeaders(),
      tags: { endpoint: "GET /patients/:id/documents" },
    }
  );
  check(listResponse, {
    "list: status 200": (response) => response.status === 200,
  });

  // 3. Patient gets the document by ID
  const getResponse = http.get(`${BASE_URL}/documents/${documentId}`, {
    headers: buildPatientHeaders(),
    tags: { endpoint: "GET /documents/:id" },
  });
  check(getResponse, {
    "get: status 200": (response) => response.status === 200,
  });

  // 4. Staff reviews the document
  const reviewResponse = http.patch(
    `${BASE_URL}/documents/${documentId}/review`,
    JSON.stringify({ status: "APPROVED", reviewNote: "Looks good" }),
    { headers: staffHeaders, tags: { endpoint: "PATCH /documents/:id/review" } }
  );
  check(reviewResponse, {
    "review: status 200": (response) => response.status === 200,
  });

  // 5. Staff links document to a medical record
  const linkResponse = http.post(
    `${BASE_URL}/documents/${documentId}/links/medical-records/record-${__VU}-${__ITER}`,
    null,
    {
      headers: staffHeaders,
      tags: { endpoint: "POST /documents/:id/links/.../medical-records/:id" },
    }
  );
  check(linkResponse, {
    "link: status 200": (response) => response.status === 200,
  });

  sleep(testType === "volume" ? 0.1 : 0.5);
}
