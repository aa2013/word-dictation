@echo off
rem 一键构建：前端产物嵌入 Go，输出单二进制 build\word-dictation.exe
setlocal
cd /d "%~dp0..\front"

if not exist node_modules (
  npm install --no-audit --no-fund
)

set NODE_OPTIONS=--openssl-legacy-provider
call npm run build
if errorlevel 1 exit /b 1

cd ..\server
if exist web\dist rmdir /s /q web\dist
xcopy /e /i /y ..\front\dist web\dist >nul
if not exist ..\build mkdir ..\build
if exist ..\build\word-dictation.exe del /q ..\build\word-dictation.exe
go build -o ..\build\word-dictation.exe
if errorlevel 1 exit /b 1

echo 构建完成: build\word-dictation.exe
