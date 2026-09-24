@echo off
setlocal
cd /d "%~dp0.."
title LAF pack for upload

echo ============================================================
echo   LAF package for server upload
echo   Project root : %CD%
echo ============================================================
echo   Will create: %CD%\..\LAF.zip
echo   Include : everything, INCLUDING dist-offline\laf-images.tar
echo   Exclude : frontend-demo\node_modules, frontend-demo\dist, .git, .idea, .vscode
echo ============================================================
echo.

if not exist "dist-offline\laf-images.tar" (
  echo [WARN] dist-offline\laf-images.tar NOT found.
  echo        Run offline\build-and-save.bat first.
  echo.
)

where tar >nul 2>nul
if errorlevel 1 (
  echo [ERROR] tar.exe not found ^(needed on Windows 10 1803+^).
  echo Use the manual method in PACK_UPLOAD.md instead.
  goto :end
)

set "OUT=%~dp0..\LAF.zip"
if exist "%OUT%" del /f /q "%OUT%"

echo Packing ... (this may take a while for ~270MB image tar)
echo ------------------------------------------------------------
tar -a -c -f "%OUT%" ^
  --exclude="frontend-demo/node_modules" ^
  --exclude="frontend-demo/dist" ^
  --exclude=".git" ^
  --exclude=".idea" ^
  --exclude=".vscode" ^
  --exclude="LAF.zip" .
echo ------------------------------------------------------------
if not exist "%OUT%" (
  echo [FAILED] zip was not created.
  goto :end
)

echo.
echo ============================================================
echo   DONE. Package:
echo   %OUT%
echo ============================================================
dir "%OUT%"
echo.
echo Next:
echo   1) Aliyun console - Workbench - "Upload File" - select LAF.zip
echo      (upload it into /root)
echo   2) On the server run:
echo        mkdir -p /opt ^&^& cd /opt
echo        unzip -o /root/LAF.zip -d /opt/ ^&^& cd /opt/LAF
echo        bash offline/load-and-run.sh
goto :end

:end
echo.
echo (window stays open)
pause
exit /b 0
