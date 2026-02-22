import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Kastomnye metriki
const responseTime = new Trend('post_user_registration_response_time');
const requestCount = new Counter('total_post_requests');
const errorCount = new Counter('total_post_errors');
const successRate = new Rate('post_success_rate');

const BASE_URL = 'http://app:8080/api/v1';

export const options = {
  stages: [
    { duration: '1m', target: 1 },  
    { duration: '3m', target: 10 },    
    { duration: '3m', target: 100 },    
    { duration: '3m', target: 1000 },
    { duration: '3m', target: 5000 }, 
  ],
  thresholds: {
    'http_req_failed': ['rate<0.005'],
    'http_req_duration': ['p(95)<500'],
  },
};

export default function () {
  // POST zapros: registratsiya novogo polzovatelya (ne trebuet vneshnikh ID)
  const timestamp = Date.now();
  const testEmail = `degradation_test_user_${timestamp}_${__VU}_${__ITER}@example.com`;
  const registerPayload = JSON.stringify({
    name: `Degradation Test User ${__VU}_${__ITER}`,
    email: testEmail,
    password: 'superSecurePassword123',
    birth_date: "1990-01-01",
  });

  const headers = { 'Content-Type': 'application/json' };
  
  const postRegisterRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers });
  requestCount.add(1);
  
  const isSuccess = check(postRegisterRes, { 
    'POST /auth/register status is 201': (r) => r.status === 201 || r.status === 200
  });
  
  if (!isSuccess) {
    errorCount.add(1);
    console.log(`POST register request failed with status: ${postRegisterRes.status}, body: ${postRegisterRes.body}`);
  }
  successRate.add(isSuccess);
  
  responseTime.add(postRegisterRes.timings.duration);

  sleep(Math.random() * 2 + 1); // Nebolshaya zaderzhka mezhdu zaprosami
}