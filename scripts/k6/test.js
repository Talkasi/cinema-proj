import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend } from 'k6/metrics';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// ===================================================================================
// 1. NASTROYKA TESTA
// ===================================================================================

// Load test scenario comment.
// Scenario note.
const trends = {
  // Scenario note.
  register: new Trend('T_01_Register'),
  login: new Trend('T_02_Login'),
  // Load test scenario comment.
  getMoviesList: new Trend('T_03_GetMoviesList'),
  getMovieDetails: new Trend('T_04_GetMovieDetails'),
  getMovieShows: new Trend('T_05_GetMovieShows'),
  // Load test scenario comment.
  postReview: new Trend('T_06_PostReview'),
};

export const options = {
  // Load test scenario comment.
  // Scenario note.
  stages: [
    { duration: '1m', target: 50 },   // Ramp up to 50 VUs over 1 minute
    { duration: '3m', target: 50 },   // Scenario note.
    { duration: '1m', target: 150 },  // Scenario note.
    { duration: '3m', target: 150 },  // Scenario note.
    { duration: '1m', target: 0 },    // Scenario note.
  ],
  // Load test scenario comment.
  // Load test scenario comment.
  thresholds: {
    'http_req_failed': ['rate<0.05'], // Error rate must stay below 5%
    'http_req_duration': ['p(95)<1500'], // Scenario note.
    // Load test scenario comment.
    'T_03_GetMoviesList_p(95)': ['p(95)<1000'], // Scenario note.
    'T_04_GetMovieDetails_p(95)': ['p(95)<500'],  // Scenario note.
    'T_06_PostReview_p(95)': ['p(95)<2000'],    // Scenario note.
  },
};

const BASE_URL = 'http://app:8080/api/v1'; // API URL inside the Docker network
const USER_PASSWORD = 'superSecurePassword123';

// ===================================================================================
// 2. SETUP FUNCTION - runs once before the test starts
// ===================================================================================

// Load test scenario comment.
// Load test scenario comment.
// Load test scenario comment.
export function setup() {
  console.log('Preparing test data...');

  const registerPayload = JSON.stringify({
    name: 'Benchmark Admin User',
    email: `admin_user_${Date.now()}@example.com`,
    password_hash: USER_PASSWORD,
    birth_date: "1990-01-01",
  });

  const headers = { 'Content-Type': 'application/json' };
  
  // Scenario note.
  const regRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers });
  check(regRes, { 'Admin user registered successfully': (r) => r.status === 201 });
  
  // Load test scenario comment.
  const loginPayload = JSON.stringify({
    email: JSON.parse(registerPayload).email,
    password_hash: USER_PASSWORD,
  });

  const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });
  check(loginRes, { 'Admin user logged in successfully': (r) => r.status === 200 });

  const token = loginRes.json('token');
  if (!token) {
    throw new Error('Status message');
  }

  console.log('Preparing test data...');
  return { authToken: token }; // Pass token to the main test function
}


// ===================================================================================
// 3. MAIN TEST FUNCTION - runs in a loop for each VU
// ===================================================================================
export default function (data) {
  // Scenario note.
  const authToken = data.authToken;
  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  //----------------------------------------------------------------
  // Load test scenario comment.
  // Scenario note.
  //----------------------------------------------------------------
  group('Unauthenticated User Flow', function () {
    // Load test scenario comment.
    const getMoviesRes = http.get(`${BASE_URL}/movies?limit=10`);
    
    check(getMoviesRes, { 'GET /movies status is 200': (r) => r.status === 200 });
    trends.getMoviesList.add(getMoviesRes.timings.duration);

    const movies = getMoviesRes.json('data');
    
    // Scenario note.
    if (movies && movies.length > 0) {
      // Pick a random movie from the list
      const randomMovie = randomItem(movies);
      
      // Load test scenario comment.
      const getMovieDetailRes = http.get(`${BASE_URL}/movies/${randomMovie.id}`);
      check(getMovieDetailRes, { 'GET /movies/{id} status is 200': (r) => r.status === 200 });
      trends.getMovieDetails.add(getMovieDetailRes.timings.duration);

      // Load test scenario comment.
      const getShowsRes = http.get(`${BASE_URL}/movie-shows?movie_id=${randomMovie.id}`);
      check(getShowsRes, { 'GET /movie-shows status is 200': (r) => r.status === 200 });
      trends.getMovieShows.add(getShowsRes.timings.duration);

    } else {
        console.log("Status message");
    }

    sleep(2); // Scenario note.
  });


  //----------------------------------------------------------------
  // Load test scenario comment.
  // Load test scenario comment.
  //----------------------------------------------------------------
  group('Authenticated User Flow', function () {
    // Load test scenario comment.
    const getMoviesRes = http.get(`${BASE_URL}/movies?limit=5`);
    const movies = getMoviesRes.json('data');

    if (movies && movies.length > 0) {
      const movieToReview = randomItem(movies);
      const reviewPayload = JSON.stringify({
          movie_id: movieToReview.id,
          user_id: "550e8400-e29b-41d4-a716-446655440001", // Scenario note.
          rating: Math.floor(Math.random() * 10) + 1, // Sluchaynyy reyting ot 1 do 10
          comment: `Otlichnyy film! Test VU: ${__VU}, Iteratsiya: ${__ITER}`
      });
      
      // Load test scenario comment.
      const postReviewRes = http.post(`${BASE_URL}/movies/${movieToReview.id}/reviews`, reviewPayload, { headers: authHeaders });
      check(postReviewRes, { 'POST /reviews status is 201': (r) => r.status === 201 });
      trends.postReview.add(postReviewRes.timings.duration);
    }
    sleep(3); // Scenario note.
  });


  //----------------------------------------------------------------
  // Load test scenario comment.
  // Scenario note.
  //----------------------------------------------------------------
  group('Authentication Flow', function () {
    const userEmail = `user_${__VU}_${__ITER}@test.com`; // Unikalnyy email dlya kazhdogo VU i iteratsii
    const registerPayload = JSON.stringify({
        name: `User ${__VU} ${__ITER}`,
        email: userEmail,
        password_hash: USER_PASSWORD,
        birth_date: "1995-05-15",
    });

    // Scenario note.
    const registerRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers: { 'Content-Type': 'application/json' } });
    check(registerRes, { 'POST /register status is 201': (r) => r.status === 201 });
    trends.register.add(registerRes.timings.duration);

    sleep(1);

    // Scenario note.
    const loginPayload = JSON.stringify({
        email: userEmail,
        password_hash: USER_PASSWORD
    });

    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers: { 'Content-Type': 'application/json' } });
    check(loginRes, { 'POST /login status is 200': (r) => r.status === 200 });
    trends.login.add(loginRes.timings.duration);
  });
}
