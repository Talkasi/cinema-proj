import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend } from 'k6/metrics';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// ===================================================================================
// 1. НАСТРОЙКА ТЕСТА
// ===================================================================================

// Создаем кастомные метрики (Trend) для измерения времени ответа каждого важного эндпоинта.
// Это позволит вам детально анализировать производительность разных частей вашего API.
const trends = {
  // Аутентификация
  register: new Trend('T_01_Register'),
  login: new Trend('T_02_Login'),
  // Чтение (легкие и средние запросы)
  getMoviesList: new Trend('T_03_GetMoviesList'),
  getMovieDetails: new Trend('T_04_GetMovieDetails'),
  getMovieShows: new Trend('T_05_GetMovieShows'),
  // Запись (тяжелые запросы)
  postReview: new Trend('T_06_PostReview'),
};

export const options = {
  // Сценарий нагрузки, имитирующий поиск "точки деградации" и работу под нагрузкой.
  // VU - Virtual User (виртуальный пользователь)
  stages: [
    { duration: '1m', target: 50 },   // 1. Плавный рост до 50 VU в течение 1 минуты
    { duration: '3m', target: 50 },   // 2. Работа на стабильной нагрузке (50 VU) в течение 3 минут
    { duration: '1m', target: 150 },  // 3. Плавный рост до 150 VU (ищем точку деградации)
    { duration: '3m', target: 150 },  // 4. Работа на пиковой нагрузке (150 VU)
    { duration: '1m', target: 0 },    // 5. Плавное снижение нагрузки до 0 для проверки восстановления
  ],
  // Пороги (thresholds) - это критерии успеха/провала теста.
  // Тест будет считаться проваленным, если хотя бы одно из условий не выполнится.
  thresholds: {
    'http_req_failed': ['rate<0.05'], // Доля ошибок должна быть меньше 5%
    'http_req_duration': ['p(95)<1500'], // 95% запросов должны выполняться быстрее 1.5 секунд
    // Персональные пороги для разных эндпоинтов
    'T_03_GetMoviesList_p(95)': ['p(95)<1000'], // Список фильмов < 1с
    'T_04_GetMovieDetails_p(95)': ['p(95)<500'],  // Детали фильма < 0.5с
    'T_06_PostReview_p(95)': ['p(95)<2000'],    // Отправка отзыва < 2с
  },
};

const BASE_URL = 'http://app:8080/api/v1'; // URL вашего API внутри Docker-сети
const USER_PASSWORD = 'superSecurePassword123';

// ===================================================================================
// 2. ФУНКЦИЯ SETUP - выполняется один раз перед началом тестов
// ===================================================================================

// В `setup` мы подготовим данные, которые будут использоваться во всем тесте.
// Например, создадим одного "тестового" пользователя, чтобы под его токеном выполнять
// действия, требующие авторизации (например, оставлять отзывы).
export function setup() {
  console.log('Подготовка тестовых данных...');

  const registerPayload = JSON.stringify({
    name: 'Benchmark Admin User',
    email: `admin_user_${Date.now()}@example.com`,
    password_hash: USER_PASSWORD,
    birth_date: "1990-01-01",
  });

  const headers = { 'Content-Type': 'application/json' };
  
  // Регистрируем админ-пользователя
  const regRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers });
  check(regRes, { 'Admin user registered successfully': (r) => r.status === 201 });
  
  // Логинимся и получаем токен
  const loginPayload = JSON.stringify({
    email: JSON.parse(registerPayload).email,
    password_hash: USER_PASSWORD,
  });

  const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });
  check(loginRes, { 'Admin user logged in successfully': (r) => r.status === 200 });

  const token = loginRes.json('token');
  if (!token) {
    throw new Error('Не удалось получить токен аутентификации в setup');
  }

  console.log('Подготовка завершена. Получен токен для тестов.');
  return { authToken: token }; // Передаем токен в основную функцию теста
}


// ===================================================================================
// 3. ОСНОВНАЯ ФУНКЦИЯ ТЕСТА - выполняется в цикле каждым VU
// ===================================================================================
export default function (data) {
  // `data` содержит результат из `setup`, то есть наш токен.
  const authToken = data.authToken;
  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  //----------------------------------------------------------------
  // СЦЕНАРИЙ 1: Поведение неавторизованного пользователя (самый частый случай)
  // Пользователь просматривает список фильмов, выбирает один и смотрит его детали.
  //----------------------------------------------------------------
  group('Unauthenticated User Flow', function () {
    // 1. Получаем первую страницу списка фильмов (средняя нагрузка)
    const getMoviesRes = http.get(`${BASE_URL}/movies?limit=10`);
    
    check(getMoviesRes, { 'GET /movies status is 200': (r) => r.status === 200 });
    trends.getMoviesList.add(getMoviesRes.timings.duration);

    const movies = getMoviesRes.json('data');
    
    // Если список фильмов пришел и он не пустой, выполняем следующие шаги
    if (movies && movies.length > 0) {
      // Выбираем случайный фильм из списка
      const randomMovie = randomItem(movies);
      
      // 2. Получаем детали этого фильма (легкая нагрузка)
      const getMovieDetailRes = http.get(`${BASE_URL}/movies/${randomMovie.id}`);
      check(getMovieDetailRes, { 'GET /movies/{id} status is 200': (r) => r.status === 200 });
      trends.getMovieDetails.add(getMovieDetailRes.timings.duration);

      // 3. Получаем расписание сеансов для этого фильма (средняя нагрузка с фильтром)
      const getShowsRes = http.get(`${BASE_URL}/movie-shows?movie_id=${randomMovie.id}`);
      check(getShowsRes, { 'GET /movie-shows status is 200': (r) => r.status === 200 });
      trends.getMovieShows.add(getShowsRes.timings.duration);

    } else {
        console.log("Не удалось получить список фильмов, пропускаем шаги.");
    }

    sleep(2); // Пользователь "думает" 2 секунды
  });


  //----------------------------------------------------------------
  // СЦЕНАРИЙ 2: Поведение авторизованного пользователя
  // Пользователь оставляет отзыв к случайному фильму.
  //----------------------------------------------------------------
  group('Authenticated User Flow', function () {
    // Снова получаем список фильмов, чтобы взять ID для отзыва
    const getMoviesRes = http.get(`${BASE_URL}/movies?limit=5`);
    const movies = getMoviesRes.json('data');

    if (movies && movies.length > 0) {
      const movieToReview = randomItem(movies);
      const reviewPayload = JSON.stringify({
          movie_id: movieToReview.id,
          user_id: "550e8400-e29b-41d4-a716-446655440001", // В реальном тесте ID надо получать динамически
          rating: Math.floor(Math.random() * 10) + 1, // Случайный рейтинг от 1 до 10
          comment: `Отличный фильм! Тест VU: ${__VU}, Итерация: ${__ITER}`
      });
      
      // 4. Отправляем отзыв (тяжелая нагрузка, запись в БД)
      const postReviewRes = http.post(`${BASE_URL}/movies/${movieToReview.id}/reviews`, reviewPayload, { headers: authHeaders });
      check(postReviewRes, { 'POST /reviews status is 201': (r) => r.status === 201 });
      trends.postReview.add(postReviewRes.timings.duration);
    }
    sleep(3); // Пауза после написания отзыва
  });


  //----------------------------------------------------------------
  // СЦЕНАРИЙ 3: Массовая регистрация и логин
  // Имитирует приток новых пользователей.
  //----------------------------------------------------------------
  group('Authentication Flow', function () {
    const userEmail = `user_${__VU}_${__ITER}@test.com`; // Уникальный email для каждого VU и итерации
    const registerPayload = JSON.stringify({
        name: `User ${__VU} ${__ITER}`,
        email: userEmail,
        password_hash: USER_PASSWORD,
        birth_date: "1995-05-15",
    });

    // 5. Регистрируем нового пользователя
    const registerRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers: { 'Content-Type': 'application/json' } });
    check(registerRes, { 'POST /register status is 201': (r) => r.status === 201 });
    trends.register.add(registerRes.timings.duration);

    sleep(1);

    // 6. Логинимся под этим новым пользователем
    const loginPayload = JSON.stringify({
        email: userEmail,
        password_hash: USER_PASSWORD
    });

    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers: { 'Content-Type': 'application/json' } });
    check(loginRes, { 'POST /login status is 200': (r) => r.status === 200 });
    trends.login.add(loginRes.timings.duration);
  });
}