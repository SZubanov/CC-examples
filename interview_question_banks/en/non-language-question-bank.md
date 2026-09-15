# Question bank: language-agnostic questions (DB, architecture, networks, infrastructure, security, algorithms)

**What this is:** a synthesis of the non-language parts of two banks - [Go bank](go-question-bank.md) (parts C, D) and [PHP bank](php-question-bank.md) (parts C, D, E). Duplicates between the banks are merged into one canonical answer; where the answers diverge or the candidate made a mistake - the corrected version (⚠️). Traps - 🪤. These questions get asked at both Go and PHP interviews alike, regardless of language.

**Sources (labels):**
- `G#` - from the Go bank: G1 credit backend · G2 mock Avito/Ozon · G3 Lamoda · G4 mock middle/senior · G5 cache live-coding · G6 mock Ozon · G7 first-response-wins.
- `P#` - from the PHP bank: P1 Top-30 Q&A · P2 275k RUB · P3 350k RUB Savin · P4 Middle/Senior 300k RUB · P5 250k RUB middle · P6 tips/mistakes.

**Legend:** 🪤 trap / provocation · ⚠️ sources diverge / candidate's mistake, corrected here · 💡 nuance · 🔧 code · ✍️ a question that wasn't in the sources.

**How to use:** cover the answer with your hand and say it out loud. Drill the terminology (JOIN, isolation levels, selectivity, at-least-once; trace/span, RED/USE, OTLP, sampling) until it's automatic: fuzzy wording at an interview stands out and sinks you even when you solved the tasks [G3].

---

# 🪤 Trap map (non-language)

| # | Provocation | Correct answer |
|---|---|---|
| 1 | "CROSS JOIN is a table intersection" | **Cartesian product** (every row × every row) [G3, P3] |
| 2 | "Read committed allows dirty reads" | No: RC **excludes** dirty reads. The real threat to balances is **lost update** [G3] |
| 3 | "HAVING can only go after GROUP BY" | It also works **without GROUP BY** (the whole result = one group); the typical spot is after [P3] |
| 4 | "An index on a 2-value column will speed up the query" | Low **selectivity** - useless [G1, P2, P4] |
| 5 | "Store money in float" | Only **integer minor units** (kopecks) [G3] |
| 6 | "Vertical scaling is splitting by tables" | Vertical = **adding resources to a single machine**; splitting (sharding/partitioning/replication) is horizontal [G2] |
| 7 | "Changing a column type via `ALTER ... TYPE` is fine" | On millions of rows it's a **heavy lock**; a new NULL column without DEFAULT + a background migration [P5] |
| 8 | "Creating an index blocks nobody" | It can block; on a live table use **`CREATE INDEX CONCURRENTLY`** [P3, P4] |
| 9 | "Exactly-once delivery is the standard" | Hard to achieve; usually **at-least-once** (duplicates) or at-most-once (losses); exactly-once requires idempotency [P4, G1] |
| 10 | "gRPC is faster because of encryption" | A binary protocol over **HTTP/2** + a `.proto` contract [G2] |
| 11 | "HTTP/2 doesn't need parsing" | Both need parsing; the win is that binary frames **know their exact length** [G6] |
| 12 | "MD5 + salt is enough for passwords" | No: collisions + **timing attacks**; you need bcrypt/argon2id [P3] |
| 13 | "Secure protects the cookie from JS theft" | **HttpOnly** protects against `document.cookie` access [P3] |
| 14 | "An atomic file write is fwrite to a stream" | **temp file + `rename()`** (atomic at the OS level) [P4] |
| 15 | "Reversing an array with a loop over all N" | Loop over **N/2**, otherwise it reverses twice; in practice `array_reverse()` [P5] |
| 16 | "OpenTelemetry is a trace store (a Jaeger equivalent)" | OTel is a **telemetry standard** (API + SDK, export over OTLP); Jaeger/Tempo are storage **backends** with a UI [✍️] |
| 17 | "Tracing replaces logging" | They complement each other: **logs are "what", metrics are "how many", traces are "where"** [✍️] |
| 18 | "DDD is about microservices" | No: DDD is about the model and boundaries; services are about deployment. You pick the boundary by the domain [✍️] |
| 19 | "An aggregate can be modified from several places" | Only through the **aggregate root**; one transaction - one aggregate [✍️] |
| 20 | "A domain event = a message in the broker" | A domain event is internal (in-process); what goes outside is an **integration event** under a contract [✍️] |
| 21 | "Publish the event right after commit" | A crash between commit and publish loses the event → **transactional outbox** [✍️] |

---

# Part A. SQL and databases

## A1. SQL constructs

**A1.1. WHERE vs HAVING.** [G2, G3, G4, P3, P4, P5]
- `WHERE` filters **rows before grouping**; `HAVING` filters **groups after GROUP BY**, by aggregates (`COUNT`, `SUM`).
- 💡 Mnemonic phrasing: "WHERE goes by rows, HAVING goes by groups".

**A1.2. HAVING without GROUP BY.** [P3] ⚠️
- Allowed: the whole result is treated as **one group**. ⚠️ The candidate answered "it complains" - the error `must appear in GROUP BY or be used in an aggregate` applies to columns in SELECT, not to HAVING itself.

**A1.3. Types of JOIN.** [G3, P3] 🪤
- **INNER** - only matching rows. **LEFT/RIGHT** - all rows from one side, NULL on the other. **FULL OUTER** - all rows from both. **CROSS** - a **Cartesian product**, not an "intersection" (candidates mixed it up).
- Users without carts: INNER won't return them; LEFT JOIN → NULL on the cart side.

**A1.4. CROSS JOIN - what it's for.** [P3] ⚠️ - generating every combination of rows from two tables. ⚠️ The candidate couldn't name a use case.

**A1.5. Top-N: the standard skeleton (know it cold).** [G3, G6] 🔧
```sql
SELECT c.email, SUM(carts.amount) AS total
FROM customers c
JOIN carts ON carts.customer_id = c.id
WHERE c.country = 'Россия'          -- filter BEFORE grouping
GROUP BY c.id
HAVING SUM(carts.amount) >= 1000    -- filter AFTER grouping
ORDER BY total DESC
LIMIT 10;
```
- Skeleton: **JOIN → WHERE → GROUP BY → HAVING → ORDER BY DESC → LIMIT**.
- Aggregates: SUM/MIN/MAX/COUNT; `string_agg(title, ', ')` - listing values with a separator (PostgreSQL).
- Everything in SELECT with GROUP BY is either grouped or an aggregate.

**A1.6. CTE (`WITH ...`).** [P3]
- A temporary "view" in memory for the duration of the query; you can create several and join them.
- Performance: keeping it in memory affects the database. Need the subquery twice → `WITH` (computed once); small/simple - a subquery is fine.

## A2. Indexes

**A2.1. What an index is, types, pros/cons.** [P4, P5, G2]
- A structure that speeds up `SELECT`; stored on disk; **slows down INSERT/UPDATE/DELETE** (recomputation). Without an index - a full scan.
- B-tree (balanced; the leaves are linked in a list - fast range scans). Types: primary key, unique, regular, composite, hash, full-text, partial, GIN/GiST (Postgres: JSONB, geo, full-text).

**A2.2. Selectivity; an index on a bool.** [G1, P2, P3, P4] 🪤
- The key concept is **selectivity**. Low-selectivity columns (a 50/50 bool, gender, `status` with two values) → an index is nearly useless ("cut it in half" - the scan is still large).
- It makes sense at high selectivity (true in ~10% - it helps; 90%+ - the optimizer will pick a full scan). In a composite index, **the most selective column goes leftmost**.
- 💡 You have to be able to name the term "selectivity" - candidates in G1/P2/P3/P4 stumbled on it.

**A2.3. When indexes hurt.** [G2, G4, P3, P4]
- Not used - wasted space; slows down INSERT/UPDATE/DELETE; on small tables a seq scan is often faster; low selectivity - useless.
- On a large table - **`CREATE INDEX CONCURRENTLY`** (without blocking writes). ⚠️ Creating an index on a live table can lock it [P3].

**A2.4. Covering index.** [P3]
- `INCLUDE (...)`: the columns are not indexed but are stored on the index leaf - once the value is found, there's no trip to the table (Index Only Scan).

**A2.5. How to check that indexes work.** [G1, G3, P3]
- **`EXPLAIN`** (or `EXPLAIN ANALYZE`): you can see `using index` and the index name, the join algorithms and order. If a DBA says "something's off with the query" - start with EXPLAIN.

**A2.6. Vacuum (PostgreSQL).** [G1] - cleaning up dead row versions (MVCC). A general understanding is enough.

## A3. Transactions and isolation

**A3.1. What a transaction is, ACID.** [P3]
- Several operations as one, "all or nothing" (atomicity). ACID = Atomicity, Consistency, Isolation, Durability. It's needed for consistency (debiting one → crediting another with no intermediate state).

**A3.2. Isolation levels.** [G1, G3, P3]
- `Read uncommitted` → `Read committed` → `Repeatable read` → `Serializable`. The PostgreSQL default is **Read committed**.
- **RC**: only committed data is visible. **RR**: a snapshot taken at transaction start, protects against non-repeatable reads; **phantoms remain**. ⚠️ Correction: in PostgreSQL, READ UNCOMMITTED is treated as READ COMMITTED (not supported separately).
- Anomalies: dirty read, non-repeatable read, phantom read, lost update.

**A3.3. Debiting a balance: a task with a catch.** [G3] 🪤 🔧
- Problem #1: **float for money is a mistake**, only integer minor units (2350 ₽ = 235000).
- Problem #2: read committed + two concurrent debits → **lost update** (both read the balance, each wrote its own). This is NOT a "dirty read" (RC excludes it) - the candidate got the term wrong.
- Solutions:
  - `SELECT ... FOR UPDATE` (holds the lock until **the end of the transaction**; when transferring between wallets, lock both rows **in a fixed order** to avoid deadlocks).
  - An atomic conditional UPDATE: `UPDATE wallets SET balance = balance - 100 WHERE user_id = 42 AND balance >= 100;`
  - `CHECK (balance >= 0)` + an unconditional UPDATE.
- Raising isolation globally is not worth it ("all correct, but performance will drop"): a targeted lock is preferable.
- **Deadlock**: two transactions hold each other's rows; avoid it by ordering locks.

## A4. Replication and scaling

**A4.1. Master-slave replication.** [G3, G4, P4, P5]
- Write to the master, read from replicas (there are usually more reads). Horizontal scaling + fault tolerance.
- Synchronous/asynchronous; with asynchronous there can be **replication lag** → eventual consistency (you need to handle the inconsistency).

**A4.2. Failover: switching the master.** [P5] ⚠️
- Yes: when the master goes down, a replica becomes the new master. Risk: with asynchronous replication, some data may not have arrived. ⚠️ The candidate drifted into describing lag, while the main scenario is fault tolerance when the DB server goes down.

**A4.3. Vertical vs horizontal.** [G2] 🪤
- **Vertical** - adding resources to one machine (CPU/memory/disk) - there's a **ceiling**. ⚠️ The candidate called "splitting by tables" vertical - wrong.
- **Horizontal**: replication (copies), sharding (splitting data by rows/tables), partitioning (a table is split inside a single DB).

**A4.4. Sharding: choosing the key.** [G3]
- **By the business access logic**: chats → by chat, orders → by region/location. Partitioning is easiest **by time** (created_at).

**A4.5. View vs materialized view.** [G4]
- **View** - a query alias, takes no space; downside: logic moves to the DB side (often an anti-pattern).
- **Materialized view** - a **cached snapshot**; updated via `REFRESH MATERIALIZED VIEW`; keeps no history.

## A5. Locks and migrations

**A5.1. Locks in a DB.** [P5, G3]
- Levels: database, table, row. At the row level: `FOR UPDATE` (exclusive), `FOR NO KEY UPDATE`, `FOR SHARE` (less exclusive).

**A5.2. Heavy migration (changing a column type).** [P5] ⚠️ 🔧
- A naive `ALTER ... TYPE` (integer → bigint) on millions of rows takes a heavy lock (in Postgres - `ACCESS EXCLUSIVE`) → downtime.
- The non-blocking path:
  1. A new column of the required type **as NULL and without DEFAULT** (in Postgres - instant, metadata only).
  2. In the code - dual writes to both columns.
  3. Backfill/convert the old data in background batches.
  4. Switch the code, drop the old column.
- Tools: gh-ost / pt-online-schema-change (MySQL), pg-osc (Postgres). ⚠️ The candidate searched for a long time (batches, dump - off the mark), the interviewer had to suggest the solution.

## A6. Normalization and schema design

**A6.1. Normal forms.** [P3]
- **1NF** - don't stuff several values into one column (split the full name). **2NF** - avoid duplication (move `resume_link`/`source` out). **3NF** - ⚠️ the candidate couldn't recall it.

**A6.2. "Companies and ads" schema.** [P2]
- Entities: `users`, `companies`, `ads`, `countries`, `tags` + junction tables for white/blacklist filters (many-to-many): `country_rules`, `tag_rules` with an allow/forbid flag.
- `status` - `varchar`, not `enum` (a new enum status requires an ALTER); `id` - UUIDv7 or unsigned bigint; soft delete via `deleted_at` (composite unique on `name` + `deleted_at`).

**A6.3. Recruiting schema (EAV).** [P3]
- Separate **"candidate" and "resume"** (one person - several resumes for different openings); interviews reference `resume_id`; the source is a separate table; dynamic fields - **EAV** (`candidate_properties` + `property_values`).
- Indexes: foreign keys + `status` (where there are many statuses) + `created`; don't index a status with two values.

**A6.4. Partner program (month closing).** [P2]
- Store **a specific date** (not a period) - otherwise you can't aggregate. The token - **encrypt reversibly** (don't hash it: you need the original to call the API).
- Month close: two date fields - `date` (business) and `source_date` (in the source); after closing, changes are forbidden (enforced in the app + `CHECK`/trigger); refunds - as a negative number.
- Updates - a background worker/cron, not a button; don't rely on a webhook (they most likely won't provide one).

**A6.5. Loading 30 million records (tax database).** [P6]
- Describe the solution design (what to use, how to build it, why), the impact of optimizations and how to profile. What's evaluated is **your thinking approach**, not a memorized answer. See also chunked/streaming processing under memory constraints.

**A6.6. Library schema (many-to-many).** [G6] 🪤
- The "Gang of Four" trap: a book with 4 authors doesn't fit into "author_id in books". Many-to-many → a junction table `book_authors`.
- Reader↔book: `reader_id` in books is **nullable** + a unique constraint on `book_id` (at most one reader).

## A7. Performance diagnostics

**A7.1. "The service is slow" - a systematic answer.** [G5]
1. Logs → find the request (ID) → **trace ID** → the whole path.
2. **The delta between logs** → the spot with the biggest delay.
3. Which queries are in that chunk → **EXPLAIN** (rows/cost, indexes, joins).
4. If it's the code - **pprof** (Go) / Xdebug (PHP); bring it up locally and run it.
- Answer from the assumption that the company has everything (Kibana, tracing). Don't pile on technologies: first exhaust the DB (indexes, materialized views), Redis is a last resort.

**A7.2. JSON columns in a DB.** [P2]
- Trade-off: JSON or separate tables. JSON is worse in speed and size; you can index it, but it's "not very snappy". The argument for normalized tables is more sensible.

---

# Part B. Architecture and system design

## B1. Architectural styles

**B1.1. MVC / DDD / hexagon / CQRS.** [P3]
- They don't contradict each other: MVC is about View/Controller, the model is the domain layer. View = a DTO in the response, controller = the HTTP layer, model = the domain.
- **CQRS** - separating reads/writes; the motivation is heavy read load (the query side reads directly from the database, bypassing the ORM).
- 💡 "MVC + DDD + hexagonal don't contradict each other" - a solid position.

**B1.2. Layered architecture.** [P3, P5]
- DDD layers: **domain** (business logic), **application** (use cases), **infrastructure** (DB, integrations), **presentation** (API/controllers).
- Patterns: onion, hexagonal (the core is business, adapters/ports around it).

**B1.3. DDD aggregate.** [P5]
- A unit of consistency: one entry point (the aggregate root); related objects change only through it.
- Details in B5.5 (transaction boundary, references by ID, one transaction = one aggregate).

## B2. Modularity

**B2.1. Coupling vs Cohesion.** [P5] ⚠️
- **Coupling** - links **between** modules, should be low. **Cohesion** - links **inside** a module, should be high (a module is responsible for its own context).
- Anti-example - god objects ("god services"). ⚠️ The candidate mixed up the terms.

**B2.2. Switching the DBMS / external services.** [P3]
- Switching the DBMS is a **super-rare process**; flexibility for it is often unnecessary. ORMs can work with different DBs, but you pay with performance.
- External services - via **an interface + inversion**: implementations behind a single interface, interchangeable through feature flags/env.
- Risk minimization: caching read queries, fallback, a clear error, API versioning, stubs for tests.

## B3. Microservices

**B3.1. What they solve and the pitfalls of extracting a feature.** [P5, G4]
- They solve: extracting a high-load module, independent deployment, horizontal scaling, a separate stack.
- Pitfalls of extracting a feature: coupling → refactoring; choosing sync/async (brokers) - async is simpler but adds complexity; data consistency between databases; DevOps load; different stacks → narrow specialists. "You can do anything, but the question is why".

**B3.2. Downsides of microservices.** [G4]
- Harder to **debug** (a request spans several services; harder to roll back; transactions across services).
- Decentralization: **you don't know what's in other services** (separate databases, separate teams).
- Interaction is slower (data overhead) vs in-process calls in a monolith; development speed is lower for an MVP; more costs and maintenance.

## B4. SOLID and patterns

**B4.1. SOLID (cheat sheet).** [P4, P5]
- S - single responsibility · O - open/closed · L - Liskov substitution (a subclass correctly replaces its parent) · I - interface segregation · D - dependency inversion (depend on abstractions; solved with interfaces + DI).
- 💡 A mature position: not every principle has to be jammed in everywhere - otherwise you get "Hello World in abstract factories" [P3].

**B4.2. Design patterns.** [P4] ⚠️
- The basic set: Singleton, Factory (Abstract/Factory Method/Simple), Decorator, Adapter, Proxy, Chain of Responsibility, Iterator, Strategy, Builder.
- Applications: **Decorator** (wrap a repository with a cache), **Builder** (DTOs), **Singleton** (service provider), **Strategy** (payment drivers), CQRS (separating reads/writes).
- ⚠️ The candidate mixed up Strategy / Abstract Factory / Factory Method: Strategy is choosing an algorithm through a common interface; Factory is creating objects.

## B5. DDD (applied): contexts, aggregates, domain events

> Added to match a job requirement ("Applied DDD practices: bounded contexts, aggregates, domain events") - a block for the interview. These questions weren't in the real recordings, so all entries are authored [✍️].
> Related: `reference/sources/design_event_sourcing_processing.md` (aggregate, hydration, projections), `design_payment_system.md` (outbox, saga, idempotency), `reference/wiki/cqrs.md`.

**B5.1. What DDD is and its two levels.** [✍️]
- DDD (Eric Evans, "Domain-Driven Design") - an approach to designing **a complex domain**: the domain model is at the center, the code speaks the language of the business. It is not a framework and not a set of patterns.
- **Strategic level**: language, context boundaries, a context map, subdomains. It answers "where the models are and how they fit together".
- **Tactical level**: entity, value object, **aggregate**, **domain event**, repository, domain service. It answers "how the model looks in code".
- Subdomains: **core** - a competitive advantage (maximum attention), **supporting** - needed but not unique, **generic** - take an off-the-shelf solution, don't write your own.
- 💡 A mature answer: "in DDD the main thing is language and boundaries, tactical patterns are secondary". The answer "aggregates and repositories" is course-level, not senior.

**B5.2. Ubiquitous language (a shared language).** [✍️]
- One vocabulary for the team and the business: the same words in conversation and in code (`Payment`, `Capture`, `Settlement`).
- In practice: if the business says "the payment is stuck" while the code has `status = 5`, every discussion requires translation - meaning gets lost, requirement errors stay invisible.
- Why: it lowers the cost of communication; a term you can't say out loud usually means the concept doesn't exist in the domain - or that two different domains have been glued into one model.
- The catch: the language works **inside its own context** - outside it, the same words mean something else (see B5.3).

**B5.3. Bounded context.** [✍️]
- The boundary inside which the model and the language are **consistent**: one word - one meaning.
- One word - different models: "Payment" in processing (statuses, retries to the PSP, UNKNOWN) and "Payment" in the ledger (entries) are different concepts; so is "Customer" in sales, support and billing. Trying to make one shared company-wide model is a source of conflicts.
- **Context ≠ module ≠ service**: context is about the model and language, a module is about code structure, a service is about deployment and scale. One context can live happily in a monolith; extracting it into a service is a consequence of a boundary you drew, not a goal of DDD.
- 🪤 "DDD is about microservices" - no: you choose the context boundary by the domain.

**B5.4. Context map: how contexts connect.** [✍️]
- A context map: which contexts exist and who influences whom (upstream/downstream).
- Relationship patterns: **Shared Kernel** (a shared part of the model), **Customer/Supplier** (customer-supplier), **Conformist** (we adapt to someone else's model), **Anticorruption Layer (ACL)** (a translator at the boundary: the foreign model doesn't leak inside), **Open Host Service / Published Language** (a public contract), **Separate Ways**.
- Applied: an ACL at the boundary with legacy and external APIs (PSP, bank) - their statuses and error codes are mapped to ours at the boundary instead of scattering across the code.
- ⚠️ Shared Kernel is often loved as "one common module for everyone" - it couples team releases: saves you at the start, costs a lot later.

**B5.5. Aggregate and aggregate root.** [✍️]
- **Aggregate** - a cluster of objects that changes as a single whole: it is the **consistency and transaction boundary**, inside which invariants always hold.
- **Aggregate root** - the single entry point: the outside references only the root and only by **ID**, state changes go only through the root's methods, internal objects are not exposed.
- Transaction rule: **one transaction - one aggregate**; between aggregates - a reference by ID and eventual consistency (via events, see B5.8).
- Keep aggregates **small**: the bigger the aggregate, the higher the lock contention (lost updates, deadlocks) and the longer the transaction.
- Example: `Payment` is an aggregate with the invariant "you can't capture a canceled payment"; ledger entries are a separate aggregate and context, synchronized via events.
- 💡 A sign of a wrong boundary: you want to modify two objects in one transaction → they're probably one aggregate; and if an aggregate has ballooned and any edits conflict - the boundary was drawn too wide.

**B5.6. Domain events.** [✍️]
- A domain event is a fact important to the business, in the **past tense**: `PaymentCaptured`, `RefundIssued`, `PayoutRequested`. Immutable; carries the aggregate id, aggregate version, time, idempotency key, actor.
- Published by **the aggregate root** at the moment of a state change; the subscribers are projections, other contexts, sagas, notifications, audit.
- Why: it decouples aggregates and contexts (no dragging foreign dependencies into a transaction), gives an explicit audit of "what happened", replaces part of the cross-aggregate transactions.
- **Domain event ≠ integration event**: a domain event is internal (in-process) and can be refactored; an integration event is a public contract to the outside (broker/bus, versioning, fewer details). Translation happens at the context boundary.

**B5.7. Reliable event publishing: transactional outbox.** [✍️]
- The problem: state in the DB and the event in the broker are two systems; we're not doing a distributed transaction.
- The solution: the event is written to an **outbox** table in the same transaction as the state; a separate worker (or CDC/Debezium) publishes it and marks published rows.
- Delivery is **at-least-once** → the consumer is **idempotent** (idempotency key + deduplication), see `design_payment_system.md`.
- Alternative: event sourcing - the event itself is the source of truth (for aggregates that need history: a payment, a wallet; reference data - no).
- 🪤 "Publish right after commit from the handler" - a crash between commit and publish loses the event; publishing before commit - the consumer sees something that doesn't exist yet.

**B5.8. Consistency between aggregates (saga).** [✍️]
- Inside an aggregate - strict consistency; **between aggregates - eventual**: event → reaction → compensation.
- A long process (debit → capture at the PSP → ledger entries → payout) is a **saga / process manager**: steps + compensations on failure (refund/void), the process state is stored separately.
- In a monolith you sometimes allow a transaction across two aggregates - that's a deliberate trade-off: the "one transaction - one aggregate" rule exists so the code survives extraction into a separate service.
- 💡 Say it out loud: "consistency is strict inside an aggregate, eventual between them" - that is exactly what the interviewer is checking.

**B5.9. When DDD is not needed.** [✍️]
- CRUD apps and simple domains (admin panels, reference data, configs) → DDD = overengineering: the abstractions cost more than they save.
- **DDD-lite** (Fowler): take the strategic part - language, boundaries, contexts; tactical patterns - only as needed.
- Microservices ≠ DDD; a "distributed monolith" is a symptom of boundaries cut along technical layers instead of the domain.
- CQRS/event sourcing are not a mandatory part of DDD: apply them where audit and history are needed.
- 💡 A mature position: "DDD is justified in a complex domain - payments, the ledger; in an admin panel I'd take DDD-lite and not multiply abstractions".

**B5.10. "How have you applied DDD?" - an answer skeleton.** [✍️]
- Skeleton: the domain → where you drew the context boundary → what you made an aggregate and which invariant it protected → which events went outside and how they were delivered → what it gave you (metric/effect).
- Terms for the EN version: bounded context, aggregate root, domain event, integration event, eventual consistency, transactional outbox, anti-corruption layer, ubiquitous language.
- `[to be filled in]` facts (project, boundaries, invariant, events, effect) - take them from your own experience; don't invent them.

---

# Part C. Networks, HTTP, brokers

## C1. Networks

**C1.1. TCP vs UDP.** [G2, P4]
- TCP guarantees delivery (connection setup, ordering, retransmission). UDP is a "raw stream" (VoIP/streaming/games): packets can be lost, but it's faster.

**C1.2. What happens when you type google.com.** [P4] ⚠️
- DNS → IP; TCP handshake; nginx/apache accepts → static files or a proxy to the backend (PHP-FPM/Go) → the application.
- ⚠️ A basic outline without TLS/HTTPS, the HTTP request/response structure, status codes (interviewer: "maybe").

**C1.3. Timeouts.** [P4] 💡
- They protect against "infinite" requests and stuck connections; nginx returns 504 when the threshold is exceeded.
- For PHP-FPM: the worker pool is limited - stuck requests exhaust the pool (self-DoS). Long operations - move to the background/queues.

## C2. HTTP

**C2.1. HTTP request structure, codes.** [G6]
- A request: method, path, protocol version; headers - metadata (Host, Content-Type); body - data. **Host** - several sites on one IP, the server knows which resource is meant.
- Codes: 200, 201 (Created), 204 (No Content), 301/302 (redirects), 404. A custom HTTP method - formally possible, bad practice.

**C2.2. HTTP/1.1 vs HTTP/2.** [G6] ⚠️
- 1.1 is a text protocol; 2 uses **binary frames**. ⚠️ Parsing exists in both (the interviewer pounced on this) - the win is that frames **know their exact length** and are read unambiguously faster.

**C2.3. gRPC.** [G2]
- Over **HTTP/2**, data in a **binary format**, contract in `.proto`, clients generated for any language, strict typing, streaming. For service-to-service communication **inside the perimeter**.
- Why not everywhere: the migration cost, HTTP/2 not available everywhere, inconvenient to test. ⚠️ Not "faster because of encryption" - the point is the binary protocol + HTTP/2.

## C3. Queues and brokers

**C3.1. Why queues.** [P4, P5]
- Move long/non-critical operations (email, notifications, logging) to an asynchronous background; the user gets an answer "here and now".

**C3.2. Kafka: partitions and replicas.** [G1]
- A **partition** is a logical sequence of messages: **ordering** + parallel processing. A **replica** is a copy of a partition **on another broker**: resilience; on failure - rebalancing.

**C3.3. Consumer groups.** [G4] - they group consumers: inside a group, **one partition is never read by two consumers**.

**C3.4. Delivery guarantees.** [G1, P4] ⚠️
- **at-least-once** (duplicates possible), **at-most-once** (losses possible), **exactly-once** (hard to achieve, requires idempotency). ⚠️ The candidate didn't know the terminology.

**C3.5. Poison message and DLQ.** [P4]
- A message fails → gets requeued → fails again: a retry counter + a limit (max attempts) → the dead letter queue (**DLQ**).

**C3.6. The broker is down - how not to lose messages.** [P3]
- In RabbitMQ - the **durable/persistent** flag when creating a queue: messages are backed up to disk. Enabled by default in Symfony.
- A payment system with polling: retries with pauses + **DLQ** - "delayed payments (bad, but not critical) instead of losses (critical)".

---

# Part D. Infrastructure

## D1. Containers and orchestration

**D1.1. VM vs container.** [G2, G6, P5]
- **Virtualization**: the hypervisor emulates a machine (its own OS on top of yours) - heavy, but **genuinely isolated**.
- **Container**: uses the host OS, syscalls almost directly - lightweight, fast, **weaker isolation**.
- Docker's upsides: lightness, **layered** images (only the changed layer is rebuilt), Docker Hub, fast deployment/scaling. The "suspicious binary" provocation: in a container you risk the host OS (isolation is incomplete) [G6].

**D1.2. Docker vs Kubernetes.** [G2, P5]
- Docker is building containers/images; K8s is **orchestration**: self-healing, load distribution, rolling updates, "so that not all pods die". For fault tolerance - 3+ nodes.

## D2. Git

**D2.1. merge vs rebase.** [P4]
- `merge` is an ordinary merge, **preserves history**. `rebase` **rewrites history** (you can lose commits); acceptable when you work alone in a branch, in a team - merge is better.

## D3. Observability: monitoring, alerting, tracing

**D3.1. Tools.** [G6] - **Prometheus** (metrics/time series), **Grafana** (dashboards/alerts), **Jaeger/Tempo** (distributed tracing backends), **OpenTelemetry** (telemetry standard: API+SDK, export over **OTLP**), **ELK/Loki** (logs). Cheat sheet: metrics - "how many", logs - "what", traces - "where".

**D3.2. Service alerting.** [G6] - error rate above a threshold (4xx - Bad Request, 401/403 - authentication; connection refused - an infrastructure problem), CPU/memory above a threshold.

**D3.3. What distributed tracing is and why.** [✍️]
- In microservices a request passes through several services; each service's logs alone don't show the path - **a trace** links all the pieces into one story.
- Concepts: **trace** - the story of one logical request; **span** - a segment of the path (a service, a DB call, an external API); **trace ID** - the common identifier; **span ID + parent/child** - the span tree.
- The main use case: in minutes find **which service** added the delay or failed the request (cascading timeouts, retries), investigate an incident by a single ID.

**D3.4. How context is passed between services (propagation).** [✍️]
- A service reads the context from the incoming call, creates a **child span**, and passes the context on with every outgoing call.
- HTTP - **W3C Trace Context**: the `traceparent` header (`{trace-id}-{span-id}-{flags}`); the older standard is X-B3 (Zipkin). gRPC - metadata. Kafka/RabbitMQ - **message headers** (the worker continues the parent's trace instead of starting a new one).
- "The trace broke" almost always means **the header wasn't passed**: a hand-rolled HTTP client without middleware, incompatible standards between services.

**D3.5. Trace sampling.** [✍️]
- Tracing **every** request is expensive: telemetry traffic is comparable to production traffic → you keep a sample.
- **Head-based** - the sample fraction is decided at the entrance (e.g. 5-10%). **Tail-based** - the decision is made after completion: you can keep **all errors and slow** traces - the thing the whole exercise was for.
- Iron rule: **never sample errors and tails (p99)**. In payments - 100% or sampling by a business key (all operations on a transaction - always).

**D3.6. OpenTelemetry.** [✍️] 🪤
- A vendor-neutral **telemetry standard**: API (application code) + SDK (collection, sampling, export) + exporters speaking **OTLP**; three signals (traces + metrics + logs) under one umbrella.
- The point: you instrument once - you can send everything to any backend (Jaeger, Tempo, Datadog, New Relic).
- 🪤 The provocation "OTel is Jaeger": OTel is the API/SDK/protocol; Jaeger/Prometheus are storage and UI **backends** (trap map #16).

**D3.7. Logs and traces: linking them.** [✍️]
- **Log correlation**: the trace ID is written into every log line (in Go - a context logger; in PHP - a Monolog processor). From a log/alert → the trace; from the trace → the logs of a specific span.
- Without the link, tracing and logs are two parallel worlds and the investigation is again "by timestamp".

**D3.8. Metrics: RED and USE.** [✍️]
- **RED** is about services and requests: **R**ate (RPS), **E**rrors (error rate), **D**uration (p50/p95/p99).
- **USE** is about resources: **U**tilization, **S**aturation (queues, pools), **E**rrors.
- Investigation: RPS is the same, latency is up, no errors → resources (USE) or code/DB (trace). Latency and errors grow together → an external service/DB - the trace will show the culprit.

**D3.9. SLO / SLI / error budget.** [✍️]
- **SLI** - a measurable metric (p99 < 500 ms; the share of successful requests). **SLO** - a target threshold (99.9%). **Error budget** - the allowed amount of SLO violations per period.
- The alert is tied to the **budget burn rate**, not to "one error" - otherwise you get alert noise.
- 💡 For a senior, the "SLI → SLO → budget" chain + why it beats threshold alerts is enough.

**D3.10. "Traces disappeared" - the order of checks.** [✍️]
1. **Propagation** - is the header arriving? (a break = a client call didn't pass it on).
2. **Sampling** - are errors and slow requests guaranteed to be recorded?
3. **Export** - are the collector/backend alive, are the OTLP endpoint and token correct?
4. Are the services' **clocks** synchronized (NTP) - otherwise the spans form a "crooked tree".
5. Are **propagation standards** between services compatible (traceparent vs X-B3 won't knit together).

**D3.11. Tracing in payments.** [✍️]
- A money operation is an end-to-end path: API → validation → antifraud → external payment system → webhook → reconciliation. The trace is an **audit key**: given the `payment_id` in span attributes and a support ticket, the whole path is visible.
- The external payment system usually doesn't participate in propagation → an "external call" span + a `payment_id ↔ trace_id` mapping in your own services.
- Retries are visible: one trace ID, several attempt spans - you immediately see "it churned 3 times before succeeding".

---

# Part E. Security (general)

**E1.1. Which vulnerabilities you've encountered and how to defend.** [P3]
- **SQL injection** (prepared statements). **XSS** - someone else's JS executes in the victim's browser. Enumeration of vulnerable files (`phpinfo()`, configs): only the entry point should be exposed publicly, turn off profilers in production. XML/ZIP bombs. Open ports (the nuclei scanner). Privilege escalation. Missing TLS (sniffing).

**E1.2. Cookie protection: flags.** [P3] ⚠️
- **`HttpOnly`** - blocks access from JS (`document.cookie`), protection against theft via XSS. **`Secure`** - HTTPS only. **`SameSite`**, **`Domain`**.
- ⚠️ The candidate named only SameSite/Domain, missed HttpOnly and Secure.

**E1.3. Why you can't use MD5 for passwords.** [P3] ⚠️
- Fast collisions + **timing attacks** (you guess characters from the response time). You need modern algorithms with a **salt**: bcrypt/argon2id. MD5 is cracked in an hour; a modern one takes ~10 years.

**E1.4. JWT: what it is and why.** [P3]
- JSON Web Token is an authorization method on par with sessions; it's needed in **distributed systems** (several servers, you can't share sessions). Three dot-separated parts (base64): header, payload, signature.

---

# Part F. Algorithms

**F1.1. Array reversal (a task with a bug).** [P5] 🪤 🔧
- A function with a bug: the loop runs over **all N** instead of **N/2** → the array reverses twice and comes back in the original order.
```php
// Bugged version: looping over all N → reverses twice
function reverseArray(array $array): array
{
    $temp = 0;
    for ($i = 0; $i < count($array); $i++) {
        $temp = $array[$i];
        $array[$i] = $array[count($array) - $i - 1];
        $array[count($array) - $i - 1] = $temp;
    }
    return $array;
}
```
- For `[1,2,3,4,5,6]` the result is `[1,2,3,4,5,6]` again. Fix: `$i < intdiv(count($array), 2)`.
- In practice - **`array_reverse()`** (in Go - write it by hand with two pointers), don't reinvent the wheel. ⚠️ The candidate botched the manual index trace (counted from one, not from zero).

---

# Part G. Behavioral and preparation

> Here is a brief checklist of behavioral topics that came up in interviews. Deep preparation (STAR, personal contribution, numbers) is separate, per your own projects.

**G1. "Tell me about yourself", motivation, deadlines, STAR.** [P1, P5, P4]
- "About yourself": background → experience/stack → relevant projects → achievements; keep it short and tailored to the specific company.
- A deadline + a comment in review: if everything works - close the business task, put the refactoring into **tech debt** (the interviewer confirmed this as the correct answer).
- STAR questions (a difficult project/problem): a real project → the difficulty → your approach → a measurable result.
- AI/LLM: docs, books; LLMs as a search engine and for general awareness, not in "solve this code for me" mode (they can produce nonsense).

**G2. Preparation: what actually works.** [P6, P5]
- Know your grade (junior/middle/senior - different questions); learn the standard checklist - the market runs on it; have practice you can talk about; set up a GitHub with a pet project (even unfinished, but solid in architecture). Grades are blurry - prepare at the maximum level.

**G3. Candidate mistakes (don't do these).** [P6]
- Cramming without understanding; ignoring security best practices; dirty code; no real examples; weak DB skills; outdated practices; poor communication; no tests; inability to admit you don't know (an honest "I don't know, but here's how I'd look it up" is valued).

**G4. Should you prepare algorithms?** [P5]
- Out of ~35 interviews, algorithms came up in ~6; if you do prepare - "Grokking Algorithms" + ~200 easy / 50 medium on LeetCode.

---

# Part H. Interview techniques (interviewer feedback)

> From the Go bank (Part E). Not questions, but behavior that decides the outcome.

1. **Think out loud** - silence in 2026 = suspicion of an LLM/Googling; live coding is a mini-lecture [G5].
2. **Summarize** if the answer drags on - the interviewer's "mush in my head" becomes your problem [G5].
3. **Clarify requirements before coding** (replica count, timeouts, case sensitivity) - most interviewers give you a plus [G7, G3].
4. **Run a primitive case before saying "I'm done"** - finding the bug yourself before declaring readiness is even better [G7].
5. **Decompose** - a function >50 lines → extract it [G7].
6. **An honest "I don't know / haven't done that" is fine** (Kafka/DB administration): the interviewer is testing depth and honesty [G1, G3, G6].
7. **Seniority = solving within constraints** ("no Redis allowed"), not piling on technologies [G5].
8. **You may be more right than the interviewer** (cache: O(1) vs O(n)) - a middle/senior interview is a reasoning exercise, not a dictation [G5].
9. **Terminology to the point of automaticity** (JOIN, HAVING/WHERE, isolation levels, selectivity) - "solved the tasks" ≠ "passed the interview" [G3].
10. **The candidate's questions** - to the point, don't overload: team composition, the lead's role, neighboring teams; don't grill for task specifics at early stages [G1, G3].
11. **Connect your experience to the interviewer's project** ("right now I'm in a similar story") - it works [G1].
12. **A code review at an interview** can be harsher than in real life - that's normal [G1].

---

*Synthesis of the non-language questions from the [Go bank](go-question-bank.md) (parts C/D) and the [PHP bank](php-question-bank.md) (parts C/D/E), 24.08.2026. Section D3 (observability, distributed tracing) - questions that were not in the sources [✍️].*
