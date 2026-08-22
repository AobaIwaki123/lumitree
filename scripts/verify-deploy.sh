#!/usr/bin/env bash
# ==============================================================================
# verify-deploy.sh - Kubernetes deployment verification script for lumitree
#
# Checks ArgoCD Application sync status, Kubernetes deployment rollout,
# live pod logs, and external Ingress endpoint connectivity.
# ==============================================================================

set -euo pipefail

NAMESPACE="lumitree"
APP_NAME="lumitree"
INGRESS_HOST="lumitree.aooba.net"

echo "========================================================"
echo "1. Checking ArgoCD Application Status..."
echo "========================================================"
if kubectl get application "$APP_NAME" -n argocd >/dev/null 2>&1; then
  kubectl get application "$APP_NAME" -n argocd -o wide
else
  echo "Note: ArgoCD Application not found in 'argocd' namespace or kubectl not configured for cluster."
fi
echo ""

echo "========================================================"
echo "2. Checking Deployment Rollout in namespace '${NAMESPACE}'..."
echo "========================================================"
if kubectl get deployment "$APP_NAME" -n "$NAMESPACE" >/dev/null 2>&1; then
  kubectl rollout status deployment/"$APP_NAME" -n "$NAMESPACE" --timeout=60s
  kubectl get pods -n "$NAMESPACE" -l app="$APP_NAME" -o wide
else
  echo "Note: Deployment '${APP_NAME}' not found in namespace '${NAMESPACE}'."
fi
echo ""

echo "========================================================"
echo "3. Testing Live Ingress Health Endpoint..."
echo "========================================================"
if curl -fsS "https://${INGRESS_HOST}/healthz" >/dev/null 2>&1; then
  echo "OK: https://${INGRESS_HOST}/healthz is responding successfully!"
  curl -s "https://${INGRESS_HOST}/healthz" | jq . || curl -s "https://${INGRESS_HOST}/healthz"
else
  echo "Warning: Could not reach https://${INGRESS_HOST}/healthz (DNS propagation or cluster ingress pending)."
fi
echo ""

echo "========================================================"
echo "Deployment verification check completed."
echo "========================================================"
