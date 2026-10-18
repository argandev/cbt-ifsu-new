#!/usr/bin/env bash
# CBT SMK v2 — backup.sh (Tim 6)
# Backup PostgreSQL (format arsip pg_dump, + timestamp) + file upload.
#
# Penggunaan:
#   ./deploy/backup.sh                       # sekali jalan
#   crontab: 0 2 * * * /opt/cbt-smk-v2/deploy/backup.sh >> /var/log/cbt-backup.log 2>&1
#
# Konfigurasi lewat env (dengan default yang masuk akal):
#   CBT_DB_NAME        (default: cbt_smk_v2)
#   CBT_DB_USER        (default: cbt)
#   CBT_BACKUP_DIR     (default: <repo>/backups)
#   CBT_UPLOAD_DIR     (default: <repo>/uploads)
#   CBT_BACKUP_KEEP_DAYS (default: 30 — file backup lebih lama dihapus)
# Koneksi memakai PGPASSWORD bila di-set, atau .pgpass / peer auth.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DB_NAME="${CBT_DB_NAME:-cbt_smk_v2}"
DB_USER="${CBT_DB_USER:-cbt}"
BACKUP_DIR="${CBT_BACKUP_DIR:-$ROOT/backups}"
UPLOAD_DIR="${CBT_UPLOAD_DIR:-$ROOT/uploads}"
KEEP_DAYS="${CBT_BACKUP_KEEP_DAYS:-30}"
TS="$(date +%Y%m%d-%H%M%S)"

mkdir -p "$BACKUP_DIR"

DB_DUMP="$BACKUP_DIR/db-$DB_NAME-$TS.dump"
echo "Backup database $DB_NAME -> $DB_DUMP"
pg_dump -U "$DB_USER" -d "$DB_NAME" -F c -f "$DB_DUMP"

if [ -d "$UPLOAD_DIR" ]; then
  UPLOAD_TAR="$BACKUP_DIR/uploads-$TS.tar.gz"
  echo "Backup upload $UPLOAD_DIR -> $UPLOAD_TAR"
  tar -czf "$UPLOAD_TAR" -C "$(dirname "$UPLOAD_DIR")" "$(basename "$UPLOAD_DIR")"
else
  echo "Direktori upload $UPLOAD_DIR tidak ada — lewati backup upload."
fi

# Catat versi aplikasi + migrasi saat backup diambil (diperlukan saat restore).
{
  echo "timestamp=$TS"
  echo "db=$DB_NAME"
  if [ -x "$ROOT/cbt-smk-v2" ]; then
    echo "binary=$("$ROOT/cbt-smk-v2" --version 2>/dev/null || echo 'tidak-diketahui')"
  fi
  echo "git=$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo 'bukan-repo-git')"
} > "$BACKUP_DIR/backup-$TS.meta.txt"

echo "Hapus backup lebih lama dari $KEEP_DAYS hari..."
find "$BACKUP_DIR" -maxdepth 1 -type f \( -name 'db-*.dump' -o -name 'uploads-*.tar.gz' -o -name 'backup-*.meta.txt' \) -mtime +"$KEEP_DAYS" -delete || true

echo "Selesai. Salin satu salinan backup ke LUAR server (flashdisk/komputer lain)."
ls -lh "$BACKUP_DIR" | tail -n 5
