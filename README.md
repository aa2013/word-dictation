# word-dictation

单词听写/默词本应用。Go 单二进制后端 + Vue2 前端，SQLite 存储，前端经 `go:embed` 打包进
后端，构建产物只有一个可执行文件。

- 后端：gin + gorm + SQLite（纯 Go 驱动 [glebarez/sqlite](https://github.com/glebarez/sqlite)，无 CGO，可任意交叉编译）
- 前端：Vue2 + Vuetify + Element UI
- 存储：SQLite（WAL 模式），版本化迁移管理表结构与种子数据
- 认证：JWT(HS512) + `session` 表吊销 + bcrypt

## 目录结构

```
word-dictation/
├── front/                  # Vue2 前端
├── server/                 # Go 后端
│   ├── main.go
│   ├── config/             # 配置加载（config.yaml，缺省自动生成）
│   ├── migration/          # 版本化迁移 + 种子数据
│   ├── handler/            # 控制器
│   ├── service/            # 业务逻辑（含在线词典翻译器）
│   ├── model/              # GORM 模型
│   ├── dto/                # 请求/响应 DTO
│   ├── middleware/         # JWT 认证 / CORS
│   ├── router/             # 路由与 SPA 静态服务（含端到端测试）
│   ├── util/               # 响应封装、JWT、随机、正则
│   └── web/                # go:embed 嵌入的前端构建产物
├── scripts/                # build.sh|bat 一键构建
```

## 构建

依赖：Go 1.21+、Node 16+。

```bash
./scripts/build.sh          # Linux / macOS / Git Bash
scripts\build.bat           # Windows

# 产物输出到 build/word-dictation(.exe)

# 或手动分步
cd front && npm install && npm run build
# 将 front/dist 拷贝到 server/web/dist 后：
cd ../server && go build -o ../build/word-dictation.exe
```

> vue-cli 4 在 Node 17+ 构建需要 `NODE_OPTIONS=--openssl-legacy-provider`，构建脚本已自动处理。

## 运行

```bash
./word-dictation.exe          # 默认端口 8099，浏览器打开 http://127.0.0.1:8099
```

- 首次运行自动生成 `config.yaml`（JWT 密钥自动随机生成）与 `data/word-dictation.db`
- 数据库迁移自动执行：建表并导入内置词库数据（9 个词库、787 个单词、7391 条释义）
- 注册验证码默认走 `email.debug` 模式（验证码打印到控制台），配置 SMTP 后发送真实邮件

### 系统初始化

首次部署（全新数据库）时，访问系统会自动跳转到 `/setup` 初始化页面，设置管理员账号与密码
（更新内置管理员用户，不新建）。初始化标记持久化在 `setting` 表：

- 前端检测到未初始化时跳转初始化页，提交前先校验状态
- 服务端 `POST /api/setup/init` 严格校验未初始化状态，**已初始化后永久返回 403**，防止接口被扫描利用

普通用户通过前端注册流程创建（验证码见 `email.debug` 说明）。

## 配置

`config.yaml`（首次运行自动生成）：

| 配置 | 说明 | 默认值 |
| --- | --- | --- |
| `port` | 监听端口 | `8099` |
| `db.path` | SQLite 数据库路径 | `data/word-dictation.db` |
| `storage.dir` | 封面图片存储目录（含内置默认封面 `default.jpg`，经 `/covers/*` 提供访问） | `data/storage` |
| `jwt.secret` | 签名密钥（自动生成，勿泄露） | 随机 |
| `jwt.tokenDays` / `jwt.rememberDays` | 登录态有效期（天），后者为勾选"记住我" | `7` / `30` |
| `email.*` | SMTP 配置；`debug: true` 时不真正发信，验证码打印到控制台 | `debug: true` |

## API

所有接口在 `/api` 前缀下，RESTful 风格，登录后携带 `Authorization: Bearer <token>`。

- 初始化：`GET /api/setup/status`、`POST /api/setup/init`（仅未初始化状态可用）
- 认证：`POST /api/auth/login`、`POST /api/auth/logout`、`GET /api/auth/me`
- 用户：`POST /api/users`（注册）、`POST /api/users/codes/*`（邮箱验证码）、`PUT /api/users/password`
- 词库：`GET/POST /api/libs`、`GET/PUT/DELETE /api/libs/:id`、`POST /api/libs/:id/adopt`
- 单词：`GET/POST /api/libs/:id/words`、`GET /api/libs/:id/random-words`、
  `PUT /api/libs/:id/default-explain`、`GET /api/words/:wordId/explains`
- 方案：`GET/POST /api/plans`、`GET /api/plans/:id/words`、`DELETE /api/plans/:id`

响应包体 `{code, message, data}`，HTTP 状态码即语义：`400` 参数错误 / `401` 未登录或密码错误 /
`403` 无权限 / `404` 不存在 / `409` 重复 / `502` 词典查询失败 / `500` 内部错误。

## 数据迁移

`server/migration/migrations/` 下按 `000N_名称.sql` 命名的 SQL 脚本在启动时按序自动执行，
记录存于 `schema_migrations` 表。新增功能直接追加迁移文件并调整 `model/`。

## 测试

```bash
cd server && go test ./...
```

覆盖：迁移与种子数据完整性、登录/登出/会话吊销、注册验证码流程、词库权限、
单词导入（mock 词典）、默认释义切换、方案 CRUD。
