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
const requestCount = new Counter('total_requests');
const errorCount = new Counter('total_errors');
const successRate = new Rate('success_rate');

const BASE_URL = 'http://app:8080/api/v1';
const USER_PASSWORD = 'superSecurePassword123';

export function setup() {
  console.log('Setup: подготовка тестовых данных...');
  
  const testEmail = `admin_user_${Date.now()}@example.com`;
  const registerPayload = JSON.stringify({
    name: 'Benchmark Admin User',
    email: testEmail,
    password: USER_PASSWORD,
    birth_date: "1990-01-01",
  });

  const headers = { 'Content-Type': 'application/json' };
  
  // Регистрация
  const regRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers });
  console.log('Register status:', regRes.status, 'Body:', regRes.body);
  
  check(regRes, { 
    'Admin user registered': (r) => r.status === 201 || r.status === 200 
  });
  
  // Логин
  const loginPayload = JSON.stringify({
    email: testEmail,
    password: USER_PASSWORD
  });

  const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });
  console.log('Login status:', loginRes.status, 'Body:', loginRes.body);
  
  check(loginRes, { 
    'Admin user logged in': (r) => r.status === 200 
  });

  let token = null;
  try {
    const loginData = JSON.parse(loginRes.body);
    token = loginData.Token || loginData.token || loginData.access_token || loginData.accessToken;
    console.log('Token found:', token ? 'YES' : 'NO');
  } catch (e) {
    console.log('Error parsing login response:', e.message);
  }

  return { 
    authToken: token,
    testEmail: testEmail,
    userPassword: USER_PASSWORD
  };
}

export const options = {
  scenarios: {
    high_load: {
      executor: 'constant-arrival-rate',
      rate: 350, // 10500 запросов / 30 секунд = 350 RPS
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 500,
      maxVUs: 2000,
    },
  },
  thresholds: {
    'http_req_failed': ['rate<0.005'], // Допускаем до 2% ошибок при пиковой нагрузке
    'http_req_duration': ['p(95)<500'],
  },
  discardResponseBodies: false,
};

export default function (data) {
  const authToken = data.authToken;
  const authHeaders = authToken ? {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  } : { 'Content-Type': 'application/json' };

  // Оптимизированное распределение операций для высокой нагрузки
  const operationType = Math.random();
  let success = false;

  if (operationType < 0.8) {
    // 80% - Самый легкий сценарий (только список фильмов)
    group('Light - Movies List Only', function () {
      const getMoviesRes = http.get(`${BASE_URL}/movies?limit=5`);
      requestCount.add(1);
      
      const checkResult = check(getMoviesRes, { 
        'GET /movies status is 200': (r) => r.status === 200
      });
      
      if (checkResult) {
        success = true;
        trends.getMoviesList.add(getMoviesRes.timings.duration);
      } else {
        errorCount.add(1);
      }
    });
    
  } else if (operationType < 0.95 && authToken) {
    // 15% - Средний сценарий (список + детали фильма)
    group('Medium - Movies List + Details', function () {
      const getMoviesRes = http.get(`${BASE_URL}/movies?limit=3`);
      requestCount.add(1);
      
      if (getMoviesRes.status === 200) {
        const movies = getMoviesRes.json('data');
        if (movies && movies.length > 0) {
          const randomMovie = randomItem(movies);
          
          // Детали фильма
          const getMovieDetailRes = http.get(`${BASE_URL}/movies/${randomMovie.id}`);
          requestCount.add(1);
          
          const checkResult = check(getMovieDetailRes, { 
            'GET /movies status is 200': (r) => r.status === 200,
            'GET /movies/{id} status is 200': (r) => r.status === 200
          });
          
          if (checkResult) {
            success = true;
            trends.getMoviesList.add(getMoviesRes.timings.duration);
            trends.getMovieDetails.add(getMovieDetailRes.timings.duration);
          } else {
            errorCount.add(1);
          }
        }
      }
    });
    
  } else {
    // 5% - Тяжелый сценарий (регистрация)
    group('Heavy - Registration Only', function () {
      const userEmail = `load_user_${__VU}_${__ITER}_${Date.now()}@test.com`;
      const registerPayload = JSON.stringify({
        name: `Load User ${__VU} ${__ITER}`,
        email: userEmail,
        password: USER_PASSWORD,
        birth_date: "1995-05-15",
      });

      const registerRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { 
        headers: { 'Content-Type': 'application/json' } 
      });
      requestCount.add(1);
      
      const checkResult = check(registerRes, { 
        'POST /register status is 201': (r) => r.status === 201 || r.status === 200 
      });
      
      if (checkResult) {
        success = true;
        trends.register.add(registerRes.timings.duration);
      } else {
        errorCount.add(1);
      }
    });
  }

  // Обновляем метрику успешности
  successRate.add(success);

  // Минимальная пауза для снижения нагрузки на систему
  sleep(0.05);
}