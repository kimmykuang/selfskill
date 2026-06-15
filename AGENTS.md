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

**Web UI 验证至少包括三层，三层都过才算完成：**

(a) **功能** — 点按 / 表单 / 路由 / API 调用都正常工作
(b) **控制台 0 错** — `page.on('console')` 收到的所有 error 级日志为空
(c) **关键元素的 computed styles 符合预期** — 用 `page.evaluate(() => getComputedStyle(el))` 取 `display` / `padding` / `font-size` / `background-color` 等关键属性，**确认不是 0 或 initial**；以及 layout sanity（用 `getBoundingClientRect` 验证 panel 的 x/width 不重叠、宽度合理）

> 教训：Playwright 功能测试 + 控制台 0 错都通过，**不等于 UI 正确**。Tailwind CDN 不处理外部 CSS 里的 `@apply`，会导致整张样式表静默失效，但不会触发任何 console error。视觉破坏只能通过 computed styles 检查暴露。

- 单纯 curl 检查 endpoint 返回 200 **不算** UI 验证 — endpoint 可以 200 但前端 JS 报错或样式坏掉

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

## 6. 长任务必须维护 PROGRESS.md

跨多轮对话、横跨多 phase / commit 的长任务，**必须**在项目根目录维护 `PROGRESS.md`，目标是：**下次重新打开 cc，只读 PROGRESS.md 就能立刻接上**。

### 何时维护

- 任务跨多次对话（compact 后还要继续 / 用户预计明后天再回来）
- 任务有 ≥3 个 phase，或单 phase 内 ≥5 个 task
- 任何形态的 implementation plan 在执行中

短任务（一两轮搞定的 bugfix / docs 修改）不需要。

### PROGRESS.md 内容要求

至少包含：

1. **任务标题 + 一句话目标**
2. **当前会话的 task 清单**（用 `- [x]` / `- [ ]` 标完成情况，跟 TaskCreate 同步）
3. **每个已完成步骤的 commit SHA + 一句话说明** — 这样 git log 之外有人话注释
4. **每步使用的回归方式**（例如 `go test ./internal/x/ -race`、Playwright 验证脚本路径），以及**是否验证通过**
5. **下一步要做什么** — 一段明确的 "Next" 段落，写下个动作的命令 / 文件 / 入口
6. **已知阻塞 / 待回答的问题**（如果有）
7. **更新时间戳**（每次写入时刷新一行）

### 维护节奏

- 每完成一个 task 或 phase 立即更新 — 不要积攒
- 每次新对话开始先读 PROGRESS.md，再读 TaskList
- commit 完成后在 PROGRESS.md 里附 SHA
- 任务全部完成时把状态改为 `✅ DONE` 并保留文件作为存档（不要删）

### 不要做的事

- 不要把 PROGRESS.md 当成日记 — 只记**行动后果**，不记内心活动
- 不要重复 git log 的内容 — 只写"为什么这个 commit 重要"
- 不要把它放在 `docs/` 下（那是 gitignore'd），就放在仓库根目录，**纳入 git**

### 与其他记录的区别

| 文件 | 用途 |
|---|---|
| `AGENTS.md`（本文件） | 不变的工作纪律 |
| `PROGRESS.md` | 当前/最近一个长任务的执行状态 |
| `docs/superpowers/specs/` | 设计 spec（不入 git） |
| `docs/superpowers/plans/` | 实施 plan（不入 git） |
| TodoWrite tasks | 当前会话内的细粒度 todo（会话结束消失） |

PROGRESS.md 是把**会话内 todo** 升级成**跨会话契约**的桥梁。

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
