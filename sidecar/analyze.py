import sys
import json
import re
from datetime import datetime
import numpy as np
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

def parse_log(filepath):
    timestamps = []
    received_vals = []
    stored_vals = []
    dbactual_vals = []
    dbactual_times = []
    fas2_time = None
    fas3_time = None

    with open(filepath, 'r') as f:
        for line in f:
            line = line.strip()

            fas_match = re.match(r'\[(\d{2}:\d{2}:\d{2})\] (FAS\d_START)', line)
            if fas_match:
                t = datetime.strptime(fas_match.group(1), "%H:%M:%S")
                if fas_match.group(2) == "FAS2_START":
                    fas2_time = t
                elif fas_match.group(2) == "FAS3_START":
                    fas3_time = t
                continue

            stats_match = re.search(r'\[STATS\] App3: ({.*?})', line)
            if stats_match:
                try:
                    data = json.loads(stats_match.group(1))
                    time_match = re.match(r'(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})', line)
                    if time_match:
                        t = datetime.strptime(time_match.group(1), "%Y-%m-%dT%H:%M:%S")
                        timestamps.append(t)
                        received_vals.append(data['received'])
                        stored_vals.append(data['stored'])
                except (json.JSONDecodeError, KeyError):
                    continue

            dbcount_match = re.search(r'\[DBCOUNT\] App3: ({.*?})', line)
            if dbcount_match:
                try:
                    data = json.loads(dbcount_match.group(1))
                    time_match = re.match(r'(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})', line)
                    if time_match and data['db_actual'] != -1:
                        t = datetime.strptime(time_match.group(1), "%Y-%m-%dT%H:%M:%S")
                        dbactual_times.append(t)
                        dbactual_vals.append(data['db_actual'])
                except (json.JSONDecodeError, KeyError):
                    continue

    # Filtrera bort initialt brus fran gamla pods
    # Hitta forsta index dar received borjar fran ett lagt varde
    start_idx = 0
    for i, v in enumerate(received_vals):
        if v <= 2:
            start_idx = i
            break

    timestamps = timestamps[start_idx:]
    received_vals = received_vals[start_idx:]
    stored_vals = stored_vals[start_idx:]

    # Samma filtrering for dbactual - ta bara med varden efter att stats-data borjar
    if timestamps and dbactual_times:
        clean_db_times = []
        clean_db_vals = []
        for t, v in zip(dbactual_times, dbactual_vals):
            if t >= timestamps[0]:
                clean_db_times.append(t)
                clean_db_vals.append(v)
        dbactual_times = clean_db_times
        dbactual_vals = clean_db_vals

    return timestamps, received_vals, stored_vals, dbactual_times, dbactual_vals, fas2_time, fas3_time


def plot_results(timestamps, received_vals, stored_vals, dbactual_times, dbactual_vals, fas2_time, fas3_time):
    if not timestamps:
        print("[!] Inga STATS-rader hittades i loggen.")
        return

    start = timestamps[0]

    if fas2_time:
        fas2_time = fas2_time.replace(year=start.year, month=start.month, day=start.day)
    if fas3_time:
        fas3_time = fas3_time.replace(year=start.year, month=start.month, day=start.day)

    seconds = [(t - start).total_seconds() for t in timestamps]
    db_seconds = [(t - start).total_seconds() for t in dbactual_times]
    fas2_sec = (fas2_time - start).total_seconds() if fas2_time else None
    fas3_sec = (fas3_time - start).total_seconds() if fas3_time else None

    # Normalisera alla varden till 0
    received_offset = received_vals[0] if received_vals else 0
    stored_offset = stored_vals[0] if stored_vals else 0
    db_offset = dbactual_vals[0] if dbactual_vals else 0

    received_norm = [v - received_offset for v in received_vals]
    stored_norm = [v - stored_offset for v in stored_vals]
    db_normalized = [v - db_offset for v in dbactual_vals]

    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 8))

    # Graf 1: App3s uppfattning - received och stored alltid identiska
    ax1.plot(seconds, received_norm, label='Mottagna paket (received)',
             color='steelblue', linewidth=2)
    ax1.plot(seconds, stored_norm, label='Sparade paket (stored, enligt app3)',
             color='green', linewidth=2, linestyle='--')
    if fas2_sec is not None:
        ax1.axvline(x=fas2_sec, color='red', linestyle='--', linewidth=1.5,
                    label='DB avstangd')
    if fas3_sec is not None:
        ax1.axvline(x=fas3_sec, color='orange', linestyle='--', linewidth=1.5,
                    label='DB aterstartad')
    ax1.set_title('App3:s uppfattning - mottagna vs "sparade" paket (ingen skillnad synlig)')
    ax1.set_xlabel('Tid (sekunder fran start)')
    ax1.set_ylabel('Antal paket')
    ax1.legend()
    ax1.grid(True, alpha=0.3)

    # Graf 2: Verkligheten - received vs faktiska rader i DB
    ax2.plot(seconds, received_norm, label='Mottagna paket (received)',
             color='steelblue', linewidth=2)
    if db_seconds:
        ax2.plot(db_seconds, db_normalized, label='Faktiskt sparade i DB (db_actual)',
                 color='green', linewidth=2, linestyle='--')
        if len(db_seconds) > 1:
            db_interp = np.interp(seconds, db_seconds, db_normalized)
            ax2.fill_between(seconds, db_interp, received_norm,
                             where=[r > d for r, d in zip(received_norm, db_interp)],
                             alpha=0.3, color='red', label='Forlorad data (gap)')
    if fas2_sec is not None:
        ax2.axvline(x=fas2_sec, color='red', linestyle='--', linewidth=1.5,
                    label='DB avstangd')
    if fas3_sec is not None:
        ax2.axvline(x=fas3_sec, color='orange', linestyle='--', linewidth=1.5,
                    label='DB aterstartad')
    ax2.set_title('Verkligheten - mottagna paket vs faktiskt sparade i databasen')
    ax2.set_xlabel('Tid (sekunder fran start)')
    ax2.set_ylabel('Antal paket')
    ax2.legend()
    ax2.grid(True, alpha=0.3)

    plt.tight_layout()
    plt.savefig('experiment_resultat.png', dpi=150)
    print("[+] Graf sparad som experiment_resultat.png")

    print(f"\n--- Sammanfattning ---")
    print(f"Totalt mottagna paket:        {received_norm[-1]}")
    print(f"App3 trodde den sparade:      {stored_norm[-1]}")
    if db_normalized:
        print(f"Faktiskt sparade i DB:        {db_normalized[-1]}")
        print(f"Forlorade paket:              {received_norm[-1] - db_normalized[-1]}")
        if received_norm[-1] > 0:
            print(f"Forlustgrad:                  "
                  f"{((received_norm[-1] - db_normalized[-1]) / received_norm[-1] * 100):.1f}%")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Anvandning: python3 analyze.py <loggfil>")
        sys.exit(1)

    filepath = sys.argv[1]
    timestamps, received, stored, db_times, db_actual, fas2, fas3 = parse_log(filepath)
    plot_results(timestamps, received, stored, db_times, db_actual, fas2, fas3)
