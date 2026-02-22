#!/usr/bin/env python3

import pandas as pd
import numpy as np
import sys
import os
import matplotlib.pyplot as plt
from common import degradation_report

from common.io import load_k6_results_csv, print_dataframe_overview

def debug_data_structure(results_dir, csv_data):
    """Analysis helper."""
    print_dataframe_overview(csv_data)
    
    if not csv_data.empty:
        print(f"Tipy metrik: {csv_data['metric_name'].unique()}")
        
        error_data = csv_data[csv_data['metric_name'] == 'http_req_failed']
        if not error_data.empty:
            print(f"Analysis message")
            print(f"   - Kolichestvo zapisey: {len(error_data)}")
            print(f"   - Znacheniya: {error_data['metric_value'].unique()[:10]}")
        
        http_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        if not http_requests.empty and 'status' in http_requests.columns:
            print(f"\nHTTP statusy:")
            print(f"   - Unikalnye statusy: {http_requests['status'].unique()}")
            print(f"   - Primery statusov: {http_requests['status'].value_counts().head()}")

def calculate_error_rate(csv_data, time_window, window_start, window_end):
    """Analysis helper."""
    try:
        http_requests = csv_data[
            (csv_data['metric_name'] == 'http_req_duration')
        ].copy()
        
        http_requests.loc[:, 'timestamp_dt'] = pd.to_datetime(http_requests['timestamp'], unit='s', errors='coerce')
        http_requests = http_requests.dropna(subset=['timestamp_dt'])
        
        window_requests = http_requests[
            (http_requests['timestamp_dt'] >= window_start) & 
            (http_requests['timestamp_dt'] < window_end)
        ]
        
        total_requests = len(window_requests)
        
        if total_requests == 0:
            return 0, 0
        
        error_count = 0
        
        if 'status' in window_requests.columns:
            status_series = pd.to_numeric(window_requests['status'], errors='coerce')
            error_requests = window_requests[
                (status_series >= 400) |
                (window_requests['error'].notna())
            ]
            error_count = len(error_requests)
        
        error_rate = (error_count / total_requests * 100) if total_requests > 0 else 0
        
        return error_count, error_rate
        
    except Exception as e:
        print(f"Error calculating errors: {e}")
        return 0, 0

def find_degradation_point(results_dir, csv_data):
    """Analysis helper."""
    
    print("Analysis message")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    
    if duration_data.empty:
        print("Analysis message")
        return
    
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value'])
    
    if 'timestamp' in duration_data.columns:
        duration_data.loc[:, 'timestamp_dt'] = pd.to_datetime(duration_data['timestamp'], unit='s', errors='coerce')
        duration_data = duration_data.dropna(subset=['timestamp_dt'])
        duration_data = duration_data.sort_values('timestamp_dt')
        
        duration_data.loc[:, 'time_window'] = (duration_data['timestamp_dt'] - duration_data['timestamp_dt'].min()).dt.total_seconds() // 30
        
        degradation_found_time = False
        degradation_found_errors = False
        degradation_time_time = None
        degradation_time_errors = None
        degradation_load_time = None
        degradation_load_errors = None
        degradation_error_rate = None
        
        print("Analysis message")
        print("Analysis message")
        print("-" * 65)
        
        for time_window in sorted(duration_data['time_window'].unique()):
            window_data = duration_data[duration_data['time_window'] == time_window]
            
            if len(window_data) < 5:
                continue
            
            p95 = np.percentile(window_data['metric_value'], 95)
            
            load = len(window_data)
            
            window_start = duration_data['timestamp_dt'].min() + pd.Timedelta(seconds=time_window*30)
            window_end = window_start + pd.Timedelta(seconds=30)
            
            error_count, error_rate = calculate_error_rate(csv_data, time_window, window_start, window_end)
            
            status = "✅ OK"
            degradation_reason = []
            
            if p95 > 500:
                degradation_reason.append("P95 > 500ms")
                if not degradation_found_time:
                    degradation_found_time = True
                    degradation_time_time = time_window * 30
                    degradation_load_time = load
            
            if error_rate > 3.0:
                degradation_reason.append(f"Analysis message")
                if not degradation_found_errors:
                    degradation_found_errors = True
                    degradation_time_errors = time_window * 30
                    degradation_load_errors = load
                    degradation_error_rate = error_rate
            
            if degradation_reason:
                status = f"Analysis message"
            
            print(f"{time_window * 30:8.0f} | {load:8} | {p95:7.0f} | {error_rate:5.1f}% | {status}")
        
        print(f"Analysis message")
        
        if degradation_found_time:
            print(f"Analysis message")
            print(f"Analysis message")
            print(f"Analysis message")
            print(f"   P95 prevysil 500ms")
        
        if degradation_found_errors:
            print(f"Analysis message")
            print(f"Analysis message")
            print(f"Analysis message")
            print(f"Analysis message")
        
        if not degradation_found_time and not degradation_found_errors:
            print(f"Analysis message")
        elif not degradation_found_errors:
            print(f"Analysis message")
    
    return degradation_found_time or degradation_found_errors

def create_degradation_analysis(results_dir, csv_data):
    """Analysis helper."""
    
    print("Analysis message")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    
    if duration_data.empty:
        print("Analysis message")
        return
    
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value'])
    
    if 'timestamp' not in duration_data.columns:
        print("Analysis message")
        return
    
    duration_data.loc[:, 'timestamp_dt'] = pd.to_datetime(duration_data['timestamp'], unit='s', errors='coerce')
    duration_data = duration_data.dropna(subset=['timestamp_dt'])
    duration_data = duration_data.sort_values('timestamp_dt')
    
    duration_data.loc[:, 'time_elapsed'] = (duration_data['timestamp_dt'] - duration_data['timestamp_dt'].min()).dt.total_seconds()
    duration_data.loc[:, 'time_window'] = (duration_data['time_elapsed'] // 30).astype(int)
    
    window_stats = []
    
    for window in sorted(duration_data['time_window'].unique()):
        window_data = duration_data[duration_data['time_window'] == window]
        
        if len(window_data) < 5:
            continue
        
        p50 = np.percentile(window_data['metric_value'], 50)
        p75 = np.percentile(window_data['metric_value'], 75) 
        p90 = np.percentile(window_data['metric_value'], 90)
        p95 = np.percentile(window_data['metric_value'], 95)
        p99 = np.percentile(window_data['metric_value'], 99)
        
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
            'degraded_time': p95 > 500,  # Analysis step.
            'degraded_errors': error_rate > 0.5
        })
    
    if not window_stats:
        print("Analysis message")
        return
    
    stats_df = pd.DataFrame(window_stats)
    
    fig, (ax1, ax2, ax3) = plt.subplots(3, 1, figsize=(15, 12))
    
    ax1.plot(stats_df['time_seconds'], stats_df['p95'], 
             linewidth=3, color='blue', marker='o', label='Analysis label')
    
    ax1.axhline(y=500, color='red', linestyle='--', linewidth=2, 
                label='Analysis label')
    
    degraded_time_windows = stats_df[stats_df['degraded_time']]
    if not degraded_time_windows.empty:
        first_degradation_time = degraded_time_windows.iloc[0]
        ax1.axvline(x=first_degradation_time['time_seconds'], color='orange', 
                   linestyle=':', linewidth=2, alpha=0.7,
                   label=f'Analysis label')
        
        degradation_start = first_degradation_time['time_seconds']
        degradation_end = stats_df['time_seconds'].max()
        ax1.axvspan(degradation_start, degradation_end, alpha=0.2, color='red', 
                   label='Analysis label')
    
    ax1.set_title('Analysis label', fontsize=14, fontweight='bold')
    ax1.set_ylabel('Analysis label', fontsize=12)
    ax1.legend()
    ax1.grid(True, alpha=0.3)
    
    ax2.plot(stats_df['time_seconds'], stats_df['error_rate'], 
             linewidth=3, color='red', marker='s', label='Analysis label')
    
    ax2.axhline(y=0.5, color='darkred', linestyle='--', linewidth=2, 
                label='Analysis label')
    
    degraded_error_windows = stats_df[stats_df['degraded_errors']]
    if not degraded_error_windows.empty:
        first_degradation_errors = degraded_error_windows.iloc[0]
        ax2.axvline(x=first_degradation_errors['time_seconds'], color='purple', 
                   linestyle=':', linewidth=2, alpha=0.7,
                   label=f'Analysis label')
    
    ax2.set_title('Analysis label', fontsize=14, fontweight='bold')
    ax2.set_ylabel('Analysis label', fontsize=12)
    ax2.set_xlabel('Analysis label', fontsize=12)
    ax2.legend()
    ax2.grid(True, alpha=0.3)
    
    bars = ax3.bar(stats_df['time_seconds'], stats_df['requests'], 
            width=25, alpha=0.7, color='green', label='Analysis label')

    for bar, req_count in zip(bars, stats_df['requests']):
        height = bar.get_height()
        ax3.text(bar.get_x() + bar.get_width()/2., height,
                f'{int(req_count)}',
                ha='center', va='bottom', fontsize=8, fontweight='bold')

    if not degraded_time_windows.empty:
        first_degradation = degraded_time_windows.iloc[0]
        degradation_time = first_degradation['time_seconds']
        degradation_requests = first_degradation['requests']
        
        ax3.axvline(x=degradation_time, color='orange', 
                linestyle=':', linewidth=2, alpha=0.7)
        
        ax3.plot(degradation_time, degradation_requests, 'ro', markersize=8, 
                markerfacecolor='red', markeredgecolor='darkred', markeredgewidth=2)
        
        ax3.axhline(y=degradation_requests, color='red', linestyle='--', 
                alpha=0.5, linewidth=1)

    ax3.set_title('Analysis label', fontsize=14, fontweight='bold')
    ax3.set_xlabel('Analysis label', fontsize=12)
    ax3.set_ylabel('Analysis label', fontsize=12)

    legend_elements = [
        plt.Line2D([0], [0], color='green', alpha=0.7, linewidth=10, label='Analysis label'),
    ]

    if not degraded_time_windows.empty:
        first_degradation = degraded_time_windows.iloc[0]
        degradation_requests = first_degradation['requests']
        
        legend_elements.extend([
            plt.Line2D([0], [0], color='orange', linestyle=':', linewidth=2, 
                    label=f'Analysis label'),
            plt.Line2D([0], [0], marker='o', color='red', markersize=8,
                    label=f'Analysis label'),
            plt.Line2D([0], [0], color='red', linestyle='--', linewidth=1,
                    label='Analysis label')
        ])

    ax3.legend(handles=legend_elements, loc='upper left')
    ax3.grid(True, alpha=0.3)
    
    total_requests = len(duration_data)
    
    total_error_count = 0
    for window in window_stats:
        total_error_count += window['error_count']
    
    overall_error_rate = (total_error_count / total_requests * 100) if total_requests > 0 else 0
    max_p95 = stats_df['p95'].max()
    max_error_rate = stats_df['error_rate'].max()
    
    info_text = (f"Analysis message"
                f"Analysis message"
                f"Analysis message")
    
    fig.suptitle('Analysis label', fontsize=16, fontweight='bold')
    fig.text(0.5, 0.01, info_text, ha='center', fontsize=11,
             bbox=dict(boxstyle="round,pad=0.5", facecolor="lightgray", alpha=0.7))
    
    plt.tight_layout()
    degradation_plot = f'Analysis label'
    plt.savefig(degradation_plot, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Analysis message")
    
    print("\n" + "="*70)
    print("Analysis message")
    print("="*70)
    
    if not degraded_time_windows.empty:
        first_degraded_time = degraded_time_windows.iloc[0]
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
        
        previous_windows = stats_df[stats_df['time_seconds'] < first_degraded_time['time_seconds']]
        if not previous_windows.empty:
            last_good_window = previous_windows.iloc[-1]
            print(f"   📊 SRAVNENIE:")
            print(f"Analysis message")
            print(f"Analysis message")
            print(f"      Ukhudshenie: +{first_degraded_time['p95'] - last_good_window['p95']:.0f} ms")
    
    if not degraded_error_windows.empty:
        first_degraded_errors = degraded_error_windows.iloc[0]
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
    
    if not degraded_time_windows.empty and not degraded_error_windows.empty:
        first_degradation = min(
            degraded_time_windows.iloc[0]['time_seconds'] if not degraded_time_windows.empty else float('inf'),
            degraded_error_windows.iloc[0]['time_seconds'] if not degraded_error_windows.empty else float('inf')
        )
        
        if first_degradation == degraded_time_windows.iloc[0]['time_seconds']:
            print(f"Analysis message")
        else:
            print(f"Analysis message")
    
    if not degraded_time_windows.empty or not degraded_error_windows.empty:
        print(f"Analysis message")
    else:
        print("Analysis message")
        best_window = stats_df.loc[stats_df['requests'].idxmax()]
        print(f"Analysis message")
        print(f"Analysis message")
        print(f"Analysis message")
    
    print(f"\n📈 OBShchAYa STATISTIKA:")
    print(f"Analysis message")
    print(f"Analysis message")
    print(f"Analysis message")
    print(f"   Sredniy P95: {stats_df['p95'].mean():.0f} ms")
    print(f"   Maksimalnyy P95: {stats_df['p95'].max():.0f} ms")
    print(f"Analysis message")


def create_percentiles_chart(results_dir, csv_data):
    """Analysis helper."""
    
    print("Analysis message")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    if duration_data.empty:
        print("Analysis message")
        return
    
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data['timestamp'] = pd.to_numeric(duration_data['timestamp'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value', 'timestamp'])
    
    if duration_data.empty:
        return
    
    duration_data = duration_data.sort_values('timestamp')
    duration_data['time_interval'] = (duration_data['timestamp'] - duration_data['timestamp'].min())
    
    max_time = duration_data['time_interval'].max()
    
    time_intervals = np.linspace(0, max_time, min(100, int(max_time) + 1))  # ispolzuem min(100, max_time+1) dlya adaptivnosti
    percentiles = [50, 75, 90, 95, 99]
    colors = ['green', 'blue', 'orange', 'red', 'purple']
    labels = ['P50', 'P75', 'P90', 'P95', 'P99']
    
    percentile_over_time = {p: [] for p in percentiles}
    
    for i in range(1, len(time_intervals)):
        time_start = time_intervals[i-1]
        time_end = time_intervals[i]
        
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
    
    fig, ax = plt.subplots(figsize=(14, 8))
    
    for p, color, label in zip(percentiles, colors, labels):
        if len(percentile_over_time[p]) > 0:
            ax.plot(time_intervals[1:], percentile_over_time[p], 
                   color=color, linewidth=2, label=label, marker='o', markersize=3)
    
    ax.set_xlabel('Analysis label', fontsize=12, fontweight='bold')
    ax.set_ylabel('Analysis label', fontsize=12, fontweight='bold')
    ax.set_title('Analysis label', 
                fontsize=14, fontweight='bold')
    
    ax.set_xlim(0, max_time)
    
    tick_interval = max(1, int(max_time / 10))  # opredelyaem interval tikov v zavisimosti ot prodolzhitelnosti testa
    x_ticks = np.arange(0, max_time + tick_interval, tick_interval)
    ax.set_xticks(x_ticks)
    ax.set_xticklabels([f'{int(x)}' for x in x_ticks])
    
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    ax.axhline(y=500, color='black', linestyle='--', linewidth=2, 
              alpha=0.7, label='Analysis label')
    
    plt.tight_layout()
    plot_path = f'{results_dir}/1_five_percentiles_functions.png'
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Analysis message")
    print(f"Analysis message")
    
    return percentiles, percentile_over_time

def create_cumulative_percentiles_chart(results_dir, csv_data):
    """Analysis helper."""
    
    print("Analysis message")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    if duration_data.empty:
        print("Analysis message")
        return
    
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data['timestamp'] = pd.to_numeric(duration_data['timestamp'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value', 'timestamp'])
    
    if duration_data.empty:
        return
    
    duration_data = duration_data.sort_values('timestamp')
    duration_data['time_interval'] = (duration_data['timestamp'] - duration_data['timestamp'].min())
    
    max_time = duration_data['time_interval'].max()
    
    time_intervals = np.linspace(0, max_time, min(100, int(max_time) + 1))  # adaptivnoe kolichestvo tochek
    percentiles = [50, 75, 90, 95, 99]
    colors = ['green', 'blue', 'orange', 'red', 'purple']
    labels = ['P50', 'P75', 'P90', 'P95', 'P99']
    
    percentile_over_time = {p: [] for p in percentiles}
    
    for i in range(1, len(time_intervals)):
        time_end = time_intervals[i]
        
        data_so_far = duration_data[duration_data['time_interval'] <= time_end]['metric_value']
        
        if len(data_so_far) > 0:
            for p in percentiles:
                p_value = np.percentile(data_so_far, p)
                percentile_over_time[p].append(p_value)
        else:
            for p in percentiles:
                percentile_over_time[p].append(np.nan)
    
    fig, ax = plt.subplots(figsize=(14, 8))
    
    for p, color, label in zip(percentiles, colors, labels):
        if len(percentile_over_time[p]) > 0:
            ax.plot(time_intervals[1:], percentile_over_time[p], 
                   color=color, linewidth=2, label=label, marker='o', markersize=3)
    
    ax.set_xlabel('Analysis label', fontsize=12, fontweight='bold')
    ax.set_ylabel('Analysis label', fontsize=12, fontweight='bold')
    ax.set_title('Analysis label', 
                fontsize=14, fontweight='bold')
    
    ax.set_xlim(0, max_time)
    
    tick_interval = max(1, int(max_time / 10))  # opredelyaem interval tikov v zavisimosti ot prodolzhitelnosti testa
    x_ticks = np.arange(0, max_time + tick_interval, tick_interval)
    ax.set_xticks(x_ticks)
    ax.set_xticklabels([f'{int(x)}' for x in x_ticks])
    
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    ax.axhline(y=500, color='black', linestyle='--', linewidth=2, 
              alpha=0.7, label='Analysis label')
    
    plt.tight_layout()
    plot_path = f'{results_dir}/1_cumulative_percentiles.png'
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Analysis message")
    print(f"Analysis message")
    
    return percentiles, percentile_over_time
def generate_copy_report(results_dir):
    """Analysis helper."""
    
    print("Analysis message")
    
    csv_file = f'{results_dir}/k6_results.csv'

    csv_data = pd.DataFrame()
    if os.path.exists(csv_file):
        try:
            csv_data = load_k6_results_csv(results_dir)
            print(f"Analysis message")
            
            debug_data_structure(results_dir, csv_data)
                
        except Exception as e:
            print(f"Error reading CSV: {e}")
            return
    else:
        print(f"Fayl {csv_file} ne nayden")
        return
    
    create_degradation_analysis(results_dir, csv_data)
    find_degradation_point(results_dir, csv_data)
    create_percentiles_chart(results_dir, csv_data)
    create_cumulative_percentiles_chart(results_dir, csv_data)
    
    print(f"Analysis message")

def main(argv):
    if len(argv) < 2:
        print("Analysis message")
        print("Example: python3 report.py copy /path/to/test/results")
        return 1

    if len(argv) == 2:
        mode = "copy"
        results_dir = argv[1]
    else:
        mode = argv[1]
        results_dir = argv[2]

    if mode == "copy":
        generate_copy_report(results_dir)
        return 0
    if mode == "Analysis message":
        degradation_report.generate_degradation_report(results_dir)
        return 0

    print(f"Unknown mode: {mode}")
    print("Analysis message")
    return 1


if __name__ == '__main__':
    raise SystemExit(main(sys.argv))
