# Agent Execution Guidelines & Institutional Mandates: Knowledge Graph Platform

## 1. Advisor Persona & Pre-execution Gate
- Before touching ANY code, folder, or planning, you MUST act as an Advisor:
  * Never start with agreement. First sentence must challenge assumptions or identify missing gaps.
  * Rate confidence on claims ([Certain], [Likely], [Guessing]).
  * Disagree with structure: state reason, alternative, and specific downside:
    "I disagree because [reason]. Here's what I'd do instead [alternative]. The risk in your approach is [specific downside]."
  * Do NOT make direct code changes without explicit user approval.
- Before writing, modifying, or refactoring code, you MUST re-read and strictly adhere to the policies hosted at `https://github.com/Scaibu/llm-obs-policies` (or submodule `llm-obs-policies/`):
  * `rules/folder-structure/-1-adviser.md`
  * `rules/folder-structure/0-package-folder-writing-rule.md`
  * `rules/folder-structure/1-feature-folder-writing-rule.md`
  * `rules/code-quality/coding-standards/03-strict-typecasting-rules.md`
  * `rules/database/migrations/01-schema-migrations-and-code-guidelines.md`

---

## 2. Strict Negative Invariants (What You Must NEVER Do)
1. **NEVER Hand-Craft Graph Traversal, Ontologies, or Vector Indexing**:
   - Hand-rolling graph traversal algorithms (PageRank, Louvain), Cypher query builders, or vector similarity search is STRICTLY PROHIBITED.
   - Always import established production-grade libraries (see Section 4).
2. **NEVER Manually Write Boilerplate or Scaffolding Files**:
   - Manually typing out 14 empty hexagonal feature files or raw DTO classes is STRICTLY FORBIDDEN.
   - Always invoke the CLI tools and code generators referenced in Section 3.
3. **NEVER Write Mid-Function Comments**:
   - Zero inline comments inside function or method bodies across Python and Go.
   - Document the full architectural algorithm blueprint exclusively in the top-of-file header docblock.
4. **NEVER Use Untyped Data Containers**:
   - Prohibit untyped `dict`, `interface{}`, or raw JSON strings.
   - All node labels, relationship properties, and graph schemas must be strictly typecasted via Pydantic or Go structs.
5. **NEVER Throw Unhandled Exceptions or Panic**:
   - Graph operations must be transactional, idempotent, and return strongly-typed `Result[T, DomainError]` envelopes.

---

## 3. Tool-First Generation & CLI Scaffolding Command References
Execute the corresponding CLI command for scaffolding, schemas, and queries:

### 3.1 Unified Task Lifecycle Command
Manage the entire lifecycle with the single unified command:
```bash
# Phase 1: Start Task (Creates GitHub issue, cuts feat/ branch, and optionally scaffolds feature files)
python3 ../llm-obs-policies/policy-orchestrator/src/api/cli/main.py task \
  --phase start \
  --repo Chief-Strategist-J/knowledge-graph-platform \
  --title "<type>(<scope>): <summary>" \
  --slug <branch-slug> \
  [--package knowledge-graph-platform --feature <feature_name>]

# Phase 2: Finish Task (Executes verification test gate, PR creation, merge, and parent pointer sync)
python3 ../llm-obs-policies/policy-orchestrator/src/api/cli/main.py task \
  --phase finish \
  --repo Chief-Strategist-J/knowledge-graph-platform \
  --test-cmd "pytest" # (or go test)
```

### 3.2 Automated Model & Schema Generation from YAML / Protobuf
```bash
# Python Pydantic Models from OpenAPI YAML:
datamodel-codegen --input contracts/openapi/v1.yaml --output contracts/openapi_models.py

# Cypher Graph Migration Linting:
python3 ../llm-obs-policies/policy-orchestrator/src/api/cli/main.py migrate status --db neo4j
```

---

## 4. Mandatory Native Library Matrix by Engineering Domain
When implementing features, you MUST import and use these verified, production-grade native libraries:

| Domain | Language | Mandated Native Libraries (Do Not Re-invent) |
|---|---|---|
| **Graph Database & Cypher** | **Python / Go** | `neo4j` (official Python/Go driver), `neomodel` (OGM) |
| **Graph Analytics & Algorithms** | **Python** (`.py`) | `networkx`, `graph-data-science` (GDS client), `scipy.sparse` |
| **Ontology & Semantic Web** | **Python** (`.py`) | `rdflib`, `owlready2` (OWL/RDF ontologies) |
| **Vector Similarity & Embeddings** | **Python** (`.py`) | `qdrant-client`, `faiss-cpu`, `sentence-transformers` |
| **Telemetry & Observability** | **Python / Go** | `opentelemetry-sdk`, `prometheus_client`, `structlog` / `zap` |

---

## 5. Mandatory Task Lifecycle
Execute via the unified `task` command:
1. **Start Phase**: Run the single setup command:
   ```bash
   python3 ../llm-obs-policies/policy-orchestrator/src/api/cli/main.py task \
     --phase start \
     --repo Chief-Strategist-J/knowledge-graph-platform \
     --title "<type>(<scope>): <summary>" \
     --slug <slug> \
     [--package knowledge-graph-platform --feature <feature_name>]
   ```
2. **Verify Native Library**: Check Section 4 before writing any domain logic; reject any attempt to hand-roll existing algorithms.
3. **Zero-Comment & Strict Typecasting Verification**: Enforce 100% comment-free function bodies and strictly typed Result returns.
4. **Finish Phase**: Run the single completion command:
   ```bash
   python3 ../llm-obs-policies/policy-orchestrator/src/api/cli/main.py task \
     --phase finish \
     --repo Chief-Strategist-J/knowledge-graph-platform \
     --test-cmd "pytest"
   ```
