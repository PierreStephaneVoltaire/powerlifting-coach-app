#!/bin/bash

# Pod Health Check Script
# Checks Kubernetes pod status and logs issues to infra-health.log

set -e

NAMESPACE="app"
LOG_FILE="infra-health.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo "=== Pod Health Check - $TIMESTAMP ===" >> "$LOG_FILE"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "ERROR: kubectl not found. Please install kubectl." | tee -a "$LOG_FILE"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo "ERROR: Namespace '$NAMESPACE' does not exist." | tee -a "$LOG_FILE"
    exit 1
fi

echo "Checking pods in namespace: $NAMESPACE" | tee -a "$LOG_FILE"

# Get all pods
PODS=$(kubectl get pods -n "$NAMESPACE" -o json)

# Check for pods not in Running state
NOT_RUNNING=$(echo "$PODS" | jq -r '.items[] | select(.status.phase != "Running") | .metadata.name')

if [ -n "$NOT_RUNNING" ]; then
    echo "WARNING: Found pods not in Running state:" | tee -a "$LOG_FILE"
    echo "$NOT_RUNNING" | tee -a "$LOG_FILE"

    # Get details for each non-running pod
    for POD in $NOT_RUNNING; do
        echo "--- Details for $POD ---" >> "$LOG_FILE"
        kubectl describe pod "$POD" -n "$NAMESPACE" >> "$LOG_FILE" 2>&1
        echo "" >> "$LOG_FILE"
    done
else
    echo "SUCCESS: All pods are in Running state." | tee -a "$LOG_FILE"
fi

# Check for CrashLoopBackOff
CRASHLOOP=$(echo "$PODS" | jq -r '.items[] | select(.status.containerStatuses[]?.state.waiting?.reason == "CrashLoopBackOff") | .metadata.name')

if [ -n "$CRASHLOOP" ]; then
    echo "CRITICAL: Found pods in CrashLoopBackOff:" | tee -a "$LOG_FILE"
    echo "$CRASHLOOP" | tee -a "$LOG_FILE"

    # Get logs for crashlooping pods
    for POD in $CRASHLOOP; do
        echo "--- Logs for $POD ---" >> "$LOG_FILE"
        kubectl logs "$POD" -n "$NAMESPACE" --tail=50 >> "$LOG_FILE" 2>&1
        echo "" >> "$LOG_FILE"

        # Get previous logs if available
        echo "--- Previous logs for $POD ---" >> "$LOG_FILE"
        kubectl logs "$POD" -n "$NAMESPACE" --previous --tail=50 >> "$LOG_FILE" 2>&1 || echo "No previous logs available" >> "$LOG_FILE"
        echo "" >> "$LOG_FILE"
    done
fi

# Check pod restarts
HIGH_RESTARTS=$(echo "$PODS" | jq -r '.items[] | select(.status.containerStatuses[]?.restartCount > 5) | .metadata.name + " (restarts: " + (.status.containerStatuses[0].restartCount | tostring) + ")"')

if [ -n "$HIGH_RESTARTS" ]; then
    echo "WARNING: Found pods with high restart counts:" | tee -a "$LOG_FILE"
    echo "$HIGH_RESTARTS" | tee -a "$LOG_FILE"
fi

# Check resource usage
echo "" >> "$LOG_FILE"
echo "Resource usage:" >> "$LOG_FILE"
kubectl top pods -n "$NAMESPACE" >> "$LOG_FILE" 2>&1 || echo "Metrics server not available" >> "$LOG_FILE"

# Summary
echo "" >> "$LOG_FILE"
TOTAL_PODS=$(echo "$PODS" | jq '.items | length')
RUNNING_PODS=$(echo "$PODS" | jq '[.items[] | select(.status.phase == "Running")] | length')

echo "Summary: $RUNNING_PODS/$TOTAL_PODS pods running" | tee -a "$LOG_FILE"
echo "===================================" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Exit with error if any issues found
if [ -n "$NOT_RUNNING" ] || [ -n "$CRASHLOOP" ]; then
    echo "Health check failed. See $LOG_FILE for details."
    exit 1
fi

echo "Health check passed."
exit 0
