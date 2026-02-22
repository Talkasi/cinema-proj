from __future__ import annotations

from pathlib import Path
from typing import Iterable

import pandas as pd


def results_path(results_dir: str, filename: str) -> Path:
    return Path(results_dir) / filename


def read_csv_if_exists(path: Path, *, low_memory: bool = False) -> pd.DataFrame:
    if not path.exists():
        return pd.DataFrame()

    return pd.read_csv(path, low_memory=low_memory)


def load_k6_results_csv(results_dir: str) -> pd.DataFrame:
    path = results_path(results_dir, "k6_results.csv")
    return read_csv_if_exists(path, low_memory=False)


def load_resource_usage_csv(results_dir: str) -> pd.DataFrame:
    path = results_path(results_dir, "resource_usage.csv")
    return read_csv_if_exists(path, low_memory=False)


def convert_numeric_columns(df: pd.DataFrame, columns: Iterable[str]) -> pd.DataFrame:
    if df.empty:
        return df

    result = df.copy()
    for column in columns:
        if column in result.columns:
            result[column] = pd.to_numeric(result[column], errors="coerce")

    return result


def print_dataframe_overview(df: pd.DataFrame) -> None:
    print("Analysis message")
    print(f"Vsego strok: {len(df)}")
    print(f"Kolonki: {list(df.columns)}")
    if not df.empty and "metric_name" in df.columns:
        print(f"Tipy metrik: {df['metric_name'].unique()}")

