import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:8000';

// ─── Scenario Definitions ────────────────────────────────────────────────────
export const options = {
  scenarios: {
    // 1. Single User Load Test: 100 req/s for 1 minute
    single_user_load: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 50,
      maxVUs: 200,
      tags: { scenario: 'single_user_load' },
    },

    // 2. Batch Endpoint Stress Test: concurrent requests with varying page sizes
    batch_stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 0 },
      ],
      tags: { scenario: 'batch_stress' },
    },

    // 3. Cache Effectiveness Test: repeated identical requests to measure hit ratio
    cache_effectiveness: {
      executor: 'constant-vus',
      vus: 20,
      duration: '1m',
      tags: { scenario: 'cache_effectiveness' },
    },
  },

  thresholds: {
    // Global thresholds
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],

    // Per-scenario thresholds
    'http_req_duration{scenario:single_user_load}': ['p(95)<500'],
    'http_req_duration{scenario:batch_stress}': ['p(95)<800'],
    'http_req_duration{scenario:cache_effectiveness}': ['p(95)<300'],
  },
};

// ─── Helpers ─────────────────────────────────────────────────────────────────
function randomUserId() {
  return Math.floor(Math.random() * 20) + 1;
}

// ─── Default Function (routes per active scenario) ───────────────────────────
export default function () {
  const scenario = __ENV.K6_SCENARIO_NAME || '';

  if (scenario === 'batch_stress') {
    runBatchStress();
  } else if (scenario === 'cache_effectiveness') {
    runCacheEffectiveness();
  } else {
    runSingleUserLoad();
  }
}

// ─── Scenario Implementations ────────────────────────────────────────────────

/** 1. Single User Load Test */
function runSingleUserLoad() {
  const userId = randomUserId();
  const res = http.get(`${BASE_URL}/users/${userId}/recommendations?limit=10`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has recommendations': (r) => {
      const body = JSON.parse(r.body);
      return Array.isArray(body.recommendations) && body.recommendations.length > 0;
    },
  });

  sleep(0.1);
}

/** 2. Batch Endpoint Stress Test — varying page sizes */
function runBatchStress() {
  const pageSizes = [5, 10, 20, 50];
  const limit = pageSizes[Math.floor(Math.random() * pageSizes.length)];
  const page = Math.floor(Math.random() * 5) + 1;

  const res = http.get(`${BASE_URL}/recommendations/batch?page=${page}&limit=${limit}`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'body is not empty': (r) => r.body.length > 0,
  });

  sleep(0.1);
}

/** 3. Cache Effectiveness Test — fixed user to maximise cache hits */
function runCacheEffectiveness() {
  // Fixed userId so the same cache key is hit repeatedly
  const userId = 1;
  const res = http.get(`${BASE_URL}/users/${userId}/recommendations?limit=10`);

  const cacheHeader = res.headers['X-Cache'] || res.headers['x-cache'] || '';

  check(res, {
    'status is 200': (r) => r.status === 200,
    'cache header set': () => cacheHeader !== '',
  });

  // No sleep — maximise repeated requests to stress cache layer
}