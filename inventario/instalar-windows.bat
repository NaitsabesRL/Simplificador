@echo off
rem ============================================================
rem  INSTALADOR - Gestion de Inventario (pollos, carnes, embutidos)
rem  DOBLE CLIC para instalar. No necesita permisos de administrador.
rem  Hace 4 cosas: copia el programa, crea acceso directo en el
rem  Escritorio y en el menu Inicio, registra el comando "inventario"
rem  en el PATH y verifica que funcione.
rem ============================================================
setlocal enabledelayedexpansion

set "NOMBRE=Inventario"
set "CARPETA=%LOCALAPPDATA%\Inventario"
set "EXE=%CARPETA%\inventario.exe"

echo ===============================================
echo   Instalacion de %NOMBRE% para Windows
echo ===============================================

rem 1) Localizar el binario junto a este script
if not exist "%~dp0inventario.exe" (
    echo [ERROR] No se encontro "inventario.exe" junto a este instalador.
    echo Copie instalar-windows.bat e inventario.exe en la misma carpeta.
    pause
    exit /b 1
)

rem 2) Copiar el programa
if not exist "%CARPETA%" mkdir "%CARPETA%"
copy /y "%~dp0inventario.exe" "%EXE%" >nul
if errorlevel 1 (
    echo [ERROR] No se pudo copiar el programa.
    pause
    exit /b 1
)
echo   [OK] Programa instalado en: %EXE%

rem 3) Agregar al PATH del usuario (sin duplicar)
set "PATHUSER="
for /f "tokens=2,*" %%a in ('reg query "HKCU\Environment" /v Path 2^>nul ^| find "REG_SZ"') do set "PATHUSER=%%b"
echo !PATHUSER! | find /i "%CARPETA%" >nul
if errorlevel 1 (
    if defined PATHUSER (set "PATHNUEVO=!PATHUSER!;%CARPETA%") else (set "PATHNUEVO=%CARPETA%")
    reg add "HKCU\Environment" /v Path /t REG_EXPAND_SZ /d "!PATHNUEVO!" /f >nul
    echo   [OK] Comando "inventario" agregado al PATH.
)

rem 4) Accesos directores (Escritorio y menu Inicio) via VBScript embebido
set "VBS=%TEMP%\instalar_inventario.vbs"
> "%VBS%" echo Set s = CreateObject("WScript.Shell")
>>"%VBS%" echo Set l = s.CreateShortcut(s.SpecialFolders("Desktop") ^& "\%NOMBRE%.lnk")
>>"%VBS%" echo l.TargetPath = "%EXE%" : l.WorkingDirectory = "%CARPETA%" : l.Description = "Gestion de inventario" : l.Save
>>"%VBS%" echo Set l2 = s.CreateShortcut(s.SpecialFolders("Programs") ^& "\%NOMBRE%.lnk")
>>"%VBS%" echo l2.TargetPath = "%EXE%" : l2.WorkingDirectory = "%CARPETA%" : l2.Save
cscript //nologo "%VBS%"
del "%VBS%" >nul 2>&1
echo   [OK] Accesos directores creados (Escritorio y menu Inicio).

rem 5) Verificacion
"%EXE%" --version >nul 2>&1
if errorlevel 1 (
    echo   [AVISO] No se pudo verificar en esta consola, pero la
    echo           instalacion de archivos se completo correctamente.
) else (
    echo   [OK] Verificado: el programa funciona.
)

echo.
echo  ============= INSTALACION COMPLETA =============
echo  * Doble clic en el icono "%NOMBRE%" del Escritorio para usarlo.
echo  * O escriba "inventario" en una terminal nueva.
echo  * Sus datos se guardan en: %%APPDATA%%\inventario
echo.
echo  Para desinstalar: borre la carpeta %CARPETA%
echo  y los accesos directores "%NOMBRE%".
echo.
pause
endlocal
