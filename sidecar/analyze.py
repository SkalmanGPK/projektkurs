import sys
import json
import re
from datetime import datetime
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

def parse_log(filepath):
    timestamps = []
    received_vals = []
    stored_vals = []
    fas2_time = None
    fas3_time = None

    with open(filepath, 'r') as f:
        for line in f:
            # Hitta fasmarkeringar
            fas_match = re.search(r'\[(\d{2}:\d{2}:\d{2})\] (FAS\d_START)', line)
            if fas_match:
                t = datetime.strptime(fas_match.group(1), "%H:%M:%S")
                if fas_match.group(2) == "FAS2_START":
                    fas2_time = t
                elif fas_match.group(2) == "FAS3_START":
                    fas3_time = t
                continue

            # Hitta STATS-rader från sidecar
            stats_match = re.search(r'\[STATS\] App3: ({.*?})', line)
            if stats_match:
                try:
                    data = json.loads(stats_match.group(1))
                    # Försök hitta tidsstämpel på raden
                    time_match = re.search(r'(\d{2}:\d{2}:\d{2})', line)
                    if time_match:
                        t = datetime.strptime(time_match.group(1), "%H:%M:%S")
                    else:
                        t = datetime.now()
                    
                    timestamps.append(t)
                    received_vals.append(data['received'])
                    stored_vals.append(data['stored'])
                except (json.JSONDecodeError, KeyError):
                    continue

    return timestamps, received_vals, stored_vals, fas2_time, fas3_time


def plot_results(timestamps, received_vals, stored_vals, fas2_time, fas3_time):
    if not timestamps:
        print("[!] Inga STATS-rader hittades i loggen.")
        return

    # Konvertera till sekunder från start
    start = timestamps[0]
    seconds = [(t - start).total_seconds() for t in timestamps]
    fas2_sec = (fas2_time - start).total_seconds() if fas2_time else None
    fas3_sec = (fas3_time - start).total_seconds() if fas3_time else None

    # Beräkna gap
    gaps = [r - s for r, s in zip(received_vals, stored_vals)]

    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 8))

    # Graf 1: received vs stored
    ax1.plot(seconds, received_vals, label='Mottagna paket (received)', color='steelblue', linewidth=2)
    ax1.plot(seconds, stored_vals, label='Sparade paket (stored)', color='green', linewidth=2)
    if fas2_sec:
        ax1.axvline(x=fas2_sec, color='red', linestyle='--', label='DB avstängd')
    if fas3_sec:
        ax1.axvline(x=fas3_sec, color='orange', linestyle='--', label='DB återstartad')
    ax1.set_title('Mottagna vs sparade paket över tid')
    ax1.set_xlabel('Tid (sekunder)')
    ax1.set_ylabel('Antal paket')
    ax1.legend()
    ax1.grid(True, alpha=0.3)

    # Graf 2: gap (förlorad data)
    ax2.fill_between(seconds, gaps, alpha=0.4, color='red', label='Förlorad data (gap)')
    ax2.plot(seconds, gaps, color='red', linewidth=2)
    if fas2_sec:
        ax2.axvline(x=fas2_sec, color='red', linestyle='--', label='DB avstängd')
    if fas3_sec:
        ax2.axvline(x=fas3_sec, color='orange', linestyle='--', label='DB återstartad')
    ax2.set_title('Gap mellan mottagna och sparade paket (förlorad data)')
    ax2.set_xlabel('Tid (sekunder)')
    ax2.set_ylabel('Antal förlorade paket')
    ax2.legend()
    ax2.grid(True, alpha=0.3)

    plt.tight_layout()
    plt.savefig('experiment_resultat.png', dpi=150)
    print("[+] Graf sparad som experiment_resultat.png")

    # Skriv ut sammanfattning
    if fas2_sec and fas3_sec:
        # Hitta index under attackfasen
        attack_gaps = [g for s, g in zip(seconds, gaps) if fas2_sec <= s <= fas3_sec]
        if attack_gaps:
            max_gap = max(attack_gaps)
            print(f"\n--- Sammanfattning ---")
            print(f"Max förlorade paket under attackfas: {max_gap}")
            print(f"Totalt mottagna paket: {received_vals[-1]}")
            print(f"Totalt sparade paket:  {stored_vals[-1]}")
            print(f"Totalt förlorade:      {received_vals[-1] - stored_vals[-1]}")
            print(f"Förlustgrad:           {((received_vals[-1] - stored_vals[-1]) / received_vals[-1] * 100):.1f}%")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Användning: python3 analyze.py <loggfil>")
        sys.exit(1)

    filepath = sys.argv[1]
    timestamps, received, stored, fas2, fas3 = parse_log(filepath)
    plot_results(timestamps, received, stored, fas2, fas3)
