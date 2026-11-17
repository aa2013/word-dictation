# word-dictation front

单词听写应用前端，Vue2 + [Vuetify2](https://v2.vuetifyjs.com/zh-Hans) + [ElementUI](https://element.eleme.cn/#/zh-CN)（表格/分页）+ Axios。

## 开发

```bash
npm install
npm run serve        # http://localhost:8080，/api 与 /covers 代理到 http://localhost:8099
```

需先启动后端（见仓库根 README）。

## 构建

```bash
npm run build        # 产物 dist/，由根目录构建脚本拷贝到 server/web/dist 嵌入后端
```

> Node 17+ 构建需 `NODE_OPTIONS=--openssl-legacy-provider`（根目录脚本已处理）。

## 结构

```
src/
├── main.js              # 入口，全局组件注册
├── router/              # 路由（/login /register /root 下含公开词库、我的词库、生成方案）
├── network/             # axios 封装与 API 定义（request.js 为统一拦截器）
├── views/               # 页面：Login / Register / Root / CommonLib / MyLib / Plan / Print
├── components/          # LibCard、DetailLib、打印预览、Dialog/SnackBar/Loading
├── store/               # Vuex（页面查询参数缓存）
└── plugins/             # Vuetify / ElementUI 注册
```
