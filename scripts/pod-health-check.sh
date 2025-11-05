#!/bin/bash

set -e

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LOG_FILE="infra-health.log"
NAMESPACES=("default" "powerlifting-coach")

echo "Running pod health check at $TIMESTAMP" >&2

check_pods() {
    local namespace=$1
    local all_pods=$(kubectl get pods -n "$namespace" -o json 2>/dev/null || echo '{"items":[]}')

    local not_ready=$(echo "$all_pods" | jq -r '.items[] | select(.status.phase != "Running" or (.status.conditions[] | select(.type == "Ready" and .status != "True"))) | .metadata.name' 2>/dev/null || echo "")

    local crashloop=$(echo "$all_pods" | jq -r '.items[] | select(.status.containerStatuses[]?.state.waiting.reason == "CrashLoopBackOff") | .metadata.name' 2>/dev/null || echo "")

    if [ -n "$not_ready" ] || [ -n "$crashloop" ]; then
        local pod_list=()

        while IFS= read -r pod; do
            [ -z "$pod" ] && continue
            pod_list+=("\"$pod\"")

            echo "Pod $pod in $namespace is not ready. Fetching logs..." >&2
            kubectl logs "$pod" -n "$namespace" --tail=50 2>&1 | head -20 >&2 || echo "Failed to fetch logs for $pod" >&2
        done <<< "$not_ready"

        while IFS= read -r pod; do
            [ -z "$pod" ] && continue
            if [[ ! " ${pod_list[@]} " =~ " \"$pod\" " ]]; then
                pod_list+=("\"$pod\"")
            fi

            echo "Pod $pod in $namespace is in CrashLoopBackOff. Fetching logs..." >&2
            kubectl logs "$pod" -n "$namespace" --tail=50 2>&1 | head -20 >&2 || echo "Failed to fetch logs for $pod" >&2
        done <<< "$crashloop"

        local pods_json=$(IFS=,; echo "${pod_list[*]}")
        echo "{\"timestamp\": \"$TIMESTAMP\", \"level\": \"error\", \"msg\": \"Unhealthy pods detected in namespace $namespace\", \"namespace\": \"$namespace\", \"pods\": [$pods_json]}" >> "$LOG_FILE"
    else
        echo "{\"timestamp\": \"$TIMESTAMP\", \"level\": \"info\", \"msg\": \"All pods healthy in namespace $namespace\", \"namespace\": \"$namespace\"}" >> "$LOG_FILE"
    fi
}

for ns in "${NAMESPACES[@]}"; do
    if kubectl get namespace "$ns" &>/dev/null; then
        check_pods "$ns"
    else
        echo "{\"timestamp\": \"$TIMESTAMP\", \"level\": \"info\", \"msg\": \"Namespace $ns does not exist, skipping\", \"namespace\": \"$ns\"}" >> "$LOG_FILE"
    fi
done

echo "Pod health check complete. Results appended to $LOG_FILE" >&2
