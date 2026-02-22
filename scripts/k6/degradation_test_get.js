import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Кастомные метрики
const responseTime = new Trend('get_movies_response_time');
const requestCount = new Counter('total_get_requests');
const errorCount = new Counter('total_get_errors');
const successRate = new Rate('get_success_rate');

const BASE_URL = 'http://app:8080/api/v1';

export const options = {
  stages: [
    { duration: '3m', target: 1000 },
    { duration: '3m', target: 2500 },
    { duration: '3m', target: 5000 },
    { duration: '3m', target: 7500 },
    { duration: '3m', target: 10000 }, 
  ],
  // thresholds: {
  //   'http_req_failed': ['rate<0.005'],
  //   'http_req_duration': ['p(95)<500'],
  // },
};

export default function () {
  // GET запрос: получение списка фильмов (не требует внешних ID)
  const randomPage = Math.floor(Math.random() * 8) + 1; // Generates a random number between 1 and 80
  const getMoviesRes = http.get(`${BASE_URL}/movies?limit=10&page=${randomPage}`);
  requestCount.add(1);
  
  const isSuccess = check(getMoviesRes, { 
    'GET /movies status is 200': (r) => r.status === 200,
  });
  
  if (!isSuccess) {
    errorCount.add(1);
    console.log(`GET request failed with status: ${getMoviesRes.status}, body: ${getMoviesRes.body}`);
  }
  successRate.add(isSuccess);
  
  responseTime.add(getMoviesRes.timings.duration);

  return 
  // sleep(Math.random() * 2 + 1); // Небольшая задержка между запросами
}