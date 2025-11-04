import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Кастомные метрики
const trends = {
  getMoviesList: new Trend('T_01_GetMoviesList'),
  getMovieDetails: new Trend('T_02_GetMovieDetails'),
  getMovieShows: new Trend('T_03_GetMovieShows'),
  postReview: new Trend('T_04_PostReview'),
  register: new Trend('T_05_Register'),
  login: new Trend('T_06_Login'),
};

// Дополнительные метрики для мониторинга стабильности
const responseTime = new Trend('get_movies_response_time');
const requestCount = new Counter('total_get_requests');
const errorCount = new Counter('total_get_errors');
const successRate = new Rate('get_success_rate');

const BASE_URL = 'http://app:8080/api/v1';

export const options = {
  // scenarios: {
  //   high_load: {
  //     executor: 'constant-arrival-rate',
  //     rate: 1900,
  //     timeUnit: '1s',
  //     duration: '3m', // Увеличиваем продолжительность для лучшего анализа
  //     preAllocatedVUs: 750, // RPS * response time per second = 1900 * 0.49
  //   },
  // },
  // discardResponseBodies: false,
  scenarios: {
    high_load: {
      executor: 'constant-arrival-rate',
      rate: 1940,
      timeUnit: '1s',
      duration: '6m', // Увеличиваем продолжительность для лучшего анализа
      preAllocatedVUs: 950, // RPS * response time per second = 1900 * 0.49
    },
  },
  discardResponseBodies: false,
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