# word-dictation（Go + Vue2 + SQLite 单体版）

单词听写/默词本应用。由原 `Java Spring Boot + MySQL + Redis + Vue2` 项目完整移植为
**Go (gin + gorm) + SQLite（纯 Go 驱动，无 CGO）+ Vue2** 的单二进制应用，前端通过
`go:embed` 打进后端，最终产物只有一个可执行文件。

## 目录结构

```
word-dictation/
├── front/                  # Vue2 前端（Vuetify + Element UI）
├── server/                 # Go 后端
│   ├── main.go
│   ├── config/             # 配置加载（config.yaml，缺省自动生成）
│   ├── migration/          # SQLite 版本化迁移（含 MySQL 数据导入的种子迁移）
│   ├── handler/            # 控制器
│   ├── service/            # 业务逻辑（含在线词典翻译器移植）
│   ├── model/              # GORM 模型
│   ├── dto/                # 请求/响应 DTO
│   ├── middleware/         # JWT 认证 / CORS
│   ├── router/             # 路由与 SPA 静态服务（含端到端测试）
│   ├── util/               # 结果封装、JWT、随机、正则
│   ├── web/                # go:embed 嵌入的前端构建产物
│   └── scripts/            # word.sql → SQLite 转换脚本（已执行，留存备查）
└── scripts/build.sh|bat    # 一键构建脚本
```

## 构建

依赖：Go 1.21+、Node 16+。

```bash
./scripts/build.sh          # Linux / macOS / Git Bash
scripts\build.bat           # Windows

# 或手动分步
cd front && npm install && npm run build
# 将 front/dist 拷贝到 server/web/dist 后：
cd ../server && go build -o word-dictation.exe
```

> 旧版 vue-cli 4 在 Node 17+ 构建需要 `NODE_OPTIONS=--openssl-legacy-provider`，
> 构建脚本已自动处理。

## 运行

```bash
./word-dictation.exe          # 默认端口 8099，浏览器打开 http://127.0.0.1:8099
```

- 首次运行自动生成 `config.yaml`（可改端口/SMTP 等，JWT 密钥自动随机生成）与 `data/word-dictation.db`
- SQLite 通过版本化迁移自动建表并导入原 MySQL 全量数据（9 个词库、787 个单词、7391 条释义、13 个方案等）
- 原库中的 `admin` 账号保留但密码未知（旧 bcrypt 哈希），建议走注册流程创建自己的账号：
  本地未配置 SMTP 时 `email.debug: true`（默认开启），验证码直接打印到控制台

## API 与认证（重新设计）

- 所有接口在 `/api` 前缀下，RESTful 风格
- 认证：`POST /api/auth/login`（bcrypt 校验）签发 **JWT(HS512)**，`Authorization: Bearer <token>` 携带；
  服务端 `session` 表保存 jti 用于登出吊销与过期清理（替代 Redis）
- 响应包体 `{code, message, data}`，HTTP 状态码即语义：
  `400` 参数错误 / `401` 未登录或密码错误 / `403` 无权限 / `404` 不存在 / `409` 重复 / `502` 翻译源失败 / `500` 内部错误
- 忘记密码/注册验证码存进程内存（5 分钟有效，1 分钟重发冷却）

## 主要改动（相对旧版）

| 旧版 | 新版 |
| --- | --- |
| MySQL | SQLite（`glebarez/sqlite` 纯 Go 驱动，免 CGO、免交叉编译） |
| Redis（会话 + 邮件验证码） | SQLite `session` 表 + 进程内 TTL 缓存 |
| RSA 加密登录 + 密钥混淆 | 明文 JSON 登录 + bcrypt（传输安全交给 HTTPS），JWT 重新设计 |
| 恒定 HTTP 200 + 自定义业务码 | 标准 HTTP 状态码 + `{code, message, data}` |
| 七牛云存储封面 | 本地磁盘 `data/storage`，经 `/covers/*` 提供服务 |
| 前端 axios 硬编码后端地址 | 生产同源相对路径 + devServer 代理，配合 embed 单体部署 |
| 无迁移机制 | `schema_migrations` 版本化迁移，后续加功能直接追加 SQL 文件 |

## 后续加功能

在 `server/migration/migrations/` 下新增 `000X_名称.sql`（重启自动执行），模型对应调整 `model/` 即可。

## 测试

```bash
cd server && go test ./...
```

覆盖：迁移与种子数据完整性（含 `o'clock` 等转义用例）、登录/登出/会话吊销、注册验证码流程、
词库增删与权限、单词导入（mock 翻译器）、重复导入 409、默认释义切换、方案 CRUD。
