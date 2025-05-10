# chasi-sreagent

**AI SRE Agent for Multi-tenant Kubernetes Environments**

Chinese(中文) README [readme-cn](README-CN.md).
---

## Introduction

`chasi-sreagent` is an open-source project aimed at building an intelligent Site Reliability Engineering (SRE) agent powered by Artificial Intelligence, specifically Large Language Models (LLMs). It is designed to address the complexities of operating applications deployed in multi-tenant Kubernetes environments, particularly those utilizing technologies like vCluster for logical isolation.

Leveraging and extending the capabilities of projects like k8sgpt, `chasi-sreagent` automates the discovery, analysis, diagnosis, and suggestion of remediation steps for both platform-level and business-level issues. A key focus is on integrating operational data from business applications through a standardized adaptation framework and supporting the use of locally deployed LLMs alongside online services.

## Features

* **Multi-tenant Awareness:** Deep understanding and support for vCluster-based architectures, capable of analyzing resources and issues across host and virtual clusters.
* **Automated Analysis & Diagnosis:** Automates the process of identifying potential issues by analyzing Kubernetes resources, logs, metrics, and business data.
* **Business Integration Framework:** Provides a low-intrusion SDK/framework for business systems to expose relevant operational data (status, logs, metrics, runbooks) to the SRE agent.
* **Intelligent Decision & Suggestion:** Utilizes LLMs, combined with multi-source data and a Retrieval Augmented Generation (RAG) knowledge base, to provide accurate diagnoses and actionable remediation suggestions.
* **Flexible LLM Support:** Supports various LLM providers, with a strong emphasis on facilitating the use of locally deployed models (e.g., via LocalAI) for data privacy and cost efficiency, alongside online services (e.g., DeepSeek).
* **Extensible Architecture:** Designed with a layered structure and clear interfaces, allowing for easy extension with custom analyzers, data sources, LLM providers, and action executors.
* **Fusable/Separable Deployment:** Can be deployed centrally in the host cluster or with components distributed within virtual clusters to suit different operational needs.

## Underlying Technologies & Inspirations

* **Kubernetes & Containers:** The core environment the agent operates within.
* **vCluster:** Provides the multi-tenant isolation context. [https://github.com/loft-sh/vcluster](https://github.com/loft-sh/vcluster)
* **k8sgpt:** Used as a base for Kubernetes resource analysis capabilities. [https://github.com/k8sgpt-ai/k8sgpt](https://github.com/k8sgpt-ai/k8sgpt)
* **LocalAI:** Key for enabling local LLM inference. [https://github.com/mudler/LocalAI](https://github.com/mudler/LocalAI)
* **Kubernetes Operator Pattern:** For managing the agent's lifecycle and configurations within Kubernetes (inspired by k8sgpt-operator). [https://github.com/k8sgpt-ai/k8sgpt-operator](https://github.com/k8sgpt-ai/k8sgpt-operator)
* **LangChain Concepts:** Ideas around building LLM agents, RAG, and tool integration (inspired by articles like [9] in architecture doc).

## Architecture

Detailed architecture can be found in [docs/architecture.md](docs/architecture.md).

## Getting Started

**(Coming Soon - Instructions on building, deploying, and configuring the agent)**

## Contributing

We welcome contributions! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file (Coming Soon) for details on how to contribute.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file (Coming Soon) for details.

---

**Project Link:** [https://github.com/turtacn/chasi-sreagent](https://github.com/turtacn/chasi-sreagent)
