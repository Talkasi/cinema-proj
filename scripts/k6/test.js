import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend } from 'k6/metrics';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// ===================================================================================
// 1. NASTROYKA TESTA
// ===================================================================================

// Sozdaem kastomnye metriki (Trend) dlya izmereniya vremeni otveta kazhdogo vazhnogo endpointa.
// Eto pozvolit vam detalno analizirovat proizvoditelnost raznykh chastey vashego API.
const trends = {
  // Autentifikatsiya
  register: new Trend('T_01_Register'),
  login: new Trend('T_02_Login'),
  // Chtenie (legkie i srednie zaprosy)
  getMoviesList: new Trend('T_03_GetMoviesList'),
  getMovieDetails: new Trend('T_04_GetMovieDetails'),
  getMovieShows: new Trend('T_05_GetMovieShows'),
  // Zapis (tyazhelye zaprosy)
  postReview: new Trend('T_06_PostReview'),
};

export const options = {
  // Stsenariy nagruzki, imitiruyuschiy poisk "tochki degradatsii" i rabotu pod nagruzkoy.
  // VU - Virtual User (virtualnyy polzovatel)
  stages: [
    { duration: '1m', target: 50 },   // 1. Plavnyy rost do 50 VU v techenie 1 minuty
    { duration: '3m', target: 50 },   // 2. Rabota na stabilnoy nagruzke (50 VU) v techenie 3 minut
    { duration: '1m', target: 150 },  // 3. Plavnyy rost do 150 VU (ischem tochku degradatsii)
    { duration: '3m', target: 150 },  // 4. Rabota na pikovoy nagruzke (150 VU)
    { duration: '1m', target: 0 },    // 5. Plavnoe snizhenie nagruzki do 0 dlya proverki vosstanovleniya
  ],
  // Porogi (thresholds) - eto kriterii uspekha/provala testa.
  // Test budet schitatsya provalennym, esli khotya by odno iz usloviy ne vypolnitsya.
  thresholds: {
    'http_req_failed': ['rate<0.05'], // Dolya oshibok dolzhna byt menshe 5%
    'http_req_duration': ['p(95)<1500'], // 95% zaprosov dolzhny vypolnyatsya bystree 1.5 sekund
    // Personalnye porogi dlya raznykh endpointov
    'T_03_GetMoviesList_p(95)': ['p(95)<1000'], // Spisok filmov < 1s
    'T_04_GetMovieDetails_p(95)': ['p(95)<500'],  // Detali filma < 0.5s
    'T_06_PostReview_p(95)': ['p(95)<2000'],    // Otpravka otzyva < 2s
  },
};

const BASE_URL = 'http://app:8080/api/v1'; // URL vashego API vnutri Docker-seti
const USER_PASSWORD = 'superSecurePassword123';

// ===================================================================================
// 2. FUNKTsIYa SETUP - vypolnyaetsya odin raz pered nachalom testov
// ===================================================================================

// V `setup` my podgotovim dannye, kotorye budut ispolzovatsya vo vsem teste.
// Naprimer, sozdadim odnogo "testovogo" polzovatelya, chtoby pod ego tokenom vypolnyat
// deystviya, trebuyuschie avtorizatsii (naprimer, ostavlyat otzyvy).
export function setup() {
  console.log('Podgotovka testovykh dannykh...');

  const registerPayload = JSON.stringify({
    name: 'Benchmark Admin User',
    email: `admin_user_${Date.now()}@example.com`,
    password_hash: USER_PASSWORD,
    birth_date: "1990-01-01",
  });

  const headers = { 'Content-Type': 'application/json' };
  
  // Registriruem admin-polzovatelya
  const regRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers });
  check(regRes, { 'Admin user registered successfully': (r) => r.status === 201 });
  
  // Loginimsya i poluchaem token
  const loginPayload = JSON.stringify({
    email: JSON.parse(registerPayload).email,
    password_hash: USER_PASSWORD,
  });

  const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });
  check(loginRes, { 'Admin user logged in successfully': (r) => r.status === 200 });

  const token = loginRes.json('token');
  if (!token) {
    throw new Error('Ne udalos poluchit token autentifikatsii v setup');
  }

  console.log('Podgotovka zavershena. Poluchen token dlya testov.');
  return { authToken: token }; // Peredaem token v osnovnuyu funktsiyu testa
}


// ===================================================================================
// 3. OSNOVNAYa FUNKTsIYa TESTA - vypolnyaetsya v tsikle kazhdym VU
// ===================================================================================
export default function (data) {
  // `data` soderzhit rezultat iz `setup`, to est nash token.
  const authToken = data.authToken;
  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  //----------------------------------------------------------------
  // STsENARIY 1: Povedenie neavtorizovannogo polzovatelya (samyy chastyy sluchay)
  // Polzovatel prosmatrivaet spisok filmov, vybiraet odin i smotrit ego detali.
  //----------------------------------------------------------------
  group('Unauthenticated User Flow', function () {
    // 1. Poluchaem pervuyu stranitsu spiska filmov (srednyaya nagruzka)
    const getMoviesRes = http.get(`${BASE_URL}/movies?limit=10`);
    
    check(getMoviesRes, { 'GET /movies status is 200': (r) => r.status === 200 });
    trends.getMoviesList.add(getMoviesRes.timings.duration);

    const movies = getMoviesRes.json('data');
    
    // Esli spisok filmov prishel i on ne pustoy, vypolnyaem sleduyuschie shagi
    if (movies && movies.length > 0) {
      // Vybiraem sluchaynyy film iz spiska
      const randomMovie = randomItem(movies);
      
      // 2. Poluchaem detali etogo filma (legkaya nagruzka)
      const getMovieDetailRes = http.get(`${BASE_URL}/movies/${randomMovie.id}`);
      check(getMovieDetailRes, { 'GET /movies/{id} status is 200': (r) => r.status === 200 });
      trends.getMovieDetails.add(getMovieDetailRes.timings.duration);

      // 3. Poluchaem raspisanie seansov dlya etogo filma (srednyaya nagruzka s filtrom)
      const getShowsRes = http.get(`${BASE_URL}/movie-shows?movie_id=${randomMovie.id}`);
      check(getShowsRes, { 'GET /movie-shows status is 200': (r) => r.status === 200 });
      trends.getMovieShows.add(getShowsRes.timings.duration);

    } else {
        console.log("Ne udalos poluchit spisok filmov, propuskaem shagi.");
    }

    sleep(2); // Polzovatel "dumaet" 2 sekundy
  });


  //----------------------------------------------------------------
  // STsENARIY 2: Povedenie avtorizovannogo polzovatelya
  // Polzovatel ostavlyaet otzyv k sluchaynomu filmu.
  //----------------------------------------------------------------
  group('Authenticated User Flow', function () {
    // Snova poluchaem spisok filmov, chtoby vzyat ID dlya otzyva
    const getMoviesRes = http.get(`${BASE_URL}/movies?limit=5`);
    const movies = getMoviesRes.json('data');

    if (movies && movies.length > 0) {
      const movieToReview = randomItem(movies);
      const reviewPayload = JSON.stringify({
          movie_id: movieToReview.id,
          user_id: "550e8400-e29b-41d4-a716-446655440001", // V realnom teste ID nado poluchat dinamicheski
          rating: Math.floor(Math.random() * 10) + 1, // Sluchaynyy reyting ot 1 do 10
          comment: `Otlichnyy film! Test VU: ${__VU}, Iteratsiya: ${__ITER}`
      });
      
      // 4. Otpravlyaem otzyv (tyazhelaya nagruzka, zapis v BD)
      const postReviewRes = http.post(`${BASE_URL}/movies/${movieToReview.id}/reviews`, reviewPayload, { headers: authHeaders });
      check(postReviewRes, { 'POST /reviews status is 201': (r) => r.status === 201 });
      trends.postReview.add(postReviewRes.timings.duration);
    }
    sleep(3); // Pauza posle napisaniya otzyva
  });


  //----------------------------------------------------------------
  // STsENARIY 3: Massovaya registratsiya i login
  // Imitiruet pritok novykh polzovateley.
  //----------------------------------------------------------------
  group('Authentication Flow', function () {
    const userEmail = `user_${__VU}_${__ITER}@test.com`; // Unikalnyy email dlya kazhdogo VU i iteratsii
    const registerPayload = JSON.stringify({
        name: `User ${__VU} ${__ITER}`,
        email: userEmail,
        password_hash: USER_PASSWORD,
        birth_date: "1995-05-15",
    });

    // 5. Registriruem novogo polzovatelya
    const registerRes = http.post(`${BASE_URL}/auth/register`, registerPayload, { headers: { 'Content-Type': 'application/json' } });
    check(registerRes, { 'POST /register status is 201': (r) => r.status === 201 });
    trends.register.add(registerRes.timings.duration);

    sleep(1);

    // 6. Loginimsya pod etim novym polzovatelem
    const loginPayload = JSON.stringify({
        email: userEmail,
        password_hash: USER_PASSWORD
    });

    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers: { 'Content-Type': 'application/json' } });
    check(loginRes, { 'POST /login status is 200': (r) => r.status === 200 });
    trends.login.add(loginRes.timings.duration);
  });
}