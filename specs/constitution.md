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
- **1.4 (YAGNI):** 只实现`spec.md`中明确要求的功能.




---

## 第2条：测试先行铁律 (Test-First Imperative) - 不可协商

**核心：** 所有新功能或Bug修复，都必须从编写一个（或多个）失败的测试开始.

- **2.1 (TDD循环):** 严格遵循“Red-Green-Refactor”循环.
- **2.2 (表格驱动):** 单元测试必须优先采用表格驱动测试（Table-Driven Tests）的风格.
- **2.3 (测试先行):** 在unit test编码完成之前, 不允许写任何功能代码.
- **2.4 (拒绝Mocks):** 优先编写integration test，使用真实的依赖.



---

## 第3条：架构原则 

**核心：** 遵循“少即是多”哲学, 绝不进行非必要的抽象，绝不引入非必要的依赖.

- **3.1 (Library-First Principle):** 
  - 必须优先使用语言标准库，例如Golang中web服务优先使用`net/http`.
  - 每一个功能都必须首先作为一个独立的库来实现, 绝不直接在应用代码中实现, 以保证代码高内聚, 低耦合.
- **3.2 (Single Responsibility):** 每个package, file, class, method都应该只做好一件事, 都必须为单一目标设计, 以方便日后有更优solution时无脑替换.
- **3.3 (CLI Interface Mandate):** 每一个库都必须通过命令行接口 (CLI) 暴露其核心功能, 以保证可测试性与可观测性.
- **3.4 (Simplicity and Anti-Abstraction):** 简单的函数和数据结构优于复杂的接口和继承体系.



---

## 第4条：编码原则

**核心：** 代码的首要目的是让人类易于理解.

- **4.1 (错误处理):** **不可协商**：所有错误或异常都必须被显式处理.
- **4.2 (无全局可变状态):**
  - **禁止**: 绝不允许使用全局变量来**传递运行时状态**（如计数器、缓存、连接池、业务对象等）
  - **允许**: 包级**只读常量**（`const` 或初始化后不可变的变量），包括：
    - 反射类型缓存（如 `reflect.TypeOf((*T)(nil)).Elem()`）
    - 正则表达式预编译（`regexp.MustCompile(...)`）
    - 时间格式、错误定义等常量
  - **要求**: 所有依赖必须通过函数参数或结构体成员显式注入（只读常量除外）



---

## 治理 (Governance)

本宪法具有最高优先级，其效力高于任何`CLAUDE.md`或单次会话中的指令。
