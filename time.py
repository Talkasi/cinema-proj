# import sqlite3
# import time
# import matplotlib.pyplot as plt
# import numpy as np
# import random
# from datetime import datetime, timedelta

# # Настройки эксперимента
# max_size = 6  # Максимальный размер таблицы (10^6 записей)
# sizes = []  # Размеры таблиц
# num_samples = 50  # Количество замеров для усреднения

# # Списки для хранения результатов
# times_without_index = []
# times_with_index = []

# # Минимальное время для отображения на графике
# MIN_TIME = 1e-9  # 1 наносекунда

# # Генерация размеров таблиц
# for size in [10**i for i in range(1, max_size + 1)]:
#     sizes.append(size)
#     # Добавляем дополнительные размеры между степенями
#     if size != 10**max_size:
#         for additional_size in range(size, size*10, size):
#             sizes.append(additional_size)

# sizes = sorted(set(sizes))  # Удаляем дубликаты и сортируем
# print(sizes, len(sizes))

# def generate_random_timestamp(start, end):
#     """Генерация случайной временной метки между start и end"""
#     delta = end - start
#     int_delta = (delta.days * 24 * 60 * 60) + delta.seconds
#     random_second = random.randrange(int_delta)
#     return start + timedelta(seconds=random_second)

# for size in sizes:
#     conn = sqlite3.connect(':memory:')
#     cursor = conn.cursor()
    
#     # Создание таблиц (упрощенная версия вашей схемы)
#     cursor.execute('''
#         CREATE TABLE movies (
#             id TEXT PRIMARY KEY,
#             title TEXT NOT NULL,
#             duration TEXT NOT NULL
#         )
#     ''')
    
#     cursor.execute('''
#         CREATE TABLE halls (
#             id TEXT PRIMARY KEY,
#             name TEXT NOT NULL
#         )
#     ''')
    
#     cursor.execute('''
#         CREATE TABLE movie_shows (
#             id TEXT PRIMARY KEY,
#             movie_id TEXT REFERENCES movies(id),
#             hall_id TEXT REFERENCES halls(id),
#             start_time TEXT NOT NULL,
#             language TEXT NOT NULL
#         )
#     ''')
    
#     # Заполнение таблиц тестовыми данными
#     # Сначала создаем фильмы и залы
#     movies = [('movie_' + str(i), f'Movie {i}', '02:30:00') for i in range(100)]
#     halls = [('hall_' + str(i), f'Hall {i}') for i in range(10)]
    
#     cursor.executemany('INSERT INTO movies VALUES (?, ?, ?)', movies)
#     cursor.executemany('INSERT INTO halls VALUES (?, ?)', halls)
    
#     # Генерируем сеансы
#     start_date = datetime(2023, 1, 1)
#     end_date = datetime(2023, 12, 31)
    
#     movie_shows = []
#     for i in range(size):
#         movie_id = random.choice(movies)[0]
#         hall_id = random.choice(halls)[0]
#         start_time = generate_random_timestamp(start_date, end_date)
#         language = random.choice(['English', 'Spanish', 'French', 'German', 'Italian', 'Русский'])
#         movie_shows.append((f'show_{i}', movie_id, hall_id, start_time.isoformat(), language))
    
#     cursor.executemany('INSERT INTO movie_shows VALUES (?, ?, ?, ?, ?)', movie_shows)
#     conn.commit()
    
#     # Тестирование запроса по hall_id и start_time БЕЗ индекса
#     times = []
#     for _ in range(num_samples):
#         hall_id = random.choice(halls)[0]
#         random_time = generate_random_timestamp(start_date, end_date)
        
#         start = time.perf_counter()
#         cursor.execute('''
#             SELECT * FROM movie_shows 
#             WHERE hall_id = ? AND start_time >= ? AND start_time <= ?
#         ''', (hall_id, 
#               (random_time - timedelta(hours=2)).isoformat(), 
#               (random_time + timedelta(hours=2)).isoformat()))
#         cursor.fetchall()
#         elapsed = max(time.perf_counter() - start, MIN_TIME)
#         times.append(elapsed)
#     avg_time = np.mean(times)
#     times_without_index.append(avg_time)
    
#     # Создание индекса
#     cursor.execute('CREATE INDEX idx_movie_shows_hall_time ON movie_shows(hall_id, start_time)')
    
#     # Тестирование запроса по hall_id и start_time С индексом
#     times = []
#     for _ in range(num_samples):
#         hall_id = random.choice(halls)[0]
#         random_time = generate_random_timestamp(start_date, end_date)
        
#         start = time.perf_counter()
#         cursor.execute('''
#             SELECT * FROM movie_shows 
#             WHERE hall_id = ? AND start_time >= ? AND start_time <= ?
#         ''', (hall_id, 
#               (random_time - timedelta(hours=2)).isoformat(), 
#               (random_time + timedelta(hours=2)).isoformat()))
#         cursor.fetchall()
#         elapsed = max(time.perf_counter() - start, MIN_TIME)
#         times.append(elapsed)
#     avg_time = np.mean(times)
#     times_with_index.append(avg_time)
    
#     conn.close()

# # Построение графика
# plt.figure(figsize=(12, 7))
# plt.plot(sizes, times_without_index, color='red', linestyle='-', label='Без индекса')
# plt.plot(sizes, times_with_index, color='green', linestyle='--', label='С индексом')
# plt.xscale('log')
# plt.yscale('log')  # Логарифмическая шкала для оси Y
# plt.xlabel('Размер таблицы (количество сеансов)')
# plt.ylabel('Среднее время выполнения запроса (сек)')
# plt.title('Производительность запросов к таблице movie_shows\n(поиск по hall_id и временному диапазону)')
# plt.legend()
# plt.grid(True, which="both", ls="--")
# plt.show()

import psycopg2
import time
import matplotlib.pyplot as plt
import numpy as np
import random
from datetime import datetime, timedelta
from psycopg2 import sql

# Настройки эксперимента
max_size = 5  # Максимальный размер таблицы (10^6 записей)
sizes = []  # Размеры таблиц
num_samples = 5  # Количество замеров для усреднения

# Списки для хранения результатов
times_without_index = []
times_with_index = []

# Минимальное время для отображения на графике
MIN_TIME = 1e-9  # 1 наносекунда

# Генерация размеров таблиц
for size in [10**i for i in range(1, max_size + 1)]:
    sizes.append(size)
    # Добавляем дополнительные размеры между степенями
    if size != 10**max_size:
        for additional_size in range(size, size*10, size):
            sizes.append(additional_size)

sizes = sorted(set(sizes))  # Удаляем дубликаты и сортируем
print(sizes, len(sizes))

def generate_random_timestamp(start, end):
    """Генерация случайной временной метки между start и end"""
    delta = end - start
    int_delta = (delta.days * 24 * 60 * 60) + delta.seconds
    random_second = random.randrange(int_delta)
    return start + timedelta(seconds=random_second)

# Подключение к PostgreSQL
conn = psycopg2.connect(
    dbname="cinema",
    user="postgres",
    password="postgres",
    host="localhost"
)
conn.autocommit = True
cursor = conn.cursor()

for size in sizes:
    print(size)
    try:
        # Создание временной схемы для изоляции экспериментов
        schema_name = f"experiment_{size}"
        cursor.execute(sql.SQL("CREATE SCHEMA IF NOT EXISTS {}").format(sql.Identifier(schema_name)))
        
        # Создание таблиц
        cursor.execute(sql.SQL('''
            CREATE TABLE IF NOT EXISTS {}.movies (
                id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                duration TEXT NOT NULL
            )
        ''').format(sql.Identifier(schema_name)))
        
        cursor.execute(sql.SQL('''
            CREATE TABLE IF NOT EXISTS {}.halls (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL
            )
        ''').format(sql.Identifier(schema_name)))
        
        cursor.execute(sql.SQL('''
            CREATE TABLE IF NOT EXISTS {}.movie_shows (
                id TEXT PRIMARY KEY,
                movie_id TEXT REFERENCES {}.movies(id),
                hall_id TEXT REFERENCES {}.halls(id),
                start_time TIMESTAMP NOT NULL,
                language TEXT NOT NULL
            )
        ''').format(
            sql.Identifier(schema_name),
            sql.Identifier(schema_name),
            sql.Identifier(schema_name)
        ))
        
        # Заполнение таблиц тестовыми данными
        # Сначала создаем фильмы и залы
        movies = [('movie_' + str(i), f'Movie {i}', '02:30:00') for i in range(100)]
        halls = [('hall_' + str(i), f'Hall {i}') for i in range(10)]
        
        cursor.executemany(
            sql.SQL('INSERT INTO {}.movies VALUES (%s, %s, %s)').format(sql.Identifier(schema_name)),
            movies
        )
        cursor.executemany(
            sql.SQL('INSERT INTO {}.halls VALUES (%s, %s)').format(sql.Identifier(schema_name)),
            halls
        )
        
        # Генерируем сеансы
        start_date = datetime(2023, 1, 1)
        end_date = datetime(2023, 12, 31)
        
        # Используем batch-вставку для ускорения
        batch_size = 10000
        for i in range(0, size, batch_size):
            batch = []
            for j in range(i, min(i + batch_size, size)):
                movie_id = random.choice(movies)[0]
                hall_id = random.choice(halls)[0]
                start_time = generate_random_timestamp(start_date, end_date)
                language = random.choice(['English', 'Spanish', 'French', 'German', 'Italian', 'Русский'])
                batch.append((f'show_{j}', movie_id, hall_id, start_time, language))
            
            cursor.executemany(
                sql.SQL('INSERT INTO {}.movie_shows VALUES (%s, %s, %s, %s, %s)').format(sql.Identifier(schema_name)),
                batch
            )
        
        # Анализируем таблицы для обновления статистики
        cursor.execute(sql.SQL('ANALYZE {}.movies').format(sql.Identifier(schema_name)))
        cursor.execute(sql.SQL('ANALYZE {}.halls').format(sql.Identifier(schema_name)))
        cursor.execute(sql.SQL('ANALYZE {}.movie_shows').format(sql.Identifier(schema_name)))
        
        # Тестирование запроса по hall_id и start_time БЕЗ индекса
        times = []
        for _ in range(num_samples):
            hall_id = random.choice(halls)[0]
            random_time = generate_random_timestamp(start_date, end_date)
            
            start = time.perf_counter()
            cursor.execute(
                sql.SQL('''
                    SELECT * FROM {}.movie_shows 
                    WHERE hall_id = %s AND start_time >= %s AND start_time <= %s
                ''').format(sql.Identifier(schema_name)),
                (hall_id, random_time - timedelta(hours=2), random_time + timedelta(hours=2))
            )
            cursor.fetchall()
            elapsed = max(time.perf_counter() - start, MIN_TIME)
            times.append(elapsed)
        avg_time = np.mean(times)
        times_without_index.append(avg_time)
        
        # Создание индекса
        cursor.execute(
            sql.SQL('CREATE INDEX IF NOT EXISTS idx_movie_shows_hall_time ON {}.movie_shows(hall_id, start_time)')
            .format(sql.Identifier(schema_name))
        )
        
        # Тестирование запроса по hall_id и start_time С индексом
        times = []
        for _ in range(num_samples):
            hall_id = random.choice(halls)[0]
            random_time = generate_random_timestamp(start_date, end_date)
            
            start = time.perf_counter()
            cursor.execute(
                sql.SQL('''
                    SELECT * FROM {}.movie_shows 
                    WHERE hall_id = %s AND start_time >= %s AND start_time <= %s
                ''').format(sql.Identifier(schema_name)),
                (hall_id, random_time - timedelta(hours=2), random_time + timedelta(hours=2))
            )
            cursor.fetchall()
            elapsed = max(time.perf_counter() - start, MIN_TIME)
            times.append(elapsed)
        avg_time = np.mean(times)
        times_with_index.append(avg_time)
        
    finally:
        # Удаление временной схемы
        cursor.execute(sql.SQL('DROP SCHEMA IF EXISTS {} CASCADE').format(sql.Identifier(schema_name)))

cursor.close()
conn.close()

# Построение графика
plt.figure(figsize=(12, 7))
plt.plot(sizes, times_without_index, color='red', linestyle='-', label='Без индекса')
plt.plot(sizes, times_with_index, color='green', linestyle='--', label='С индексом')
plt.xscale('log')
plt.yscale('log')  # Логарифмическая шкала для оси Y
plt.xlabel('Размер таблицы (количество сеансов)')
plt.ylabel('Среднее время выполнения запроса (сек)')
plt.title('Производительность запросов к таблице movie_shows\n(поиск по hall_id и временному диапазону)')
plt.legend()
plt.grid(True, which="both", ls="--")
plt.show()