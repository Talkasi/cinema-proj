#!/usr/bin/env python3
# analyze_degradation.py

import json
import pandas as pd
import numpy as np
import sys
import os
from datetime import datetime, timedelta
import matplotlib.pyplot as plt
import seaborn as sns

def debug_data_structure(results_dir, csv_data):
    """Функция для отладки структуры данных"""
    print("\nДЕБАГ СТРУКТУРЫ ДАННЫХ:")
    print(f"Всего строк: {len(csv_data)}")
    print(f"Колонки: {list(csv_data.columns)}")
    
    if not csv_data.empty:
        print(f"Типы метрик: {csv_data['metric_name'].unique()}")
        
        # Смотрим на данные http_reqs
        http_reqs = csv_data[csv_data['metric_name'] == 'http_reqs']
        if not http_reqs.empty:
            print(f"HTTP_REQS данные:")
            print(f"   - Количество записей: {len(http_reqs)}")
            print(f"   - Значения: {http_reqs['metric_value'].unique()[:10]}")  # первые 10 значений
            if 'timestamp' in http_reqs.columns:
                print(f"   - Временные метки: {http_reqs['timestamp'].head(3)}")

# def create_response_time_plots(results_dir, csv_data):
#     """Создает графики времени ответа"""
    
#     if csv_data.empty:
#         print("Нет данных k6 для анализа времени ответа")
#         return
    
#     print("Анализ времени ответа и пропускной способности...")
    
#     # Фильтруем только HTTP запросы
#     http_data = csv_data[csv_data['metric_name'].str.contains('http_req', na=False)]
    
#     if http_data.empty:
#         print("Нет данных HTTP запросов")
#         return
    
#     # Создаем фигуру с 4 графиками
#     fig, axes = plt.subplots(2, 2, figsize=(16, 12))
#     fig.suptitle('Анализ времени ответа и пропускной способности', fontsize=16, fontweight='bold')
    
#     # График 1: Распределение времени ответа по типам запросов
#     try:
#         # Группируем по URL и вычисляем статистику
#         url_stats = http_data.groupby('metric_name').agg({
#             'metric_value': ['count', 'mean', 'median', 'min', 'max', 'std']
#         }).round(2)
        
#         url_stats.columns = ['count', 'mean', 'median', 'min', 'max', 'std']
#         url_stats = url_stats.sort_values('mean', ascending=False)
        
#         # Барплот среднего времени ответа по типам запросов
#         bars = axes[0, 0].bar(range(len(url_stats)), url_stats['mean'], 
#                              color='skyblue', alpha=0.7, edgecolor='black')
#         axes[0, 0].set_title('Среднее время ответа по типам запросов')
#         axes[0, 0].set_ylabel('Время (мс)')
#         axes[0, 0].set_xlabel('Типы запросов')
#         axes[0, 0].set_xticks(range(len(url_stats)))
#         axes[0, 0].set_xticklabels([name.replace('http_req_duration{', '').replace('}', '') 
#                                    for name in url_stats.index], rotation=45, ha='right')
        
#         # Добавляем значения на столбцы
#         for i, bar in enumerate(bars):
#             height = bar.get_height()
#             axes[0, 0].text(bar.get_x() + bar.get_width()/2., height + 5,
#                            f'{height:.0f}ms', ha='center', va='bottom', fontsize=8)
        
#         axes[0, 0].grid(True, alpha=0.3)
        
#     except Exception as e:
#         axes[0, 0].text(0.5, 0.5, f'Ошибка: {e}', 
#                        ha='center', va='center', transform=axes[0, 0].transAxes)
    
#     # График 2: Время ответа по перцентилям
#     try:
#         # Вычисляем перцентили для всех запросов
#         all_requests = http_data[http_data['metric_name'] == 'http_req_duration']['metric_value']
        
#         percentiles = [50, 75, 90, 95, 99]
#         percentile_values = [np.percentile(all_requests, p) for p in percentiles]
        
#         bars = axes[0, 1].bar(range(len(percentiles)), percentile_values,
#                              color=['lightgreen', 'lightblue', 'orange', 'red', 'darkred'],
#                              alpha=0.7, edgecolor='black')
        
#         axes[0, 1].set_title('Время ответа по перцентилям (все запросы)')
#         axes[0, 1].set_ylabel('Время (мс)')
#         axes[0, 1].set_xlabel('Перцентиль')
#         axes[0, 1].set_xticks(range(len(percentiles)))
#         axes[0, 1].set_xticklabels([f'P{p}' for p in percentiles])
        
#         # Добавляем значения
#         for i, bar in enumerate(bars):
#             height = bar.get_height()
#             axes[0, 1].text(bar.get_x() + bar.get_width()/2., height + 5,
#                            f'{height:.0f}ms', ha='center', va='bottom', fontweight='bold')
        
#         axes[0, 1].grid(True, alpha=0.3)
        
#     except Exception as e:
#         axes[0, 1].text(0.5, 0.5, f'Ошибка: {e}', 
#                        ha='center', va='center', transform=axes[0, 1].transAxes)
    
#     # График 3: Динамика времени ответа во время теста
#     try:
#         # Группируем по времени (если есть временные метки)
#         if 'timestamp' in http_data.columns:
#             time_data = http_data.copy()
#             time_data['timestamp'] = pd.to_datetime(time_data['timestamp'])
#             time_data = time_data.sort_values('timestamp')
            
#             # Скользящее среднее для сглаживания
#             window_size = min(50, len(time_data) // 10)
#             if window_size > 1:
#                 time_data['smooth_duration'] = time_data['metric_value'].rolling(
#                     window=window_size, center=True).mean()
                
#                 axes[1, 0].plot(time_data['timestamp'], time_data['smooth_duration'],
#                                linewidth=2, color='purple', alpha=0.7, label='Сглаженное время')
            
#             axes[1, 0].scatter(time_data['timestamp'], time_data['metric_value'],
#                               alpha=0.3, s=1, color='blue', label='Отдельные запросы')
            
#             axes[1, 0].set_title('Динамика времени ответа во время теста')
#             axes[1, 0].set_ylabel('Время ответа (мс)')
#             axes[1, 0].set_xlabel('Время теста')
#             axes[1, 0].legend()
#             axes[1, 0].grid(True, alpha=0.3)
            
#             # Поворачиваем метки времени для лучшей читаемости
#             plt.setp(axes[1, 0].xaxis.get_majorticklabels(), rotation=45)
#         else:
#             # Если нет временных меток, используем порядковый номер
#             axes[1, 0].plot(range(len(http_data)), http_data['metric_value'],
#                            alpha=0.5, linewidth=1, color='blue')
#             axes[1, 0].set_title('Время ответа по порядку запросов')
#             axes[1, 0].set_ylabel('Время ответа (мс)')
#             axes[1, 0].set_xlabel('Номер запроса')
#             axes[1, 0].grid(True, alpha=0.3)
            
#     except Exception as e:
#         axes[1, 0].text(0.5, 0.5, f'Ошибка: {e}', 
#                        ha='center', va='center', transform=axes[1, 0].transAxes)
    
#     # График 4: Гистограмма распределения времени ответа
#     try:
#         durations = http_data[http_data['metric_name'] == 'http_req_duration']['metric_value']
        
#         axes[1, 1].hist(durations, bins=50, alpha=0.7, color='teal', edgecolor='black')
#         axes[1, 1].set_title('Распределение времени ответа')
#         axes[1, 1].set_ylabel('Количество запросов')
#         axes[1, 1].set_xlabel('Время ответа (мс)')
#         axes[1, 1].grid(True, alpha=0.3)
        
#         # Добавляем вертикальные линии для статистики
#         mean_duration = durations.mean()
#         median_duration = durations.median()
#         p95_duration = np.percentile(durations, 95)
        
#         axes[1, 1].axvline(mean_duration, color='red', linestyle='--', 
#                           label=f'Среднее: {mean_duration:.0f}ms')
#         axes[1, 1].axvline(median_duration, color='green', linestyle='--', 
#                           label=f'Медиана: {median_duration:.0f}ms')
#         axes[1, 1].axvline(p95_duration, color='orange', linestyle='--', 
#                           label=f'P95: {p95_duration:.0f}ms')
#         axes[1, 1].legend()
        
#     except Exception as e:
#         axes[1, 1].text(0.5, 0.5, f'Ошибка: {e}', 
#                        ha='center', va='center', transform=axes[1, 1].transAxes)
    
#     plt.tight_layout()
#     response_time_plot = f'{results_dir}/response_time_analysis.png'
#     plt.savefig(response_time_plot, dpi=300, bbox_inches='tight')
#     plt.close()
    
#     print(f"График времени ответа сохранен: {response_time_plot}")


def create_response_time_plots(results_dir, csv_data):
    """Создает графики времени ответа"""
    
    if csv_data.empty:
        print("Нет данных k6 для анализа времени ответа")
        return
    
    print("Анализ времени ответа...")
    
    # Создаем фигуру с 2 графиками друг под другом
    fig, axes = plt.subplots(2, 1, figsize=(14, 10))
    fig.suptitle('Анализ времени ответа', fontsize=16, fontweight='bold')
    
    # График 1: Время ответа по перцентилям
    try:
        completed_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        
        if not completed_requests.empty:
            percentiles = [50, 75, 90, 95, 99]
            percentile_values = [np.percentile(completed_requests['metric_value'], p) for p in percentiles]
            
            bars = axes[0].bar(range(len(percentiles)), percentile_values,
                             color=['lightgreen', 'lightblue', 'orange', 'red', 'darkred'],
                             alpha=0.7, edgecolor='black')
            
            axes[0].set_title('Время ответа по перцентилям (завершенные запросы)')
            axes[0].set_ylabel('Время (мс)')
            axes[0].set_xlabel('Перцентиль')
            axes[0].set_xticks(range(len(percentiles)))
            axes[0].set_xticklabels([f'P{p}' for p in percentiles])
            
            for i, bar in enumerate(bars):
                height = bar.get_height()
                axes[0].text(bar.get_x() + bar.get_width()/2., height + 5,
                           f'{height:.0f}ms', ha='center', va='bottom', fontweight='bold')
            
            axes[0].grid(True, alpha=0.3)
        else:
            axes[0].text(0.5, 0.5, 'Нет данных о завершенных запросах', 
                        ha='center', va='center', transform=axes[0].transAxes)
        
    except Exception as e:
        axes[0].text(0.5, 0.5, f'Ошибка перцентилей: {e}', 
                    ha='center', va='center', transform=axes[0].transAxes)
    
    # График 2: Гистограмма распределения времени ответа
    try:
        completed_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        
        if not completed_requests.empty:
            durations = completed_requests['metric_value']
            
            axes[1].hist(durations, bins=50, alpha=0.7, color='teal', edgecolor='black')
            axes[1].set_title('Распределение времени ответа (завершенные запросы)')
            axes[1].set_ylabel('Количество запросов')
            axes[1].set_xlabel('Время ответа (мс)')
            axes[1].grid(True, alpha=0.3)
            
            mean_duration = durations.mean()
            median_duration = durations.median()
            p95_duration = np.percentile(durations, 95)
            
            axes[1].axvline(mean_duration, color='red', linestyle='--', 
                           label=f'Среднее: {mean_duration:.0f}ms')
            axes[1].axvline(median_duration, color='green', linestyle='--', 
                           label=f'Медиана: {median_duration:.0f}ms')
            axes[1].axvline(p95_duration, color='orange', linestyle='--', 
                           label=f'P95: {p95_duration:.0f}ms')
            axes[1].legend()
        else:
            axes[1].text(0.5, 0.5, 'Нет данных о завершенных запросах', 
                        ha='center', va='center', transform=axes[1].transAxes)
        
    except Exception as e:
        axes[1].text(0.5, 0.5, f'Ошибка гистограммы: {e}', 
                    ha='center', va='center', transform=axes[1].transAxes)
    
    plt.tight_layout()
    response_time_plot = f'{results_dir}/response_time_analysis.png'
    plt.savefig(response_time_plot, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"График времени ответа сохранен: {response_time_plot}")


def create_container_plots(results_dir, resource_data):
    """Создает отдельные детальные графики для каждого контейнера"""
    
    if resource_data.empty:
        print("Нет данных ресурсов для создания графиков контейнеров")
        return
    
    containers = resource_data['container'].unique()
    
    for container in containers:
        if container not in ['go_app', 'postgres_db']:
            continue
            
        container_data = resource_data[resource_data['container'] == container]
        if container_data.empty:
            continue
            
        print(f"Создаю графики для контейнера: {container}")
        
        # Создаем фигуру с 2 подграфиками один под другим
        fig, axes = plt.subplots(2, 1, figsize=(15, 12))
        fig.suptitle(f'Детальный анализ контейнера: {container}', fontsize=16, fontweight='bold')
        
        # Подготовка данных с РЕАЛЬНЫМ временем
        container_data = container_data.copy()
        
        # Преобразуем timestamp в реальное время (секунды от начала теста)
        if 'timestamp' in container_data.columns:
            container_data['timestamp'] = pd.to_numeric(container_data['timestamp'], errors='coerce')
            container_data = container_data.dropna(subset=['timestamp'])
            # Конвертируем в секунды от начала теста
            start_time = container_data['timestamp'].min()
            real_time = (container_data['timestamp'] - start_time) / 1000  # в секундах
        else:
            # Если нет timestamp, используем индекс как приближение
            real_time = range(len(container_data))
        
        # График 1: Использование CPU (верхний)
        try:
            container_data['cpu_numeric'] = container_data['cpu_percent'].str.replace('%', '').astype(float)
            axes[0].plot(real_time, container_data['cpu_numeric'], 
                        linewidth=2, color='red', marker='o', markersize=2)
            axes[0].set_title(f'{container} - Использование CPU')
            axes[0].set_ylabel('CPU %')
            axes[0].set_xlabel('Время тестирования (секунды)')  # ИСПРАВЛЕНО!
            axes[0].grid(True, alpha=0.3)
            
            # Добавляем статистику
            cpu_mean = container_data['cpu_numeric'].mean()
            cpu_max = container_data['cpu_numeric'].max()
            axes[0].axhline(y=cpu_mean, color='blue', linestyle='--', alpha=0.7, 
                           label=f'Среднее: {cpu_mean:.1f}%')
            axes[0].axhline(y=cpu_max, color='orange', linestyle='--', alpha=0.7, 
                           label=f'Максимум: {cpu_max:.1f}%')
            axes[0].legend()
        except Exception as e:
            axes[0].text(0.5, 0.5, f'Ошибка CPU: {e}', 
                        ha='center', va='center', transform=axes[0].transAxes)
        
        # График 2: Использование памяти (нижний)
        try:
            def parse_memory(mem_str):
                try:
                    used = mem_str.split('/')[0].strip()
                    used_num = float(''.join(filter(lambda x: x.isdigit() or x == '.', used)))
                    return used_num
                except:
                    return 0
            
            container_data['mem_used'] = container_data['mem_usage'].apply(parse_memory)
            
            axes[1].plot(real_time, container_data['mem_used'], 
                        linewidth=2, color='green', marker='s', markersize=2)
            axes[1].set_title(f'{container} - Использование памяти')
            axes[1].set_ylabel('Память (MB)')
            axes[1].set_xlabel('Время тестирования (секунды)')  # ИСПРАВЛЕНО!
            axes[1].grid(True, alpha=0.3)
            
            # Добавляем статистику для памяти
            mem_mean = container_data['mem_used'].mean()
            mem_max = container_data['mem_used'].max()
            axes[1].axhline(y=mem_mean, color='blue', linestyle='--', alpha=0.7, 
                           label=f'Среднее: {mem_mean:.1f} MB')
            axes[1].axhline(y=mem_max, color='orange', linestyle='--', alpha=0.7, 
                           label=f'Максимум: {mem_max:.1f} MB')
            axes[1].legend()
            
        except Exception as e:
            axes[1].text(0.5, 0.5, f'Ошибка памяти: {e}', 
                        ha='center', va='center', transform=axes[1].transAxes)
        
        plt.tight_layout()
        container_plot_path = f'{results_dir}/container_{container}_analysis.png'
        plt.savefig(container_plot_path, dpi=300, bbox_inches='tight')
        plt.close()
        
        print(f"Графики сохранены: {container_plot_path}")

def create_plots(results_dir, csv_data, resource_data):
    """Создает все графики производительности"""
    create_response_time_plots(results_dir, csv_data)
    create_container_plots(results_dir, resource_data)

def analyze_degradation(results_dir):
    print("Детальный анализ точки деградации...")
    
    # Загрузка результатов
    csv_file = f'{results_dir}/k6_results.csv'
    resource_file = f'{results_dir}/resource_usage.csv'
    
    # Загружаем CSV данные k6
    csv_data = pd.DataFrame()
    if os.path.exists(csv_file):
        try:
            csv_data = pd.read_csv(csv_file, low_memory=False)
            
            # Преобразуем числовые колонки
            if 'metric_value' in csv_data.columns:
                csv_data['metric_value'] = pd.to_numeric(csv_data['metric_value'], errors='coerce')
            
            print(f"Загружено {len(csv_data)} записей из k6 CSV")
            
            # Отладка структуры данных
            debug_data_structure(results_dir, csv_data)
                
        except Exception as e:
            print(f"Ошибка при чтении CSV: {e}")
    
    # Загружаем данные ресурсов
    resource_data = pd.DataFrame()
    if os.path.exists(resource_file):
        try:
            resource_data = pd.read_csv(resource_file)
            print(f"Загружено {len(resource_data)} записей о ресурсах")
            
        except Exception as e:
            print(f"Ошибка при чтении ресурсов: {e}")

    # Создаем графики
    create_plots(results_dir, csv_data, resource_data)

    print(f"Анализ завершен! Все графики сохранены в: {results_dir}")

if __name__ == '__main__':
    if len(sys.argv) > 1:
        analyze_degradation(sys.argv[1])
    else:
        print("Usage: python3 analyze_degradation.py <results_directory>")