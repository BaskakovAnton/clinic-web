#!/bin/sh
set -e

if [ "${AUTO_MIGRATE}" = "true" ] || [ "${AUTO_MIGRATE}" = "1" ]; then
  MIG_URL="${MIGRATION_DATABASE_URL:-$DATABASE_URL}"
  if [ -z "$MIG_URL" ]; then
    echo "AUTO_MIGRATE set but no MIGRATION_DATABASE_URL/DATABASE_URL" >&2
    exit 1
  fi

  EXISTS="$(psql "$MIG_URL" -tAc "SELECT to_regclass('public.patients')" | tr -d '[:space:]')"
  if [ "$EXISTS" = "patients" ]; then
    echo "Schema already present (patients exists); skipping sborka migrate."
  else
    echo "Running sborka migrations..."
    for f in \
      /app/sborka/01_schema.sql \
      /app/sborka/02_views.sql \
      /app/sborka/03_roles.sql \
      /app/sborka/04_seed.sql \
      /app/sborka/06_amvera_bootstrap.sql
    do
      echo "==> $f"
      psql "$MIG_URL" -v ON_ERROR_STOP=1 -f "$f"
    done
    echo "Migrations finished."
  fi

  # Always ensure connect + passwords (idempotent)
  echo "Ensuring Amvera login grants..."
  psql "$MIG_URL" -v ON_ERROR_STOP=1 -f /app/sborka/06_amvera_bootstrap.sql
fi

exec /app/clinic
