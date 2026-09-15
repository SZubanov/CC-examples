# Question bank: real Go interviews (synthesis of 7 transcripts)

**What this is:** a synthesis of seven transcripts of real and mock interviews (YouTube videos). Duplicate questions are merged into one canonical answer; where the sources disagree or a source is wrong, the corrected version is given. Trap questions and provocations are marked 🪤. The questions that are not about Go (DBs, infrastructure, networks) are moved out into separate Parts C and D and collected in the [language-agnostic bank](non-language-question-bank.md).

**Sources:**
- S1 - real interview, credit backend (49 min)
- S2 - mock 300-400k, Avito/Ozon questions
- S3 - interview at Lamoda (Go + DB)
- S4 - mock middle/senior 300+
- S5 - live coding "a cache in Go", middle+/senior
- S6 - mock Ozon 400-450k
- S7 - first-response-wins in a PostgreSQL cluster
- T1 - my own Go theory notes

**Legend:** 🪤 trap question / provocation (they mix you up on purpose) · ⚠️ sources disagree / an error in the source, fixed here · 📎 cross-reference or note · 💡 nuance · 🔧 code · ✍️ a question that wasn't in the exports · 🧠 mechanism - an explanation of "why it works this way" (not facts): read during prep when your answer isn't coming together.

**How to use:** cover the answer with your hand and say it out loud. The code from Part B - write it by hand. Terminology (JOIN, isolation levels, traps) - down to automatic recall: "floating" phrasing is noticeable in an interview (S3).

---

# 🪤 Trap map (a cheat sheet of "what they try to slip you")

| # | Provocation | Correct answer |
|---|---|---|
| 1 | "Write/read on a nil channel - panic?" | No: **blocks forever**. Only `close(nil)` panics [S2] |
| 2 | "Closed a channel a second time?" | **Panic** (a send on a closed one is also a panic; a read returns zero value + `ok=false`) [S2] |
| 3 | "Does append without returning the result modify the slice?" | No: the local copy of the header is modified. Accessing `s[1]` with len=1 is a **panic**, not garbage [S1] |
| 4 | "defer prints x=10, then x=20 - what will it output?" | **10**: arguments are evaluated at the moment the defer call is made [S2, S5] |
| 5 | "Does defer always run?" | No: on a hard power-off/kill nothing runs [S5] |
| 6 | "read committed → a dirty read is possible" | No: RC **excludes** dirty reads. The real threat to balances is a **lost update** [S3 ⚠️] |
| 7 | "cross join is an intersection" | No: a **Cartesian product** (every row × every row) [S3 ⚠️] |
| 8 | "vertical scaling is splitting by tables" | No: vertical = **adding resources to a single machine** (there is a ceiling). Splitting (replication/sharding/partitioning) is horizontal [S2 ⚠️] |
| 9 | "sync.Map is better for reads, map+mutex for writes" | Half-true: sync.Map wins in read-heavy workloads **with a stable key set** (write-once-read-many). The universal read-heavy choice - map+RWMutex; write-heavy - map+Mutex [S4 ⚠️, T1] |
| 10 | "context.WithSignal appeared in Go 1.21" | No such API **exists**. Signals into a context: `signal.NotifyContext` (os/signal, Go 1.16). Both the source and the candidate are wrong [S4 ⚠️] |
| 11 | "A semaphore on atomics" | An atomic **can count, but cannot wait** (there is no blocking). A semaphore = a buffered channel [S6] |
| 12 | "Goroutines in a loop print i: 0..4 in order?" | Before Go 1.22 - capture by reference (all print the last value). Since 1.22 - a separate variable per iteration. Without a WaitGroup **nothing** is printed [S6] |
| 13 | "A list of URLs → N goroutines per URL" | With 10,000+ you hit the **file descriptors** → worker pool [S1] |
| 14 | "Network I/O = file I/O" | No: network is the netpoller (the goroutine sleeps, **doesn't hold an OS thread**); a file is a blocking syscall (**holds the thread**) [S1] |
| 15 | "You can return nil from a function returning int" | **Doesn't compile**. For "empty" use zero value + an `ok` flag [S5] |
| 16 | "An index on a field with 2 values will speed up the query" | No: **low selectivity** - the index is useless [S1] |
| 17 | "Store money in a float" | Never: **only integer minimal units** (kopecks) [S3] |
| 18 | "append always grows ×2" | ×2 up to 256 elements (Go 1.18; before 1.18 the threshold was 1024), then a smooth transition to ×1.25 [S2 ⚠️, verified] |
| 19 | "The 200th cache item → slot 200" | 200 % 200 = **0** (slots 0..199) [S5] |
| 20 | "Two goroutines write responses into a shared slice → a mutex is needed" | If you write **by index** (one cell per URL) - no mutex needed, no races [S3] |
| 21 | "not found from a replica = the cluster is dead" | No: not found is **not a node failure**, retrying it is pointless, it's a separate branch [S7] |
| 22 | "Closing a channel = waking one goroutine" | Closing unblocks **all** readers at once, no matter how many there are [S6] |
| 23 | "Concurrency = parallelism" | No: concurrency is about **efficient CPU utilization while waiting (I/O bound)**; parallelism also exists on OS threads [S4] |
| 24 | "gRPC is faster because of encryption" | A slip of the tongue: it's faster because of the **binary protocol over HTTP/2** and the `.proto` contract [S2 ⚠️] |
| 25 | "HTTP/2 doesn't need parsing" | Both need parsing; the win is that binary frames **know their exact length** [S6 ⚠️] |

---

# Part A. Go core (theory)

## A1. Goroutines and the scheduler

> A1 vocabulary: **task** = goroutine (G), **worker** = OS thread (M), **workbench** = an execution slot (P - called "processor" in articles, but it is NOT a CPU; workbenches ≈ the number of cores). The 🧠 lines below use only these words.

**A1.1. How does a goroutine differ from an OS thread?** [S2, S4, S6, T1]
- Stack: a goroutine is ~2 KB and **grows on demand**; a thread is ~1 MB and **fixed** (many people forget this point - always mention it).
- Scheduler: for threads it's the OS kernel (syscall, ~1-10 µs to switch); for goroutines - the **runtime, userspace** (~200 ns).
- The key point: goroutines **still run on OS threads** (they are mapped onto them) - there are just many of them per thread.
- You can launch millions; there is no hard limit. You'll hit the **memory** wall (idle goroutines sit in memory) or the **CPU** wall (for computing ones - as many cores as you have, that's how many run) [S6].
- 🧠 **Mechanism:** a goroutine is not a kernel object but a record in process memory (stack pointer, program counter, state) - the runtime manages its life itself. The stack starts at ~2 KB on the heap and grows: on overflow the runtime allocates a new one and copies. An OS thread is a kernel object: `clone()` hands out a fixed ~1 MB stack; switching threads = entering the kernel and back (syscall, ~1-10 µs). Switching tasks inside the runtime is saving/restoring a few registers, no kernel involved (~200 ns). Hence "millions of goroutines, thousands of threads": a task is a record in memory, a thread is an expensive kernel object.

**A1.2. Concurrency vs parallelism? Why do we even need fast context switching?** [S4] 🪤
- Concurrency ≠ parallelism. Parallelism works perfectly well on OS threads (Java, Python without asyncio).
- Fast switching is needed **because of I/O bound** workloads: a thread sent a request (network/DB/file) and waits; with thousands of I/Os per second, switches at the OS-thread level are too expensive. Go is built around this problem. Analogues: coroutines, green threads, fibers.
- **I/O bound vs CPU bound**: CPU-bound - the processor computes all the time (a flat, high utilization graph); I/O-bound - rare spikes of "sent → waits → got a response".
- The classic failure: knowing the "surface" (what a goroutine is) without digging deeper. Whoever can tell this depth is "a head above the majority".
- 🧠 **Mechanism:** the whole point is waiting on I/O. A task waiting for a network response is parked: the worker sets it aside and in hundreds of nanoseconds picks up the next ready task. With OS threads, every I/O wait = a trip into the kernel + blocking + a context switch - thousands of those per second eat CPU. Go keeps many lightweight tasks on a small number of workers: while one waits on the network, others compute on the same core. Hence: CPU-bound work barely benefits from concurrency - it needs parallelism (the number of cores); fast switching matters exactly for I/O-bound work.

**A1.3. Preemption: has the scheduler's behavior changed?** [S2]
- Before Go 1.14 - cooperative multitasking: a goroutine decided for itself when to yield.
- Since Go 1.14 - **asynchronous preemption**: a runtime monitor; if a goroutine has been stuck longer than a threshold (~10 ms) and won't yield, it is forcibly taken off the thread.
- 🧠 **Mechanism:** before 1.14 the scheduler switched only at "safe points" (function call, channel, syscall) - cooperatively: an empty infinite loop with no calls would hog a core forever. Since 1.14 the system monitor (sysmon) checks roughly every ~10 ms whether a task has "hung" and initiates preemption: the thread is interrupted by a signal at the nearest safe point, and the scheduler inserts a switch. This is not a "kill": the state stays consistent, control is simply taken away.

**A1.4. The GMP model - what is the scheduler made of?** [S4, T1]
- **G** - goroutine; **M** - machine, an OS thread; **P** - processor: **not a physical CPU**, but an internal Go entity with a **local queue** of ready goroutines. The number of Ps = GOMAXPROCS.
- Queues: a local one (per P) + a **global** one (shared).
- Picking a goroutine: with probability **1/61** from the global queue → from the local one → if empty, **work stealing** from other local queues → netpoller.
- Blocking I/O: a G goes into a syscall on an M → the M detaches from its P → the P picks up another G; the returned G gets queued. The processor doesn't idle.
- 🧠 **Mechanism:** the letters are entities: G - a task (goroutine), M - a worker (OS thread), P - a workbench (a "workplace": a local queue of ready tasks + service resources; "processor" in articles is NOT a CPU). The invariant: a working worker has a workbench; without a workbench no task executes. There are ≈ as many workbenches as cores (GOMAXPROCS) - about 8 tasks execute at once, the rest sleep in queues. The workbench exists so there is no single global scheduler lock: each workbench spins its own queue, and when empty it **steals** tasks from its neighbors (work stealing) - auto-balancing. Why a worker detaches during a syscall: it has blocked in the kernel, and keeping the workbench attached to it means an idle workplace; the worker goes off to wait alone, and the workbench is picked up by another task on another worker. The returning syscall puts the task back into a queue - a free workbench/worker will be found.

**A1.5. The netpoller and sockets.** [S4, S1] 🪤
- The netpoller handles **network** requests: it wakes a goroutine (puts it in runnable) only when the **socket is ready**. Network I/O does not block an OS thread.
- **File I/O is a blocking syscall**: the goroutine holds an OS thread for the whole read. That's why 10,000 HTTP requests are fine, while 10,000 file reads will hit the OS-thread wall [S1].
- A socket is one of the kinds of **IPC** (inter-process communication) [S4].
- 🧠 **Mechanism:** network fds in Go are non-blocking (`O_NONBLOCK`): read/write instantly answer "no data". A task that lacks data leaves a note with the secretary (the netpoller): "wake me when a letter arrives in my socket" - and goes to sleep. The secretary listens to all sockets at once as a single list (epoll/kqueue at the runtime level); the kernel reports "socket ready" → the secretary puts the task into the ready queue. The worker wasn't sitting with it all this time - it was executing others. The analogy: a secretary with a switchboard instead of a hundred separate phones. Files can't do this (you can't register them with epoll): a file read is an honest blocking syscall, and the task sleeps in the kernel together with its worker. Hence: 10,000 HTTP requests are fine, 10,000 file reads will hit the OS-thread wall.

**A1.6. How many goroutines, and how to limit them?** [S6, S1, T1]
- CPU-bound → ≈ GOMAXPROCS (semaphore/worker pool); I/O-bound → hundreds/thousands, the limit = the throughput of the external system.
- 10,000 URLs → **worker pool** (file descriptors) [S1].
- Limiting parallelism: a semaphore on a channel, `errgroup.Group.SetLimit` [T1].
- 🧠 **Mechanism:** the limit is determined by the resource each task consumes. CPU-bound: there is useful work for exactly the number of cores - about GOMAXPROCS tasks execute at once, and any extra just causes switches (and cache misses: the "hot" stack gets evicted) → semaphore/worker pool ≈ GOMAXPROCS. I/O-bound: a task sleeps almost all the time and doesn't eat CPU → the limit is external: the external system's descriptors, RPS, connections. Technical ceilings: memory (≥ 2 KB per task: a million ≈ 2+ GB) and fds (ulimit - that's why 10,000 URLs require a worker pool).

**A1.7. The loop variable in goroutines.** [S6] 🪤
- Before Go 1.22 - capture by reference: all goroutines see the last value. **Since Go 1.22 there is a separate variable per iteration** - the problem went away on its own. You must know this.
- Bonus provocation: "why does the last iteration print first in ~90% of cases?" - the scheduler puts a freshly created goroutine into the **priority slot** (the data is "hot" in the CPU cache, starting it is cheaper). A rare question; an honest "I don't know" is fine [S6].
- 🧠 **Mechanism:** before 1.22 the loop variable was declared once for the whole loop (a single memory cell) - the closure captured its address, so all tasks read the same last value. Since 1.22 each iteration has its own variable (the equivalent of an implicit `i := i` in the body). "The last one prints first": a new task goes into **runnext** (the workbench's priority slot) - its stack and code are "hot" in the cache, and starting it immediately is cheaper than waiting for the end of the queue.

**A1.8. IPC - ways of inter-process communication.** [S4]
- Sockets, **files**, **signals**, pipes, semaphores. Be able to name at least **three** (files, signals, sockets) - they like to grill you like a senior even at the middle level. Go has signals (SIGTERM and others).
- 🧠 **Mechanism:** each process has its own address space (OS isolation) - "sharing" is only possible through an intermediary: a file in the filesystem, a socket/pipe through the kernel (buffering + wakeups), a signal (asynchronous notification, no data), a semaphore/message queue (System V/POSIX), or shared memory (mmap). A socket is system-level IPC: the network between machines works through it too. Go channels are synchronization inside a single process (a shared address space), so strictly speaking they are not IPC.

## A2. Channels and CSP

**A2.1. What kinds of channels exist, and what are they for?** [S2, T1]
- Purpose: passing data between goroutines + a synchronization primitive.
- **Unbuffered**: a send blocks until a reader takes the value (a synchronous hand-off).
- **Buffered**: you write until the buffer fills up. A buffer is not added without a reason; it fits worker pools, prefetching, and cases where "handing off fast matters more" [S4].
- 🧠 **Mechanism:** a channel is a runtime structure (hchan): a ring buffer + a mutex + lists of waiting senders/receivers. Unbuffered: there is no buffer, so a send **does not complete until the receiver takes the value** - either immediately (the receiver is already waiting), or the sender gets parked in the wait list. Buffered: a send puts the value into the buffer and completes; it only parks when the buffer is full. In other words, a channel combines a queue and blocking for free - which is why semaphores and worker pools are built on channels.

**A2.2. Nil channel, closed channel, double close.** [S2] 🪤
- A nil channel: **send and receive block forever** (a classic deadlock). **Only closing** a nil channel panics. The candidate in S2 got it wrong ("panic on write") - that's the trap.
- Reading from a closed buffered channel: first the remainder of the buffer, then **zero value + `ok=false`**.
- A send on a closed channel - panic. Closing an already closed one - panic.
- Only the sender closes; with multiple senders, don't close at all (sync.Once / context) [T1].
- 🧠 **Mechanism:** a nil channel is a null pointer to hchan: there is nothing to dereference, so by design the runtime blocks the operation forever (this is also the trick for "disabling" a select branch). `close(nil)` dereferences nil → panic. Closing changes the hchan state (the closed flag) and wakes everyone waiting: receivers drain the buffer, then get zero value + `ok=false`. A send after close violates the invariant "after close there will be no data" and is caught by a panic: the sender may not know about the close (a close/send race), and the panic protects against silent corruption.

**A2.3. select: order of the branches?** [S4, T1]
- **Random** (picked at random), although it looks syntactically like switch. `default` makes it non-blocking. A branch with a nil channel is never ready - that's how a branch is "disabled".
- 🧠 **Mechanism:** select gathers all the operations and checks each for readiness "right now" (send: there is room in the buffer or a waiting receiver; receive: there is data or the channel is closed). If several are ready → a **random** pick: deliberately, so that the programmer cannot and should not rely on the order (otherwise everyone would write code for their "favorite" branch). If none are ready → it registers in the wait queues of all the channels and parks; the first channel to free up wakes it. `default` is "none are ready → move on".

**A2.4. A buffered channel saves you from two classic problems.** [S4, S7]
- **The "reader upstream" deadlock**: the reading goroutine waits before the writer has put anything in - even a buffer of size 1 solves it [S4].
- **Goroutine leaks**: with first-response-wins, an unbuffered channel means the remaining goroutines hang on the send forever (an unbuffered channel with no reader blocks). A buffer of `len(replicas)` - nobody ends up blocked, and the GC removes the unread values [S7].
- 🧠 **Mechanism:** both cases follow from one thing: an unbuffered send requires **a ready receiver at the same moment**. The "reader upstream" deadlock: the reader parked on receive before the writer started writing, and the writer with no buffer waits for another reader - nobody moves. A buffer of size 1 gives the phases asymmetry: the send completes even if the reader isn't ready yet. The leak: after the first response the reader is gone, while K−1 goroutines are still sending - there is no unbuffered handler anymore, nobody will receive → they hang forever. A buffer sized for all N: every send completes (there is room), the goroutines finish, and the GC collects the unread values.

**A2.5. Waking all readers without knowing how many there are.** [S6] 🪤
- **Closing the channel** unblocks all readers at once (a thousand, doesn't matter). An idiomatic "go/stop everyone" signal. Managing goroutines: a signal channel (pause/resume), a stop channel in select (shut down individually), context (shut down everyone at once when the service stops) [S6, S6 Q12].
- 🧠 **Mechanism:** close doesn't "send a message", it changes the hchan state and **takes all** receivers off the wait list at once (each gets zero + `ok=false`). That's why it's a natural broadcast: you don't need to know the number of readers. A regular send is point-to-point (it wakes one). A WaitGroup is no good as a signal: it waits for completion, it doesn't "wake". context works the same way internally: cancellation = closing the internal channel (done), and everyone listening on `ctx.Done()` wakes up.

**A2.6. CSP - the synchronization concept in Go.** [S4]
- The classic approach: shared memory + mutexes. Go is **CSP**: "do not communicate by sharing memory - **share memory by communicating**". Three parts: goroutines, channels, select (random choice). The theorem was formulated long before Go. A frequent "growth point" - learn the phrasing.
- **WaitGroup vs channels**: a WaitGroup is just waiting for completion; channels are for collecting results or a cancellation channel. You can replace a WaitGroup with channels (counting completion signals) [S2].
- 🧠 **Mechanism:** in the shared-memory model, synchronization = locking around a shared structure ("whoever grabbed the mutex first owns it"). In CSP there is no shared mutable state: data is passed as messages, and the receiver **solely owns** what it received - there is no need to synchronize access. A channel is the only point of connection, and the transfer itself is the synchronization (the send completed = the receiver took it). Hence the phrasing: "don't communicate through shared memory - share memory through communication", i.e. instead of guarding a shared variable, hand it over to its new owner.

**A2.7. Fan-in: merging N channels into one.** [S6] 🔧
- One goroutine per input channel reads and writes into the shared output. A "conductor" goroutine: `wg.Wait()` → `close(out)`. A micro-optimization: `wg.Add(len(chs))` once instead of Add in a loop (the number of channels is known in advance) [S6, S7].
- **A channel send with a timeout**: `select { case ch <- v: ... case <-time.After(1*time.Second): ... }` - so you don't hang forever if there is no one to read [S6].
- 🧠 **Mechanism:** fan-in = N→1: each input channel is read by its own goroutine (why not one select for all: the set of channels may be dynamic, while select branches are fixed; one goroutine per channel scales). All of them write to a single output. Who closes: the "conductor" waits for `wg.Wait()` - once all readers finish, nobody writes anymore, and only then `close(out)`; closing earlier = a send on a closed channel (panic), not closing = the consumers of the output channel hang forever. A timeout on the send: if the consumer is slow, the producer must not block indefinitely.

## A3. Synchronization

**A3.1. The full set of primitives.** [S2]
- **Atomics** (sync/atomic) - counters; the typed `atomic.Int/atomic.Bool` - since Go 1.19; faster than a mutex, but they cover only simple operations [S2, T1].
- **Mutex** - exclusive access; **RWMutex** - parallel reads + exclusive writes (when there are noticeably more readers).
- **WaitGroup** - waiting for goroutines (Add - before starting, defer Done, Wait at the end); **Once** - one-time initialization; **Cond** - complex wait-signal conditions. Reciting the full set (Once/Cond are often forgotten) is a strong answer.
- **Semaphore** - there is no built-in one in Go; it's implemented via a channel (or an atomic counter, but see the trap below).
- 🧠 **Mechanism:** the primitives cover different problems. An atomic - a single operation on a word of memory without locking (in hardware, a CPU lock instruction). A Mutex - a critical section of several operations: locking = parking in a wait queue. An RWMutex - parallel reads: an atomic reader counter + a separate waiting writer. A WaitGroup - a counter of unfinished tasks + "wake when it hits zero". Once - an atomic flag on the fast path + a mutex on the slow one. Cond - condition-based parking with the mutex released (Wait) and targeted/mass wakeups (Signal/Broadcast). A semaphore = "count and wait" - exactly what a channel gives out of the box.

**A3.2. An atomic ≠ a semaphore.** [S6] 🪤
- An atomic is just a counter: it can increment safely, but **it cannot wait**. A semaphore must both count and **stop** a goroutine once the limit is exhausted.
- A semaphore done "properly": **a buffered channel with n slots** - a send = taking a slot (the goroutine waits by itself), a receive from the channel = releasing it. A channel gives both the counter and the blocking for free [S6, S3].
- 🧠 **Mechanism:** an atomic can only change a value atomically (add/fetch) - it physically has no "wait": once the limit is exhausted it will keep incrementing, nobody stops. A semaphore must be able to **block** a goroutine when the limit is exhausted. A buffered channel: a send to a full buffer parks the sender (that is exactly waiting for a slot = throttle), a receive frees a slot and wakes a waiting sender. The buffer slots are the counter; the parking/waking is the blocking. Both halves are already implemented in the channel runtime.

**A3.3. Worker pool vs semaphore ("horses and gates").** [S4]
- **Worker pool** - we limit the number of executors: N workers take tasks from a channel. **Semaphore** - there are a million executors, but **no more than K pass at the same time** (throughput). The interviewer's analogy, and he loves it in real interviews.
- Worker pool in code: `go func(){ for job := range jobs {...} }()` × N [S4].
- Dynamic pool scaling: expansion - launch new workers as the queue grows; contraction - a stop channel in select (individually) or context (all at once) [S6].
- 🧠 **Mechanism:** the difference is in who waits. Worker pool: N goroutines spin `for job := range jobs` forever; tasks wait in the channel queue until a worker frees up; the limit is on **executors** (streams of execution), whose count is fixed. Semaphore: there can be a million executors (one goroutine per task), but at any moment **≤ K are active**: each one sends into the semaphore channel before working (taking a slot) and receives after (releasing it); the rest are stuck on the send and wait. The limit is on **throughput** (how many tasks run at once), not on the number of execution streams.

**A3.4. RWMutex vs Mutex.** [S2, S4, S5]
- RWMutex: **reads do not block each other**; a write waits for all readers and blocks everything. "RWMutex performs worse on writes" - if it were better everywhere, Mutex would have been removed [S4].
- Choosing: read-heavy → RWMutex; reads ≈ writes → Mutex (simpler API); writes more often → Mutex [S5].
- 🧠 **Mechanism:** inside an RWMutex: an atomic reader counter + a writer flag/queue. RLock atomically increments the counter (new RLocks pass while there is no waiting writer); RUnlock decrements it. Lock marks "writers are waiting" (new RLocks queue up), then waits until the reader counter hits 0; Unlock wakes everyone waiting. Net result: reads share one mutex hold between themselves (they are parallel), but every RLock/RUnlock is an atomic operation + memory barriers: a bit more expensive per operation than a plain Mutex. So the win only comes when readers noticeably outnumber writers; on writes an RWMutex is always worse (it waits for all readers).

**A3.5. sync.Map vs map + mutex.** [S4, T1] 🪤
- The candidate in S4 mixed it up - don't repeat it. Correct: `sync.Map` is optimized for **read-heavy work with a stable key set** (write-once-read-many, disjoint key sets); a plain **map + RWMutex** is the universal option for read-heavy; **map + Mutex** for write-heavy. `sync.Map` is not a "replacement everywhere".
- 🧠 **Mechanism:** sync.Map keeps two levels: a read map (reads via atomics, no locking at all) and a dirty map (under a mutex). If a read hits the read map → no locking at all; a miss → take the mutex, refresh read from dirty (migration). Hence the winning conditions: keys are stable (the read layer always hits), writes are rare. Constant insertion of new keys → permanent misses and read/dirty migrations - more expensive than map+Mutex. A map+RWMutex reads in parallel without copying data; Lock writes exclusively.

**A3.6. Data race.** [S2, T1]
- Definition: ≥2 goroutines access the same memory, and **at least one of them writes**. Consequences: overwrites (a "broken" counter).
- Detection: `go test -race` (it even finds bugs in the stdlib). Protection: synchronization primitives + running with `-race`.
- The difference from a race condition: a data race is unsynchronized access (caught by the detector); a race condition is a logical ordering bug (not caught) [T1].
- 🧠 **Mechanism:** a data race is about memory: two goroutines in parallel on different cores, and without synchronization there is neither ordering nor visibility (each works with its own cache); a write can get "lost" - what remains is a mixed/undefined state. The detector (go test -race) instruments every memory access and builds a happens-before graph: an unordered "write/write" or "write/read" pair on the same address → a report. A race condition is about logic: even without a data race, the result depends on the unpredictable order of events (two goroutines checked a flag and both proceeded) - the detector won't help; it's cured by design.

**A3.7. A goroutine stuck in a computation / how to "kill" a goroutine.** [S2, S6]
- Preemption (see A1.3). Control: a signal channel (an empty struct), closing a channel = to everyone, context = to everyone on cancellation. You cannot forcibly "kill" a goroutine from the outside - only cooperatively, through these mechanisms.
- 🧠 **Mechanism:** the runtime has no "kill a goroutine" API, and it cannot safely exist: yanking a goroutine out mid-computation means abandoning its stack, locked mutexes, and defers - you cannot roll back the state from the outside without corruption. So only cooperation: the goroutine itself checks a signal in select (`case <-stop: return`), close(stop) wakes all listeners (A2.5), context threads through the call stack (`ctx.Done()` is checked at every level). As for one stuck in a computation (not in I/O): preemption takes it off the core (A1.3) but doesn't cancel it - it returns control only once it reaches its own check; you still can't "kill" it from the outside.

## A4. Slices, maps, strings

**A4.1. Array vs slice; internals; append.** [S2, S1, T1] 🪤
- A slice = a header `{ptr, len, cap}` + an underlying array; an array has a fixed length.
- Append: if cap allows - write into the same array; if not - allocate a new one: **×2 up to 256 elements**, then a smooth transition to ×1.25 (Go 1.18; before 1.18 the threshold was 1024 and growth was strictly ×1.25; in S2 the interviewer said "from 128 to 256" - imprecise) ⚠️.
- **The "what will it print" task** [S1]: `make([]int, 1, 2)` → `[0]`; `appendOne(nums)` without returning the result **doesn't change the outer slice** (a local copy of the header) → `[0]` again; `nums[1]` with len=1 → **panic index out of range**. A classic comprehension check.
- Two slices over one array (via `s[1:3]`): their len/cap diverge, and an append to one can modify the "other" slice until reallocation. The cure: `copy` [S2, T1].

**A4.2. The map: internals and peculiarities.** [S2, T1]
- O(1) lookup/insert/delete (average). Under the hood: **buckets of 8 elements**, `tophash` (the high bits of the hash - fast filtering), keys and values are stored **separately** (alignment), collisions, incremental evacuation on growth (load factor > 6.5).
- A rare detail that makes you stand out: a map **does not shrink back** after growth - if it has ballooned, it's easier to recreate and copy it over.
- A concurrent write → `fatal error: concurrent map writes` (protection against corruption) [T1]. Iteration order is randomized on purpose [T1].
- **Set**: `map[T]struct{}` - an empty struct weighs 0 bytes [S2].

**A4.3. Strings, runes, bytes.** [S3]
- `s[i]` returns a **byte** (part of a character; for Cyrillic it's "garbage" ASCII). A rune is a Unicode code point (it can take up several bytes). To work with characters - `[]rune(s)`.
- A case-insensitive palindrome: compare via `strings.ToLower`, but return the word **as it was** in the output (we preserve the output's case) [S3].
- 🧠 **Mechanism:** a string in Go is an immutable sequence of bytes (essentially a `[]byte` you can't write to): there are no "characters" inside, only UTF-8 encoding. All "transformations" are a choice of how to read the bytes: `s[i]` is a raw byte **without decoding** (for Cyrillic, a piece of a 2-byte character); `for i, r := range s` is a UTF-8 decoder, `r` is a rune (int32, a code point), `i` is the byte offset (a rune can take 1-4 bytes); `[]rune(s)` decodes everything at once into a slice of code points; `[]byte(s)` copies the bytes; `string(r)`, `string(b)` do the reverse encoding. `len(s)` = the number of **bytes**, not characters.
- In an interview: frequency counting via `s[i]` is correct only for ASCII input; to the question "what if it's Unicode?" → `range` + `map[rune]int` (or `[]rune(s)`). This continues in A4.5.

**A4.4. Zipping two slices.** [S6] 🔧
- The edge case is different lengths: iterate up to `min(len(a), len(b))` (going out of bounds is a panic). `make([][2]int, 0, minLen)` - reserve the capacity up front.

**A4.5. Map syntax: counting idioms.** [✍️] 🔧
- `map` is a reserved name: the type is written `map[K]V`. To create one for writing - `m := make(map[K]V)`; with data - `m := map[string]int{"a": 1}`. A nil map: you can read (`m[k]` → zero value), **you cannot write - panic**.
- **Reading:** `v := m[k]` returns the value's zero value if the key is missing. If you need to know whether the key exists → `v, ok := m[k]`: `ok` is a `bool`, checked with `if !ok`, **not `ok != nil`**.
- **A counter without an existence check:** `m[k]++` works for a missing key (0 → 1). This is the frequency-counting idiom - no need for `_, ok := m[k]` first.
- **Set:** `map[T]struct{}`, inserting with `s[k] = struct{}{}` - **two things**: the type `struct{}` and the literal `struct{}{}` (an empty struct weighs 0 bytes); checking with `_, ok := s[k]`.
- **Comparing two maps:** `maps.Equal(a, b)` (the `maps` package, Go 1.21+) is the simplest path; manually: first `len(a) != len(b)` → false, then `for k, v := range a { if b[k] != v { return false } }` (a second pass is not needed: equal lengths + all of a's keys match).
- **Iterating:** `for k, v := range m` gives the **(key, value)** pair in that order. Keys only → `for k := range m` (the only form of range with a single argument). **The mistake from 08.09:** `for _, letter := range m` - `letter` got the **counter** (the value) and the character key itself was thrown away → `m[byte(letter)]` addresses entries by the counter number, both maps yield 0 → the comparison is "always equal" → true on a non-anagram.
- 🧠 **Mechanism:** a map is a hash table (buckets of 8, A4.2): indexing `m[k]` means hashing the key and searching a bucket, not "taking by index". A missing key on read is not an error - you get the value type's zero value (otherwise you'd have to panic on every miss). The increment `m[k]++` = read-modify-write: read the zero value (0) → wrote 1. Comparing maps manually via range is because a map has no `==` operator (other than comparison with nil): the order and type of keys are arbitrary, so comparison is only possible element-wise. range returns (key, value) pairs because a map entry is addressed by key: a value without its key cannot be matched to an entry; `for k := range m` is shorthand when only the key is needed.
**A4.6. Map keys: comparable types; an array as a signature.** [✍️] 🔧
- A map key can be **any comparable type** (`==`/`!=` are defined): bool, numbers (int/uint/float/complex), string, pointers, channels, interfaces, **fixed-length arrays** (`[26]int`, `[2]string`), **structs of comparable fields**. **Not allowed**: slices, maps, functions - a compile error `invalid map key type`.
- The interface gotcha: `map[any]V` compiles, but panics at runtime if the key's dynamic type is not comparable: `m[[]int{1,2}] = v` → "hash of unhashable type []int". A struct or array with an uncomparable field (a slice inside) - also not allowed, that's a compile error.
- A pointer / channel as a key compares **by identity** (the address), not by value: `&x` and `&y` with identical contents are different keys.
- The "canonical key" trick (Group Anagrams): `key := [26]int{}`; `for _, letter := range str { key[letter-'a']++ }`; for anagrams the counters match **by value** → one map entry. This is a signature of the string: identical signatures = the same set of letters.
- The sorting variant (sorted letters → a string key) solves the same problem **without** any Go-specific trick: it works in any language, O(n·k log k) - valid in an interview; `[26]int` - linear O(n·k), for ASCII lowercase (Unicode is wider - A4.3).
- 🧠 **Mechanism:** a map finds an entry by the **hash of the key**, so the key must be hashable and comparable: `==` and the hash are determined by value. An array: the length is fixed, the hash and `==` run over all elements, two equal arrays are one key. A slice: the header `{ptr, len, cap}` with mutable contents, value equality is not defined (slices can only be compared with nil) - that's why it's not allowed as a key. The letter counter is a minimal immutable digest of a string: identical counters = one signature.

## A5. Errors, panic, defer

**A5.1. Error handling.** [S2, T1]
- Functions return a value + an error; handle it right after the call. `errors.Is` - comparison along the chain of wrappers (sentinel errors); `errors.As` - extract a concrete type. `fmt.Errorf("...: %w", err)` - wrapping. Don't log and return at the same time - either one or the other.

**A5.2. Error vs panic; recover.** [S2]
- An error is incorrect behavior that **can be handled** (respond to the user, retry). A panic means further execution is unsafe/impossible, the "end point". Go's principle: **don't panic**, panics are rare.
- `recover()` works only inside a deferred function (and in the same goroutine). Middleware catches a panic that bubbled up from the layers - the service keeps running [S2, T1].

**A5.3. defer - all the nuances.** [S2, S5, S6] 🪤
- Purpose: guaranteed closing of a resource / unlocking of a mutex **on function exit**, even on a panic (the key word is "guaranteed").
- **Arguments are evaluated at the moment the defer statement is called**, not when it executes: x=10, `defer print(x)`, x=20 → prints **10** [S2].
- Multiple defers are **LIFO** (the last one written runs first) [S5].
- Don't use it in loops (the overhead accumulates in hot loops; in older versions it was more expensive, now it's almost free) [S6].
- Downsides: worse readability - the call is in one place, the execution at the end of the function [S6].
- "Is there a case when defer doesn't run?" - a **hard machine shutdown** (the power was cut). A provocation question during live coding [S5] 🪤.

**A5.4. Zero value vs nil.** [S5] 🪤
- `return nil` for a primitive (int) **doesn't compile**. For "no key" use `item{}, false`. If the value itself can be nil - "empty" is indistinguishable from "no key", so pass a flag up a level.

## A6. Pointers, types, OOP

**A6.1. Pointers.** [S4]
- A pointer is a variable holding a memory address. **There is no separate "reference" abstraction in Go** (unlike C++).
- Size: **8 bytes on a 64-bit** system, 4 on a 32-bit one (a question half of the candidates trip on).
- Why: (1) access to one object from different places without copying; (2) memory savings. Name both arguments.

**A6.2. Argument passing.** [S2]
- In Go **everything is copied** (pass by value; a pointer is also copied as a value). Receiver: value - doesn't mutate the original; pointer - mutates, saves copying [T1].

**A6.3. OOP in Go.** [S2]
- Break it into the three pillars (a format that lands well): **encapsulation** - the case of the first letter (exported or not); **inheritance** - embedding (composition, not "is-a"); **polymorphism** - interfaces + **duck typing** (implemented the methods = implemented the interface).

**A6.4. Compile-time check that an interface is implemented.** [S2]
```go
var _ SomeInterface = (*SomeStruct)(nil)
```
- If it doesn't implement it → a build error. Useful in a codebase; the candidate couldn't reproduce it from memory - it didn't count against him, but learn it.

**A6.5. First-class functions.** [S2] - functions can be passed as an argument and returned. Methods are defined with a `*T`/`T` receiver.

**A6.6. The interface canon: io.Reader / io.Writer.** [✍️]
- `io.Reader` = "can emit a stream of bytes": its only method is `Read(p []byte) (n int, err error)`. `io.Writer` = "can accept one": `Write(p []byte) (n int, err error)`.
- Implementation is **implicit** (duck typing): strings.Reader, bytes.Buffer/bytes.Reader, os.File, resp.Body, a network connection - all of these are io.Reader. A matching method is enough - there are no `implements` declarations.
- `io.EOF` is the normal end of a stream, **not an error** (handle it with `errors.Is(err, io.EOF)`).
- Accept the narrowest interface you need (Reader, not *bytes.Buffer): testability + reuse.
- Hence: in `http.NewRequestWithContext` the body is accepted as an io.Reader precisely for that reason - any source of bytes, including streaming a large file without loading it into memory.

## A7. Context

**A7.1. Role and types.** [S2, S4, T1]
- The main use cases: **graceful shutdown** (shut down cleanly on a signal), cancellation, timeouts, passing request-scoped values (**a trace ID - yes**; business parameters of functions - no, that's not accepted practice) [S2, S4].
- Kinds: `WithCancel`, `WithTimeout`/`WithDeadline`, `WithValue`; `context.Background()` at the root; pass it as the first argument, don't store it in structs [T1].
- Signals into a context: **`signal.NotifyContext`** (os/signal, Go 1.16) 🪤 - "context.WithSignal" **does not exist** (in S4 both the candidate and the analysis got this wrong) [S4 ⚠️, verified].

**A7.2. Time-bounding a request.** [S2]
- HTTP: a timeout at the client level **or** `context.WithTimeout` + `http.NewRequestWithContext`.
- Choosing the value: **p99 of the call time + a small margin**, but capped from above by the **service SLA** (if a response takes >3 sec, that's already an incident, so a timeout above 3 sec is pointless). A default to aim at is ~10 sec [S2].

**A7.3. Context in concurrent code.** [S7]
- Before a retry, check `ctx.Err() != nil` - shorter than a whole select on `ctx.Done()` (the candidate didn't know this - remember it).
- On the first successful response - **`cancel()`** (stops the other goroutines' retries).

**A7.4. http.NewRequestWithContext: the signature and the body (io.Reader).** [✍️]
- Signature: `http.NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)` - **the third parameter is an `io.Reader`, a stream of bytes, not a string/[]byte**.
- `nil` = no body (GET). Body present (POST/PUT): `bytes.NewBuffer(json)`, `strings.NewReader(...)`, `os.File`, `resp.Body`.
- A live-coding reminder: in B1/B2 you already wrote `..., url, nil)` - nil works, but you have to be able to name the parameter's type and what it means.

## A8. GC, tooling, the downsides of Go

**A8.1. GC.** [S4, T1] - stack and heap; the GC cleans the heap (mark-and-sweep, tri-color, concurrent, low pauses). You don't have to explain the tri-color algorithm - they'll stop you. Avoid unnecessary allocations (value receivers, buffer reuse).

**A8.2. Tooling ("general awareness" questions).** [S4] - there is no right answer, they're checking experience:
- Linter: **golangci-lint**. Routing: **chi**. Logging: the platform logger/zerolog. Tests: **testify**; mocks: **gomock/minimock** (generation - mockgen via go:generate), the interviewer likes mockery - "any colour works".
- DB: **pgx** (raw SQL); sqlx is ok, but boilerplate; sql builders (squirrel) - nobody likes them: the code is harder to read and debug than a plain SELECT with two JOINs.

**A8.3. The downsides of Go.** [S4]
- Verbose error handling/boilerplate (the candidate honestly added: it's also a plus - it's explicit).
- Frameworks are not accepted → projects are **completely different** (hand-rolled, someone does hexagonal, somewhere an internal/ with 100 files) → **a long onboarding**. In big tech (Ozon/Avito) microservice generators save the day.

---

# Part B. Concurrency tasks and live coding

## B1. Parallel HTTP requests + status codes [S1, S2] 🔧

**The task:** a list of URLs, requests in parallel, print the status codes; don't ignore errors. [S2]

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
var wg sync.WaitGroup
for _, u := range urls {
    wg.Add(1)
    go func(url string) {          // url as an argument - not a loop-variable capture!
        defer wg.Done()
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        if err != nil { log.Printf("create request: %v", err); return }
        resp, err := http.DefaultClient.Do(req)
        if err != nil { log.Printf("do request: %v", err); return }
        defer resp.Body.Close()
        fmt.Println(url, resp.StatusCode)   // what the candidate forgot - printing the status
    }(u)
}
wg.Wait()
```
- The candidate's shortfall in S2: he didn't print the status code. Verdict: "a good solution, it even scales".
- Two URLs → you can go "head-on": two goroutines + a WaitGroup (S1: the candidate fumbled with the syntax - the interviewer: "don't overcomplicate it").
- 10,000 URLs → **worker pool** (descriptors!) [S1].
- A channels-based alternative: goroutines send statuses+errors into a channel, a separate reader goroutine prints; close the channel after all finish [S2].

## B2. In parallel, but no more than K; responses in URL order [S3] 🔧

```go
func fetchCodes(urls []string, k int) []int {
    codes := make([]int, len(urls)) // writing by index - order preserved, NO mutex needed
    sem := make(chan struct{}, k)   // semaphore: at most k at a time
    var wg sync.WaitGroup
    for i, u := range urls {
        wg.Add(1)
        go func(idx int, url string) {
            defer wg.Done()
            sem <- struct{}{}          // took a slot
            defer func() { <-sem }()   // released it
            resp, err := http.Get(url)
            if err != nil { codes[idx] = 0; return }
            codes[idx] = resp.StatusCode
            resp.Body.Close()
        }(i, u)
    }
    wg.Wait()
    return codes
}
```
- **The key insight**: writing by index (each URL has its own cell) - no races, no mutex needed. The candidate suggested a mutex - it works, but the interviewer was waiting for index-based writing [S3].
- **A typical muscle-memory mistake** (02.09, a repeat from scratch): at the end of the function `wg.Done()` instead of `wg.Wait()` - the muscle memory of the "defer wg.Done()" pattern kicks in and overrides the "wait for completion before return" node. Cured by self-review: "Done/Wait: where do I finish a goroutine, where do I wait for everyone?".
- **Hardening the canon** (02.09): the canon above uses a bare `http.Get` - in an interview it's better to do as in B1: `http.NewRequestWithContext(ctx, …, nil)` with a 10 sec timeout: the context cancels hung goroutines instead of waiting for them forever.
- 10 million URLs: the resources will hold, but with a small K the time grows proportionally (batches of K) [S3].
- An alternative: a worker pool with a fixed number of workers - the downside: you have to synchronize the response order (solved by writing by index) [S3].

## B3. A cache in Go (a full live-coding case) [S5] 🔧

**The spec is revealed stage by stage:** "the service is slow" → "a 20-million-row table" → "no Redis, no ClickHouse" → "only Go code and the DB" → the candidate formulates it himself: **a hot cache = a map + RWMutex**. Seniority = you set the spec for yourself.

**The mistakes they caught the candidate making (learn them - they're all traps):**
1. Forgot `make(map...)` in the constructor - the map is not initialized.
2. Forgot `return` in Get ("sure you didn't lose anything?").
3. Didn't handle the map's second result (`ok`) - the client gets a zero value instead of "no key".
4. `return nil` for a primitive type - doesn't compile (zero value + ok).
5. Put an `int` into main instead of an `item` (confusion about the value type).
6. `200 % 200 = 0`, not 1 - modular arithmetic.

```go
type item struct { value int }
type cache struct {
    mu   sync.RWMutex
    data map[int]item
    count int
}
func newCache() *cache { return &cache{data: make(map[int]item)} } // make is mandatory

func (c *cache) Set(key int, it item) {
    c.mu.Lock(); defer c.mu.Unlock()
    c.data[key] = it
}
func (c *cache) Get(key int) (item, bool) {
    c.mu.RLock(); defer c.mu.RUnlock()
    v, ok := c.data[key]
    if !ok { return item{}, false }
    return v, true
}
```
- Convention: Go **doesn't use Get/Set prefixes** in method names.
- **A 200-element limit**: a counter + `slot := count % 200`, keys 0..199 - a ring buffer, **O(1) per insert**. The interviewer suggested "find the oldest and delete it" (O(n)) - the candidate turned out to be more right than the interviewer, that's fine: a middle/senior interview is a discussion, not a dictation.
- The RWMutex choice: read-heavy; if writes were more frequent - a plain Mutex (simpler and faster).
- The main principle: a map touched from different goroutines - only under a lock. Talk the solution through out loud - in 2026, silence = suspicion of a neural net/Googling.
- Before the code - a plan out loud, after the code - run it in main (ideally a unit test).

## B4. A distributed request: first-response-wins [S7] 🔧

**The task:** a PostgreSQL cluster, synchronous replication (the data is identical everywhere). We go to all replicas concurrently and wait for the first one. There is a sentinel not-found error (no data does not mean the replica is dead).

**The key decisions they asked to justify:**
1. **`errors.Is` instead of `==`** - not-found may be wrapped [S7, S2 on errors.Is in general].
2. **A buffered channel of `len(replicas)`** - without a buffer: the first goroutine wrote, we returned the response, and the remaining 4-5 **hang forever** on the send (a goroutine leak). With a buffer they write and the GC cleans up.
3. **`wg.Add(len(replicas))` once** - the number is known in advance, removing the overhead of atomic increments in the loop.
4. **Retries - only for node failures, not for not-found** (retrying not-found is pointless - "you're looking for Vasya, and Vasya is nowhere to be found"; for a replica's non-not-found errors - `continue`, not `return`).
5. **We don't write errors from unreachable replicas into the channel** (we log and skip them) - otherwise they can't be told apart from a result.
6. **`cancel()` on the first successful response** - stops the retries of the other goroutines. The check before a retry - `ctx.Err() != nil` (shorter than select).
7. **A timeout for the whole request** (7-8 sec): select on `<-respCh` / `<-ctx.Done()` → a separate meaningful error `ErrTimeout`, not "cluster unavailable".
8. **not-found ≠ "the cluster is dead"** 🪤: if all replicas returned not-found (the channel is closed by the conductor after wg.Wait), we return not-found, and "cluster unavailable" - only when there is no answer at all.

**Interviewer feedback (what moves the needle):** (1) clarify the requirements before coding (number of replicas, timeouts) - most interviewers give you a plus for it; (2) run the primitive case before saying "I'm done"; (3) decompose functions over 50 lines.

## B5. Reversing words, don't touch palindromes [S3] 🔧

```go
func process(s string) string {
    words := strings.Split(s, " ")
    for i, w := range words {
        if isPalindrome(w) { continue }
        words[i] = reverseWord(w)
    }
    return strings.Join(words, " ")
}
```
- Palindrome - case-insensitive (`strings.ToLower`), and in the output we return the word **as it was** ("Ара" → "Ара").
- Cyrillic → we work with `[]rune`; `s[i]` is a byte.
- Complexity: the naive "reverse a word and search for it in the input" - **O(n²)**; via Split - **O(n)**. Computing the complexity yourself is half the solution - interviewers value that.
- A typical beginner's mistake: "walking over the string" ≠ "reversing the string" (copied in the same order).

## B6. The "what will it print" task: goroutines in a loop [S6]

- Without a WaitGroup - **nothing** (main finishes before the goroutines).
- With a WaitGroup - the numbers 0..4 in **random order**.
- Before Go 1.22 - all print the last value (capture by reference); since 1.22 - fixed.
- In ~90% of cases the four prints first - the priority slot of a fresh goroutine (see A1.7).

## B7. Code review: the Payments service [S1]

**What to look for (the candidate's + the interviewer's comments):**
1. Unhandled errors (in several places).
2. Resources (client/connection/response body) - close them via **defer**; without it, an early return leaks.
3. Don't hard-depend on exactly **200** - correctly you wait for 2xx (201, 202…) (debatable, a matter of taste).
4. Too generic an error ("payment system error") - **log the status and the response body**.
5. "Empty string = bad" - it may be a valid response; an explicit decision is needed.
6. Add **retries** (the next question).
- A behavioral nuance: in a real review, comments are written more delicately than in an interview.
- **A TCP client and an unstable network**: retries with a **growing pause** (exponential backoff) between attempts [S1, S7: a fixed pause is also acceptable, chosen empirically].

---

# Part C. Databases (not Go - a separate category) 🧭

## C1. SQL: basic constructs

**C1.1. WHERE vs HAVING.** [S2, S3, S4] - WHERE filters **rows before grouping**; HAVING - **groups after GROUP BY** (in S3: it can also be used without grouping like WHERE, but its typical place is after). A mnemonic phrasing: "WHERE - a filter on rows, HAVING - a filter on groups".

**C1.2. Kinds of JOIN.** [S3] 🪤
- **Inner** - only matching rows. **Left/Right** - all rows of one side, NULL on the other. **Full outer** - all rows of both. **Cross** - a **Cartesian product** (every × every), not an "intersection" (the candidate got confused - get the definitions down to automatic recall).
- Users without carts: an inner join won't return them; a left join → NULL for the cart [S3].

**C1.3. Top-N: the standard skeleton (to the point of automatic recall!).** [S3, S6]
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
- The skeleton: **JOIN → WHERE → GROUP BY → HAVING → ORDER BY DESC → LIMIT** [S3, S6].
- Aggregates: SUM, MIN, MAX, COUNT; `string_agg(title, ', ')` - listing with a separator (PostgreSQL; S3: the candidate wasn't sure of the name - refresh it) [S3, S2].
- Everything in SELECT with a GROUP BY must be either grouped or an aggregate [S2].

## C2. Indexes

**C2.1. Why they exist and how they're built.** [S2] - a separate structure (most often a B-tree; faster than a sequential scan); the search condition must **match the index**, otherwise it won't kick in; an index on 2 values is useless (see C2.2).

**C2.2. Selectivity.** [S1] 🪤 - "added an index for the query - EXPLAIN shows no speedup". The reason: **low selectivity** (a field with two values: "we'll cut it in half" - the scan is still large). You must be able to name the term (the candidate couldn't - a minus). B-tree: values are ordered; on low-selectivity fields the gain is minimal. Verdict: an index is effective with high selectivity.

**C2.3. When indexes hurt.** [S2, S4]
- Not used - it takes up space for nothing.
- Slows down **INSERT/UPDATE/DELETE** (recalculation).
- Small tables (1-10 thousand rows) - a seq scan is often faster.
- Low selectivity - useless.
- Migrations: an index on a large table - **`CREATE INDEX CONCURRENTLY`** (without locking writes); in the candidate's team non-concurrent indexes are banned - a good signal [S4].
- First understand which queries are executed - then the index (composite, selectivity) [S4].

**C2.4. "The DBA says something's wrong with your query" - where do you start?** [S3, S1] - **EXPLAIN**: the algorithms and order (e.g. merge join), whether indexes are used → rewrite the query around the index, or create/redo the index.

**C2.5. Vacuum (PostgreSQL).** [S1] - cleaning up dead row versions (MVCC). A general understanding is enough.

## C3. Transactions, isolation, locks

**C3.1. Isolation levels.** [S1, S3]
- **Read Committed** - the default in PostgreSQL; only what others committed is visible. ⚠️ A correction to S1: "read uncommitted only in MySQL" - imprecise: in PostgreSQL READ UNCOMMITTED **is treated as READ COMMITTED** (it isn't supported separately).
- **Repeatable Read** - a snapshot taken at the moment the transaction starts; protects against dirty reads and non-repeatable reads; **phantom rows remain** (S1: the candidate slipped in terminology, but the gist is right).
- **Serializable** - the maximum, expensive.

**C3.2. Debiting a balance: a task with a catch.** [S3] 🪤
- Problem #1: **float for money is a mistake** - only integer minimal units (kopecks: 2350 ₽ = 235000).
- Problem #2: read committed + two parallel debits → a **lost update**: both read the balance and each wrote its own. This is NOT a "dirty read" (RC actually excludes it) - the candidate in S3 got the term wrong, the interviewer noted it.
- Solutions:
  - A targeted row lock: `SELECT ... FOR UPDATE` (the lock is held **until the end of the transaction** - COMMIT, not until the end of the SELECT; when transferring between wallets, lock both rows **in a fixed order** to prevent deadlocks).
  - An atomic UPDATE with a condition: `UPDATE wallets SET balance = balance - 100 WHERE user_id = 42 AND balance >= 100;` - the database checks it itself.
  - `CHECK (balance >= 0)` + an unconditional UPDATE.
- Raising the isolation level globally - "everything will be clean, but performance will sag": a targeted lock is preferable.
- **Deadlock**: two transactions hold each other's rows; the DB/application detects the mutual block. The classic definition + how to avoid it (lock ordering).

## C4. Scaling, replication, views

**C4.1. Vertical vs horizontal.** [S2] 🪤
- **Vertical** - adding resources to a single machine (CPU/memory/disk) - there is a **ceiling**. The candidate in S2 called splitting by tables vertical - wrong (that's horizontal).
- **Horizontal**: **replication** (copies of the data on servers), **sharding** (splitting the data: by rows/tables - e.g. by region), **partitioning** (a table inside one DB is split into parts by range/list; logically one table, physically several files; queries get faster).

**C4.2. Sharding: how to choose the key?** [S3] - **by the business access logic**: chats → by chat (all of a chat's messages in one place), orders → by region/location. Partitioning is easiest to do **by time** (created_at).

**C4.3. Replication.** [S3, S4]
- Schemes: master-slave, master-master; synchronous/asynchronous (per consistency requirements).
- Synchronization: shipping the log (WAL/binary log) or the data (snapshots/dumps); **push** (the master sends on its own) / **pull** (replicas fetch); a **hybrid cascade** - the master sends to one node, which fans it out to the rest.
- Choosing a new master: the candidate honestly didn't know - the interviewer accepted it. **Learn at least the names**: Raft (a consensus algorithm), PostgreSQL failover tools (Patroni), elections in MySQL clusters. An honest "I don't know" is fine, but for senior positions it's better to know the names [S3].

**C4.4. View vs materialized view.** [S4]
- **View** - an alias for a query, it takes up no space; the downside: moving logic to the DB side (like stored procedures) - in many cases an **anti-pattern**.
- **Materialized view** - a **cached snapshot**: the query runs at creation time, and on access it serves the latest snapshot; refreshed with **`REFRESH MATERIALIZED VIEW`**; it doesn't keep history. The "redis → caching" association is the right one.

**C4.5. Diagnosing "the service is slow" (a systematic answer).** [S5]
1. Logs → find the request (ID) → **trace ID** → the whole path.
2. **The delta between log entries** → the spot with the biggest delay.
3. Which queries are in that segment → **EXPLAIN** (rows/cost, indexes, joins).
4. If it's Go code - **pprof**; if you have access - reproduce it locally/on the staging environment and run it.
- Answer from the assumption that the company has everything (Kibana, tracing) - not "I'll go dig through file logs over FTP" [S5].
- A 20-million-row table: EXPLAIN → indexes → partitioning; if you need part of the data - **LIMIT**; "three rows out of 20 million" - don't build a Redis setup [S5].
- **Don't multiply technologies**: exhaust the DB first (indexes, materialized views), Redis is the last resort; constraints ("no Redis, no ClickHouse") are normal in enterprise, and seniority = solving under constraints [S5].

**C4.6. Design: a library schema.** [S6] 🪤
- **The "Gang of Four" trap**: a book with four authors doesn't fit the "author_id in books" schema (one-to-many). Many-to-many → a **junction table** `book_authors` (a "book-author" pair = a row).
- Reader↔book: `reader_id` in books **nullable** (the book is free) + a unique constraint on `book_id` (at most one reader).
- Queries: "books on loan" - `WHERE reader_id IS NOT NULL`; "top most-read authors" - JOIN → WHERE (only on loan) → GROUP BY → COUNT → ORDER BY DESC → LIMIT 3.

---

# Part D. Infrastructure, networks, brokers (not Go - a separate category) 🧭

## D1. Containers, virtualization, Kubernetes

**D1.1. VM vs container.** [S2, S6]
- **Virtualization**: a hypervisor emulates hardware/a machine (its own OS on top of yours, syscall translation) - heavy, but **truly isolated**.
- **A container**: uses the host OS (chroot + similar mechanisms), syscalls are almost direct - **the same host "behind a curtain"**: lightweight, fast, **weaker isolation** [S6].
- Docker's advantages: lightweight, images in **layers** (only the changed layer is rebuilt), Docker Hub, fast deploy/scaling.
- **The "suspicious binary" provocation**: in a **VM** - a container is not fully isolated, you risk the host OS [S6].

**D1.2. Docker vs Kubernetes.** [S2] - Docker is creating containers/images; K8s is **orchestration** (control over containers in pods: it makes sure "all pods didn't die and the service didn't go down"; addressing a pod by URL, policies, load balancers) - the level of large systems.

## D2. Networks

**D2.1. TCP vs UDP.** [S2] - TCP: ordering and delivery confirmation (requests, data transfer); UDP: **can lose packets, but is faster** - streaming, games, video calls.

**D2.2. HTTP.** [S6]
- Request structure: method, path, protocol version; headers are metadata (Host, Content-Type); body is the data.
- **Host**: several sites can live on one IP → the server knows which resource is being addressed.
- A custom HTTP method: formally possible, **bad practice**.
- Codes: 200, 201 (Created), 204 (No Content), 301/302 (redirects), 404. (The candidate in S6 mixed up the numbers - don't repeat it; "204 vs 2011" is a transcription slip.)
- **HTTP/1.1 vs HTTP/2**: 1.1 is a text protocol; 2 uses **binary frames**. Both need parsing (the interviewer picked on that) - the win is that frames **know their exact length** and are read unambiguously faster [S6] ⚠️.

**D2.3. gRPC.** [S2]
- Over **HTTP/2**, the data is a **binary format** (fast), the contract is a **`.proto`** file (methods, request/response), clients are generated for any language, strict typing, **streaming** (bidirectional). For service communication **inside the perimeter** (user → HTTP, internally - gRPC).
- Why not everywhere: the difficulty of migrating, HTTP/2 isn't everywhere, it's inconvenient to test/you can't see what's flying.
- ⚠️ Don't say "faster because of encryption" - that's the candidate's slip; the point is the binary protocol + HTTP/2.

## D3. Kafka

**D3.1. Partitions and replicas.** [S1]
- **A partition** - a logical sequence of messages: **ordering** + **parallelizing** processing.
- **A replica** - a copy of a partition **on another broker**: resilience; when a broker fails - rebalancing.

**D3.2. Consumer groups.** [S4] - they unite consumers: within a group **one partition is not read by two consumers**; partitions and groups are used together.

**D3.3. Delivery guarantees.** [S1] - at-least-once / at-most-once (be able to name them; honestly admit deep configuration gaps - the interviewer is checking depth, "if you haven't done it, no big deal").
- 💡 If you haven't worked with Kafka in production, that's normal: honestly say "I haven't had to administer it", keep the terminology, and where possible steer the answer toward familiar analogues (RabbitMQ: queues/exchanges, ack, prefetch).

## D4. Monitoring and alerting

**D4.1. Tools.** [S6] - **Prometheus** (metrics/time series), **Grafana** (dashboards), **Jaeger** (distributed tracing - where the request got stuck), **OpenTelemetry** (the telemetry standard: API+SDK, export over OTLP). The canon on this topic with answers - [language-agnostic bank](non-language-question-bank.md), D3.

**D4.2. Service alerting.** [S6] - requests/errors: the error rate above a threshold (4xx - Bad Request, 401/403 - authorization; connection refused - an infrastructure problem), CPU/memory above a threshold.

**D4.3. Tracing in Go: how to wire it up.** [✍️]
- The standard: `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` - it wraps **both the server and the client** (http.Server / http.Client); for gRPC - `otelgrpc`. Export - OTLP to Jaeger/Tempo/a collector.
- ⚠️ Hole #1: the server is wrapped, but **hand-rolled outgoing calls are not** → the trace breaks at the first external call.
- ⚠️ Hole #2: **the `context.Context` with the span is not passed** into goroutines/DB calls - manual spans on SQL and background tasks.
- Propagation by default - **W3C Trace Context**; for a Zipkin environment - `tracecontext + b3` (a textual combination of propagators).
- The canon on the topic - [language-agnostic bank](non-language-question-bank.md), D3 (D3.4 propagation, D3.5 sampling).

**D4.4. pprof: profiles and scenarios.** [✍️]
- `net/http/pprof` + `go tool pprof`: **CPU** (a 30 sec profile), **heap** (allocations), **goroutine** (leaks/stuck ones), **block/mutex** (contention).
- Scenarios: latency without errors → a CPU profile. Memory growing → heap at **two points in time** (a diff). Goroutines piling up → goroutine + the `runtime.NumGoroutine()` metric on a dashboard.
- 📎 pprof is already in the systematic "the service is slow" answer (C4.5) - here we cover the profiles individually.

## D5. Microservices: the downsides

**D5.1. The drawbacks of microservices.** [S4]
- Harder to **debug** (a request spans several services; harder to roll back; transactions across services).
- Decentralization: **you don't know what's in the other services** (separate databases, separate teams).
- **Interaction is slower** (data overhead) vs inter-thread calls in a monolith.
- **Development speed is lower** for an MVP; a monolith is easier to deploy.
- Business: more costs, longer support, more developers. (The interviewer's joke: "context switching also happens in the developers' heads between services".)

---

# Part E. Interview behavior (from interviewers' feedback)

1. **Talk it through out loud.** Code "strictly top to bottom without logic" plus silence in 2026 = suspicion of a neural net/Googling [S5]. Live coding is a mini-lecture.
2. **Sum up** if your answer drags on - the interviewer's "head full of mush" becomes your problem [S5].
3. **Clarify the requirements before coding** (minimum/maximum replicas, timeouts, a case-sensitive palindrome or not) - most interviewers give a plus for it [S7, S3].
4. **Run the primitive case** before saying "I'm done": found the bug yourself - great; found it before declaring readiness - even better [S7].
5. **Decompose**: a function has grown past 50 lines - extract it/suggest it [S7].
6. **An honest "I don't know/haven't done that" is fine** (administering Kafka/DBs, scheduler mechanics): the interviewer is checking depth and honesty; don't claim experience you don't have [S1, S3, S6].
7. **Seniority = solving under constraints** ("no Redis"), not multiplying technologies, answering from the assumption that "the company has everything" (Kibana, EXPLAIN) [S5].
8. **You might turn out to be more right than the interviewer** (the cache: O(1) vs O(n)) - a middle/senior interview is a discussion, not a dictation [S5].
9. **Terminology down to automatic recall**: kinds of joins, HAVING/WHERE, isolation levels, selectivity - "floating" phrasing is noticeable and sinks you even with the tasks solved ("solved the tasks" ≠ "passed the interview") [S3].
10. **The candidate's questions** - to the point and not overwhelming: team composition, the lead's role, testers, neighboring teams; at early stages don't pry for specifics about the tasks - a "black box" [S1, S3].
11. **Connect your experience to the interviewer's project** ("right now I'm in a similar story") - it works [S1].
12. **Code review in an interview** can be harsher than in real life - that's normal; write comments to the point [S1].

---

*The synthesis was done on 23.08.2026. As new transcripts come in, questions are added here with a source label.*
