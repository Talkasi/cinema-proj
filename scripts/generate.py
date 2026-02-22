import csv
import uuid
import random
from datetime import datetime, timedelta, date
from faker import Faker
import logging
import os

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class CSVDataGenerator:
    def __init__(self, output_dir: str = "test_data"):
        self.output_dir = output_dir
        self.fake = Faker(['ru_RU', 'en_US'])
        
        os.makedirs(self.output_dir, exist_ok=True)
        
        # self.config = {
        #     'users_count': 500000,
        #     'genres_count': 20,
        #     'movies_count': 1000,
        #     'screen_types_count': 5,
        #     'halls_count': 10,
        #     'seat_types_count': 4,
        #     'seats_per_hall_min': 80,
        #     'seats_per_hall_max': 150,
        #     'movie_shows_total': 70000,
        #     'reviews_per_user_min': 1,
        #     'reviews_per_user_max': 250,
        #     'ticket_purchase_probability': 0.3
        # }

        self.config = {
            'users_count': 3500,           # база постоянных клиентов + разовые посетители
            'genres_count': 15,            # все основные жанры + несколько нишевых
            'movies_count': 80,            # фильмы за 6-12 месяцев проката
            'screen_types_count': 4,       # 2D, 3D, IMAX, 4DX/Премиум
            'halls_count': 8,              # 8 залов - кинотеатр выше среднего
            'seat_types_count': 4,         # Эконом, Стандарт, Комфорт, VIP
            'seats_per_hall_min': 70,      # минимальный зал
            'seats_per_hall_max': 180,     # максимальный зал (есть один большой)
            'movie_shows_total': 300,      # сеансов на неделю (40-50 в день)
            'reviews_per_user_min': 0,
            'reviews_per_user_max': 12,    # активные пользователи пишут больше отзывов
            'ticket_purchase_probability': 0.20  # 20% вероятность покупки - популярное место
        }
                
        self.genre_ids = []
        self.movie_ids = []
        self.user_ids = []
        self.hall_ids = []
        self.screen_type_ids = []
        self.seat_type_ids = []
        self.seats_by_hall = {}
        self.movies_by_id = {}
        self.movie_shows_by_hall = {}

    def generate_genres(self):
        """Генерация жанров с учетом CHECK constraints"""
        logger.info(f"Генерация {self.config['genres_count']} жанров...")
        
        genres_data = [
            "Action", "Adventure", "Animation", "Comedy", "Crime", "Documentary", 
            "Drama", "Fantasy", "Historical", "Horror", "Musical", "Mystery", 
            "Romance", "Science Fiction", "Thriller", "Western", "Biographical",
            "Family", "War", "Sports"
        ]
        
        with open(f'{self.output_dir}/genres.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'name', 'description'])
            
            for genre_name in genres_data[:self.config['genres_count']]:
                genre_id = str(uuid.uuid4())
                # Генерируем описание, который проходит CHECK constraint
                description = self.fake.text(max_nb_chars=300).strip()
                while not description:  # Убеждаемся, что не пустой
                    description = self.fake.text(max_nb_chars=300).strip()
                
                writer.writerow([genre_id, genre_name, description])
                self.genre_ids.append(genre_id)
        
        logger.info("Жанры созданы")

    def generate_screen_types(self):
        """Генерация типов экранов"""
        logger.info(f"Генерация {self.config['screen_types_count']} типов экранов...")
        
        screen_types = [
            {"name": "Standard", "price": 1.0},
            {"name": "IMAX", "price": 1.8},
            {"name": "4DX", "price": 2.2},
            {"name": "Dolby Cinema", "price": 1.5},
            {"name": "VIP", "price": 2.5}
        ]
        
        with open(f'{self.output_dir}/screen_types.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'name', 'description', 'price_modifier'])
            
            for screen_type in screen_types[:self.config['screen_types_count']]:
                screen_type_id = str(uuid.uuid4())
                description = self.fake.text(max_nb_chars=200).strip()
                while not description:
                    description = self.fake.text(max_nb_chars=200).strip()
                
                writer.writerow([
                    screen_type_id, 
                    screen_type["name"], 
                    description, 
                    screen_type["price"]
                ])
                self.screen_type_ids.append(screen_type_id)
        
        logger.info("Типы экранов созданы")

    def generate_seat_types(self):
        """Генерация типов мест"""
        logger.info(f"Генерация {self.config['seat_types_count']} типов мест...")
        
        seat_types = [
            {"name": "Standard", "price": 1.0},
            {"name": "Comfort", "price": 1.3},
            {"name": "VIP", "price": 1.8},
            {"name": "Handicap", "price": 1.0}
        ]
        
        with open(f'{self.output_dir}/seat_types.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'name', 'description', 'price_modifier'])
            
            for seat_type in seat_types[:self.config['seat_types_count']]:
                seat_type_id = str(uuid.uuid4())
                description = self.fake.text(max_nb_chars=200).strip()
                while not description:
                    description = self.fake.text(max_nb_chars=200).strip()
                
                writer.writerow([
                    seat_type_id, 
                    seat_type["name"], 
                    description, 
                    seat_type["price"]
                ])
                self.seat_type_ids.append(seat_type_id)
        
        logger.info("Типы мест созданы")

    def generate_users(self):
        """Генерация пользователей с учетом CHECK constraints"""
        logger.info(f"Генерация {self.config['users_count']} пользователей...")
        
        with open(f'{self.output_dir}/users.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'name', 'email', 'password_hash', 'birth_date', 'is_admin'])
            
            used_emails = set()
            
            for i in range(self.config['users_count']):
                user_id = str(uuid.uuid4())
                
                # Генерируем имя, которое проходит CHECK constraint
                name = self.fake.name().strip()
                while not name or len(name) > 50:
                    name = self.fake.name().strip()
                
                # Генерируем уникальный email
                email = self.fake.unique.email()
                while email in used_emails or len(email) > 100:
                    email = self.fake.unique.email()
                used_emails.add(email)
                
                password_hash = self.fake.sha256()
                
                # Дата рождения от 12 до 90 лет назад
                birth_date = self.fake.date_of_birth(minimum_age=12, maximum_age=90)
                
                is_admin = 'true' if random.random() < 0.01 else 'false'
                
                writer.writerow([
                    user_id, name, email, password_hash, 
                    birth_date.isoformat(), is_admin
                ])
                self.user_ids.append(user_id)
                
                if i % 1000 == 0:
                    logger.info(f"Создано {i} пользователей")
        
        logger.info("Пользователи созданы")

    def generate_halls(self):
        """Генерация кинозалов"""
        logger.info(f"Генерация {self.config['halls_count']} кинозалов...")
        
        with open(f'{self.output_dir}/halls.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'screen_type_id', 'name', 'description'])
            
            for i in range(self.config['halls_count']):
                hall_id = str(uuid.uuid4())
                screen_type_id = random.choice(self.screen_type_ids)
                name = f"Hall {i+1}"
                description = self.fake.text(max_nb_chars=300).strip()
                
                writer.writerow([hall_id, screen_type_id, name, description])
                self.hall_ids.append(hall_id)
                self.seats_by_hall[hall_id] = []
                self.movie_shows_by_hall[hall_id] = []
        
        logger.info("Кинозалы созданы")

    def generate_seats(self):
        """Генерация мест в кинозалах"""
        logger.info("Генерация мест...")
        
        with open(f'{self.output_dir}/seats.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'hall_id', 'seat_type_id', 'row_number', 'seat_number'])
            
            for hall_id in self.hall_ids:
                seats_count = random.randint(
                    self.config['seats_per_hall_min'], 
                    self.config['seats_per_hall_max']
                )
                rows = random.randint(8, 15)
                seats_per_row = seats_count // rows
                
                for row in range(1, rows + 1):
                    actual_seats_in_row = seats_per_row + (1 if row <= seats_count % rows else 0)
                    for seat_num in range(1, actual_seats_in_row + 1):
                        seat_id = str(uuid.uuid4())
                        seat_type_id = random.choice(self.seat_type_ids)
                        
                        seat_data = {
                            'id': seat_id,
                            'seat_type_id': seat_type_id,
                            'row_number': row,
                            'seat_number': seat_num
                        }
                        
                        writer.writerow([
                            seat_id, hall_id, seat_type_id, row, seat_num
                        ])
                        self.seats_by_hall[hall_id].append(seat_data)
        
        logger.info("Места созданы")

    def generate_movies(self):
        """Генерация фильмов с учетом всех CHECK constraints"""
        logger.info(f"Генерация {self.config['movies_count']} фильмов...")
        
        age_limits = [0, 6, 12, 16, 18]
        
        with open(f'{self.output_dir}/movies.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow([
                'id', 'title', 'duration', 'description', 
                'age_limit', 'box_office_revenue', 'release_date'
            ])
            
            for i in range(self.config['movies_count']):
                movie_id = str(uuid.uuid4())
                
                # Title с проверкой
                title = self.fake.catch_phrase().strip()
                while not title or len(title) > 200:
                    title = self.fake.catch_phrase().strip()
                
                # Duration больше 00:00:00
                hours = random.randint(1, 3)
                minutes = random.randint(0, 59)
                duration = f"{hours:02d}:{minutes:02d}:00"
                
                # Description с проверкой
                description = self.fake.text(max_nb_chars=500).strip()
                while not description:
                    description = self.fake.text(max_nb_chars=500).strip()
                
                age_limit = random.choice(age_limits)
                box_office_revenue = round(random.uniform(0, 1000000000), 2)
                
                # Release date может быть в будущем
                release_date = self.fake.date_between_dates(
                    date_start=date(1920, 1, 1),
                    date_end=date(2026, 12, 31)  # Может быть в будущем
                )
                
                movie_data = {
                    'id': movie_id,
                    'title': title,
                    'duration': duration,
                    'release_date': release_date
                }
                
                writer.writerow([
                    movie_id, title, duration, description, age_limit,
                    box_office_revenue, release_date.isoformat()
                ])
                self.movie_ids.append(movie_id)
                self.movies_by_id[movie_id] = movie_data
                
                if i % 100 == 0:
                    logger.info(f"Создано {i} фильмов")
        
        logger.info("Фильмы созданы")

    def generate_movies_genres(self):
        """Генерация связи фильмов и жанров"""
        logger.info("Генерация связи фильмов и жанров...")
        
        with open(f'{self.output_dir}/movies_genres.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['movie_id', 'genre_id'])
            
            for movie_id in self.movie_ids:
                # Добавление жанров к фильму (1-4 жанра)
                movie_genres = random.sample(
                    self.genre_ids, 
                    random.randint(1, min(4, len(self.genre_ids)))
                )
                
                for genre_id in movie_genres:
                    writer.writerow([movie_id, genre_id])
        
        logger.info("Связи фильмов и жанров созданы")

    def calculate_movie_end_time(self, movie_id, start_time):
        """Вычисление времени окончания фильма с учетом уборки"""
        movie_data = self.movies_by_id[movie_id]
        duration_parts = movie_data['duration'].split(':')
        duration_td = timedelta(
            hours=int(duration_parts[0]),
            minutes=int(duration_parts[1]),
            seconds=int(duration_parts[2])
        )
        return start_time + duration_td + timedelta(minutes=10)  # +10 минут на уборку

    def is_show_time_available(self, hall_id, new_start_time, new_movie_id):
        """Проверка доступности времени для сеанса (аналог триггера check_movie_show_conflict)"""
        if hall_id not in self.movie_shows_by_hall:
            return True
            
        new_end_time = self.calculate_movie_end_time(new_movie_id, new_start_time)
        
        for existing_show in self.movie_shows_by_hall[hall_id]:
            existing_movie_id = existing_show['movie_id']
            existing_start_time = existing_show['start_time']
            existing_end_time = self.calculate_movie_end_time(existing_movie_id, existing_start_time)
            
            # Проверка пересечения временных интервалов
            if (new_start_time < existing_end_time and new_end_time > existing_start_time):
                return False
                
        return True

    def generate_movie_shows(self):
        """Генерация сеансов со полностью случайным распределением"""
        logger.info(f"Генерация ~{self.config['movie_shows_total']} сеансов...")
        
        languages = ['English', 'Spanish', 'French', 'German', 'Italian', 'Русский']
        shows_created = 0
        
        with open(f'{self.output_dir}/movie_shows.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow(['id', 'movie_id', 'hall_id', 'start_time', 'language'])
            
            max_attempts = self.config['movie_shows_total'] * 3
            attempts = 0
            
            while shows_created < self.config['movie_shows_total'] and attempts < max_attempts:
                attempts += 1
                
                show_id = str(uuid.uuid4())
                movie_id = random.choice(self.movie_ids)
                hall_id = random.choice(self.hall_ids)
                
                # Случайное время в ближайшие 90 дней
                start_time = datetime.now().replace(
                    hour=0, minute=0, second=0, microsecond=0
                ) + timedelta(
                    days=random.randint(0, 90),
                    hours=random.randint(7, 24),  # с 7 утра до полуночи
                    minutes=random.choice([0, 10, 20, 30, 40, 50])
                )
                
                # Проверяем доступность времени
                if self.is_show_time_available(hall_id, start_time, movie_id):
                    language = random.choice(languages)
                    
                    show_data = {
                        'id': show_id,
                        'movie_id': movie_id,
                        'start_time': start_time,
                        'language': language
                    }
                    
                    if hall_id not in self.movie_shows_by_hall:
                        self.movie_shows_by_hall[hall_id] = []
                    self.movie_shows_by_hall[hall_id].append(show_data)
                    
                    writer.writerow([
                        show_id, movie_id, hall_id, 
                        start_time.isoformat(), language
                    ])
                    
                    shows_created += 1
                    
                    if shows_created % 100 == 0:
                        logger.info(f"Создано {shows_created} сеансов")
    
        # Статистика
        if self.movie_shows_by_hall:
            shows_distribution = {f"Hall {i+1}": len(shows) for i, (hall_id, shows) in enumerate(self.movie_shows_by_hall.items())}
            avg_shows = shows_created / len(self.movie_shows_by_hall)
            logger.info(f"Распределение сеансов: {shows_distribution}")
            logger.info(f"Всего создано {shows_created} сеансов")
            logger.info(f"Среднее на зал: {avg_shows:.1f}, Min: {min(shows_distribution.values())}, Max: {max(shows_distribution.values())}")

    def generate_tickets(self):
        """Генерация билетов с учетом всех constraints и триггеров"""
        logger.info("Генерация билетов...")
        
        base_price = 300.0
        tickets_created = 0
        
        with open(f'{self.output_dir}/tickets.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow([
                'id', 'movie_show_id', 'seat_id', 'user_id', 
                'ticket_Status', 'price'
            ])
            
            # Собираем все сеансы
            all_shows = []
            for hall_shows in self.movie_shows_by_hall.values():
                all_shows.extend(hall_shows)
            
            for show in all_shows:
                show_id = show['id']
                hall_id = None
                
                # Находим hall_id для этого сеанса
                for h_id, shows in self.movie_shows_by_hall.items():
                    if any(s['id'] == show_id for s in shows):
                        hall_id = h_id
                        break
                
                if not hall_id or hall_id not in self.seats_by_hall:
                    continue
                
                # Для каждого места в зале создаем билет
                for seat_data in self.seats_by_hall[hall_id]:
                    ticket_id = str(uuid.uuid4())
                    seat_id = seat_data['id']
                    
                    # Определяем статус билета
                    if random.random() < self.config['ticket_purchase_probability']:
                        user_id = random.choice(self.user_ids)
                        status = random.choices(
                            ['Purchased', 'Reserved'],
                            weights=[0.7, 0.3]  # 70% купленных, 30% забронированных
                        )[0]
                    else:
                        user_id = None
                        status = 'Available'
                    
                    # Расчет цены с учетом модификаторов
                    price = round(base_price * random.uniform(0.8, 1.5), 2)
                    
                    writer.writerow([
                        ticket_id, show_id, seat_id, user_id, status, price
                    ])
                    tickets_created += 1
                    
                    if tickets_created % 1000 == 0:
                        logger.info(f"Создано {tickets_created} билетов")
        
        logger.info(f"Всего создано {tickets_created} билетов")

    def generate_reviews(self):
        """Генерация отзывов с учетом UNIQUE constraint"""
        logger.info("Генерация отзывов...")
        
        reviews_created = 0
        used_review_combinations = set()
        
        with open(f'{self.output_dir}/reviews.csv', 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow([
                'id', 'user_id', 'movie_id', 'rating', 'review_comment'
            ])
            
            for user_id in self.user_ids:
                num_reviews = random.randint(
                    self.config['reviews_per_user_min'], 
                    self.config['reviews_per_user_max']
                )
                
                for _ in range(num_reviews):
                    movie_id = random.choice(self.movie_ids)
                    review_key = f"{user_id}_{movie_id}"
                    
                    # Проверяем уникальность комбинации user_id + movie_id
                    if review_key in used_review_combinations:
                        continue
                    
                    review_id = str(uuid.uuid4())
                    rating = random.randint(1, 10)
                    
                    # Комментарий может быть пустым, но если есть - должен быть не пустой
                    if random.random() > 0.3:
                        review_comment = self.fake.text(max_nb_chars=200).strip()
                        while not review_comment:  # Если получился пустой - генерируем заново
                            review_comment = self.fake.text(max_nb_chars=200).strip()
                    else:
                        review_comment = ''
                    
                    writer.writerow([
                        review_id, user_id, movie_id, rating, review_comment
                    ])
                    
                    used_review_combinations.add(review_key)
                    reviews_created += 1
                    
                    if reviews_created % 1000 == 0:
                        logger.info(f"Создано {reviews_created} отзывов")
        
        logger.info(f"Всего создано {reviews_created} отзывов")

    def generate_all_data(self):
        """Основной метод генерации всех данных"""
        try:
            # Порядок важен из-за foreign keys
            self.generate_genres()
            self.generate_screen_types()
            self.generate_seat_types()
            self.generate_users()
            self.generate_halls()
            self.generate_seats()
            self.generate_movies()
            self.generate_movies_genres()
            self.generate_movie_shows()
            self.generate_tickets()
            self.generate_reviews()
            
            logger.info("Все CSV файлы успешно сгенерированы!")
            
        except Exception as e:
            logger.error(f"Ошибка при генерации данных: {e}")
            raise

def main():
    generator = CSVDataGenerator("../../data")
    generator.generate_all_data()

if __name__ == "__main__":
    main()