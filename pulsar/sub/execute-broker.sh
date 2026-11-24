#!/usr/bin/env bash
set -euo pipefail

PULSAR_HOME="/mnt/d/pulsar/apache-pulsar-5.0.0-M1"
ADMIN_URL="${PULSAR_ADMIN_URL:-http://localhost:8080}"
ADMIN="$PULSAR_HOME/bin/pulsar-admin"

if [[ ! -x "$PULSAR_HOME/bin/pulsar-daemon" || ! -x "$ADMIN" ]]; then
  echo "Pulsar commands not found under $PULSAR_HOME/bin"
  exit 1
fi

for command in curl jq; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "$command is required."
    exit 1
  fi
done

broker_running() {
  curl --silent --fail --max-time 2 \
    "$ADMIN_URL/admin/v2/clusters" >/dev/null 2>&1
}

if broker_running; then
  echo "Pulsar is already running."
else
  echo "Starting Pulsar standalone..."
  "$PULSAR_HOME/bin/pulsar-daemon" start standalone

  for _ in {1..60}; do
    if broker_running; then
      break
    fi
    sleep 1
  done

  if ! broker_running; then
    echo "Pulsar did not become ready. Check $PULSAR_HOME/logs."
    exit 1
  fi
fi

echo "Truncating existing persistent topics..."

#for tenant in $("$ADMIN" tenants list | jq -r '.[]'); do
#for tenant in $("$ADMIN" tenants list); do
#  while IFS= read -r namespace; do
#    [[ -z "$namespace" ]] && continue
#
#    namespace_name="${namespace#*/}"
#    mapfile -t topics < <(
#      {
#        "$ADMIN" topics list "$namespace" | jq -r '.[]'
#        "$ADMIN" topics list "$namespace"
#        "$ADMIN" topics list-partitioned-topics "$namespace" | jq -r '.[]'
#        "$ADMIN" topics list-partitioned-topics "$namespace"
#      } | sort -u
#    )
#    for topic in "${topics[@]}"; do
#      [[ "$topic" == persistent://* ]] || continue
#      echo "Truncating $topic"
#      "$ADMIN" topics truncate "$topic"
#    done
#  done < <("$ADMIN" namespaces list "$tenant" | jq -r '.[]')
#  done < <("$ADMIN" namespaces list "$tenant")
#done
echo "Truncating 'aina' <ALERT: Always check the 'topic name' configured in config.json>"
"$ADMIN" topics truncate "aina"

#echo "Finished. Topic names were retained."