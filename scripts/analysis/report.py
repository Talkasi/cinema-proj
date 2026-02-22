#!/usr/bin/env python3
# degradation_analysis.py

import pandas as pd
import numpy as np
import sys
import os
import matplotlib.pyplot as plt
from common import degradation_report

from common.io import load_k6_results_csv, print_dataframe_overview

def debug_data_structure(results_dir, csv_data):
    """Funktsiya dlya otladki struktury dannykh"""
    print_dataframe_overview(csv_data)
    
    if not csv_data.empty:
        print(f"Tipy metrik: {csv_data['metric_name'].unique()}")
        
        # Smotrim na dannye oshibok
        error_data = csv_data[csv_data['metric_name'] == 'http_req_failed']
        if not error_data.empty:
            print(f"\nHTTP_REQ_FAILED dannye:")
            print(f"   - Kolichestvo zapisey: {len(error_data)}")
            print(f"   - Znacheniya: {error_data['metric_value'].unique()[:10]}")
        
        # Smotrim na statusy HTTP zaprosov
        http_requests = csv_data[csv_data['metric_name'] == 'http_req_duration']
        if not http_requests.empty and 'status' in http_requests.columns:
            print(f"\nHTTP statusy:")
            print(f"   - Unikalnye statusy: {http_requests['status'].unique()}")
            print(f"   - Primery statusov: {http_requests['status'].value_counts().head()}")

def calculate_error_rate(csv_data, time_window, window_start, window_end):
    """Pravilno vychislyaet protsent oshibok dlya vremennogo intervala"""
    try:
        # Nakhodim vse HTTP zaprosy v etom intervale
        http_requests = csv_data[
            (csv_data['metric_name'] == 'http_req_duration')
        ].copy()
        
        # Konvertiruem timestamp dlya filtratsii
        http_requests.loc[:, 'timestamp_dt'] = pd.to_datetime(http_requests['timestamp'], unit='s', errors='coerce')
        http_requests = http_requests.dropna(subset=['timestamp_dt'])
        
        # Filtruem po vremennomu oknu
        window_requests = http_requests[
            (http_requests['timestamp_dt'] >= window_start) & 
            (http_requests['timestamp_dt'] < window_end)
        ]
        
        total_requests = len(window_requests)
        
        if total_requests == 0:
            return 0, 0
        
        # Schitaem oshibki po statusam (4xx, 5xx) ili nalichiyu error
        error_count = 0
        
        if 'status' in window_requests.columns:
            # Konvertiruem statusy v chislovoy format bez preduprezhdeniy
            status_series = pd.to_numeric(window_requests['status'], errors='coerce')
            error_requests = window_requests[
                (status_series >= 400) |  # HTTP oshibki
                (window_requests['error'].notna())  # Est tekst oshibki
            ]
            error_count = len(error_requests)
        
        error_rate = (error_count / total_requests * 100) if total_requests > 0 else 0
        
        return error_count, error_rate
        
    except Exception as e:
        print(f"Error calculating errors: {e}")
        return 0, 0

def find_degradation_point(results_dir, csv_data):
    """Nakhodit tochku degradatsii proizvoditelnosti po vremeni otveta i oshibkam"""
    
    print("Poisk tochki degradatsii...")
    
    # Filtruem dannye
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    
    if duration_data.empty:
        print("Net dannykh o vremeni otveta")
        return
    
    # Podgotavlivaem dannye
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value'])
    
    # Gruppiruem po vremennym intervalam
    if 'timestamp' in duration_data.columns:
        duration_data.loc[:, 'timestamp_dt'] = pd.to_datetime(duration_data['timestamp'], unit='s', errors='coerce')
        duration_data = duration_data.dropna(subset=['timestamp_dt'])
        duration_data = duration_data.sort_values('timestamp_dt')
        
        # Sozdaem vremennye intervaly (kazhdye 30 sekund)
        duration_data.loc[:, 'time_window'] = (duration_data['timestamp_dt'] - duration_data['timestamp_dt'].min()).dt.total_seconds() // 30
        
        # Analiziruem kazhdyy vremennoy interval
        degradation_found_time = False
        degradation_found_errors = False
        degradation_time_time = None
        degradation_time_errors = None
        degradation_load_time = None
        degradation_load_errors = None
        degradation_error_rate = None
        
        print("\nAnaliz vremennykh intervalov:")
        print("Vremya(sek) | Zaprosov | P95(ms) | Oshibok | Status")
        print("-" * 65)
        
        for time_window in sorted(duration_data['time_window'].unique()):
            window_data = duration_data[duration_data['time_window'] == time_window]
            
            if len(window_data) < 5:  # Minimum 5 zaprosov dlya statistiki
                continue
            
            # Vychislyaem P95 dlya etogo intervala
            p95 = np.percentile(window_data['metric_value'], 95)
            
            # Otsenivaem nagruzku (kolichestvo zaprosov v intervale)
            load = len(window_data)
            
            # Vychislyaem protsent oshibok dlya etogo intervala
            window_start = duration_data['timestamp_dt'].min() + pd.Timedelta(seconds=time_window*30)
            window_end = window_start + pd.Timedelta(seconds=30)
            
            error_count, error_rate = calculate_error_rate(csv_data, time_window, window_start, window_end)
            
            # Proveryaem usloviya degradatsii
            status = "✅ OK"
            degradation_reason = []
            
            if p95 > 500:
                degradation_reason.append("P95 > 500ms")
                if not degradation_found_time:
                    degradation_found_time = True
                    degradation_time_time = time_window * 30
                    degradation_load_time = load
            
            if error_rate > 3.0:
                degradation_reason.append(f"Oshibok > 0.5% ({error_rate:.1f}%)")
                if not degradation_found_errors:
                    degradation_found_errors = True
                    degradation_time_errors = time_window * 30
                    degradation_load_errors = load
                    degradation_error_rate = error_rate
            
            if degradation_reason:
                status = f"❌ DEGRADATsIYa ({', '.join(degradation_reason)})"
            
            print(f"{time_window * 30:8.0f} | {load:8} | {p95:7.0f} | {error_rate:5.1f}% | {status}")
        
        # Vyvodim rezultaty
        print(f"\n📊 REZULTATY POISKA DEGRADATsII:")
        
        if degradation_found_time:
            print(f"🚨 DEGRADATsIYa PO VREMENI OTVETA:")
            print(f"   Vremya: {degradation_time_time} sekund ot nachala testa")
            print(f"   Nagruzka: ~{degradation_load_time} zaprosov/30sek")
            print(f"   P95 prevysil 500ms")
        
        if degradation_found_errors:
            print(f"🚨 DEGRADATsIYa PO OShIBKAM:")
            print(f"   Vremya: {degradation_time_errors} sekund ot nachala testa")
            print(f"   Nagruzka: ~{degradation_load_errors} zaprosov/30sek")
            print(f"   Protsent oshibok: {degradation_error_rate:.1f}%")
        
        if not degradation_found_time and not degradation_found_errors:
            print(f"✅ Degradatsiya ne obnaruzhena v predelakh testa")
        elif not degradation_found_errors:
            print(f"ℹ️  Oshibok ne obnaruzheno (vse zaprosy uspeshny)")
    
    return degradation_found_time or degradation_found_errors

def create_degradation_analysis(results_dir, csv_data):
    """Sozdaet analiz dlya poiska tochki degradatsii po vremeni otveta i oshibkam"""
    
    print("Sozdanie analiza tochki degradatsii...")
    
    # Filtruem dannye vremeni otveta
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    
    if duration_data.empty:
        print("Net dannykh o vremeni otveta")
        return
    
    # Podgotavlivaem dannye
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value'])
    
    if 'timestamp' not in duration_data.columns:
        print("Net vremennykh metok dlya analiza")
        return
    
    # Konvertiruem timestamp
    duration_data.loc[:, 'timestamp_dt'] = pd.to_datetime(duration_data['timestamp'], unit='s', errors='coerce')
    duration_data = duration_data.dropna(subset=['timestamp_dt'])
    duration_data = duration_data.sort_values('timestamp_dt')
    
    # Sozdaem vremennye intervaly (kazhdye 30 sekund)
    duration_data.loc[:, 'time_elapsed'] = (duration_data['timestamp_dt'] - duration_data['timestamp_dt'].min()).dt.total_seconds()
    duration_data.loc[:, 'time_window'] = (duration_data['time_elapsed'] // 30).astype(int)
    
    # Analiziruem kazhdyy interval
    window_stats = []
    
    for window in sorted(duration_data['time_window'].unique()):
        window_data = duration_data[duration_data['time_window'] == window]
        
        if len(window_data) < 5:  # Minimum 5 zaprosov dlya statistiki
            continue
        
        # Statistika vremeni otveta
        p50 = np.percentile(window_data['metric_value'], 50)
        p75 = np.percentile(window_data['metric_value'], 75) 
        p90 = np.percentile(window_data['metric_value'], 90)
        p95 = np.percentile(window_data['metric_value'], 95)
        p99 = np.percentile(window_data['metric_value'], 99)
        
        # Statistika oshibok
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
            'degraded_time': p95 > 500,  # Degradatsiya po vremeni
            'degraded_errors': error_rate > 0.5  # Degradatsiya po oshibkam
        })
    
    if not window_stats:
        print("Nedostatochno dannykh dlya analiza")
        return
    
    stats_df = pd.DataFrame(window_stats)
    
    # Sozdaem grafik analiza degradatsii
    fig, (ax1, ax2, ax3) = plt.subplots(3, 1, figsize=(15, 12))
    
    # Grafik 1: P95 po vremeni
    ax1.plot(stats_df['time_seconds'], stats_df['p95'], 
             linewidth=3, color='blue', marker='o', label='P95 vremya otveta')
    
    # Liniya poroga degradatsii po vremeni
    ax1.axhline(y=500, color='red', linestyle='--', linewidth=2, 
                label='Porog degradatsii (500ms)')
    
    # Zakrashivaem zonu degradatsii po vremeni
    degraded_time_windows = stats_df[stats_df['degraded_time']]
    if not degraded_time_windows.empty:
        first_degradation_time = degraded_time_windows.iloc[0]
        ax1.axvline(x=first_degradation_time['time_seconds'], color='orange', 
                   linestyle=':', linewidth=2, alpha=0.7,
                   label=f'Degradatsiya vremeni: {first_degradation_time["time_seconds"]}sek')
        
        # Zakrashivaem period degradatsii
        degradation_start = first_degradation_time['time_seconds']
        degradation_end = stats_df['time_seconds'].max()
        ax1.axvspan(degradation_start, degradation_end, alpha=0.2, color='red', 
                   label='Period degradatsii')
    
    ax1.set_title('Vremya otveta P95 i porog degradatsii', fontsize=14, fontweight='bold')
    ax1.set_ylabel('P95 vremya otveta (ms)', fontsize=12)
    ax1.legend()
    ax1.grid(True, alpha=0.3)
    
    # Grafik 2: Protsent oshibok po vremeni
    ax2.plot(stats_df['time_seconds'], stats_df['error_rate'], 
             linewidth=3, color='red', marker='s', label='Protsent oshibok')
    
    # Liniya poroga degradatsii po oshibkam
    ax2.axhline(y=0.5, color='darkred', linestyle='--', linewidth=2, 
                label='Porog degradatsii (0.5% oshibok)')
    
    # Zakrashivaem zonu degradatsii po oshibkam
    degraded_error_windows = stats_df[stats_df['degraded_errors']]
    if not degraded_error_windows.empty:
        first_degradation_errors = degraded_error_windows.iloc[0]
        ax2.axvline(x=first_degradation_errors['time_seconds'], color='purple', 
                   linestyle=':', linewidth=2, alpha=0.7,
                   label=f'Degradatsiya oshibok: {first_degradation_errors["time_seconds"]}sek')
    
    ax2.set_title('Protsent oshibok i porog degradatsii', fontsize=14, fontweight='bold')
    ax2.set_ylabel('Protsent oshibok (%)', fontsize=12)
    ax2.set_xlabel('Vremya ot nachala testa (sekundy)', fontsize=12)
    ax2.legend()
    ax2.grid(True, alpha=0.3)
    
    # Grafik 3: Kolichestvo zaprosov po vremeni (nagruzka)
    bars = ax3.bar(stats_df['time_seconds'], stats_df['requests'], 
            width=25, alpha=0.7, color='green', label='Zaprosov za 30sek')

    # Dobavlyaem znacheniya poverkh stolbtsov
    for bar, req_count in zip(bars, stats_df['requests']):
        height = bar.get_height()
        ax3.text(bar.get_x() + bar.get_width()/2., height,
                f'{int(req_count)}',
                ha='center', va='bottom', fontsize=8, fontweight='bold')

    # Otmechaem tochki degradatsii na grafike nagruzki
    if not degraded_time_windows.empty:
        first_degradation = degraded_time_windows.iloc[0]
        degradation_time = first_degradation['time_seconds']
        degradation_requests = first_degradation['requests']
        
        # Dobavlyaem vertikalnuyu liniyu
        ax3.axvline(x=degradation_time, color='orange', 
                linestyle=':', linewidth=2, alpha=0.7)
        
        # Dobavlyaem tochku na grafike
        ax3.plot(degradation_time, degradation_requests, 'ro', markersize=8, 
                markerfacecolor='red', markeredgecolor='darkred', markeredgewidth=2)
        
        # Dobavlyaem gorizontalnuyu liniyu k znacheniyu Y
        ax3.axhline(y=degradation_requests, color='red', linestyle='--', 
                alpha=0.5, linewidth=1)

    # DOBAVLYaEM PODPISI OSEY I LEGENDU S INFORMATsIEY O ZAPROSAKh
    ax3.set_title('Nagruzka (kolichestvo zaprosov po vremeni)', fontsize=14, fontweight='bold')
    ax3.set_xlabel('Vremya ot nachala testa (sekundy)', fontsize=12)
    ax3.set_ylabel('Kolichestvo zaprosov', fontsize=12)

    # Sozdaem kastomnuyu legendu s informatsiey o degradatsii
    legend_elements = [
        plt.Line2D([0], [0], color='green', alpha=0.7, linewidth=10, label='Zaprosov za 30sek'),
    ]

    if not degraded_time_windows.empty:
        first_degradation = degraded_time_windows.iloc[0]
        degradation_requests = first_degradation['requests']
        
        legend_elements.extend([
            plt.Line2D([0], [0], color='orange', linestyle=':', linewidth=2, 
                    label=f'Degradatsiya: {first_degradation["time_seconds"]:.0f}sek'),
            plt.Line2D([0], [0], marker='o', color='red', markersize=8,
                    label=f'Zaprosov pri degradatsii: {degradation_requests}'),
            plt.Line2D([0], [0], color='red', linestyle='--', linewidth=1,
                    label='Uroven nagruzki pri degradatsii')
        ])

    ax3.legend(handles=legend_elements, loc='upper left')
    ax3.grid(True, alpha=0.3)
    
    # Obshchaya informatsiya
    total_requests = len(duration_data)
    
    # Schitaem obshchee kolichestvo oshibok
    total_error_count = 0
    for window in window_stats:
        total_error_count += window['error_count']
    
    overall_error_rate = (total_error_count / total_requests * 100) if total_requests > 0 else 0
    max_p95 = stats_df['p95'].max()
    max_error_rate = stats_df['error_rate'].max()
    
    info_text = (f"Vsego zaprosov: {total_requests:,} | "
                f"Vsego oshibok: {total_error_count} ({overall_error_rate:.1f}%) | "
                f"Maks P95: {max_p95:.0f}ms | Maks oshibok: {max_error_rate:.1f}%")
    
    fig.suptitle('Analiz tochki degradatsii proizvoditelnosti', fontsize=16, fontweight='bold')
    fig.text(0.5, 0.01, info_text, ha='center', fontsize=11,
             bbox=dict(boxstyle="round,pad=0.5", facecolor="lightgray", alpha=0.7))
    
    plt.tight_layout()
    degradation_plot = f'{results_dir}/1_degradation_analysis.png'
    plt.savefig(degradation_plot, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Grafik analiza degradatsii sokhranen: {degradation_plot}")
    
    # Vyvodim detalnyy otchet
    print("\n" + "="*70)
    print("OTChET O POISKE TOChKI DEGRADATsII")
    print("="*70)
    
    # Degradatsiya po vremeni otveta
    if not degraded_time_windows.empty:
        first_degraded_time = degraded_time_windows.iloc[0]
        print(f"🚨 DEGRADATsIYa PO VREMENI OTVETA:")
        print(f"   Vremya: {first_degraded_time['time_seconds']} sekund ot nachala testa")
        print(f"   Nagruzka: {first_degraded_time['requests']} zaprosov za 30 sekund")
        print(f"   P95 vremya otveta: {first_degraded_time['p95']:.0f} ms")
        print(f"   P99 vremya otveta: {first_degraded_time['p99']:.0f} ms")
        
        # Nakhodim interval pered degradatsiey dlya sravneniya
        previous_windows = stats_df[stats_df['time_seconds'] < first_degraded_time['time_seconds']]
        if not previous_windows.empty:
            last_good_window = previous_windows.iloc[-1]
            print(f"   📊 SRAVNENIE:")
            print(f"      Do degradatsii: P95 = {last_good_window['p95']:.0f} ms")
            print(f"      V moment degradatsii: P95 = {first_degraded_time['p95']:.0f} ms")
            print(f"      Ukhudshenie: +{first_degraded_time['p95'] - last_good_window['p95']:.0f} ms")
    
    # Degradatsiya po oshibkam
    if not degraded_error_windows.empty:
        first_degraded_errors = degraded_error_windows.iloc[0]
        print(f"🚨 DEGRADATsIYa PO OShIBKAM:")
        print(f"   Vremya: {first_degraded_errors['time_seconds']} sekund ot nachala testa")
        print(f"   Nagruzka: {first_degraded_errors['requests']} zaprosov za 30 sekund")
        print(f"   Protsent oshibok: {first_degraded_errors['error_rate']:.1f}%")
        print(f"   Kolichestvo oshibok: {first_degraded_errors['error_count']}")
    
    if not degraded_time_windows.empty and not degraded_error_windows.empty:
        # Kakaya degradatsiya proizoshla pervoy
        first_degradation = min(
            degraded_time_windows.iloc[0]['time_seconds'] if not degraded_time_windows.empty else float('inf'),
            degraded_error_windows.iloc[0]['time_seconds'] if not degraded_error_windows.empty else float('inf')
        )
        
        if first_degradation == degraded_time_windows.iloc[0]['time_seconds']:
            print(f"\n📅 PERVIChNAYa DEGRADATsIYa: po vremeni otveta")
        else:
            print(f"\n📅 PERVIChNAYa DEGRADATsIYa: po oshibkam")
    
    if not degraded_time_windows.empty or not degraded_error_windows.empty:
        print(f"\n💡 VYVOD: Sistema nachinaet degradirovat pri nagruzke ~{max(stats_df['requests'])} zaprosov/30sek")
    else:
        print("✅ Degradatsiya ne obnaruzhena")
        best_window = stats_df.loc[stats_df['requests'].idxmax()]
        print(f"   Maksimalnaya nagruzka: {best_window['requests']} zaprosov za 30 sekund")
        print(f"   P95 pri maks. nagruzke: {best_window['p95']:.0f} ms")
        print(f"   Oshibok pri maks. nagruzke: {best_window['error_rate']:.1f}%")
    
    print(f"\n📈 OBShchAYa STATISTIKA:")
    print(f"   Vsego proanalizirovano intervalov: {len(stats_df)}")
    print(f"   Vsego zaprosov: {total_requests:,}")
    print(f"   Vsego oshibok: {total_error_count} ({overall_error_rate:.1f}%)")
    print(f"   Sredniy P95: {stats_df['p95'].mean():.0f} ms")
    print(f"   Maksimalnyy P95: {stats_df['p95'].max():.0f} ms")
    print(f"   Maksimalnyy protsent oshibok: {stats_df['error_rate'].max():.1f}%")


def create_percentiles_chart(results_dir, csv_data):
    """Sozdaet grafik s 5 lineynymi funktsiyami pertsentiley po vremeni testirovaniya"""
    
    print("Sozdanie grafika s 5 funktsiyami pertsentiley...")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    if duration_data.empty:
        print("Net dannykh o vremeni otveta")
        return
    
    # Podgotovka dannykh
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data['timestamp'] = pd.to_numeric(duration_data['timestamp'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value', 'timestamp'])
    
    if duration_data.empty:
        return
    
    # Sortiruem po vremeni i sozdaem vremennye intervaly
    duration_data = duration_data.sort_values('timestamp')
    duration_data['time_interval'] = (duration_data['timestamp'] - duration_data['timestamp'].min())
    
    # Ispolzuem vse dostupnoe vremya, ubiraem ogranichenie na 600 sekund
    max_time = duration_data['time_interval'].max()
    # duration_data = duration_data[duration_data['time_interval'] <= max_time]  # teper ispolzuem vse dannye
    
    # Razbivaem na vremennye intervaly do max_time
    time_intervals = np.linspace(0, max_time, min(100, int(max_time) + 1))  # ispolzuem min(100, max_time+1) dlya adaptivnosti
    percentiles = [50, 75, 90, 95, 99]
    colors = ['green', 'blue', 'orange', 'red', 'purple']
    labels = ['P50', 'P75', 'P90', 'P95', 'P99']
    
    # Vychislyaem pertsentili dlya kazhdogo vremennogo intervala
    percentile_over_time = {p: [] for p in percentiles}
    
    for i in range(1, len(time_intervals)):
        time_start = time_intervals[i-1]
        time_end = time_intervals[i]
        
        # Dannye za tekushchiy interval vremeni
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
    
    # Sozdaem grafik
    fig, ax = plt.subplots(figsize=(14, 8))
    
    # Risuem 5 lineynykh funktsiy
    for p, color, label in zip(percentiles, colors, labels):
        if len(percentile_over_time[p]) > 0:
            ax.plot(time_intervals[1:], percentile_over_time[p], 
                   color=color, linewidth=2, label=label, marker='o', markersize=3)
    
    # Nastroyka grafika s pravilnymi podpisyami
    ax.set_xlabel('Vremya testirovaniya (sekundy)', fontsize=12, fontweight='bold')
    ax.set_ylabel('Vremya otveta (millisekundy)', fontsize=12, fontweight='bold')
    ax.set_title('Dinamika pertsentiley vremeni otveta vo vremya testirovaniya', 
                fontsize=14, fontweight='bold')
    
    # Ustanavlivaem pravilnye predely po osi X
    ax.set_xlim(0, max_time)
    
    # Adaptivnye podpisi tikov na osi X (kazhdye 100 sekund ili drugoe znachenie v zavisimosti ot obshchego vremeni)
    tick_interval = max(1, int(max_time / 10))  # opredelyaem interval tikov v zavisimosti ot prodolzhitelnosti testa
    x_ticks = np.arange(0, max_time + tick_interval, tick_interval)
    ax.set_xticks(x_ticks)
    ax.set_xticklabels([f'{int(x)}' for x in x_ticks])
    
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    # Dobavlyaem liniyu poroga degradatsii
    ax.axhline(y=500, color='black', linestyle='--', linewidth=2, 
              alpha=0.7, label='Porog degradatsii (500 ms)')
    
    plt.tight_layout()
    plot_path = f'{results_dir}/1_five_percentiles_functions.png'
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Grafik s 5 funktsiyami pertsentiley sokhranen: {plot_path}")
    print(f"Diapazon vremeni: 0-{max_time} sekund")
    
    return percentiles, percentile_over_time

def create_cumulative_percentiles_chart(results_dir, csv_data):
    """Sozdaet grafik s nakoplennymi pertsentilyami (vse dannye do tekushchego momenta)"""
    
    print("Sozdanie grafika s nakoplennymi pertsentilyami...")
    
    duration_data = csv_data[csv_data['metric_name'] == 'http_req_duration']
    if duration_data.empty:
        print("Net dannykh o vremeni otveta")
        return
    
    # Podgotovka dannykh (tochno kak v rabochem primere)
    duration_data = duration_data.copy()
    duration_data['metric_value'] = pd.to_numeric(duration_data['metric_value'], errors='coerce')
    duration_data['timestamp'] = pd.to_numeric(duration_data['timestamp'], errors='coerce')
    duration_data = duration_data.dropna(subset=['metric_value', 'timestamp'])
    
    if duration_data.empty:
        return
    
    # Sortiruem po vremeni i sozdaem vremennye intervaly (tochno kak v rabochem primere)
    duration_data = duration_data.sort_values('timestamp')
    duration_data['time_interval'] = (duration_data['timestamp'] - duration_data['timestamp'].min())
    
    # Ispolzuem vse dostupnoe vremya, ubiraem ogranichenie na 600 sekund
    max_time = duration_data['time_interval'].max()
    # duration_data = duration_data[duration_data['time_interval'] <= max_time]  # teper ispolzuem vse dannye
    
    # Razbivaem na vremennye intervaly do max_time (uvelichivaem tochnost dlya bolshego kolichestva tochek)
    time_intervals = np.linspace(0, max_time, min(100, int(max_time) + 1))  # adaptivnoe kolichestvo tochek
    percentiles = [50, 75, 90, 95, 99]
    colors = ['green', 'blue', 'orange', 'red', 'purple']
    labels = ['P50', 'P75', 'P90', 'P95', 'P99']
    
    # IZMENENIE: vychislyaem NAKOPLENNYE pertsentili dlya kazhdogo vremennogo intervala
    percentile_over_time = {p: [] for p in percentiles}
    
    for i in range(1, len(time_intervals)):
        time_end = time_intervals[i]
        
        # Berem VSE dannye do etogo momenta vremeni (IZMENENIE!)
        data_so_far = duration_data[duration_data['time_interval'] <= time_end]['metric_value']
        
        if len(data_so_far) > 0:
            for p in percentiles:
                p_value = np.percentile(data_so_far, p)
                percentile_over_time[p].append(p_value)
        else:
            for p in percentiles:
                percentile_over_time[p].append(np.nan)
    
    # Sozdaem grafik (tochno kak v rabochem primere)
    fig, ax = plt.subplots(figsize=(14, 8))
    
    # Risuem 5 lineynykh funktsiy (tochno kak v rabochem primere)
    for p, color, label in zip(percentiles, colors, labels):
        if len(percentile_over_time[p]) > 0:
            ax.plot(time_intervals[1:], percentile_over_time[p], 
                   color=color, linewidth=2, label=label, marker='o', markersize=3)
    
    # Nastroyka grafika (tochno kak v rabochem primere)
    ax.set_xlabel('Vremya testirovaniya (sekundy)', fontsize=12, fontweight='bold')
    ax.set_ylabel('Vremya otveta (millisekundy)', fontsize=12, fontweight='bold')
    ax.set_title('Nakoplennye pertsentili vremeni otveta (vse dannye do tekushchego momenta)', 
                fontsize=14, fontweight='bold')
    
    # Ustanavlivaem pravilnye predely po osi X
    ax.set_xlim(0, max_time)
    
    # Adaptivnye podpisi tikov na osi X (kazhdye 100 sekund ili drugoe znachenie v zavisimosti ot obshchego vremeni)
    tick_interval = max(1, int(max_time / 10))  # opredelyaem interval tikov v zavisimosti ot prodolzhitelnosti testa
    x_ticks = np.arange(0, max_time + tick_interval, tick_interval)
    ax.set_xticks(x_ticks)
    ax.set_xticklabels([f'{int(x)}' for x in x_ticks])
    
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    # Dobavlyaem liniyu poroga degradatsii
    ax.axhline(y=500, color='black', linestyle='--', linewidth=2, 
              alpha=0.7, label='Porog degradatsii (500 ms)')
    
    plt.tight_layout()
    plot_path = f'{results_dir}/1_cumulative_percentiles.png'
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    print(f"Grafik s nakoplennymi pertsentilyami sokhranen: {plot_path}")
    print(f"Diapazon vremeni: 0-{max_time} sekund")
    
    return percentiles, percentile_over_time
def generate_copy_report(results_dir):
    """Osnovnaya funktsiya analiza degradatsii"""
    
    print("Analiz tochki degradatsii proizvoditelnosti...")
    
    # Zagruzka rezultatov
    csv_file = f'{results_dir}/k6_results.csv'

    csv_data = pd.DataFrame()
    if os.path.exists(csv_file):
        try:
            csv_data = load_k6_results_csv(results_dir)
            print(f"Zagruzheno {len(csv_data)} zapisey iz k6 CSV")
            
            # Otladka struktury dannykh
            debug_data_structure(results_dir, csv_data)
                
        except Exception as e:
            print(f"Error reading CSV: {e}")
            return
    else:
        print(f"Fayl {csv_file} ne nayden")
        return
    
    # Zapuskaem vse analizy
    create_degradation_analysis(results_dir, csv_data)
    find_degradation_point(results_dir, csv_data)
    create_percentiles_chart(results_dir, csv_data)
    create_cumulative_percentiles_chart(results_dir, csv_data)
    
    print(f"\n✅ Analiz zavershen! Vse grafiki sokhraneny v: {results_dir}")

def main(argv):
    if len(argv) < 2:
        print("Usage: python3 report.py <copy|degradation> <results_directory>")
        print("Example: python3 report.py copy /path/to/test/results")
        return 1

    if len(argv) == 2:
        # Backward-compatible invocation for already moved local scripts:
        mode = "copy"
        results_dir = argv[1]
    else:
        mode = argv[1]
        results_dir = argv[2]

    if mode == "copy":
        generate_copy_report(results_dir)
        return 0
    if mode == "degradation":
        degradation_report.generate_degradation_report(results_dir)
        return 0

    print(f"Unknown mode: {mode}")
    print("Supported modes: copy, degradation")
    return 1


if __name__ == '__main__':
    raise SystemExit(main(sys.argv))
