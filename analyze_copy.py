#!/usr/bin/env python3
# degradation_analysis.py

import pandas as pd
import numpy as np
import sys
import os
import matplotlib.pyplot as plt

def debug_data_structure(results_dir, csv_data):
    """Функция для отладки структуры данных"""
    print("\nДЕБАГ СТРУКТУРЫ ДАННЫХ:")
    print(f"Всего строк: {len(csv_data)}")
    print(f"Колонки: {list(csv_data.columns)}")
    
    if not csv_data.empty:
        print(f"Типы метрик: {csv_data['metric_name'].unique()}")
        
        # Смотрим на данные ошибок
        error_data = csv_data[csv_data['metric_name'] == 'http_req_failed']
        if not error_data.empty:
            print(f"\nHTTP_REQ_FAILED данные:")
            print(f"   - Количество записей: {len(error_data)}")
            print(f"   - Значения: {error_data['metric_value'].unique()[:10]}")
        
        # Смотрим на статусы HTTP запросов
        http_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        if not http_requests.empty and 'status' in http_requests.columns:
            print(f"\nHTTP статусы:")
            print(f"   - Уникальные статусы: {http_requests['status'].unique()}")
            print(f"   - Примеры статусов: {http_requests['status'].value_counts().head()}")

def calculate_error_rate(csv_data, time_window, window_start, window_end):
    """Правильно вычисляет процент ошибок для временного интервала"""
    try:
        # Находим все HTTP запросы в этом интервале
        http_requests = csv_data[
            (csv_data['metric_name'] == 'http_req_duration')
        ].copy()
        
        # Конвертируем timestamp для фильтрации
        http_requests.loc[:, 'timestamp_dt'] = pd.to_datetime(http_requests['timestamp'], unit='s', errors='coerce')
        http_requests = http_requests.dropna(subset=['timestamp_dt'])
        
        # Фильтруем по временному окну
        window_requests = http_requests[
            (http_requests['timestamp_dt'] >= window_start) & 
            (http_requests['timestamp_dt'] < window_end)
        ]
        
        total_requests = len(window_requests)
        
        if total_requests == 0:
            return 0, 0
        
        # Считаем ошибки по статусам (4xx, 5xx) или наличию error
        error_count = 0
        
        if 'status' in window_requests.columns:
            # Конвертируем статусы в числовой формат без предупреждений
            status_series = pd.to_numeric(window_requests['status'], errors='coerce')
            error_requests = window_requests[
                (status_series >= 400) |  # HTTP ошибки
                (window_requests['error'].notna())  # Есть текст ошибки
            ]
            error_count = len(error_requests)
        
        error_rate = (error_count / total_requests * 100) if total_requests > 0 else 0
        
        return error_count, error_rate
        
    except Exception as e:
        print(f"Ошибка при расчете ошибок: {e}")
        return 0, 0

def find_degradation_point(results_dir, csv_data):
    """Находит точку деградации производительности по времени ответа и ошибкам"""
    
    print("Поиск точки деградации...")
    
    # Фильтруем данные
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    
    if duration_data.empty:
        print("Нет данных о времени ответа")
        return
    
    # Подготавливаем данные
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value'])
    
    # Группируем по временным интервалам
    if 'timestamp' in duration_data.columns:
        duration_data.loc[:, 'timestamp_dt'] = pd.to_datetime(duration_data['timestamp'], unit='s', errors='coerce')
        duration_data = duration_data.dropna(subset=['timestamp_dt'])
        duration_data = duration_data.sort_values('timestamp_dt')
        
        # Создаем временные интервалы (каждые 30 секунд)
        duration_data.loc[:, 'time_window'] = (duration_data['timestamp_dt'] - duration_data['timestamp_dt'].min()).dt.total_seconds() // 30
        
        # Анализируем каждый временной интервал
        degradation_found_time = False
        degradation_found_errors = False
        degradation_time_time = None
        degradation_time_errors = None
        degradation_load_time = None
        degradation_load_errors = None
        degradation_error_rate = None
        
        print("\nАнализ временных интервалов:")
        print("Время(сек) | Запросов | P95(ms) | Ошибок | Статус")
        print("-" * 65)
        
        for time_window in sorted(duration_data['time_window'].unique()):
            window_data = duration_data[duration_data['time_window'] == time_window]
            
            if len(window_data) < 5:  # Минимум 5 запросов для статистики
                continue
            
            # Вычисляем P95 для этого интервала
            p95 = np.percentile(window_data['metric_value'], 95)
            
            # Оцениваем нагрузку (количество запросов в интервале)
            load = len(window_data)
            
            # Вычисляем процент ошибок для этого интервала
            window_start = duration_data['timestamp_dt'].min() + pd.Timedelta(seconds=time_window*30)
            window_end = window_start + pd.Timedelta(seconds=30)
            
            error_count, error_rate = calculate_error_rate(csv_data, time_window, window_start, window_end)
            
            # Проверяем условия деградации
            status = "✅ OK"
            degradation_reason = []
            
            if p95 > 500:
                degradation_reason.append("P95 > 500ms")
                if not degradation_found_time:
                    degradation_found_time = True
                    degradation_time_time = time_window * 30
                    degradation_load_time = load
            
            if error_rate > 3.0:
                degradation_reason.append(f"Ошибок > 0.5% ({error_rate:.1f}%)")
                if not degradation_found_errors:
                    degradation_found_errors = True
                    degradation_time_errors = time_window * 30
                    degradation_load_errors = load
                    degradation_error_rate = error_rate
            
            if degradation_reason:
                status = f"❌ ДЕГРАДАЦИЯ ({', '.join(degradation_reason)})"
            
            print(f"{time_window * 30:8.0f} | {load:8} | {p95:7.0f} | {error_rate:5.1f}% | {status}")
        
        # Выводим результаты
        print(f"\n📊 РЕЗУЛЬТАТЫ ПОИСКА ДЕГРАДАЦИИ:")
        
        if degradation_found_time:
            print(f"🚨 ДЕГРАДАЦИЯ ПО ВРЕМЕНИ ОТВЕТА:")
            print(f"   Время: {degradation_time_time} секунд от начала теста")
            print(f"   Нагрузка: ~{degradation_load_time} запросов/30сек")
            print(f"   P95 превысил 500ms")
        
        if degradation_found_errors:
            print(f"🚨 ДЕГРАДАЦИЯ ПО ОШИБКАМ:")
            print(f"   Время: {degradation_time_errors} секунд от начала теста")
            print(f"   Нагрузка: ~{degradation_load_errors} запросов/30сек")
            print(f"   Процент ошибок: {degradation_error_rate:.1f}%")
        
        if not degradation_found_time and not degradation_found_errors:
            print(f"✅ Деградация не обнаружена в пределах теста")
        elif not degradation_found_errors:
            print(f"ℹ️  Ошибок не обнаружено (все запросы успешны)")
    
    return degradation_found_time or degradation_found_errors

def create_degradation_analysis(results_dir, csv_data):
    """Создает анализ для поиска точки деградации по времени ответа и ошибкам"""
    
    print("Создание анализа точки деградации...")
    
    # Фильтруем данные времени ответа
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    
    if duration_data.empty:
        print("Нет данных о времени ответа")
        return
    
    # Подготавливаем данные
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value'])
    
    if 'timestamp' not in duration_data.columns:
        print("Нет временных меток для анализа")
        return
    
    # Конвертируем timestamp
    duration_data.loc[:, 'timestamp_dt'] = pd.to_datetime(duration_data['timestamp'], unit='s', errors='coerce')
    duration_data = duration_data.dropna(subset=['timestamp_dt'])
    duration_data = duration_data.sort_values('timestamp_dt')
    
    # Создаем временные интервалы (каждые 30 секунд)
    duration_data.loc[:, 'time_elapsed'] = (duration_data['timestamp_dt'] - duration_data['timestamp_dt'].min()).dt.total_seconds()
    duration_data.loc[:, 'time_window'] = (duration_data['time_elapsed'] // 30).astype(int)
    
    # Анализируем каждый интервал
    window_stats = []
    
    for window in sorted(duration_data['time_window'].unique()):
        window_data = duration_data[duration_data['time_window'] == window]
        
        if len(window_data) < 5:  # Минимум 5 запросов для статистики
            continue
        
        # Статистика времени ответа
        p50 = np.percentile(window_data['metric_value'], 50)
        p75 = np.percentile(window_data['metric_value'], 75) 
        p90 = np.percentile(window_data['metric_value'], 90)
        p95 = np.percentile(window_data['metric_value'], 95)
        p99 = np.percentile(window_data['metric_value'], 99)
        
        # Статистика ошибок
        window_start = duration_data['timestamp_dt'].min() + pd.Timedelta(seconds=window*30)
        window_end = window_start + pd.Timedelta(seconds=30)
        
        error_count, error_rate = calculate_error_rate(csv_data, window, window_start, window_end)
        
        total_requests = len(window_data)
        
        window_stats.append({
            'time_window': window,
            'time_seconds': window * 30,
            'requests': total_requests,
            'p50': p50,
            'p75': p75, 
            'p90': p90,
            'p95': p95,
            'p99': p99,
            'error_count': error_count,
            'error_rate': error_rate,
            'degraded_time': p95 > 500,  # Деградация по времени
            'degraded_errors': error_rate > 0.5  # Деградация по ошибкам
        })
    
    if not window_stats:
        print("Недостаточно данных для анализа")
        return
    
    stats_df = pd.DataFrame(window_stats)
    
    # Создаем график анализа деградации
    fig, (ax1, ax2, ax3) = plt.subplots(3, 1, figsize=(15, 12))
    
    # График 1: P95 по времени
    ax1.plot(stats_df['time_seconds'], stats_df['p95'], 
             linewidth=3, color='blue', marker='o', label='P95 время ответа')
    
    # Линия порога деградации по времени
    ax1.axhline(y=500, color='red', linestyle='--', linewidth=2, 
                label='Порог деградации (500ms)')
    
    # Закрашиваем зону деградации по времени
    degraded_time_windows = stats_df[stats_df['degraded_time']]
    if not degraded_time_windows.empty:
        first_degradation_time = degraded_time_windows.iloc[0]
        ax1.axvline(x=first_degradation_time['time_seconds'], color='orange', 
                   linestyle=':', linewidth=2, alpha=0.7,
                   label=f'Деградация времени: {first_degradation_time["time_seconds"]}сек')
        
        # Закрашиваем период деградации
        degradation_start = first_degradation_time['time_seconds']
        degradation_end = stats_df['time_seconds'].max()
        ax1.axvspan(degradation_start, degradation_end, alpha=0.2, color='red', 
                   label='Период деградации')
    
    ax1.set_title('Время ответа P95 и порог деградации', fontsize=14, fontweight='bold')
    ax1.set_ylabel('P95 время ответа (ms)', fontsize=12)
    ax1.legend()
    ax1.grid(True, alpha=0.3)
    
    # График 2: Процент ошибок по времени
    ax2.plot(stats_df['time_seconds'], stats_df['error_rate'], 
             linewidth=3, color='red', marker='s', label='Процент ошибок')
    
    # Линия порога деградации по ошибкам
    ax2.axhline(y=0.5, color='darkred', linestyle='--', linewidth=2, 
                label='Порог деградации (0.5% ошибок)')
    
    # Закрашиваем зону деградации по ошибкам
    degraded_error_windows = stats_df[stats_df['degraded_errors']]
    if not degraded_error_windows.empty:
        first_degradation_errors = degraded_error_windows.iloc[0]
        ax2.axvline(x=first_degradation_errors['time_seconds'], color='purple', 
                   linestyle=':', linewidth=2, alpha=0.7,
                   label=f'Деградация ошибок: {first_degradation_errors["time_seconds"]}сек')
    
    ax2.set_title('Процент ошибок и порог деградации', fontsize=14, fontweight='bold')
    ax2.set_ylabel('Процент ошибок (%)', fontsize=12)
    ax2.set_xlabel('Время от начала теста (секунды)', fontsize=12)
    ax2.legend()
    ax2.grid(True, alpha=0.3)
    
    # График 3: Количество запросов по времени (нагрузка)
    bars = ax3.bar(stats_df['time_seconds'], stats_df['requests'], 
            width=25, alpha=0.7, color='green', label='Запросов за 30сек')

    # Добавляем значения поверх столбцов
    for bar, req_count in zip(bars, stats_df['requests']):
        height = bar.get_height()
        ax3.text(bar.get_x() + bar.get_width()/2., height,
                f'{int(req_count)}',
                ha='center', va='bottom', fontsize=8, fontweight='bold')

    # Отмечаем точки деградации на графике нагрузки
    if not degraded_time_windows.empty:
        first_degradation = degraded_time_windows.iloc[0]
        degradation_time = first_degradation['time_seconds']
        degradation_requests = first_degradation['requests']
        
        # Добавляем вертикальную линию
        ax3.axvline(x=degradation_time, color='orange', 
                linestyle=':', linewidth=2, alpha=0.7)
        
        # Добавляем точку на графике
        ax3.plot(degradation_time, degradation_requests, 'ro', markersize=8, 
                markerfacecolor='red', markeredgecolor='darkred', markeredgewidth=2)
        
        # Добавляем горизонтальную линию к значению Y
        ax3.axhline(y=degradation_requests, color='red', linestyle='--', 
                alpha=0.5, linewidth=1)

    # ДОБАВЛЯЕМ ПОДПИСИ ОСЕЙ И ЛЕГЕНДУ С ИНФОРМАЦИЕЙ О ЗАПРОСАХ
    ax3.set_title('Нагрузка (количество запросов по времени)', fontsize=14, fontweight='bold')
    ax3.set_xlabel('Время от начала теста (секунды)', fontsize=12)
    ax3.set_ylabel('Количество запросов', fontsize=12)

    # Создаем кастомную легенду с информацией о деградации
    legend_elements = [
        plt.Line2D([0], [0], color='green', alpha=0.7, linewidth=10, label='Запросов за 30сек'),
    ]

    if not degraded_time_windows.empty:
        first_degradation = degraded_time_windows.iloc[0]
        degradation_requests = first_degradation['requests']
        
        legend_elements.extend([
            plt.Line2D([0], [0], color='orange', linestyle=':', linewidth=2, 
                    label=f'Деградация: {first_degradation["time_seconds"]:.0f}сек'),
            plt.Line2D([0], [0], marker='o', color='red', markersize=8,
                    label=f'Запросов при деградации: {degradation_requests}'),
            plt.Line2D([0], [0], color='red', linestyle='--', linewidth=1,
                    label='Уровень нагрузки при деградации')
        ])

    ax3.legend(handles=legend_elements, loc='upper left')
    ax3.grid(True, alpha=0.3)
    
    # Общая информация
    total_requests = len(duration_data)
    
    # Считаем общее количество ошибок
    total_error_count = 0
    for window in window_stats:
        total_error_count += window['error_count']
    
    overall_error_rate = (total_error_count / total_requests * 100) if total_requests > 0 else 0
    max_p95 = stats_df['p95'].max()
    max_error_rate = stats_df['error_rate'].max()
    
    info_text = (f"Всего запросов: {total_requests:,} | "
                f"Всего ошибок: {total_error_count} ({overall_error_rate:.1f}%) | "
                f"Макс P95: {max_p95:.0f}ms | Макс ошибок: {max_error_rate:.1f}%")
    
    fig.suptitle('Анализ точки деградации производительности', fontsize=16, fontweight='bold')
    fig.text(0.5, 0.01, info_text, ha='center', fontsize=11,
             bbox=dict(boxstyle="round,pad=0.5", facecolor="lightgray", alpha=0.7))
    
    plt.tight_layout()
    degradation_plot = f'{results_dir}/1_degradation_analysis.png'
    plt.savefig(degradation_plot, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"График анализа деградации сохранен: {degradation_plot}")
    
    # Выводим детальный отчет
    print("\n" + "="*70)
    print("ОТЧЕТ О ПОИСКЕ ТОЧКИ ДЕГРАДАЦИИ")
    print("="*70)
    
    # Деградация по времени ответа
    if not degraded_time_windows.empty:
        first_degraded_time = degraded_time_windows.iloc[0]
        print(f"🚨 ДЕГРАДАЦИЯ ПО ВРЕМЕНИ ОТВЕТА:")
        print(f"   Время: {first_degraded_time['time_seconds']} секунд от начала теста")
        print(f"   Нагрузка: {first_degraded_time['requests']} запросов за 30 секунд")
        print(f"   P95 время ответа: {first_degraded_time['p95']:.0f} ms")
        print(f"   P99 время ответа: {first_degraded_time['p99']:.0f} ms")
        
        # Находим интервал перед деградацией для сравнения
        previous_windows = stats_df[stats_df['time_seconds'] < first_degraded_time['time_seconds']]
        if not previous_windows.empty:
            last_good_window = previous_windows.iloc[-1]
            print(f"   📊 СРАВНЕНИЕ:")
            print(f"      До деградации: P95 = {last_good_window['p95']:.0f} ms")
            print(f"      В момент деградации: P95 = {first_degraded_time['p95']:.0f} ms")
            print(f"      Ухудшение: +{first_degraded_time['p95'] - last_good_window['p95']:.0f} ms")
    
    # Деградация по ошибкам
    if not degraded_error_windows.empty:
        first_degraded_errors = degraded_error_windows.iloc[0]
        print(f"🚨 ДЕГРАДАЦИЯ ПО ОШИБКАМ:")
        print(f"   Время: {first_degraded_errors['time_seconds']} секунд от начала теста")
        print(f"   Нагрузка: {first_degraded_errors['requests']} запросов за 30 секунд")
        print(f"   Процент ошибок: {first_degraded_errors['error_rate']:.1f}%")
        print(f"   Количество ошибок: {first_degraded_errors['error_count']}")
    
    if not degraded_time_windows.empty and not degraded_error_windows.empty:
        # Какая деградация произошла первой
        first_degradation = min(
            degraded_time_windows.iloc[0]['time_seconds'] if not degraded_time_windows.empty else float('inf'),
            degraded_error_windows.iloc[0]['time_seconds'] if not degraded_error_windows.empty else float('inf')
        )
        
        if first_degradation == degraded_time_windows.iloc[0]['time_seconds']:
            print(f"\n📅 ПЕРВИЧНАЯ ДЕГРАДАЦИЯ: по времени ответа")
        else:
            print(f"\n📅 ПЕРВИЧНАЯ ДЕГРАДАЦИЯ: по ошибкам")
    
    if not degraded_time_windows.empty or not degraded_error_windows.empty:
        print(f"\n💡 ВЫВОД: Система начинает деградировать при нагрузке ~{max(stats_df['requests'])} запросов/30сек")
    else:
        print("✅ Деградация не обнаружена")
        best_window = stats_df.loc[stats_df['requests'].idxmax()]
        print(f"   Максимальная нагрузка: {best_window['requests']} запросов за 30 секунд")
        print(f"   P95 при макс. нагрузке: {best_window['p95']:.0f} ms")
        print(f"   Ошибок при макс. нагрузке: {best_window['error_rate']:.1f}%")
    
    print(f"\n📈 ОБЩАЯ СТАТИСТИКА:")
    print(f"   Всего проанализировано интервалов: {len(stats_df)}")
    print(f"   Всего запросов: {total_requests:,}")
    print(f"   Всего ошибок: {total_error_count} ({overall_error_rate:.1f}%)")
    print(f"   Средний P95: {stats_df['p95'].mean():.0f} ms")
    print(f"   Максимальный P95: {stats_df['p95'].max():.0f} ms")
    print(f"   Максимальный процент ошибок: {stats_df['error_rate'].max():.1f}%")


def create_percentiles_chart(results_dir, csv_data):
    """Создает график с 5 линейными функциями перцентилей по времени тестирования"""
    
    print("Создание графика с 5 функциями перцентилей...")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    if duration_data.empty:
        print("Нет данных о времени ответа")
        return
    
    # Подготовка данных
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data['timestamp'] = pd.to_numeric(duration_data['timestamp'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value', 'timestamp'])
    
    if duration_data.empty:
        return
    
    # Сортируем по времени и создаем временные интервалы
    duration_data = duration_data.sort_values('timestamp')
    duration_data['time_interval'] = (duration_data['timestamp'] - duration_data['timestamp'].min())
    
    # Используем все доступное время, убираем ограничение на 600 секунд
    max_time = duration_data['time_interval'].max()
    # duration_data = duration_data[duration_data['time_interval'] <= max_time]  # теперь используем все данные
    
    # Разбиваем на временные интервалы до max_time
    time_intervals = np.linspace(0, max_time, min(100, int(max_time) + 1))  # используем min(100, max_time+1) для адаптивности
    percentiles = [50, 75, 90, 95, 99]
    colors = ['green', 'blue', 'orange', 'red', 'purple']
    labels = ['P50', 'P75', 'P90', 'P95', 'P99']
    
    # Вычисляем перцентили для каждого временного интервала
    percentile_over_time = {p: [] for p in percentiles}
    
    for i in range(1, len(time_intervals)):
        time_start = time_intervals[i-1]
        time_end = time_intervals[i]
        
        # Данные за текущий интервал времени
        interval_data = duration_data[
            (duration_data['time_interval'] >= time_start) & 
            (duration_data['time_interval'] < time_end)
        ]['metric_value']
        
        if len(interval_data) > 0:
            for p in percentiles:
                p_value = np.percentile(interval_data, p)
                percentile_over_time[p].append(p_value)
        else:
            for p in percentiles:
                percentile_over_time[p].append(np.nan)
    
    # Создаем график
    fig, ax = plt.subplots(figsize=(14, 8))
    
    # Рисуем 5 линейных функций
    for p, color, label in zip(percentiles, colors, labels):
        if len(percentile_over_time[p]) > 0:
            ax.plot(time_intervals[1:], percentile_over_time[p], 
                   color=color, linewidth=2, label=label, marker='o', markersize=3)
    
    # Настройка графика с правильными подписями
    ax.set_xlabel('Время тестирования (секунды)', fontsize=12, fontweight='bold')
    ax.set_ylabel('Время ответа (миллисекунды)', fontsize=12, fontweight='bold')
    ax.set_title('Динамика перцентилей времени ответа во время тестирования', 
                fontsize=14, fontweight='bold')
    
    # Устанавливаем правильные пределы по оси X
    ax.set_xlim(0, max_time)
    
    # Адаптивные подписи тиков на оси X (каждые 100 секунд или другое значение в зависимости от общего времени)
    tick_interval = max(1, int(max_time / 10))  # определяем интервал тиков в зависимости от продолжительности теста
    x_ticks = np.arange(0, max_time + tick_interval, tick_interval)
    ax.set_xticks(x_ticks)
    ax.set_xticklabels([f'{int(x)}' for x in x_ticks])
    
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    # Добавляем линию порога деградации
    ax.axhline(y=500, color='black', linestyle='--', linewidth=2, 
              alpha=0.7, label='Порог деградации (500 ms)')
    
    plt.tight_layout()
    plot_path = f'{results_dir}/1_five_percentiles_functions.png'
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"График с 5 функциями перцентилей сохранен: {plot_path}")
    print(f"Диапазон времени: 0-{max_time} секунд")
    
    return percentiles, percentile_over_time

def create_cumulative_percentiles_chart(results_dir, csv_data):
    """Создает график с накопленными перцентилями (все данные до текущего момента)"""
    
    print("Создание графика с накопленными перцентилями...")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    if duration_data.empty:
        print("Нет данных о времени ответа")
        return
    
    # Подготовка данных (точно как в рабочем примере)
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data['timestamp'] = pd.to_numeric(duration_data['timestamp'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value', 'timestamp'])
    
    if duration_data.empty:
        return
    
    # Сортируем по времени и создаем временные интервалы (точно как в рабочем примере)
    duration_data = duration_data.sort_values('timestamp')
    duration_data['time_interval'] = (duration_data['timestamp'] - duration_data['timestamp'].min())
    
    # Используем все доступное время, убираем ограничение на 600 секунд
    max_time = duration_data['time_interval'].max()
    # duration_data = duration_data[duration_data['time_interval'] <= max_time]  # теперь используем все данные
    
    # Разбиваем на временные интервалы до max_time (увеличиваем точность для большего количества точек)
    time_intervals = np.linspace(0, max_time, min(100, int(max_time) + 1))  # адаптивное количество точек
    percentiles = [50, 75, 90, 95, 99]
    colors = ['green', 'blue', 'orange', 'red', 'purple']
    labels = ['P50', 'P75', 'P90', 'P95', 'P99']
    
    # ИЗМЕНЕНИЕ: вычисляем НАКОПЛЕННЫЕ перцентили для каждого временного интервала
    percentile_over_time = {p: [] for p in percentiles}
    
    for i in range(1, len(time_intervals)):
        time_end = time_intervals[i]
        
        # Берем ВСЕ данные до этого момента времени (ИЗМЕНЕНИЕ!)
        data_so_far = duration_data[duration_data['time_interval'] <= time_end]['metric_value']
        
        if len(data_so_far) > 0:
            for p in percentiles:
                p_value = np.percentile(data_so_far, p)
                percentile_over_time[p].append(p_value)
        else:
            for p in percentiles:
                percentile_over_time[p].append(np.nan)
    
    # Создаем график (точно как в рабочем примере)
    fig, ax = plt.subplots(figsize=(14, 8))
    
    # Рисуем 5 линейных функций (точно как в рабочем примере)
    for p, color, label in zip(percentiles, colors, labels):
        if len(percentile_over_time[p]) > 0:
            ax.plot(time_intervals[1:], percentile_over_time[p], 
                   color=color, linewidth=2, label=label, marker='o', markersize=3)
    
    # Настройка графика (точно как в рабочем примере)
    ax.set_xlabel('Время тестирования (секунды)', fontsize=12, fontweight='bold')
    ax.set_ylabel('Время ответа (миллисекунды)', fontsize=12, fontweight='bold')
    ax.set_title('Накопленные перцентили времени ответа (все данные до текущего момента)', 
                fontsize=14, fontweight='bold')
    
    # Устанавливаем правильные пределы по оси X
    ax.set_xlim(0, max_time)
    
    # Адаптивные подписи тиков на оси X (каждые 100 секунд или другое значение в зависимости от общего времени)
    tick_interval = max(1, int(max_time / 10))  # определяем интервал тиков в зависимости от продолжительности теста
    x_ticks = np.arange(0, max_time + tick_interval, tick_interval)
    ax.set_xticks(x_ticks)
    ax.set_xticklabels([f'{int(x)}' for x in x_ticks])
    
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    # Добавляем линию порога деградации
    ax.axhline(y=500, color='black', linestyle='--', linewidth=2, 
              alpha=0.7, label='Порог деградации (500 ms)')
    
    plt.tight_layout()
    plot_path = f'{results_dir}/1_cumulative_percentiles.png'
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"График с накопленными перцентилями сохранен: {plot_path}")
    print(f"Диапазон времени: 0-{max_time} секунд")
    
    return percentiles, percentile_over_time
def analyze_degradation(results_dir):
    """Основная функция анализа деградации"""
    
    print("Анализ точки деградации производительности...")
    
    # Загрузка результатов
    csv_file = f'{results_dir}/k6_results.csv'
    
    # Загружаем CSV данные k6
    csv_data = pd.DataFrame()
    if os.path.exists(csv_file):
        try:
            csv_data = pd.read_csv(csv_file, low_memory=False)
            print(f"Загружено {len(csv_data)} записей из k6 CSV")
            
            # Отладка структуры данных
            debug_data_structure(results_dir, csv_data)
                
        except Exception as e:
            print(f"Ошибка при чтении CSV: {e}")
            return
    else:
        print(f"Файл {csv_file} не найден")
        return
    
    # Запускаем все анализы
    create_degradation_analysis(results_dir, csv_data)
    find_degradation_point(results_dir, csv_data)
    create_percentiles_chart(results_dir, csv_data)
    create_cumulative_percentiles_chart(results_dir, csv_data)
    
    print(f"\n✅ Анализ завершен! Все графики сохранены в: {results_dir}")

if __name__ == '__main__':
    if len(sys.argv) > 1:
        analyze_degradation(sys.argv[1])
    else:
        print("Usage: python3 degradation_analysis.py <results_directory>")
        print("Example: python3 degradation_analysis.py /path/to/test/results")