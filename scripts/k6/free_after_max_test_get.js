import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Kastomnye metriki
const trends = {
  getMoviesList: new Trend('T_01_GetMoviesList'),
  getMovieDetails: new Trend('T_02_GetMovieDetails'),
  getMovieShows: new Trend('T_03_GetMovieShows'),
  postReview: new Trend('T_04_PostReview'),
  register: new Trend('T_05_Register'),
  login: new Trend('T_06_Login'),
};

// Dopolnitelnye metriki dlya monitoringa stabilnosti
const responseTime = new Trend('get_movies_response_time');
const requestCount = new Counter('total_get_requests');
const errorCount = new Counter('total_get_errors');
const successRate = new Rate('get_success_rate');

const BASE_URL = 'http://app:8080/api/v1';

export const options = {
  scenarios: {
    // High load phase: 2500 RPS for 3 minutes
    high_load_phase: {
      executor: 'constant-arrival-rate',
      rate: 2500,
      timeUnit: '1s',
      duration: '3m',
      preAllocatedVUs: 2000, // Conservative estimate for 2500 RPS with potential slower response times
    },
    // Normal load phase: 800 RPS after the high load phase
    normal_load_phase: {
      executor: 'constant-arrival-rate',
      rate: 800,
      timeUnit: '1s',
      duration: '3m',
      preAllocatedVUs: 1000, // Sufficient for 800 RPS
      startTime: '3m', // Start after the high load phase completes
    },
  },
  discardResponseBodies: false,
};

export default function () {
  // Load test scenario comment.
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
  // Load test scenario comment.
}