@echo off

REM Переход в директорию скрипта
cd /d %~dp0

REM Сборка образа
echo ================================
echo 1. Building Docker image...
echo ================================
docker build -f Dockerfile.test -t rocks-tests ../../..
IF ERRORLEVEL 1 (
    echo.
    echo Error build image.
    pause
    EXIT /B %ERRORLEVEL%
)

REM Запуск контейнера
echo.
echo ================================
echo 2. Running Docker container...
echo ================================
docker run --rm rocks-tests
IF ERRORLEVEL 1 (
    echo.
    echo Error run cantainer.
) ELSE (
    echo.
    echo Success.
)
