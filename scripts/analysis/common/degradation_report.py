#!/usr/bin/env python3

import json
import pandas as pd
import numpy as np
import sys
import os
from datetime import datetime, timedelta
import matplotlib.pyplot as plt
import seaborn as sns

from common.io import (
    convert_numeric_columns,
    load_k6_results_csv,
    load_resource_usage_csv,
    print_dataframe_overview,
)

def debug_data_structure(results_dir, csv_data):
    """Analysis helper."""
    print_dataframe_overview(csv_data)
    
    if not csv_data.empty:
        print(f"Tipy metrik: {csv_data['metric_name'].unique()}")
        
        http_reqs = csv_data[csv_data['metric_name'] == 'http_reqs']
        if not http_reqs.empty:
            print(f"Analysis message")
            print(f"   - Kolichestvo zapisey: {len(http_reqs)}")
            print(f"   - Znacheniya: {http_reqs['metric_value'].unique()[:10]}")  # pervye 10 znacheniy
            if 'timestamp' in http_reqs.columns:
                print(f"   - Vremennye metki: {http_reqs['timestamp'].head(3)}")

def create_response_time_plots(results_dir, csv_data):
    """Analysis helper."""
    
    if csv_data.empty:
        print("Analysis message")
        return
    
    print("Analysis message")
    
    fig, axes = plt.subplots(2, 1, figsize=(14, 10))
    fig.suptitle('Analysis label', fontsize=16, fontweight='bold')
    
    try:
        completed_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        
        if not completed_requests.empty:
            percentiles = [50, 75, 90, 95, 99]
            percentile_values = [np.percentile(completed_requests['metric_value'], p) for p in percentiles]
            
            bars = axes[0].bar(range(len(percentiles)), percentile_values,
                             color=['lightgreen', 'lightblue', 'orange', 'red', 'darkred'],
                             alpha=0.7, edgecolor='black')
            
            axes[0].set_title('Analysis label')
            axes[0].set_ylabel('Analysis label')
            axes[0].set_xlabel('Pertsentil')
            axes[0].set_xticks(range(len(percentiles)))
            axes[0].set_xticklabels([f'P{p}' for p in percentiles])
            
            for i, bar in enumerate(bars):
                height = bar.get_height()
                axes[0].text(bar.get_x() + bar.get_width()/2., height + 5,
                           f'{height:.0f}ms', ha='center', va='bottom', fontweight='bold')
            
            axes[0].grid(True, alpha=0.3)
        else:
            axes[0].text(0.5, 0.5, 'Analysis label', 
                        ha='center', va='center', transform=axes[0].transAxes)
        
    except Exception as e:
        axes[0].text(0.5, 0.5, f'Error pertsentiley: {e}', 
                    ha='center', va='center', transform=axes[0].transAxes)
    
    try:
        completed_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        
        if not completed_requests.empty:
            durations = completed_requests['metric_value']
            
            axes[1].hist(durations, bins=50, alpha=0.7, color='teal', edgecolor='black')
            axes[1].set_title('Analysis label')
            axes[1].set_ylabel('Analysis label')
            axes[1].set_xlabel('Analysis label')
            axes[1].grid(True, alpha=0.3)
            
            mean_duration = durations.mean()
            median_duration = durations.median()
            p95_duration = np.percentile(durations, 95)
            
            axes[1].axvline(mean_duration, color='red', linestyle='--', 
                           label=f'Srednee: {mean_duration:.0f}ms')
            axes[1].axvline(median_duration, color='green', linestyle='--', 
                           label=f'Mediana: {median_duration:.0f}ms')
            axes[1].axvline(p95_duration, color='orange', linestyle='--', 
                           label=f'P95: {p95_duration:.0f}ms')
            axes[1].legend()
        else:
            axes[1].text(0.5, 0.5, 'Analysis label', 
                        ha='center', va='center', transform=axes[1].transAxes)
        
    except Exception as e:
        axes[1].text(0.5, 0.5, f'Error gistogrammy: {e}', 
                    ha='center', va='center', transform=axes[1].transAxes)
    
    plt.tight_layout()
    response_time_plot = f'{results_dir}/response_time_analysis.png'
    plt.savefig(response_time_plot, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Analysis message")


def create_container_plots(results_dir, resource_data):
    """Analysis helper."""
    
    if resource_data.empty:
        print("Analysis message")
        return
    
    containers = resource_data['container'].unique()
    
    for container in containers:
        if container not in ['go_app', 'postgres_db']:
            continue
            
        container_data = resource_data[resource_data['container'] == container]
        if container_data.empty:
            continue
            
        print(f"Analysis message")
        
        fig, axes = plt.subplots(2, 1, figsize=(15, 12))
        fig.suptitle(f'Analysis label', fontsize=16, fontweight='bold')
        
        container_data = container_data.copy()
        
        if 'timestamp' in container_data.columns:
            container_data['timestamp'] = pd.to_numeric(container_data['timestamp'], errors='coerce')
            container_data = container_data.dropna(subset=['timestamp'])
            start_time = container_data['timestamp'].min()
            real_time = (container_data['timestamp'] - start_time) / 1000
        else:
            real_time = range(len(container_data))
        
        try:
            container_data['cpu_numeric'] = container_data['cpu_percent'].str.replace('%', '').astype(float)
            axes[0].plot(real_time, container_data['cpu_numeric'], 
                        linewidth=2, color='red', marker='o', markersize=2)
            axes[0].set_title(f'{container} - Ispolzovanie CPU')
            axes[0].set_ylabel('CPU %')
            axes[0].set_xlabel('Analysis label')  # ISPRAVLENO!
            axes[0].grid(True, alpha=0.3)
            
            cpu_mean = container_data['cpu_numeric'].mean()
            cpu_max = container_data['cpu_numeric'].max()
            axes[0].axhline(y=cpu_mean, color='blue', linestyle='--', alpha=0.7, 
                           label=f'Srednee: {cpu_mean:.1f}%')
            axes[0].axhline(y=cpu_max, color='orange', linestyle='--', alpha=0.7, 
                           label=f'Maksimum: {cpu_max:.1f}%')
            axes[0].legend()
        except Exception as e:
            axes[0].text(0.5, 0.5, f'Error CPU: {e}', 
                        ha='center', va='center', transform=axes[0].transAxes)
        
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
            axes[1].set_title(f'{container} - Ispolzovanie pamyati')
            axes[1].set_ylabel('Pamyat (MB)')
            axes[1].set_xlabel('Analysis label')  # ISPRAVLENO!
            axes[1].grid(True, alpha=0.3)
            
            mem_mean = container_data['mem_used'].mean()
            mem_max = container_data['mem_used'].max()
            axes[1].axhline(y=mem_mean, color='blue', linestyle='--', alpha=0.7, 
                           label=f'Srednee: {mem_mean:.1f} MB')
            axes[1].axhline(y=mem_max, color='orange', linestyle='--', alpha=0.7, 
                           label=f'Maksimum: {mem_max:.1f} MB')
            axes[1].legend()
            
        except Exception as e:
            axes[1].text(0.5, 0.5, f'Error pamyati: {e}', 
                        ha='center', va='center', transform=axes[1].transAxes)
        
        plt.tight_layout()
        container_plot_path = f'{results_dir}/container_{container}_analysis.png'
        plt.savefig(container_plot_path, dpi=300, bbox_inches='tight')
        plt.close()
        
        print(f"Analysis message")

def create_plots(results_dir, csv_data, resource_data):
    """Analysis helper."""
    create_response_time_plots(results_dir, csv_data)
    create_container_plots(results_dir, resource_data)

def generate_degradation_report(results_dir):
    print("Analysis message")
    
    csv_file = f'{results_dir}/k6_results.csv'
    resource_file = f'{results_dir}/resource_usage.csv'
    
    csv_data = pd.DataFrame()
    if os.path.exists(csv_file):
        try:
            csv_data = load_k6_results_csv(results_dir)
            csv_data = convert_numeric_columns(csv_data, ['metric_value'])
            
            print(f"Analysis message")
            
            debug_data_structure(results_dir, csv_data)
                
        except Exception as e:
            print(f"Error reading CSV: {e}")
    
    resource_data = pd.DataFrame()
    if os.path.exists(resource_file):
        try:
            resource_data = load_resource_usage_csv(results_dir)
            print(f"Analysis message")
            
        except Exception as e:
            print(f"Error reading resources: {e}")

    create_plots(results_dir, csv_data, resource_data)

    print(f"Analysis message")
