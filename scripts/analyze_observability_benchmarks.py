#!/usr/bin/env python3
"""
Aggregates Go benchmark outputs for observability (tracing/logging) toggles.

The script reads *.txt artifacts produced by the benchmark runner scripts,
computes average time/op, memory/op, and allocs/op per benchmark, and writes
an easy-to-compare text report.
"""

from __future__ import annotations

import argparse
import datetime as dt
import re
import statistics
import sys
from collections import defaultdict
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Sequence, Set, Tuple

# Go can emit microseconds with the unicode micro sign; normalize to ASCII.
MICRO_SIGN = "\u00b5"

TIME_MULTIPLIERS = {
    "ns": 1,
    "us": 1_000,
    "ms": 1_000_000,
    "s": 1_000_000_000,
}

BYTE_MULTIPLIERS = {
    "B": 1,
    "kB": 1024,
    "MB": 1024 * 1024,
    "GB": 1024 * 1024 * 1024,
}

BENCH_RE = re.compile(
    r"^(?P<name>Benchmark\S+)-\d+\s+\d+\s+"
    r"(?P<time_value>[\d.]+)\s+(?P<time_unit>ns|us|ms|s)/op\s+"
    r"(?P<bytes_value>[\d.]+)\s*(?P<bytes_unit>B|kB|MB|GB)?/op\s+"
    r"(?P<allocs>[\d.]+)\s+allocs/op"
)


class BenchSample:
    def __init__(self, name: str, time_ns: float, bytes_per_op: float, allocs_per_op: float, source: Path):
        self.name = name
        self.time_ns = time_ns
        self.bytes_per_op = bytes_per_op
        self.allocs_per_op = allocs_per_op
        self.source = source


def parse_bench_line(line: str, source: Path) -> Optional[BenchSample]:
    """Parse a single go test benchmark output line."""
    normalized = line.replace(MICRO_SIGN, "u").strip()
    match = BENCH_RE.match(normalized)
    if not match:
        return None

    time_unit = match.group("time_unit")
    bytes_unit = match.group("bytes_unit") or "B"

    time_ns = float(match.group("time_value")) * TIME_MULTIPLIERS.get(time_unit, 1)
    bytes_per_op = float(match.group("bytes_value")) * BYTE_MULTIPLIERS.get(bytes_unit, 1)
    allocs_per_op = float(match.group("allocs"))

    return BenchSample(
        name=match.group("name"),
        time_ns=time_ns,
        bytes_per_op=bytes_per_op,
        allocs_per_op=allocs_per_op,
        source=source,
    )


def discover_files(paths: Sequence[Path]) -> List[Path]:
    """Expand provided paths into a list of files to parse."""
    files: List[Path] = []
    for path in paths:
        if path.is_file():
            files.append(path)
        elif path.is_dir():
            files.extend(sorted(path.glob("*.txt")))
    return files


def collect_samples(files: Sequence[Path]) -> List[BenchSample]:
    samples: List[BenchSample] = []
    for file in files:
        for line in file.read_text(encoding="utf-8", errors="ignore").splitlines():
            sample = parse_bench_line(line, file)
            if sample:
                samples.append(sample)
    return samples


def aggregate_samples(samples: Sequence[BenchSample]) -> Tuple[Dict[str, Dict[str, List[float]]], Dict[str, Set[Path]]]:
    metrics: Dict[str, Dict[str, List[float]]] = defaultdict(lambda: {"time_ns": [], "bytes": [], "allocs": []})
    sources: Dict[str, Set[Path]] = defaultdict(set)

    for sample in samples:
        metrics[sample.name]["time_ns"].append(sample.time_ns)
        metrics[sample.name]["bytes"].append(sample.bytes_per_op)
        metrics[sample.name]["allocs"].append(sample.allocs_per_op)
        sources[sample.name].add(sample.source)

    return metrics, sources


def average(values: Sequence[float]) -> float:
    return statistics.mean(values) if values else 0.0


def format_ns(ns_value: float) -> str:
    if ns_value >= 1_000_000:
        return f"{ns_value / 1_000_000:.2f} ms"
    if ns_value >= 1_000:
        return f"{ns_value / 1_000:.2f} us"
    return f"{ns_value:.0f} ns"


def format_bytes(bytes_value: float) -> str:
    if bytes_value >= 1024 * 1024:
        return f"{bytes_value / (1024 * 1024):.2f} MB"
    if bytes_value >= 1024:
        return f"{bytes_value / 1024:.2f} kB"
    return f"{bytes_value:.0f} B"


def render_report(
    metrics: Dict[str, Dict[str, List[float]]],
    sources: Dict[str, Set[Path]],
    files: Sequence[Path],
) -> str:
    table_lines = []
    table_lines.append(f"{'Benchmark':58} {'time/op':>14} {'mem/op':>14} {'allocs/op':>10}")

    summary: Dict[str, Dict[str, float]] = {}
    for name in sorted(metrics.keys()):
        times = metrics[name]["time_ns"]
        mems = metrics[name]["bytes"]
        allocs = metrics[name]["allocs"]

        summary[name] = {
            "time_ns": average(times),
            "bytes": average(mems),
            "allocs": average(allocs),
        }

        table_lines.append(
            f"{name:58} "
            f"{format_ns(summary[name]['time_ns']):>14} "
            f"{format_bytes(summary[name]['bytes']):>14} "
            f"{summary[name]['allocs']:10.2f}"
        )

    report_lines: List[str] = []
    report_lines.append("Per-benchmark averages (time/op, memory/op, allocs/op):")
    report_lines.extend(table_lines)

    return "\n".join(report_lines)


def main(argv: Optional[Sequence[str]] = None) -> int:
    parser = argparse.ArgumentParser(description="Aggregate observability benchmark outputs and build a report.")
    parser.add_argument(
        "paths",
        nargs="*",
        default=["artifacts/observability"],
        help="Files or directories with go test benchmark output (*.txt).",
    )
    parser.add_argument(
        "--out",
        dest="out",
        default=None,
        help="Where to write the report. Defaults to artifacts/observability/observability-report-<timestamp>.md",
    )
    args = parser.parse_args(argv)

    paths = [Path(p) for p in args.paths]
    files = discover_files(paths)
    if not files:
        print("No benchmark files found to analyze.", file=sys.stderr)
        return 1

    samples = collect_samples(files)
    if not samples:
        print("No benchmark samples parsed from provided files.", file=sys.stderr)
        return 1

    metrics, sources = aggregate_samples(samples)
    report = render_report(metrics, sources, files)

    if args.out:
        out_path = Path(args.out)
    else:
        timestamp = dt.datetime.now().strftime("%Y%m%d-%H%M%S")
        out_path = Path("artifacts/observability") / f"observability-report-{timestamp}.md"

    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(report, encoding="utf-8")
    print(f"Report written to {out_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
