# chasi-sreagent

**AI SRE Agent for Multi-tenant Kubernetes Environments**
English(英文) README [readme-en](README.md).

**AI SRE 智能代理，专为多租户 Kubernetes 环境设计**

---

## Introduction / 项目简介

`chasi-sreagent` is an open-source project aimed at building an intelligent Site Reliability Engineering (SRE) agent powered by Artificial Intelligence, specifically Large Language Models (LLMs). It is designed to address the complexities of operating applications deployed in multi-tenant Kubernetes environments, particularly those utilizing technologies like vCluster for logical isolation.

Leveraging and extending the capabilities of projects like k8sgpt, `chasi-sreagent` automates the discovery, analysis, diagnosis, and suggestion of remediation steps for both platform-level and business-level issues. A key focus is on integrating operational data from business applications through a standardized adaptation framework and supporting the use of locally deployed LLMs alongside online services.

`chasi-sreagent` 是一个开源项目，旨在构建一个由人工智能（特别是大型语言模型 LLM）驱动的智能站点可靠性工程 (SRE) 代理。它被设计用来解决在多租户 Kubernetes 环境中部署应用所带来的复杂性，尤其是在利用 vCluster 等技术进行逻辑隔离的场景下。

`chasi-sreagent` 利用并扩展了 k8sgpt 等项目的能力，自动化地发现、分析、诊断并提供针对平台层面和业务层面问题的处置建议。项目的关键在于通过一个标准化的适配框架集成来自业务系统的运营数据，并支持使用本地部署的 LLM 以及在线服务。

## Features / 特性

* **Multi-tenant Awareness:** Deep understanding and support for vCluster-based architectures, capable of analyzing resources and issues across host and virtual clusters.
    **多租户感知:** 深度理解并支持基于 vCluster 的架构，能够分析宿主机集群和虚拟集群中的资源和问题。
* **Automated Analysis & Diagnosis:** Automates the process of identifying potential issues by analyzing Kubernetes resources, logs, metrics, and business data.
    **自动化分析与诊断:** 通过分析 Kubernetes 资源、日志、指标和业务数据，自动化识别潜在问题。
* **Business Integration Framework:** Provides a low-intrusion SDK/framework for business systems to expose relevant operational data (status, logs, metrics, runbooks) to the SRE agent.
    **业务系统集成框架:** 提供一套低侵入性的 SDK/框架，使业务系统能够向 SRE 代理暴露相关的运营数据（状态、日志、指标、Runbook）。
* **Intelligent Decision & Suggestion:** Utilizes LLMs, combined with multi-source data and a Retrieval Augmented Generation (RAG) knowledge base, to provide accurate diagnoses and actionable remediation suggestions.
    **智能决策与建议:** 利用 LLM 能力，结合多源数据和 RAG (检索增强生成) 知识库，提供准确的诊断结论和可行的处置建议。
* **Flexible LLM Support:** Supports various LLM providers, with a strong emphasis on facilitating the use of locally deployed models (e.g., via LocalAI) for data privacy and cost efficiency, alongside online services (e.g., DeepSeek).
    **灵活的 LLM 支持:** 支持多种 LLM 提供商，并特别强调支持本地部署的模型（例如通过 LocalAI），以满足数据隐私和成本效率的需求，同时也支持在线服务（例如 DeepSeek）。
* **Extensible Architecture:** Designed with a layered structure and clear interfaces, allowing for easy extension with custom analyzers, data sources, LLM providers, and action executors.
    **可扩展架构:** 采用分层设计和清晰的接口，方便扩展自定义分析器、数据源、LLM 提供商和动作执行器。
* **Fusable/Separable Deployment:** Can be deployed centrally in the host cluster or with components distributed within virtual clusters to suit different operational needs.
    **可融合/可分离部署:** 支持在宿主机集群集中部署，或将组件分布到虚拟集群中部署，以适应不同的运维需求。

## Underlying Technologies & Inspirations / 基础技术与灵感来源

* **Kubernetes & Containers:** The core environment the agent operates within.
    **Kubernetes & 容器:** 代理操作的核心环境。
* **vCluster:** Provides the multi-tenant isolation context. [https://github.com/loft-sh/vcluster](https://github.com/loft-sh/vcluster)
    **vCluster:** 提供多租户隔离的上下文。
* **k8sgpt:** Used as a base for Kubernetes resource analysis capabilities. [https://github.com/k8sgpt-ai/k8sgpt](https://github.com/k8sgpt-ai/k8sgpt)
    **k8sgpt:** 用作 Kubernetes 资源分析能力的基础。
* **LocalAI:** Key for enabling local LLM inference. [https://github.com/mudler/LocalAI](https://github.com/mudler/LocalAI)
    **LocalAI:** 支持本地 LLM 推理的关键。
* **Kubernetes Operator Pattern:** For managing the agent's lifecycle and configurations within Kubernetes (inspired by k8sgpt-operator). [https://github.com/k8sgpt-ai/k8sgpt-operator](https://github.com/k8sgpt-ai/k8sgpt-operator)
    **Kubernetes Operator 模式:** 用于在 Kubernetes 中管理代理的生命周期和配置。
* **LangChain Concepts:** Ideas around building LLM agents, RAG, and tool integration (inspired by articles like [9] in architecture doc).
    **LangChain 概念:** 关于构建 LLM 代理、RAG 和工具集成的思想。

## Architecture / 架构设计

Detailed architecture can be found in [docs/architecture.md](docs/architecture.md).
详细架构设计请参阅 [docs/architecture.md](docs/architecture.md)。

## Getting Started / 快速开始

**(Coming Soon - Instructions on building, deploying, and configuring the agent)**
（即将推出 - 关于如何构建、部署和配置代理的说明）

## Contributing / 贡献

We welcome contributions! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file (Coming Soon) for details on how to contribute.
欢迎贡献！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md) 文件（即将推出）了解如何贡献的详细信息。

## License / 许可证

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file (Coming Soon) for details.
本项目采用 Apache 2.0 许可证。详细信息请参阅 [LICENSE](LICENSE) 文件（即将推出）。

---

**Project Link:** [https://github.com/turtacn/chasi-sreagent](https://github.com/turtacn/chasi-sreagent)
