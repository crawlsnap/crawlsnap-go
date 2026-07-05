#!/usr/bin/env bash
# Refresh the typed models in models/ from the public OpenAPI contract.
#
# Only the models are regenerated; the hand-written facade (client.go, base.go,
# resources.go, errors.go, types.go, version.go) is never touched. A throwaway
# full client is generated into a temp dir and only its scan-data models are
# copied in (the envelope/response wrappers are handled by the facade).
#
# Source of truth: ../crawlsnap-contracts/dist/crawlsnap-v1.yaml
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONTRACTS="${CONTRACTS_DIR:-$REPO_ROOT/../crawlsnap-contracts}"
SPEC="$CONTRACTS/dist/crawlsnap-v1.yaml"
TMP="$REPO_ROOT/.gen-tmp"

# 1. Bundle the contract (inlines the data-product schemas).
( cd "$CONTRACTS" && make bundle )

# 2. Generate a throwaway full client (generator version pinned in openapitools.json).
rm -rf "$TMP"
npx -y @openapitools/openapi-generator-cli@latest generate \
  -i "$SPEC" -g go -o "$TMP" \
  --additional-properties=packageName=models,isGoSubmodule=true,hideGenerationTimestamp=true,withGoMod=false,generateInterfaces=false \
  >/dev/null

# 3. Replace the typed payload models in place; leave the facade untouched.
rm -f "$REPO_ROOT"/models/*.go
cp "$TMP"/model_ioc_*_scan_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_pulse_*_scan_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_subdo_snap_scan_data.go "$REPO_ROOT/models/"
# SportSnap payload models + their nested sub-schemas.
cp "$TMP"/model_channel_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_channel_schedule_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_channel_schedule_entry.go "$REPO_ROOT/models/"
cp "$TMP"/model_match_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_match_event.go "$REPO_ROOT/models/"
cp "$TMP"/model_match_stat.go "$REPO_ROOT/models/"
cp "$TMP"/model_match_status.go "$REPO_ROOT/models/"
cp "$TMP"/model_country_channels_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_country_channel.go "$REPO_ROOT/models/"
cp "$TMP"/model_country_broadcast.go "$REPO_ROOT/models/"
cp "$TMP"/model_daily_schedule_data.go "$REPO_ROOT/models/"
cp "$TMP"/model_competition_schedule.go "$REPO_ROOT/models/"
cp "$TMP"/model_competition_ref.go "$REPO_ROOT/models/"
cp "$TMP"/model_scheduled_match.go "$REPO_ROOT/models/"
cp "$TMP"/model_broadcast_channel.go "$REPO_ROOT/models/"
cp "$TMP"/model_broadcast_right.go "$REPO_ROOT/models/"
cp "$TMP"/model_team_ref.go "$REPO_ROOT/models/"
cp "$TMP"/model_score.go "$REPO_ROOT/models/"
cp "$TMP"/model_lineup.go "$REPO_ROOT/models/"
cp "$TMP"/model_lineups.go "$REPO_ROOT/models/"
cp "$TMP"/utils.go "$REPO_ROOT/models/"
rm -rf "$TMP"

# 4. Tidy + format.
( cd "$REPO_ROOT" && gofmt -w models/ && go build ./... )

echo "==> Refreshed models/ from $SPEC (facade left intact)"
