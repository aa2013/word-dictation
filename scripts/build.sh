#!/usr/bin/env bash
# 一键构建：前端产物嵌入 Go，输出单二进制 server/word-dictation.exe
set -e
cd "$(dirname "$0")/../front"

if [ ! -d node_modules ]; then
  npm install --no-audit --no-fund
fi

# 旧版 vue-cli(webpack4) 在 Node 17+ 需要 legacy OpenSSL provider
NODE_OPTIONS=--openssl-legacy-provider npm run build

cd ../server
rm -rf web/dist
cp -r ../front/dist web/dist
go build -o word-dictation.exe
echo "构建完成: server/word-dictation.exe"
