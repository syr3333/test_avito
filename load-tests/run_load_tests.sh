#!/bin/bash

set -euo pipefail

VENV_DIR="load-tests/.venv"
RESULTS_DIR="load-tests/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
declare -a SUMMARY_LINES=()

require_python() {
    if ! command -v python3 >/dev/null 2>&1; then
        echo "Python 3 is not installed" >&2
        echo "Download: https://www.python.org/downloads/" >&2
        exit 1
    fi
}

require_server() {
    if ! curl -s -f http://localhost:8080/health >/dev/null 2>&1; then
        echo "Server is not reachable on http://localhost:8080" >&2
        echo "Start it first (e.g. go run main.go)" >&2
        exit 1
    fi
}

setup_venv() {
    if [ ! -d "$VENV_DIR" ]; then
        python3 -m venv "$VENV_DIR"
    fi
    # shellcheck source=/dev/null
    source "$VENV_DIR/bin/activate"
}

install_deps() {
    pip install -q --upgrade pip >/dev/null
    pip install -q -r load-tests/requirements.txt >/dev/null
}

extract_summary() {
    local stats_file=$1
    if [ ! -f "$stats_file" ]; then
        echo "stats unavailable"
        return
    fi
    LC_ALL=C awk -F, '$2 == "Aggregated" {printf "%s requests, %s failures, avg %.1f ms, %.2f req/s", $3, $4, $6, $10}' "$stats_file"
}

run_scenario() {
    local slug=$1
    local title=$2
    local users=$3
    local spawn_rate=$4
    local duration=$5

    local base_path="${RESULTS_DIR}/${slug}_${TIMESTAMP}"
    local html_file="${base_path}.html"
    local csv_base="$base_path"
    local log_file="${base_path}.log"

    if locust \
        -f load-tests/locustfile.py \
        --host=http://localhost:8080 \
        --users "$users" \
        --spawn-rate "$spawn_rate" \
        --run-time "${duration}s" \
        --headless \
        --only-summary \
        --html "$html_file" \
        --csv "$csv_base" \
        >"$log_file" 2>&1; then
        local stats_file="${csv_base}_stats.csv"
        local summary
        summary=$(extract_summary "$stats_file")
        SUMMARY_LINES+=("${title}: ${summary}")
    else
        cat "$log_file" >&2
        exit 1
    fi
}

print_summary() {
    echo "Load tests finished at ${TIMESTAMP}"
    for line in "${SUMMARY_LINES[@]}"; do
        echo "  - ${line}"
    done
    echo ""
}

require_python
require_server
setup_venv
install_deps
mkdir -p "$RESULTS_DIR"

run_scenario "load_test" "Load test (50 users / 20s)" 50 25 20

print_summary

deactivate
