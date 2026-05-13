# DeepSight

个人能力深度分析助手。通过聊天对话自动分析用户的技术能力和项目经历，生成结构化的能力图谱和发展建议。

## 背景

个人知识和技能随着时间在学习与项目中不断累积，但却无法有效地沉淀为可视化的能力画像。在和人交流、写简历、回顾成长时，这些内容都是隐性的，无法直观审视。

DeepSight 的目标是解决这一问题：用户可以上传项目资料创建个人知识库，与 AI 自由交流项目经历，系统会从对话中自动沉淀项目中的知识、经验和能力。基于深度研究 Agent，系统能对聊天内容和项目文档进行多维度分析，生成包含能力评估、项目复盘、行业对标和发展建议的完整分析报告。

## 核心功能

- **RAG 知识库问答**：上传文档构建个人知识库，基于 pgvector 语义检索和 SSE 流式推送实现实时对话
- **深度能力分析**：三阶段 Agent（Plan → Execute → Report）对聊天内容和文档进行多维度分析，自动生成能力画像
- **联网行业对标**：集成 Tavily WebSearch，对比用户能力与行业趋势，给出有数据支撑的发展建议
- **可追溯报告**：每条分析结论附带证据来源（文档片段或聊天记录），杜绝 AI 幻觉

## 技术架构

| 层级 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| 数据库 | PostgreSQL + pgvector (向量检索) |
| 缓存 | Redis |
| 消息队列 | RabbitMQ (异步文件处理) |
| 对象存储 | MinIO / S3 |
| LLM | OpenAI 兼容协议 (阿里云 DashScope) |
| 前端 | React 18 + TypeScript + Zustand + Vite |
| 部署 | Kubernetes + Docker |

## 项目结构

```text
.
├── cmd/deepsight/              # 应用入口
├── configs/                    # 配置文件
├── internal/
│   ├── agent/                  # Agent 层
│   │   ├── research/           # 深度研究 Agent (Planner/Executor/Reporter)
│   │   └── tools/              # 工具层 (WebSearch/ExperienceExtract)
│   ├── app/                    # 应用启动 & 依赖注入
│   ├── config/                 # 配置加载
│   ├── database/               # 基础设施客户端 (PG/Redis/MQ/S3)
│   ├── dto/                    # 请求/响应结构体
│   ├── handler/                # HTTP 处理器
│   ├── middleware/              # JWT & CORS 中间件
│   ├── model/                  # GORM 数据模型
│   ├── repository/             # 数据访问层
│   ├── service/                # 业务逻辑层
│   ├── util/                   # 工具 (JWT/Chunker/Parser)
│   └── worker/                 # 异步文件处理 Worker
├── web/                        # 前端 React 应用
├── .kube/                      # Kubernetes 部署配置
└── scripts/                    # 构建脚本
```

## 快速开始

### 前置条件

- Go 1.23+
- PostgreSQL 16+ (with pgvector extension)
- Redis
- RabbitMQ
- MinIO (or other S3-compatible storage)
- Node.js 18+

### 配置

编辑 `configs/application.yml`，填入数据库、Redis、RabbitMQ、MinIO 连接信息以及 LLM API Key。

### 启动后端

```bash
go run ./cmd/deepsight
```

### 启动前端

```bash
cd web
npm install
npm run dev
```

### 构建

```bash
# 后端
go build -o deepsight ./cmd/deepsight

# 前端
cd web && npm run build
```

## License

MIT
