@echo off
setlocal enabledelayedexpansion
cd /d "%~dp0.."
title LAF offline image build

echo ============================================================
echo   LAF offline image build and export
echo   Project root : %CD%
echo ============================================================
echo   NOTE: pulling mysql:8.0 (about 600MB) can take SEVERAL MINUTES.
echo         The window may look quiet for a while - that is normal.
echo         Do NOT close it. Layer progress lines will appear.
echo ============================================================
echo.

where docker >nul 2>nul
if errorlevel 1 (
  echo [ERROR] "docker" command not found. Install Docker Desktop and start it.
  goto :end
)
docker info >nul 2>nul
if errorlevel 1 (
  echo [ERROR] Docker engine is not running. Open Docker Desktop until "Engine running".
  goto :end
)

echo [STEP 1] Pulling base images (official Docker Hub first) ...
echo ============================================================
for %%I in (mysql:8.0 golang:1.26.5-alpine node:20-alpine nginx:1.27-alpine alpine:3.20) do call :pull "%%I"
echo.

echo [STEP 2] docker compose build ...
echo ============================================================
docker compose build
if errorlevel 1 (
  echo.
  echo [FAILED] docker compose build returned an error ^(see above^).
  goto :end
)
echo.

echo [STEP 3] docker save -o dist-offline\laf-images.tar ...
echo ============================================================
if not exist "dist-offline" mkdir "dist-offline"
docker save -o "dist-offline\laf-images.tar" laf-backend:latest laf-frontend:latest mysql:8.0
if errorlevel 1 (
  echo.
  echo [FAILED] docker save returned an error ^(see above^).
  goto :end
)

if not exist "dist-offline\laf-images.tar" (
  echo.
  echo [FAILED] tar file was NOT created.
  goto :end
)

echo.
echo ============================================================
echo   DONE. Output: %CD%\dist-offline\laf-images.tar
echo ============================================================
dir "dist-offline\laf-images.tar"
echo.
echo Next:
echo   1) Zip the whole project (include dist-offline; may exclude frontend-demo\node_modules)
echo   2) Upload to server via Aliyun Workbench "Upload File" into /root
echo   3) On server run:  cd /opt/LAF ^&^& bash offline/load-and-run.sh
goto :end

rem ============================================================
rem  :pull  ^<image^>  -> official Docker Hub first, mirror as fallback
rem ============================================================
:pull
set "IMG=%~1"
echo.
echo   --- pulling !IMG! from Docker Hub official ...
docker pull !IMG!
if not errorlevel 1 (
  echo   [OK] !IMG!
  exit /b 0
)
echo   official failed, try mirrors for !IMG! ...
docker pull docker.m.daocloud.io/library/!IMG!
if not errorlevel 1 (
  docker tag docker.m.daocloud.io/library/!IMG! !IMG!
  echo   [OK] !IMG! ^(daocloud^)
  exit /b 0
)
docker pull docker.1panel.live/library/!IMG!
if not errorlevel 1 (
  docker tag docker.1panel.live/library/!IMG! !IMG!
  echo   [OK] !IMG! ^(1panel^)
  exit /b 0
)
echo   [WARN] !IMG! not pre-pulled; compose build will retry.
exit /b 0

:end
echo.
echo (window stays open)
pause
exit /b 0
