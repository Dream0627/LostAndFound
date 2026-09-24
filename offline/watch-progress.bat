@echo off
title LAF build progress monitor
setlocal
echo ============================================================
echo   LAF build progress monitor
echo   Refresh every 5 seconds. Keep the build window running too.
echo   Close THIS window to stop monitoring.
echo ============================================================
timeout /t 2 >nul
:loop
cls
echo ==== %DATE% %TIME% ====
echo.
echo --- docker images (a row appears once an image is fully pulled) ---
docker images
echo.
echo --- docker system df ---
docker system df
echo.
echo --- build.log tail (if exists) ---
if exist "dist-offline\build.log" (
  powershell -NoProfile -Command "Get-Content -Tail 8 'dist-offline\build.log'"
) else (
  echo (no build.log in this folder; make sure you run monitor from project root)
)
echo.
echo ------------------------------------------------------------
echo  Reading:
echo   - MySQL layers are big; "docker images" may stay EMPTY for
echo     several minutes while pulling. That is NORMAL.
echo   - If a mysql:8.0 row appears, the pull finished.
echo   - If NOTHING changes for 10+ minutes, the pull is stuck:
echo     see MIRROR_FIX.md to set a registry mirror.
echo ------------------------------------------------------------
timeout /t 5 >nul
goto loop
