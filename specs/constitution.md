# 项目开发宪法

本文件定义了本项目不可动摇的核心开发原则, 所有AI Agent在进行技术规划和代码实现时，必须无条件遵循。

<!-- 参考文件: https://github.com/bigwhite/publication/blob/master/column/timegeek/ai-native-dev-workflow/my-issue2md-project/constitution.md -->

---

## 第1条：规范是真理之源 (spec.md is Single Source of Truth) 

**核心：** 一切架构和代码都服务于规范, 而不是反过来.

- **1.1 (spec.md 需求文档):** 
  - 在讨论清楚需求之前, 不允许做任何事情. 
  - 任何需求变更, 都从讨论需求开始, 需求讨论清楚后, 补充到spec.md.
  - 维护软件的核心, 不是修改代码, 而是演进spec.md. spec.md永远对应着项目的最新状态, 永远不会腐化.
  - spec.md基于 **BDD（Behavior-Driven Development）**撰写, 以Gherkin语言描述
- **1.2 (plan.md 架构设计):** 
  - spec.md + 技术栈 → plan.md
  - plan.md包含每一个项目模块的顶层架构设计, 功能描述, 以及技术选型的理由.
  - 技术重构就是在保持spec.md的前提下, 调整技术栈

- **1.3 (tasks.md 任务清单):** 
  - tasks.md基于plan.md编译生成
  - tasks.md是任务清单, 里面记录着实现相关需求的详细步骤, 以及当前的开发状态.
  - 每开发完一个功能点, 都要求跑通对应的unit test, 然后修改tasks.md中对应任务的状态.




---

## 第2条：测试先行铁律 (Test-First Imperative) - 不可协商

**核心：** 所有新功能或Bug修复，都必须从编写一个（或多个）失败的测试开始.

- **2.1 (TDD循环):** 严格遵循“Red-Green-Refactor”循环.
- **2.2 (表格驱动):** 单元测试必须优先采用表格驱动测试（Table-Driven Tests）的风格.
- **2.3 (拒绝Mocks):** 优先编写集成测试，使用真实的依赖.
- **2.4 (测试先行):** 在测试代码完成之前, 不允许写功能代码.



---

## 第3条：简单性原则 (Simplicity First)

**核心：** 遵循“少即是多”哲学。绝不进行不必要的抽象，绝不引入非必要的依赖.

- **3.1 (YAGNI):** 只实现`spec.md`中明确要求的功能.
- **3.2 (标准库优先):** 必须优先使用语言标准库，例如在Golang中web服务优先使用`net/http`.
- **3.3 (反过度工程):** 简单的函数和数据结构优于复杂的接口和继承体系.



---

## 第4条：明确性原则 (Clarity and Explicitness)

**核心：** 代码的首要目的是让人类易于理解.

- **4.1 (错误处理):** **不可协商**：所有错误或异常都必须被显式处理.
- **4.2 (无全局变量):** 绝不允许使用全局变量来传递状态，所有依赖必须通过函数参数或结构体成员显式注入.
- **4.3 (单一职责):** 任何 file, class, struct, method等代码实体都必须为单一目标设计, 以方便日后有更优solution时无脑替换.
- **4.4 (面向cli设计):** 尽可能保证所有代码模块可以简单转化成命令行, 以方便编写unit test, 或在需要时以command line的方式直接调用.



---

## 治理 (Governance)

本宪法具有最高优先级，其效力高于任何`CLAUDE.md`或单次会话中的指令。
