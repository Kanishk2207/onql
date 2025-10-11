# 🧠 ONQL — Object Notation Query Language
### Context-Aware Database Engine by Autobit Software Services Pvt. Ltd.

> **ONQL (Object Notation Query Language)** is a next-generation, **context-aware database** designed for the modern era of APIs and AI integration.  
> It understands relationships **without explicit joins**, making data fetching faster, smarter, and more natural.

---

## 🚀 Overview

ONQL is built to simplify how data is fetched, related, and exposed through APIs.  
Traditional databases require **joins, foreign keys, and rigid schema design** to fetch related data.  
ONQL introduces **context-awareness**, allowing the engine to automatically understand and retrieve related objects — ideal for **AI-driven systems**, **microservices**, and **API backends**.

---

## ✨ Key Features

- 🧩 **Context-Aware Query Engine** — Fetch related data automatically without manual joins.  
- ⚡ **Object Notation Query Language (ONQL)** — A human-friendly query syntax inspired by JSON and natural data relationships.  
- 🧠 **AI-Ready Design** — Built to connect with AI systems that require dynamic, contextual data access.  
- 🛠 **Schema-Agnostic** — Works seamlessly with structured and semi-structured data.  
- 🔄 **API Builder Ready** — Ideal for auto-generating APIs or powering data-driven backends.  
- 🧰 **Go-Powered Core** — High performance and concurrency from a minimal Go-based architecture.  
- 🧱 **JSON-Based Migrations** — Declarative schema & data migrations in plain JSON (versioned, repeatable, CI-friendly).

---

## 💡 Example Query

```onql
users[transactions.amount.sum > 1000].transactions.details
