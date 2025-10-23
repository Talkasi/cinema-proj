#!/bin/bash

# Проверяем, что номер прогона был передан
if [ -z "$1" ]; then
  echo "Ошибка: Укажите номер прогона. Пример: ./run_single_test.sh 1"
  exit 1
fi

RUN_ID=$1
RESULTS_DIR="run_${RUN_ID}"
K6_SCRIPT_PATH="/scripts/test.js" # Путь внутри контейнера k6

# Создаем папку для результатов этого прогона
mkdir -p $RESULTS_DIR

echo "--- Запуск прогона №${RUN_ID} ---"

# --- ШАГ 1: Запускаем сбор статистики ресурсов в фоновом режиме ---
echo "Запуск сбора статистики docker stats..."
STATS_FILE="${RESULTS_DIR}/resource_usage.csv"
echo "timestamp,container_name,cpu_percent,mem_usage" > $STATS_FILE

(while true; do
  TIMESTAMP=$(date +%s)
  # Собираем статистику только для нужных контейнеров go_app и postgres_db
  docker stats --no-stream --format "$TIMESTAMP,{{.Name}},{{.CPUPerc}},{{.MemUsage}}" go_app postgres_db >> $STATS_FILE
  sleep 1
done) &
STATS_PID=$!

# --- ШАГ 2: Запускаем тест k6 и ждем его завершения ---
echo "Запуск теста k6..."
# Используем 'docker-compose run --rm', чтобы создать временный контейнер для теста
docker-compose run --rm k6 run $K6_SCRIPT_PATH --out json="/scripts/performance_results.json"
# Копируем результат из контейнера на хост
docker cp k6_runner:/scripts/performance_results.json "${RESULTS_DIR}/performance_results.json"


# --- ШАГ 3: Останавливаем сбор статистики ---
echo "Остановка сбора статистики docker stats..."
kill $STATS_PID

echo "--- Прогон №${RUN_ID} завершен. Результаты в папке ${RESULTS_DIR} ---"