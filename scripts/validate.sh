#!/bin/bash
# Validate Rust library mappings for security issues
# Uses: gh CLI for repo checks, OSV API for vulnerability checks

set -euo pipefail

MAPPINGS_FILE="${1:-cmd/rinku/mappings.json}"
PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Check dependencies
command -v gh >/dev/null 2>&1 || { echo "Error: gh CLI is required"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "Error: jq is required"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "Error: curl is required"; exit 1; }

# Extract repo path from GitHub URL
# Handles: https://github.com/owner/repo, https://github.com/owner/repo/tree/main/...
parse_repo() {
    local url="$1"
    echo "$url" | sed 's|https://github.com/||' | cut -d'/' -f1-2
}

# Extract crate name (best guess from repo name)
parse_crate_name() {
    local url="$1"
    basename "$(echo "$url" | sed 's|/tree/.*||')"
}

# Check repository legitimacy via gh CLI
check_repo() {
    local repo="$1"
    local result
    local status="pass"
    local details=""

    # Get repo info
    if ! result=$(gh repo view "$repo" --json name,isArchived,pushedAt,stargazerCount 2>&1); then
        echo "fail|Repository not found or inaccessible"
        return
    fi

    local archived=$(echo "$result" | jq -r '.isArchived')
    local stars=$(echo "$result" | jq -r '.stargazerCount')
    local pushed_at=$(echo "$result" | jq -r '.pushedAt')

    # Check if archived
    if [[ "$archived" == "true" ]]; then
        status="warn"
        details="archived"
    fi

    # Check last activity (warn if >2 years old)
    if [[ -n "$pushed_at" && "$pushed_at" != "null" ]]; then
        local pushed_ts=$(date -d "$pushed_at" +%s 2>/dev/null || echo "0")
        local now_ts=$(date +%s)
        local two_years=$((2 * 365 * 24 * 60 * 60))

        if (( now_ts - pushed_ts > two_years )); then
            status="warn"
            details="${details:+$details, }inactive >2yr"
        fi
    fi

    # Format stars
    local stars_fmt
    if (( stars >= 1000 )); then
        stars_fmt="$(echo "scale=1; $stars/1000" | bc)k"
    else
        stars_fmt="$stars"
    fi

    echo "${status}|${stars_fmt} stars${details:+, $details}"
}

# Check vulnerabilities via OSV API
check_vulns() {
    local crate="$1"
    local response

    response=$(curl -s -X POST "https://api.osv.dev/v1/query" \
        -H "Content-Type: application/json" \
        -d "{\"package\":{\"name\":\"$crate\",\"ecosystem\":\"crates.io\"}}" 2>/dev/null)

    # Check if response has vulns
    local vuln_count=$(echo "$response" | jq -r '.vulns | length // 0' 2>/dev/null)

    if [[ "$vuln_count" == "0" || -z "$vuln_count" ]]; then
        echo "pass|no known vulnerabilities"
        return
    fi

    # Get severity info
    local vuln_ids=$(echo "$response" | jq -r '.vulns[].id' 2>/dev/null | head -3 | tr '\n' ', ' | sed 's/,$//')

    # Check for high severity
    local has_high=$(echo "$response" | jq -r '.vulns[].database_specific.severity // empty' 2>/dev/null | grep -iE 'high|critical' || true)

    if [[ -n "$has_high" ]]; then
        echo "fail|$vuln_count vulns: $vuln_ids"
    else
        echo "warn|$vuln_count vulns (low/moderate): $vuln_ids"
    fi
}

# Main validation loop
echo "Validating Rust library targets from $MAPPINGS_FILE..."
echo ""

# Get unique target URLs (skip <None>)
urls=$(jq -r '.mappings[].target[]' "$MAPPINGS_FILE" 2>/dev/null | grep -v '<None>' | sort -u)

total=$(echo "$urls" | wc -l)
current=0

while IFS= read -r url; do
    [[ -z "$url" ]] && continue
    current=$((current + 1))

    repo=$(parse_repo "$url")
    crate=$(parse_crate_name "$url")

    # Skip org pages (no repo)
    if [[ ! "$repo" =~ / ]] || [[ "$repo" =~ /$ ]]; then
        echo -e "${YELLOW}[SKIP]${NC} $url (org page, not a repo)"
        continue
    fi

    # Check repo legitimacy
    repo_result=$(check_repo "$repo")
    repo_status=$(echo "$repo_result" | cut -d'|' -f1)
    repo_details=$(echo "$repo_result" | cut -d'|' -f2-)

    # Check vulnerabilities
    vuln_result=$(check_vulns "$crate")
    vuln_status=$(echo "$vuln_result" | cut -d'|' -f1)
    vuln_details=$(echo "$vuln_result" | cut -d'|' -f2-)

    # Determine overall status
    if [[ "$repo_status" == "fail" || "$vuln_status" == "fail" ]]; then
        status="FAIL"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        color="$RED"
    elif [[ "$repo_status" == "warn" || "$vuln_status" == "warn" ]]; then
        status="WARN"
        WARN_COUNT=$((WARN_COUNT + 1))
        color="$YELLOW"
    else
        status="PASS"
        PASS_COUNT=$((PASS_COUNT + 1))
        color="$GREEN"
    fi

    # Print result
    echo -e "${color}[$status]${NC} $repo ($repo_details) - $vuln_details"

    # Rate limiting: be nice to APIs
    sleep 0.5

done <<< "$urls"

# Summary
echo ""
echo "========================================"
echo -e "Summary: ${GREEN}$PASS_COUNT passed${NC}, ${YELLOW}$WARN_COUNT warnings${NC}, ${RED}$FAIL_COUNT failed${NC}"
echo "========================================"

# Exit with error if any failures
if (( FAIL_COUNT > 0 )); then
    exit 1
fi
