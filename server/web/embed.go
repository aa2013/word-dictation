// Package web 嵌入前端构建产物（front/dist 拷贝到 web/dist 后由 go:embed 打包）。
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
