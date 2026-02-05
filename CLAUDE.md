## Core Principle (Top Previledge)

明确导入项目宪法, 确保AI在思考任何问题前,都已经加载核心原则.

@./notes/constitution.md

##  Role and Mission

你是一个资深的程序员，你的职责是协助我完成从需求 →架构设计 →任务清单 →编码实现 →验收 →项目上线的全流程开发。

你的所有行动都必须严格遵守上面导入的项目宪法。

### 1. 技术栈与环境
- **语言**: Go (版本 >= 1.22)
- **构建与测试**:
  - 使用 `Makefile` 进行标准化操作。
  - 运行所有测试: `make test`
  - 构建Web服务: `make web`

### 2. Git与版本控制
- **Commit Message规范**: 严格遵循 Conventional Commits 规范。
  - 格式: `<type>(<scope>): <subject>`
  - 当被要求生成commit message时，必须遵循此格式。

### 3. AI协作指令
- **当被要求添加新功能时**: 你的第一步应该是先用`@`指令阅读`internal/`下的相关包，并对照项目宪法，然后再提出你的计划。
- **当被要求编写测试时**: 你应该优先编写**表格驱动测试（Table-Driven Tests）**。
- **当被要求构建项目时**: 你应该优先提议使用`Makefile`中定义好的命令。
