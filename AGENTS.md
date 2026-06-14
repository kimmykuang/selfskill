# AGENTS.md

这份文档面向在本仓库迭代代码的 AI agent / 开发者，定义工作纪律和不可越线项。

---

## 1. 测试纪律

- **每个模块必须有 test 覆盖。** 新建 package 不允许只有 `package x` 加几个函数、没有 `_test.go`。
- **改动既有逻辑后必须跑单测回归。** 即使你认为"显然不影响其他地方"，也跑 `go test ./... -race`。回归只看本包不算数 — 跨包契约（store / service / web handlers / cli）很容易被无意识地破坏。
- **新增逻辑必须配 test。** 新 endpoint、新 service 方法、新 store 方法都要附 `_test.go`，覆盖至少 happy path + 一个错误分支。
- **TDD 优先。** 写新功能时先写 failing test，跑一遍确认它真的失败，再写实现。绕过这步的代码会偷偷少覆盖。

## 2. Commit 前 checklist

每次 `git commit` 之前，必须确认：

- [ ] `go build ./...` 通过
- [ ] `go test ./... -race` 全部通过（或：仅相关包，但要在 PR 合并前跑过全量）
- [ ] `go vet ./...` 干净
- [ ] 改动是单一意图的 commit（不要把无关重构夹带）
- [ ] 没有遗留 `TODO` / `FIXME` 占位符
- [ ] 没有调试用的 `fmt.Println` / `console.log`

## 3. 安全：禁止提交敏感数据

**绝对禁止**把以下内容 commit 进 git，无论是否是 private repo：

- 密码、API key、access token、secret
- `.env` 文件含真实凭证
- 私钥（`.pem`、`.key`、SSH private key）
- session cookie / OAuth token
- 数据库连接串含密码

实践：

- 在写 prompt / skill / 文档时如果需要演示，**用占位符**（如 `<YOUR_API_KEY>`、`xxx`、`example.com`）
- `~/ss/` 目录会被用户 push 到 GitHub，prompt frontmatter 里的真凭证就是泄漏 — 在写 prompt 时尤其注意
- commit 前用 `git diff --staged` 自查；可疑字符串先问

## 4. Web UI 改动的验证

仓库内置 Web UI（`/internal/web/`、`/web/static/`），改动 UI 时：

- **使用本机已安装的 Playwright + Chromium**
- 配置：**直接用本机 Chrome**（`channel: "chrome"`），**跳过 Chromium 下载**
  - 例：`browser = await playwright.chromium.launch({channel: "chrome", headless: true})`
- 验证至少包括：页面能渲染、主要交互不报 console 错、改动的功能点能跑通
- 单纯 curl 检查 endpoint 返回 200 不算 UI 验证 — endpoint 可以 200 但前端 JS 报错

## 5. 主干 context 是稀缺资源

主干（top-level Claude session）的 context window 一旦被噪音填满就会下降智力。**以下三类工作默认派给 sub agent**，主干只接收结论：

### 5a. 跑全套 test suite（尤其失败时输出大）

- 主干**只接收**："X tests passed / Y failed (test names: ...)"
- agent 内部跑 `go test ./... -race -v`，分析失败，**只回报结论 + failing test 名 + 简要原因**
- 主干不要直接执行可能失败的 `go test`，也不要读完整的测试 log

### 5b. 安装 / 下载大体积工具

- **主干永远不要直接 `install`**：playwright + chromium、Xcode CLT、模型权重、`npm install`、`pip install` 等
- 下载日志几百到几千行，污染上下文严重
- 派 agent 完成安装，agent 只回："installed at <path>，version <x>，可用"

### 5c. 持久化的环境探测

- 端口 / 进程 / 已安装 package / 系统状态扫描 — 这些命令输出大且对当前 task 多数无关
- 派 agent 探完，agent 一句话总结："port 18484 空闲" / "Chrome 已装在 /Applications/Google Chrome.app"

### 派发模板

```
派给 general-purpose agent：
- 任务：<具体目标，比如 "跑全部测试，告诉我哪些失败">
- 不要返回完整 log
- 只回 ≤200 字结论 + 关键 ID（test 名 / 错误码 / 路径）
```

主干读到的只能是浓缩信息。如果你（agent）作为主干跑了重命令产生大量输出，下次先派 sub agent。

---

## 项目结构速记

```
cmd/ss/                  # CLI 入口
internal/
├── cli/                 # cobra 命令 → service
├── web/                 # http handlers → service
├── service/             # 共享 use-case 层（CLI + Web 都调）
├── skill/ prompt/ group/      # 文件树 CRUD
├── plugin/              # plugin store + registry + loader
│   └── cc_compat/       # 唯一接触 ~/.claude/plugins/installed_plugins.json 的地方
├── linker/  installer/  compare/   # 既有能力
└── frontmatter/         # YAML frontmatter parser
web/static/              # Alpine.js + Tailwind CDN 前端，零构建
docs/superpowers/        # spec & implementation plan（gitignored）
```

## 不变量（任何改动都不能破坏）

- **`installed_plugins.json` 只能由 `internal/plugin/cc_compat/` 读写**。其他包用 `cc_compat.Registry` 接口。
- **`~/ss/` 是文件树真值**。不引索引、不引数据库、不上 daemon。
- **on-disk 格式与 `~/.claude/` 严格对齐**。不发明 ss-only 的 frontmatter key（除了已记录的 `source`）。
- **不引入构建步骤**。前端保持 CDN + 静态文件 + `//go:embed`。
