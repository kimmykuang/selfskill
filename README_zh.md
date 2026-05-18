# ss — Claude Code 的 Skills 与 Plugin 管理工具

`ss` 是一个本地 CLI 工具，用于管理 [Claude Code](https://docs.anthropic.com/en/docs/claude-code) 的 skills、prompts 和 plugins。它解决了 context 膨胀的问题——通过分组机制按需加载 skills，而非全量铺开。

## 核心功能

- **Skill 管理** — 从 URL 或本地目录安装 skills，分组管理，通过软链接按需加载/卸载
- **Plugin 管理** — 完全接管 CC 的 plugin 系统：添加 marketplace、安装/加载/卸载 plugins
- **Prompt 库** — 存储和搜索可复用的 prompts，支持 tags 和元数据
- **分组系统** — 将 skills 和 plugins 打包为 group，一条命令加载整组
- **Web UI** — 浏览器管理界面，支持 prompt 编辑和 plugin 操作
- **单二进制** — 无运行时依赖，静态资源内嵌

## 安装

```bash
# 从源码安装
go install github.com/kimmykuang/selfskill/cmd/ss@latest

# 或本地构建
git clone https://github.com/kimmykuang/selfskill.git
cd selfskill
make build
# 二进制在 ./bin/ss
```

## 快速开始

```bash
# 从本地目录安装 skill
ss install skill ./my-skills/video-summary/

# 创建分组并添加 skill
ss group create work
ss group add work video-summary

# 加载分组（在 ~/.claude/skills/ 创建软链接）
ss load work

# 查看当前激活状态
ss status

# 卸载
ss unload work
```

## Plugin 管理

```bash
# 添加 marketplace（clone 仓库获取 plugin 元数据）
ss plugin add-marketplace https://github.com/anthropics/claude-plugins-official.git

# 查看可安装的 plugins
ss plugin list-available

# 安装 plugin（下载到 ~/ss/plugins/，不加载）
ss plugin install superpowers@claude-plugins-official

# 加载到 CC（创建软链接 + 写入 installed_plugins.json）
ss plugin load superpowers

# 从 CC 卸载
ss plugin unload superpowers

# 将 plugin 加入分组
ss group add --plugin work superpowers
ss load work   # 同时加载 skills 和 plugins
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `ss install skill <source>` | 从 URL 或本地路径安装 skill |
| `ss install prompt <source>` | 从 URL 或本地路径安装 prompt |
| `ss skill list` | 列出已安装的 skills |
| `ss skill remove <name>` | 删除 skill |
| `ss prompt list` | 列出已安装的 prompts |
| `ss prompt remove <id>` | 删除 prompt |
| `ss group create <name>` | 创建分组 |
| `ss group add <group> <skill>` | 向分组添加 skill |
| `ss group add --plugin <group> <plugin>` | 向分组添加 plugin |
| `ss group remove <group> <item>` | 从分组移除 |
| `ss group list` | 列出所有分组 |
| `ss group delete <name>` | 删除分组 |
| `ss load <group>` | 加载分组（skills + plugins） |
| `ss unload <group>` | 卸载分组 |
| `ss unload --all` | 卸载所有 ss 管理的 skills |
| `ss status` | 显示当前激活的 skills 和 plugins |
| `ss plugin add-marketplace <url>` | 添加 plugin marketplace |
| `ss plugin list-available` | 列出可安装的 plugins |
| `ss plugin install <plugin>@<marketplace>` | 下载 plugin |
| `ss plugin load <plugin>` | 激活 plugin 到 CC |
| `ss plugin unload <plugin>` | 从 CC 停用 plugin |
| `ss plugin list` | 列出 ss 管理的 plugins |
| `ss plugin skills <plugin>` | 列出 plugin 中的 skills |
| `ss plugin update <plugin>` | git pull 更新到最新版本 |
| `ss plugin remove <plugin>` | 删除 plugin 源文件 |
| `ss web [--port 8484]` | 启动 Web 管理界面 |

## 目录结构

```
~/ss/
├── skills/           # 已安装的 skill 目录
│   └── video-summary/
│       ├── SKILL.md
│       ├── assets/
│       └── scripts/
├── prompts/          # Prompt markdown 文件
├── groups/           # 分组 YAML 配置
├── plugins/          # Plugin 源文件（git clone）
│   └── claude-plugins-official/
│       └── superpowers/
│           └── 5.1.0/
├── marketplaces/     # Marketplace 仓库（用于发现 plugins）
└── ss.yaml           # 全局配置
```

## 工作原理

**Skills** 通过软链接加载：
```
~/.claude/skills/video-summary -> ~/ss/skills/video-summary/
```

**Plugins** 通过软链接 + JSON 注册加载：
```
~/.claude/plugins/cache/mp/plugin/ver -> ~/ss/plugins/mp/plugin/ver/
+ 在 ~/.claude/plugins/installed_plugins.json 中添加条目
```

CC 启动时读取这些路径——skills 无需重启即生效，plugins 需要重启 CC。

## Web UI

```bash
ss web --port 8484
# 打开 http://localhost:8484
```

页面：Prompts（增删改查 + 搜索）、Skills（只读）、Plugins（加载/卸载）、Status（链接状态）。

## 开发

```bash
make build    # 构建二进制
make test     # 运行测试（含 race detector）
make clean    # 清理构建产物
```

## 技术栈

| 用途 | 选择 |
|------|------|
| 语言 | Go 1.25+ |
| CLI 框架 | cobra |
| 配置格式 | YAML (gopkg.in/yaml.v3) |
| HTTP 路由 | net/http (Go 1.22+ 方法路由) |
| 前端 | Alpine.js + Tailwind CSS CDN |
| 静态嵌入 | go:embed |

## License

MIT
