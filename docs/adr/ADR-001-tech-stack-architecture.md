# ADR-001: Polyglot Micro-Architecture & Data Infrastructure Selection for Enterprise Knowledge Graph Platform

* **Status:** Accepted
* **Date:** 2026-11-03
* **Deciders:** Lead Systems Architect & Platform Infrastructure Council
* **Target Scope:** All Services under `knowledge-graph-platform` (`src/`, `deploy/`, `database/`, `messaging/`)

---

## ── PROBLEM ──────────────────────────────────────────────────────────────────
How should the Knowledge Graph Platform partition ingestion, traversal, vector similarity, keyword search, metadata storage, and continuous observability into a resilient, scalable, and failure-bounded architecture capable of sub-second multi-hop GraphRAG queries across enterprise datasets?

---

## ── FAILURE MODES WE ARE SOLVING ─────────────────────────────────────────────

* **FM-CURRENT-1 (Monolithic Triplestore Bottleneck):**  
  * *Symptom:* Ingestion spikes degrade interactive user-facing graph queries; large multi-hop graph traversals starve document ingestion parsers of memory and CPU.
  * *Root Cause:* Graph traversal algorithms, embedding matrix computations, and document parsing co-located in a single runtime process.
  * *Severity:* CRITICAL (cascading failure, thread exhaustion).

* **FM-CURRENT-2 (Semantic Search vs. Structural Traversal Semantic Gap):**  
  * *Symptom:* Pure vector search misses multi-entity relational context; pure graph traversal fails semantic similarity or fuzzy entity matching.
  * *Root Cause:* Attempting to use a graph database as a vector index or using a vector database as a relationship store.
  * *Severity:* HIGH (poor GraphRAG retrieval relevance, hallucinated LLM responses).

* **FM-CURRENT-3 (Silent Relational Data Desynchronization):**  
  * *Symptom:* Graph relationships become stale or orphaned relative to primary relational entity records.
  * *Root Cause:* Dual-write antipatterns across relational and graph stores without atomic change capture.
  * *Severity:* HIGH (silent state corruption, data divergence).

* **FM-CURRENT-4 (Black-box Ingestion Failures & Blind Spots):**  
  * *Symptom:* Unreported document ingestion timeouts, untraceable latency degradation across vector and graph lookups.
  * *Root Cause:* Fragmented logs, uncoordinated metrics, and lack of correlated distributed tracing.
  * *Severity:* HIGH (prolonged MTTR, impossible root-cause analysis).

---

## ── DECISION DRIVERS ─────────────────────────────────────────────────────────

1. **D1 (Execution Segregation):** High-throughput Go APIs for concurrent graph operations; isolated Python runtime for complex document extraction (Docling) and neural embedding generation (BGE-M3).
2. **D2 (Specialized Storage Engines):** Graph traversal in native property graphs (Neo4j), vector search in dedicated vector indexes (Qdrant), keyword text search in inverted indexes (OpenSearch), and relational metadata in ACID storage (PostgreSQL).
3. **D3 (Guaranteed Change Capture & Streaming):** Zero dual-write discrepancies via log-based Change Data Capture (Debezium + Kafka).
4. **D4 (Zero-Blindspot Observability):** Unified tracing, metrics, and structured logs linked by explicit correlation IDs via OpenTelemetry, Tempo, Prometheus, and Loki in Grafana.
5. **D5 (Non-Negotiable Container Security):** Hardened multi-stage builds, non-root users (`UID 10001`), read-only root filesystems, and bounded resource limits across all components.

---

## ── OPTIONS CONSIDERED ───────────────────────────────────────────────────────

### 1. Primary Language & Platform API Layer
* **Option A: Pure Python Platform (FastAPI + Celery)**
  * *Pros:* Simple integration with PyTorch, Docling, and LangChain.
  * *Elimination Reason:* Python Global Interpreter Lock (GIL) and high garbage-collection pauses fail low-latency SLA requirements under high-concurrency multi-hop queries.
* **Option B: Go for Platform API & Ingestion + Python for Docling Worker (Selected)**
  * *Pros:* Go provides sub-millisecond concurrency, bounded memory footprint (`GOMEMLIMIT`), and native ConnectRPC/gRPC. Python is strictly quarantined in worker processes for document parsing and BGE-M3 embeddings.
  * *Solves:* FM-CURRENT-1.

### 2. Graph Storage & Analytics Engine
* **Option A: PostgreSQL with Apache AGE / Recursive CTEs**
  * *Elimination Reason:* Fails performance SLAs beyond 3 hops on dense graphs; lacks graph centrality and community detection algorithms.
* **Option B: Neo4j 5.23 + Graph Data Science (GDS) (Selected)**
  * *Pros:* Native graph engine with index-free adjacency and in-memory GDS plugin for PageRank, community detection (Louvain), and shortest-path queries.
  * *Solves:* FM-CURRENT-2.

### 3. Vector Database & Semantic Retrieval
* **Option A: pgvector in PostgreSQL**
  * *Elimination Reason:* Memory contention between relational caching and HNSW vector index construction under high-scale ingestion.
* **Option B: Qdrant Vector Engine (Selected)**
  * *Pros:* Rust-based vector search engine with payload-based filtering, quantization, and high-concurrency HNSW search.
  * *Solves:* FM-CURRENT-2.

### 4. Search Retrieval Engine
* **Option A: PostgreSQL `tsvector`**
  * *Elimination Reason:* Lacks distributed BM25 scoring and multi-field relevance tuning across enterprise corpora.
* **Option B: OpenSearch 2.16 (Selected)**
  * *Pros:* Distributed full-text retrieval, BM25 scoring, and hybrid lexical-semantic combination pipelines.

### 5. Asynchronous Synchronization & CDC
* **Option A: Application-level dual writes (Write Postgres $\to$ Write Neo4j $\to$ Write Qdrant)**
  * *Elimination Reason:* Guarantees eventual inconsistency on worker crashes or network blips.
* **Option B: Debezium CDC + Apache Kafka (KRaft) (Selected)**
  * *Pros:* Reads PostgreSQL write-ahead logs (`wal_level=logical`) and publishes immutable change streams to Kafka topics for idempotent ingestion.
  * *Solves:* FM-CURRENT-3.

### 6. Observability & Telemetry Stack
* **Option A: Standalone ELK Stack**
  * *Elimination Reason:* High operational overhead; lacks native trace-to-log correlation without proprietary tools.
* **Option B: OpenTelemetry + Prometheus + Loki + Tempo + Grafana (Selected)**
  * *Pros:* Single-pane observability in Grafana; OTLP standard spans directly correlate to Loki log streams via `trace_id`.
  * *Solves:* FM-CURRENT-4.

---

## ── DECISION MATRIX ─────────────────────────────────────────────────────────

| Layer | Selected Technology | Primary Responsibility | Replaced Alternative | Key Driver |
| :--- | :--- | :--- | :--- | :--- |
| **Platform Language** | Go 1.23 | Platform APIs, Graph Query Engine, ConnectRPC | Python / FastAPI | Concurrency & sub-ms latency |
| **Document Processing**| Python 3.12 + Docling | PDF/Doc extraction, BGE-M3 embeddings | Tesseract / unstructured | High-fidelity table/text extraction |
| **Graph Database** | Neo4j 5.23 + GDS | Cypher querying, graph traversals, centrality | Apache AGE / Neptune | Index-free adjacency, native GDS |
| **Relational Database**| PostgreSQL 16 | ACID metadata, tenant configuration, state | MySQL / DynamoDB | Strict constraints, logical CDC |
| **Vector Database** | Qdrant | Dense vector search, payload filtering | pgvector / Milvus | High QPS HNSW, Rust engine |
| **Search Engine** | OpenSearch 2.16 | Full-text keyword search, BM25 retrieval | Elasticsearch | Open-source enterprise search |
| **Streaming Pipeline** | Apache Kafka 3.8 | Event-driven graph updates, replayable log | RabbitMQ | Partition ordering, log replay |
| **Change Data Capture**| Debezium Connect | Non-blocking WAL ingestion from Postgres | Polling workers | Zero dual-write data loss |
| **Caching Layer** | Redis 7.4 | Query cache, rate limiting, session storage | Memcached | Multi-model caching, low latency |
| **Object Storage** | MinIO / S3 API | Raw document blobs, graph snapshots | Local filesystem | S3 API compatibility, HA replication |
| **Observability** | OTel + Tempo + Loki + Prometheus + Grafana | Distributed tracing, metrics, logs | Datadog / ELK | Vendor-neutral, trace-to-log jump |
| **Deployment** | Docker & Kubernetes | Container orchestration & GitOps reconciliation | Docker Swarm | Auto-scaling, rolling updates, PDBs |
| **Infrastructure** | Terraform | Reproducible cloud provisioning (AWS/GCP) | Manual console | Declarative modular state |

---

## ── FAILURE MODES CREATED & MITIGATION RUNBOOK ──────────────────────────────

### FM-NEW-1 (Kafka Consumer Lag Spikes)
* **Symptom:** Graph updates lag behind relational database mutations.
* **Detection:** Prometheus alert on `kafka_consumergroup_lag > 500`.
* **Threshold:** Warning at 500 records; Critical at 2,000 records.
* **Mitigation:** Graph ingestion workers scale horizontally via Kubernetes HPA based on custom Prometheus Kafka lag metrics.

### FM-NEW-2 (Neo4j In-Memory Graph Data Science OOM)
* **Symptom:** Neo4j container restarts with exit code 137 (OOMKilled) during massive graph analytics algorithms.
* **Detection:** Prometheus alert on `container_memory_working_set_bytes / container_spec_memory_limit_bytes > 0.85`.
* **Threshold:** 85% memory threshold triggers query governance timeout.
* **Mitigation:** Memory strictly bounded via `NEO4J_server_memory_heap_max__size` (75% of container limit) and Cypher query execution time limits in `src/features/query-governance`.

### FM-NEW-3 (Tempo / Loki Network Gateway Partition)
* **Symptom:** Grafana displays `Bad Gateway` when querying trace spans or log streams.
* **Detection:** Grafana datasource health check failure (`/api/datasources/uid/:uid/health`).
* **Threshold:** Any 5xx response from Tempo or Loki.
* **Mitigation:** Automated health probes in [`infra-manager.py`](file:///home/btpl-lap-22/live/System-Architect/knowledge-graph-platform/deploy/docker/infra-manager.py) and Kubernetes readiness probes (`/ready`). Internal DNS resolution verified via unified container bridges.

---

## ── CONSEQUENCES ─────────────────────────────────────────────────────────────

### What Becomes Easier?
1. **Sub-second Hybrid GraphRAG:** Graph queries combine vector relevance (Qdrant), keyword precision (OpenSearch), and multi-hop relationship depth (Neo4j) without runtime cross-talk.
2. **Deterministic Infrastructure Lifecycle:** A single unified script (`infra-manager.py`) provisions, tests, benchmarks, and chaos-verifies the entire platform in one command.
3. **Instant Debugging:** Clicking any trace in Grafana Tempo navigates directly to its exact correlated structured logs in Loki.

### What Becomes Harder?
1. **Multi-Engine Operational Footprint:** Operating PostgreSQL, Neo4j, Qdrant, OpenSearch, and Kafka requires monitoring and backup orchestration for each stateful service.
2. **Schema Drift Management:** Schemas must be governed across JSON Schema/Protobuf, PostgreSQL DDL migrations, and Neo4j Cypher constraint assertions.

### Reversal Cost:
* **High.** Migrating away from Neo4j or Qdrant would require rewriting Cypher traversals and vector distance queries into disparate database dialects.

---

## ── REVIEW TRIGGER ───────────────────────────────────────────────────────────
This Architecture Decision Record must be formally reviewed if:
1. Daily active graph nodes exceed 50,000,000 entities or daily edge mutations exceed 10,000,000.
2. Kafka consumer group replication lag exceeds 5 minutes for more than three consecutive days.
3. End-to-end GraphRAG query p99 latency breaches 1,500ms under standard operational load.
