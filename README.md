# word-dictation（Go + Vue2 + SQLite 单体版）

单词听写/默词本应用。由原 `Java Spring Boot + MySQL + Redis + Vue2` 项目完整移植为
**Go (gin + gorm) + SQLite（纯 Go 驱动，无 CGO）+ Vue2** 的单二进制应用，前端通过
`go:embed` 打进后端，最终产物只有一个可执行文件。

## 目录结构

```
word-dictation/
├── front/                  # Vue2 前端（Vuetify + Element UI）
│   └── src/...
├── server/                 # Go 后端
│   ├── main.go
│   ├── config/             # 配置加载（config.yaml，缺省自动生成）
│   ├── migration/          # SQLite 版本化迁移（含 MySQL 数据导入的种子迁移）
│   │   └── migrations/     # 0001 建表、0002 导入 word.sql 全量数据
│   ├── handler/            # 控制器（user/lib/word/plan）
│   ├── service/            # 业务逻辑（含在线词典翻译器移植）
│   ├── model/              # GORM 模型
│   ├── dto/                # 响应 DTO
│   ├── middleware/         # JWT 认证 / CORS
│   ├── util/               # 结果码、RSA、随机、正则
│   ├── web/                # go:embed 嵌入的前端构建产物
│   └── scripts/            # word.sql → SQLite 转换脚本（已执行，留存备查）
└── scripts/build.sh|bat    # 一键构建脚本
```

## 构建

依赖：Go 1.21+、Node 14+（建议 Node 16~22）。

```bash
# 一键构建（前端 + 后端单二进制）
./scripts/build.sh          # Linux / macOS / Git Bash
scripts\build.bat           # Windows

# 或手动分步
cd front && npm install && npm run build
# 构建产物 front/dist 拷贝到 server/web/dist 后：
cd ../server && go build -o word-dictation.exe
```

> 注意：旧版 vue-cli 4 在 Node 17+ 构建需要 `NODE_OPTIONS=--openssl-legacy-provider`，
> 构建脚本已自动处理。

## 运行

```bash
./word-dictation.exe          # 默认端口 8099，浏览器打开 http://127.0.0.1:8099
```

- 首次运行自动在工作目录生成 `config.yaml`（可改端口/邮箱 SMTP 等）与 `data/word-dictation.db`
- SQLite 数据库通过版本化迁移自动建表并导入原 MySQL 全量数据（9 个词库、约 780 个单词、全部释义/方案等）
- 原管理员账号保留（`admin`，密码同原系统 bcrypt 哈希）

## 主要改动（相对旧版）

| 旧版 | 新版 |
| --- | --- |
| MySQL | SQLite（纯 Go 驱动 `glebarez/sqlite`，免 CGO、免交叉编译烦恼） |
| Redis（token 会话 + 邮件验证码） | SQLite `token` 表 + 进程内 TTL 缓存 |
| Spring Security 过滤器链 | gin 中间件（错误码与响应包体结构保持不变） |
| 七牛云存储封面 | 本地磁盘 `data/storage`，经 `/covers/*` 提供服务 |
| 前端 axios 硬编码 `http://localhost:8099` | 生产模式同源相对路径，配合 embed 单体部署 |
| 无迁移机制 | `schema_migrations` 版本化迁移，后续加功能直接追加 SQL 文件 |

## 后续加功能

在 `server/migration/migrations/` 下新增 `000X_xxx.sql`（重启自动执行），模型对应调整 `model/` 即可。
