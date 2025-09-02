#!/usr/bin/python

# The intention of this script is to test the reconnection logic of gRPC/WebSocket by killing the established connection to an endpoint.
# It only supports ipv4 and depends on `ss` and `sudo`
# Completely vibe coded but gets the job done

import argparse
import os
import re
import subprocess
import sys

def run_cmd(cmd):
    return subprocess.check_output(cmd, text=True, stderr=subprocess.STDOUT)

def parse_connections():
    out = run_cmd(["ss", "-H", "-4", "-t", "-n", "-p", "state", "established"])
    lines = [l.strip() for l in out.splitlines() if l.strip()]
    conns = []
    for line in lines:
        # Expected columns: State Recv-Q Send-Q Local_Address:Port Peer_Address:Port Process
        parts = line.split()
        if len(parts) != 5:
            continue

        state = parts[0]
        local = parts[2]
        peer = parts[3]
        proc_field = parts[4]

        if ":" not in local or ":" not in peer:
            continue
        try:
            lip, lport = local.rsplit(":", 1)
            rip, rport = peer.rsplit(":", 1)
        except ValueError:
            continue

        procs = re.findall(r'\("([^"]+)",pid=(\d+)', proc_field)
        conns.append({
            "state": state,
            "local_ip": lip,
            "local_port": lport,
            "remote_ip": rip,
            "remote_port": rport,
            "procs": [{"name": n, "pid": int(pid)} for n, pid in procs],
            "raw": line,
        })
    return conns

def matches(conn, proc_substr, dport):
    if proc_substr:
        names = [p["name"].lower() for p in conn["procs"]]
        if not any(proc_substr in n for n in names):
            return False

    if dport and conn["remote_port"] != str(dport):
        return False

    return True

def kill_connection(conn, dry_run=False):
    # Build: ss -K src <local_ip> sport = :<local_port> dst <remote_ip> dport = :<remote_port>
    cmd = [
        "ss", "-K", "src {0} sport = :{1} dst {2} dport = :{3}".format(conn["local_ip"], conn['local_port'], conn["remote_ip"], conn['remote_port'])
    ]
    print("Executing:", " ".join(cmd))


    if dry_run:
        return 0

    print("List of killed connections:\n")

    try:
        subprocess.check_call(cmd)
        return 0
    except subprocess.CalledProcessError as e:
        return e.returncode

def main():
    ap = argparse.ArgumentParser(description="Kill TCP connections by process name and port using ss -K")
    ap.add_argument("-l", "--list",  action='store_true', help="Lists all established connections")
    ap.add_argument("-p", "--process", help="Process name substring to match (e.g. 'firefox', 'slack')")
    g = ap.add_mutually_exclusive_group()
    g.add_argument("--port", type=int, help="Match by remote/destination port")
    ap.add_argument("--dry-run", action="store_true", help="Show what would be killed without executing")
    args = ap.parse_args()

    conns = parse_connections()
    if args.list:
        for con in conns:
            print("Remote Port: {0}, Process: {1}".format(con["remote_port"], con["procs"]))
        return

    proc_substr = args.process.lower()
    matches_list = [c for c in conns if matches(c, proc_substr,  args.port)]

    if not matches_list:
        print("No matching established connections found.", file=sys.stderr)
        sys.exit(1)

    print("Found connections:")
    for i, c in enumerate(matches_list, 1):
        names = ",".join({p["name"] for p in c["procs"]}) or "unknown"
        print(f" [{i}] {c['local_ip']}:{c['local_port']} -> {c['remote_ip']}:{c['remote_port']} procs={names}")

    print("")

    rc = 0

    print("If you get the error: RTNETLINK answers: Invalid argument, don't worry. It should work regardless.")


    for c in matches_list:
        print("\n")
        rc |= kill_connection(c, dry_run=args.dry_run)
        print("\n")
    sys.exit(rc)

if __name__ == "__main__":
    if os.geteuid() != 0:
        print("Note: run this script with sudo for ss -K to succeed.", file=sys.stderr)
    main()

