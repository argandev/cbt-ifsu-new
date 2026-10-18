@echo off
REM =====================================================================
REM CBT SMK v2 — build.bat (fallback Windows TANPA make)
REM Jalankan dari ROOT repo, di cmd.exe atau PowerShell:
REM   build.bat
REM Hasil: .\cbt-smk-v2.exe (frontend ter-embed, siap jalan).
REM
REM PENTING: matikan cbt-smk-v2.exe yang sedang jalan dulu (Ctrl+C)
REM SEBELUM menjalankan file ini — Windows mengunci file .exe yang
REM sedang jalan, dan build akan gagal "Access is denied" bila tidak
REM dimatikan. Bila build gagal karena itu, tutup server lama lalu
REM jalankan build.bat lagi.
REM
REM Alternatif bila punya Git Bash / mingw32-make:  make build-windows
REM (lihat CARA-JALAN-WINDOWS.md).
REM =====================================================================
setlocal
cd /d "%~dp0"

echo [1/4] Cek prasyarat (go + npm)...
where go >nul 2>&1
if errorlevel 1 (
  echo GAGAL: perintah 'go' tidak ditemukan. Install Go 1.26+ dari https://go.dev/dl/ lalu buka cmd baru.
  goto :fail
)
where npm >nul 2>&1
if errorlevel 1 (
  echo GAGAL: perintah 'npm' tidak ditemukan. Install Node.js LTS dari https://nodejs.org/ lalu buka cmd baru.
  goto :fail
)
go version
node --version

echo [2/4] Build frontend (npm install + npm run build)...
cd frontend
call npm install
if errorlevel 1 (
  echo GAGAL: 'npm install' di frontend gagal. Lihat pesan error di atas.
  cd /d "%~dp0"
  goto :fail
)
call npm run build
if errorlevel 1 (
  echo GAGAL: 'npm run build' di frontend gagal. Lihat pesan error di atas.
  cd /d "%~dp0"
  goto :fail
)
cd /d "%~dp0"

echo [3/4] Salin frontend/dist -^> backend/dist (agar ter-embed ke binary)...
if exist backend\dist rmdir /S /Q backend\dist
xcopy /E /I /Y frontend\dist backend\dist >nul
if errorlevel 1 (
  echo GAGAL: gagal menyalin frontend\dist ke backend\dist.
  goto :fail
)

echo [4/4] Build binary Windows (GOOS=windows GOARCH=amd64)...
cd backend
set GOOS=windows
set GOARCH=amd64
set GOFLAGS=-trimpath
go build -o ..\cbt-smk-v2.exe .\cmd\server
if errorlevel 1 (
  echo GAGAL: 'go build' gagal.
  echo HINT: bila pesannya "Access is denied" / file sedang dipakai, berarti
  echo cbt-smk-v2.exe lama masih jalan. Matikan dulu (Ctrl+C di jendela
  echo server lama, atau tutup jendelanya), lalu jalankan build.bat lagi.
  cd /d "%~dp0"
  goto :fail
)
cd /d "%~dp0"

echo.
echo SUKSES: .\cbt-smk-v2.exe berhasil dibuat.
echo Jalankan:  cbt-smk-v2.exe
echo Lalu buka: http://localhost:9988
exit /b 0

:fail
echo.
echo BUILD GAGAL. Perbaiki error di atas, lalu jalankan build.bat lagi.
pause
exit /b 1
