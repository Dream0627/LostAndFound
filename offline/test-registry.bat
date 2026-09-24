@echo off
setlocal
cd /d "%~dp0.."
title LAF registry connectivity test

echo ============================================================
echo   Docker registry connectivity test  (fast, about 1 minute)
echo   Project root : %CD%
echo ============================================================
echo.

where docker >nul 2>nul
if errorlevel 1 ( echo [ERROR] docker not found. Install and start Docker Desktop. & goto :end )
docker info >nul 2>nul
if errorlevel 1 ( echo [ERROR] Docker engine not running. Open Docker Desktop. & goto :end )

echo [TEST 1] Docker Hub default - tiny image hello-world ...
echo ------------------------------------------------------------
docker pull hello-world
echo ------------------------------------------------------------
echo TEST1 exit = %errorlevel%
echo.

echo [TEST 2] China mirror daocloud - hello-world ...
echo ------------------------------------------------------------
docker pull docker.m.daocloud.io/library/hello-world
echo ------------------------------------------------------------
echo TEST2 exit = %errorlevel%
echo.

echo [TEST 3] China mirror 1panel - hello-world ...
echo ------------------------------------------------------------
docker pull docker.1panel.live/library/hello-world
echo ------------------------------------------------------------
echo TEST3 exit = %errorlevel%
echo.

echo [LOCAL IMAGES]
docker images
echo.

echo HOW TO READ THIS:
echo   - If TEST1 finished in a few seconds: network is OK.
echo     Then big images (mysql:8.0 ~600MB) are just SLOW - please wait.
echo   - If all tests hang / show timeout / TLS errors: network or proxy problem.
echo     See MIRROR_FIX.md (set registry-mirrors in Docker Desktop).

:end
echo.
echo (window stays open)
pause
exit /b 0
