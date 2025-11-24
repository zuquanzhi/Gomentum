# 🚀 Gomentum: Project Requirements Document

> **Project Name**: Gomentum (Go + Momentum)
> **Slogan**: Build momentum with a Go-powered, MCP-native planning agent.
> **Version**: v0.1.0 (MVP)
> **Author**: zuquanzhi
> **Status**: Ready for Development

-----

## 1\. 项目概述 (Executive Summary)

**Gomentum** 是一个基于 **Go 语言** 构建的终端（CLI）智能规划助手。它利用 **Google adk-go** 接入大模型（Gemini）作为推理大脑，并采用行业前沿的 **MCP (Model Context Protocol)** 协议来管理工具集。

### 1.1 核心价值

  * **消除决策瘫痪**：将模糊的自然语言目标（如“这周学完 Go 并发”）自动拆解为可执行的时间表。
  * **后端技术练兵**：在一个项目中完整实践 Go 的核心特性（Goroutines、Channels、Interfaces、File I/O）。
  * **拥抱前沿标准**：通过实现 MCP 协议，使项目具备极高的可拓展性和“高星项目”气质。

-----

## 2\. 功能需求 (Functional Requirements)

### 2.1 交互界面 (User Interface)

  * **纯终端交互 (CLI)**：无需前端页面。
  * **REPL 模式**：启动后显示 `Gomentum >` 提示符，持续接收用户输入，直到用户输入 `exit`。
  * **流式反馈**：在 AI 思考或执行工具时，终端应有简单的状态提示（如 "Thinking..." 或 "✅ Scheduled"）。

### 2.2 核心功能模块

| 功能 ID | 功能名称 | 详细描述 | 涉及技术点 |
| :--- | :--- | :--- | :--- |
| **F-01** | **智能任务拆解** | 用户输入宏观目标，Agent 自动推理并生成多个具体的时间段任务。 | ADK (Reasoning), Prompt Engineering |
| **F-02** | **日程存储 (内存)** | 将生成的任务暂存在内存中，支持增删改查。 | Go Slices, Structs, Pointers |
| **F-03** | **异步提醒** | 用户要求提醒时，后台启动倒计时，时间到后在终端高亮提示。 | **Goroutines**, `time.Sleep` |
| **F-04** | **文档导出** | 将当前的规划结果导出为 Markdown 文件。 | `os` Package, File I/O |
| **F-05** | **MCP 工具化** | 所有上述功能（F-02\~F-04）必须封装为 MCP Tool 标准。 | `mark3labs/mcp-go` |

-----

-----

这份文档现在可以作为你的**开发圣经**。祝你在构建 **Gomentum** 的过程中获得乐趣，并在简历上留下浓墨重彩的一笔！
