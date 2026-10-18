#!/usr/bin/env bash
# CBT SMK v2 — start.sh (Tim 6)
# Build frontend + binary Go, lalu jalankan aplikasi.
# Dipakai untuk jalan manual / server tanpa systemd.
# Untuk produksi sekolah, pakai unit systemd deploy/cbt-v2.service.
#
# Prasyarat: Go, Node.js + npm, PostgreSQL, file .env terisi.
# Jalankan dari root repo:  ./deploy/start.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [ ! -f .env ]; then
  echo "ERROR: file .env tidak ditemukan. Salin dulu: cp .env.example .env lalu isi DATABASE_URL & SESSION_SECRET." >&2
  exit 1
fi

echo "[1/3] Build frontend..."
npm --prefix frontend install
npm --prefix frontend run build

echo "[2/3] Build binary Go..."
# Modul Go ada di backend/ (backend/go.mod) — build dari sana.
(cd backend && go build -o ../cbt-smk-v2 ./cmd/server)

echo "[3/3] Menjalankan ./cbt-smk-v2 (port dari env PORT, default 9988)..."
exec ./cbt-smk-v2
