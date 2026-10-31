# TODO — Knowledge Graph Platform (Senior/Staff System Architect prep)

Topics: Neo4j, graph algorithms, GraphRAG, entity resolution, retrieval evaluation.
Primary company fit: **Neo4j** (this is your showcase repo), **Anthropic** (RAG, agents, evals, grounding/citations), **Google** (Knowledge Graph, Spanner Graph, search/ranking), Stripe (fraud/risk graphs, entity linking of merchants/accounts).

Current state: scaffold + 25 feature `context.yml`; only `src/features/graph-data-modeling/index.go` has code. `database/migrations/cypher` is empty.

Every WP has an **AC**. Not done until demonstrated by test/benchmark/artifact.

---

## 0. Definition of done for "interview-ready"
- [ ] Can explain Neo4j internals (storage, page cache, transactions, clustering) and justify graph vs relational for a given workload
- [ ] Working end-to-end demo: ingest documents -> extract entities/relations -> resolve -> store -> hybrid retrieve -> cited answer, with an evaluation report
- [ ] Can show a measured case where GraphRAG beats vector-only RAG AND a case where it doesn't
- [ ] Can design the system for 100M+ documents and discuss limits honestly

## 1. Architecture artifacts (`docs/`)
- [ ] `00-requirements.md`: use cases (enterprise Q&A, lineage, fraud rings, recommendations); NFRs with numbers (ingest 1k docs/s, p95 multi-hop (3 hops) < 300 ms, p95 answer < 4 s, freshness < 60 s from source change, retrieval recall@10 >= target)
- [ ] `01-c4.md`: Context/Container/Component; data-flow diagram ingest vs query paths
- [ ] `02-graph-model.md`: ontology (labels, relationship types + directions, properties, constraints), modeling rules (relationship types for verbs, nodes for things, intermediate nodes for n-ary/time/provenance), anti-patterns (supernodes, property-as-label, generic `RELATED_TO`), versioning/temporal model (valid-from/valid-to), query-driven modeling table (each query -> pattern -> index)
- [ ] `03-provenance-model.md`: `(:Chunk)-[:MENTIONS]->(:Entity)`, `(:Document)-[:HAS_CHUNK]->(:Chunk)`, edge properties: source, extractor_version, confidence, extracted_at
- [ ] `04-retrieval-design.md`: vector index + full-text + graph expansion; fusion (RRF vs weighted), re-ranking, context budget packing; local vs global (community-summary) search
- [ ] `05-capacity-model.md`: nodes/relationships x record size (Neo4j fixed-size records: ~15 B node, ~34 B relationship, plus properties) -> store size; page cache = hot working set; heap for transactions; GDS projection memory formula; backup size/time
- [ ] `06-threat-model.md`: Cypher injection, prompt injection via ingested documents (indirect injection!), tenant leakage through traversal, data poisoning, exfiltration through citations, PII
- [ ] `07-evaluation-plan.md`: datasets, metrics, thresholds, regression gates
- [ ] `08-runbooks/`: slow query, page-cache thrash, leader election, ingest backlog, bad extractor rollback, entity merge reversal
- [ ] ADRs:
  - [ ] ADR-001 Property graph (Neo4j) vs RDF/SPARQL vs relational vs Spanner Graph vs Neptune/TigerGraph
  - [ ] ADR-002 Neo4j topology: single / primary+secondaries cluster / Infinigraph property sharding; read scaling; write limits
  - [ ] ADR-003 Where vectors live: Neo4j vector index vs dedicated vector DB (consistency, ops, filtering)
  - [ ] ADR-004 Entity resolution pipeline & reversibility
  - [ ] ADR-005 Schema governance workflow
  - [ ] ADR-006 Extraction: LLM vs rule-based vs hybrid; cost/accuracy/latency
  - [ ] ADR-007 Tenancy: database-per-tenant vs label/property security vs separate clusters
  - [ ] ADR-008 Sync from systems of record: CDC/outbox vs batch ETL

## 2. Work packages

### WP-1 Graph modeling (`graph-data-modeling`)
- [ ] Extend `index.go`: model definitions, validation (allowed label/rel combos), constraint generator
- [ ] Model 3 domains to practice: org/people/projects; service dependency graph (use `internal-developer-platform` catalog); transactions/fraud ring
- [ ] **AC:** for each, a table of top-10 queries with PROFILE output showing index use and db hits; refactor model once based on PROFILE

### WP-2 Schema governance (`schema-governance`)
- [ ] Constraints (uniqueness, existence, node key), range/text/full-text/vector indexes as code in `database/migrations/cypher` with ordered, idempotent (`IF NOT EXISTS`) migrations and a migration ledger node
- [ ] Review workflow for ontology change; backward-compat rules; deprecation of labels
- [ ] **AC:** apply migrations on empty DB and on a populated DB; both converge

### WP-3 Ingestion (`dynamic-entity-ingestion`, `relationship-extraction`, `idempotent-graph-writes`)
- [ ] Pipeline: parse -> chunk (size/overlap experiments) -> embed -> extract entities/relations (structured output with JSON schema) -> validate -> write
- [ ] Idempotent writes: deterministic IDs (hash of canonical key), `MERGE` on constrained key, `ON CREATE/ON MATCH SET`, batched `UNWIND $rows` (1k-10k rows per tx), deadlock retry with backoff and sorted keys to avoid lock-order deadlocks
- [ ] Rate-limit/budget on LLM calls, caching by content hash, retries, dead-letter for failed docs
- [ ] **AC:** re-ingesting the same corpus yields identical graph (node/rel counts and hashes); 1M-chunk load benchmark recorded

### WP-4 Entity resolution (`entity-resolution`)
- [ ] Pipeline: normalize -> blocking (reduce O(n^2)) -> pairwise features (Jaro-Winkler, embedding cosine, shared neighbours) -> classifier/threshold -> cluster (connected components / Louvain) -> canonical entity with `(:Mention)-[:RESOLVED_TO]->(:Entity)`
- [ ] Merge is a graph relationship, not destructive — enables unmerge; human review queue for the grey zone
- [ ] **AC:** labeled set of >= 500 pairs; report precision/recall/F1 and chosen threshold; demonstrate unmerge restoring prior state

### WP-5 Sync + reconciliation (`event-driven-graph-updates`, `cdc-integration`, `graph-reconciliation`)
- [ ] Consume events from `event-driven-platform`; apply as idempotent upserts; deletes as soft-delete/tombstone then GC; out-of-order handling via version/`updated_at` guard (`WHERE n.version < $v`)
- [ ] Reconciler: diff source of truth vs graph by key ranges + checksums
- [ ] **AC:** inject missed events; reconciler detects and repairs; drift metric exported

### WP-6 Graph algorithms (`centrality-algorithms`, `community-detection`, `shortest-path-queries`, `similarity-algorithms`)
- [ ] Use GDS: PageRank, betweenness (approximate), Louvain/Leiden, label propagation, WCC, node similarity, kNN, Dijkstra/A*/Yen's k-shortest
- [ ] Understand: projection (native vs Cypher), memory estimate (`gds.*.estimate`), modes (stream/stats/mutate/write)
- [ ] **AC:** run each on a 1M-node graph; record runtime + memory; write a "when to use which" table with a real example each

### WP-7 Traversal + governance (`multi-hop-traversal`, `query-governance`)
- [ ] Bounded variable-length patterns, `shortestPath`, path filters pushed early, avoid Cartesian products, use `PROFILE` db-hits and eager operators
- [ ] Governance: transaction timeout, max rows, parameterized queries only, read-only role for retrieval path, query allow-list for LLM-generated Cypher with validation (EXPLAIN-first) and sandbox
- [ ] **AC:** malicious/expensive query corpus (injection, unbounded expansion, cartesian) is rejected or bounded

### WP-8 GraphRAG (`graphrag`, `context-expansion`, `hybrid-retrieval`)
- [ ] Baselines first: BM25-only, vector-only, then hybrid, then graph-expanded
- [ ] Seed selection (top-k chunks/entities) -> expand N hops with relevance pruning (edge weights, PPR) -> pack context under token budget with dedupe and ordering
- [ ] Global search: Leiden communities -> LLM summaries per community (hierarchical) -> map-reduce answer (Microsoft GraphRAG approach)
- [ ] Text2Cypher path for precise structured questions, with schema-in-prompt and validation
- [ ] **AC:** ablation table (baseline vs each addition) on your eval set with quality, latency and cost columns

### WP-9 Grounding + provenance (`citation-provenance-validation`, `data-provenance`)
- [ ] Answer format: claims with citation ids -> verifier checks each claim is supported by the cited chunk (NLI/LLM-judge) -> unsupported claims removed or flagged
- [ ] **AC:** citation precision/recall measured; hallucinated-claim rate reported before/after verifier

### WP-10 Evaluation (`retrieval-evaluation`)
- [ ] Build golden set: >= 150 questions across single-hop, multi-hop, aggregation, "no answer" cases; include adversarial ones
- [ ] Metrics: recall@k, MRR, nDCG for retrieval; faithfulness, answer correctness, citation accuracy, abstention accuracy for generation; latency/cost
- [ ] CI regression gate; judge-model calibration against human labels (report agreement)
- [ ] **AC:** `make eval` produces a report; PRs that regress beyond threshold fail

### WP-11 Security + tenancy (`graph-authorization`, `multi-tenant-graph-isolation`)
- [ ] Neo4j RBAC (roles, privileges incl. label/property-based access control where available), database-per-tenant for strong isolation, composite databases/fabric for cross-tenant admin queries
- [ ] Prevent leakage through shared canonical entities (ER across tenants is a leak!) — resolve within tenant only
- [ ] **AC:** red-team test: tenant A attempts to reach tenant B data via traversal, ER, embeddings, cache; all blocked

### WP-12 Ops (`graph-observability`, `backup-recovery`)
- [ ] Metrics: query latency percentiles, page-cache hit ratio, checkpoint duration, tx rate, replication lag, GDS memory, ingest lag
- [ ] `neo4j-admin database backup/restore` (online), restore drill; consistency check
- [ ] **AC:** restore to new cluster, run smoke queries, record RTO

### WP-13 Plumbing
- [ ] Fill `messaging/schemas`; verify Helm/Terraform; decide `rust-modules` (e.g. blocking/similarity hot path) or `cpp-modules` — or delete

## 3. Deep-dive concepts (Neo4j interview core)
- [ ] Native graph storage: fixed-size record stores (node, relationship, property, string/array), relationship chains (doubly-linked per node), **index-free adjacency** and what it does and does not buy you
- [ ] Dense nodes: relationship groups per type/direction; why supernodes hurt and mitigations
- [ ] Page cache vs OS cache; heap usage; transaction state memory; `server.memory.*` sizing
- [ ] Transaction log, checkpointing, recovery; ACID guarantees; isolation level (read-committed), locks, deadlocks
- [ ] Cypher execution: planner (cost-based), runtimes (slotted/pipelined/parallel), operators (NodeIndexSeek, Expand(All/Into), Eager, CartesianProduct), `PROFILE` reading, hints
- [ ] Indexing: range, text, point, full-text (Lucene), vector (HNSW), composite; when planner uses them
- [ ] Clustering: primaries (Raft) and secondaries, leader election, read routing, causal consistency via bookmarks, `neo4j://` vs `bolt://` routing drivers
- [ ] Scaling: read scaling via secondaries; write scaling limits; property sharding (Infinigraph) / composite databases (Fabric); know current official guidance (verify on docs)
- [ ] Bolt protocol, driver session/transaction functions (managed retries), connection pooling
- [ ] GDS architecture: in-memory projections, compressed adjacency, parallel algorithms, memory estimation
- [ ] Vector search: embeddings, ANN (HNSW parameters M, efConstruction, ef), recall/latency trade-off, filtered search, quantization
- [ ] Knowledge graph theory: RDF vs LPG, ontologies, reasoning, SHACL vs constraints, entity linking, KG embeddings (TransE etc.) at overview level
- [ ] Graph algorithms complexity: BFS/DFS, Dijkstra, PageRank iteration, Louvain modularity, Betweenness O(VE)
- [ ] Neo4j ecosystem: Aura, Kafka Connector, Spark connector, `neo4j-graphrag` Python package, LangChain/LlamaIndex integrations, MCP server for Neo4j, Bloom/Browser

## 4. Failure drills
- [ ] D1 supernode traversal blow-up (cap/partition/intermediate nodes)
- [ ] D2 concurrent ingest -> duplicate entities / lock deadlock
- [ ] D3 bad extractor version pollutes graph -> rollback by `extractor_version`
- [ ] D4 LLM-hallucinated relationship passes gate
- [ ] D5 wrong entity merge -> unmerge procedure
- [ ] D6 GDS projection OOM; long job starves OLTP
- [ ] D7 cross-tenant leak attempt
- [ ] D8 leader failover mid-batch -> idempotent retry
- [ ] D9 stale facts after source update
- [ ] D10 indirect prompt injection in a document ("ignore previous instructions ...") — extraction and answer path must not obey it
- [ ] D11 embedding model change -> reindex strategy (dual index, cutover)
- [ ] D12 page cache thrash after data growth
- **AC:** expected behavior, automated repro, runbook

## 5. Experiments (record in `docs/09-benchmarks.md`)
- [ ] Load: batch size x parallelism -> throughput; lock contention point
- [ ] Query: 1/2/3/5-hop latency vs fan-out; effect of indexes/hints
- [ ] Page cache sizing curve (hot set fits vs not)
- [ ] Vector index: recall@10 vs latency vs HNSW params
- [ ] GraphRAG ablation (WP-8); chunk size/overlap; hops vs quality vs token cost
- [ ] LLM extraction cost/accuracy: model size vs F1 on your labeled set
- **AC:** each lists hypothesis, setup, result, decision

## 6. Company-specific mapping

### Neo4j
- [ ] Read Neo4j Operations Manual (clustering, memory config, backup), Cypher Manual (planning, indexes), GDS Manual (algorithms, memory estimation), GenAI/GraphRAG docs
- [ ] Rebuild a canonical demo: "support knowledge graph + GraphRAG chatbot" using `neo4j-graphrag`; compare with your implementation and note differences
- [ ] Be able to discuss product thinking: who buys graph DBs (fraud, supply chain, KG for AI, identity), why they churn, where Postgres+pgvector is "good enough"
- [ ] Practice customer-facing architecture: sizing workshop (nodes/rels/queries -> instance sizing), migration from RDBMS (ETL, model translation), proof-of-value plan
- [ ] Solutions-architect style questions: "customer's 3-hop query is slow" — PROFILE-driven playbook; "customer has a supernode" — model change options
- [ ] Contribute something visible (small PR/blog/example) to a Neo4j open-source repo — strong signal

### Anthropic
- [ ] RAG for Claude: context window budgeting, long-context vs retrieval trade-offs, prompt caching to cut cost, citations (Claude citations feature) and grounding, structured outputs/tool use for extraction
- [ ] Evals-first mindset: build and trust evals before iterating; LLM-as-judge calibration; regression testing of prompts
- [ ] Agent design: tool design (Cypher tool, search tool), error messages that help the model recover, bounded loops, human-in-the-loop (read "Building effective agents", "Writing tools for agents")
- [ ] MCP: build a small MCP server exposing graph queries safely (read-only, schema-aware, row/time limits) — directly demonstrates relevant skill
- [ ] Safety: prompt injection (direct + indirect), data exfiltration via tools/links, least-privilege tools, audit; discuss mitigations with trade-offs rather than claiming solved
- [ ] Be ready for "why Anthropic": mission, safety research, responsible scaling — authentic, specific

### Google
- [ ] Know Google Knowledge Graph (entities, Freebase heritage), Spanner Graph, and how Search uses entities; compare to your LPG model
- [ ] Large-scale indexing and ranking basics (inverted index, sharding, scatter-gather, tail latency, hedged requests, "The Tail at Scale")
- [ ] Practice design: "design a knowledge panel service", "design entity resolution for 1B records", "design a recommendation system using a social graph"

### Stripe
- [ ] Fraud/risk graph: accounts, cards, devices, IPs as nodes; ring detection via WCC/Louvain; real-time feature lookup latency budget; explain false-positive cost
- [ ] Entity resolution for merchants (names, addresses, URLs) — directly applicable to onboarding/KYB

## 7. Mock interview bank
1. Model a social network in Neo4j and write the friend-of-friend recommendation query; then fix it for celebrities. *Strong: intermediate nodes/fan-out limits, precompute, bounded expansion.*
2. Why is my 3-hop query slow? *Strong: PROFILE, db hits, Eager/Cartesian, missing index/anchor, direction, parameterization, supernode, page cache.*
3. Design GraphRAG for a 10M-document enterprise. *Strong: ingest pipeline, ER, hybrid retrieval, community summaries, ACLs, freshness, eval, cost.*
4. When is a graph DB the wrong choice? *Strong: aggregate analytics, simple key-value, fixed shallow joins, huge write throughput without traversals.*
5. Explain index-free adjacency. *Strong: pointer chasing in relationship chains vs index lookups per join; O(local) vs O(log n); caveats for dense nodes.*
6. How does Neo4j clustering work? *Strong: Raft primaries, secondaries, leader writes, bookmarks, routing driver, failure scenarios.*
7. Design entity resolution at scale. *Strong: blocking, features, clustering, active learning, reversible merges, incremental updates.*
8. How do you evaluate a RAG system? *Strong: retrieval vs generation metrics, golden set construction, judge calibration, online signals, regression gate.*
9. Prevent prompt injection in a document-ingesting agent. *Strong: treat content as data, isolate privileges, no tool access from untrusted context without confirmation, output validation, monitoring; admit residual risk.*
10. Design a fraud ring detector. *Strong: graph features, batch + streaming, thresholds, explainability, human review, feedback loop.*
11. Migrate a relational schema to a graph. *Strong: query-driven modeling, join tables to relationships, staged dual-run, validation, rollback.*
12. Scale writes beyond one primary. *Strong: batching, partition by tenant/domain across databases, composite DBs, property sharding, accept trade-offs.*

## 8. Reading list
1. Neo4j manuals: Operations, Cypher, GDS, GenAI; "Graph Databases" (Robinson, Webber, Eifrem) 2nd ed.
2. "From Local to Global: A Graph RAG Approach to Query-Focused Summarization" (Microsoft)
3. "The Tail at Scale" (Dean & Barroso)
4. HNSW paper (Malkov & Yashunin); DDIA ch. 2-3 (data models, storage)
5. Anthropic: "Building effective agents", "Contextual Retrieval", "Writing tools for agents", MCP specification
6. Entity resolution: Christen "Data Matching"; Splink docs
7. Louvain / Leiden papers (overview), PageRank paper

## 9. Weekly cadence
- 1 WP + AC; 1 ADR; 1 experiment; 1 timed mock (alternate: modeling/query, GraphRAG design, general system design); contribute to eval set daily (10 questions)
