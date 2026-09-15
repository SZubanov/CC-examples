# Question bank: real PHP interviews (synthesis of 6 transcripts)

**What this is:** a synthesis of six transcripts of real and mock interviews (YouTube videos). Duplicate questions are merged into one canonical answer; where sources disagree or the candidate got it wrong, a corrected version is given. Trap questions are marked 🪤. Non-PHP questions (databases, architecture, networking, security, algorithms) are moved to parts C, D, E and marked 🧭 - consolidated in the [language-agnostic bank](non-language-question-bank.md) (a synthesis of the Go and PHP banks).

**Sources:**
- S1 - 30 typical questions (CareerGuidex, lecture-style list)
- S2 - 275k RUB: code review + DB + microservice (Mikhail Savin)
- S3 - 350k RUB: review + DB + SQL + PHP theory + architecture + security (Mikhail Savin)
- S4 - Middle/Senior 300k RUB (worklifebackend)
- S5 - 250k RUB, 2h, real interview: PHP + Laravel + SQL + Docker + algorithms (worklifebackend)
- S6 - advice and candidate mistakes (Phantomas Popuas)

**Legend:** 🪤 trap question / provocation (they try to confuse you on purpose) · ⚠️ sources disagree / error in the source, corrected here · 💡 nuance · 🔧 code · 🧭 non-PHP question (blocks C/D/E) · ✍️ a question that was not in the dumps.

**How to use it:** cover the answer with your hand and say it out loud. Code (prepared statements, array reversal, atomic writes) - write it by hand. Terminology (JOIN, isolation levels, HAVING/WHERE, selectivity) - down to muscle memory: "fuzzy" phrasing is noticeable in an interview and sinks you even when you have solved the tasks (S3).

---

# 🪤 Trap map (PHP)

| # | Provocation | Correct answer |
|---|---|---|
| 1 | "Will `strlen('Вася')` return 4?" | No: **8** - `strlen` counts bytes, not characters (UTF-8: Cyrillic letters are 2 bytes each). For characters, use `mb_strlen()` [S4] |
| 2 | "Are `1 == '1'` and `1 === '1'` both true?" | `==` is true (type juggling), `===` is **false** (type and value, no coercion) [S1] |
| 3 | "Does `isset()` only check whether a variable exists?" | No: it also checks `!= null`, plus magic type coercion that cannot be turned off by `strict_types` in the calling file [S2] |
| 4 | "Will `json_encode` throw an exception on error?" | No: it **silently returns `null`/`false`**. Check `json_last_error()` or pass `JSON_THROW_ON_ERROR` [S2, S3] |
| 5 | "Are PHP arguments always passed by value?" | Almost: everything except objects is by value; **objects are always by reference**; arrays are by value, but with copy-on-write [S3] |
| 6 | "Do `include` and `require` differ only in syntax?" | `include` - **warning** and we carry on; `require` - **fatal error** and we stop [S1] |
| 7 | "Is it safe to `unserialize` untrusted data?" | No - **object injection** risk (`__wakeup`/`__unserialize`). Pass `allowed_classes` [S3] |
| 8 | "A child class added a required parameter to a parent method - fine?" | No - an **LSP** violation: code that worked with the parent will break. Fixed with a default value [S3, S4] |
| 9 | "`composer.lock` does not need to be committed" | It does: it pins exact versions and gives a reproducible install [S4, S5] |
| 10 | "Composer does not work without `composer.lock`" | It does: `install` will recreate the lock from `composer.json` - the risk is drifting versions, not a crash [S5] |
| 11 | "`final` can be applied to class properties" | No: `final` applies to classes, methods and (since 8.1) constants, but **not to properties** [S5] |
| 12 | "Traits are a multithreading mechanism" | No: traits are **code reuse** under single inheritance [S5] |
| 13 | "`self` and `static` are synonyms" | No: `self` is early binding (where it is written), `static` is **late** binding (late static binding, the class that was actually called) [S4] |
| 14 | "Middleware is called before the route is resolved" | Global runs before, route middleware runs **after** the route is matched, terminable runs after the response. Middleware ≠ validation [S5] |
| 15 | "Reversing an array with a loop over all N works" | No: you have to iterate over **N/2**, otherwise it reverses twice and comes back to the original order. In practice, use `array_reverse()` [S5] |
| 16 | "An index on a bool column will speed up the query" | Only at **high selectivity**; 50/50 or 90%+ of one value - useless [S4, S2] |
| 17 | "CROSS JOIN is the intersection of tables" | No: it is the **Cartesian product** (every row × every row) [S3] |
| 18 | "HAVING can only be used after GROUP BY" | Like WHERE, it can be used without GROUP BY; the typical place is after it, over aggregates [S3] |
| 19 | "Changing a column type with `ALTER ... TYPE` is fine" | On millions of rows it takes a heavy lock. The right way: a new NULL column without a DEFAULT + dual writes + a background backfill [S5] |
| 20 | "Creating an index on a live table blocks nobody" | It can block; on a large table use `CREATE INDEX CONCURRENTLY` [S3, S4] |
| 21 | "Exactly-once delivery is the standard, enabled everywhere" | It is hard to achieve; usually you get **at-least-once** (duplicates) or at-most-once (message loss); exactly-once requires idempotency [S4] |
| 22 | "Verifying JWT in PHP on every request is fine under high load" | Parsing and signature verification is CPU-heavy; first think about the **edge (nginx/OpenResty)**, not a separate auth microservice [S5] |
| 23 | "MD5 + salt is enough for passwords" | No: fast collisions + **timing attacks**; you need bcrypt/argon2id [S3] |
| 24 | "Only Secure protects a cookie from JS theft" | No: protection from theft via JS (`document.cookie`) is **HttpOnly** [S3] |
| 25 | "An atomic file write is done with fwrite to a stream" | No: fwrite flushes in chunks. The right way: **temp file + `rename()`** (atomic at the OS level) [S4] |
| 26 | "Is `0 == "abc"` true, since PHP coerces types?" | In PHP 8 it is **false**: a number vs a non-numeric string is compared as strings (in PHP 7 it was true - the string was cast to 0). A related bug: `in_array(0, ["foo"])` - PHP 7: true, PHP 8: false [RFC PHP 8] |

---

# Part A. PHP core (theory)

## A1. Language, types, operators

**A1.1. What is PHP and what are its main capabilities?** [S1]
- An open-source server-side scripting language created primarily for web development: it processes requests on the server, generates dynamic pages, works with databases.
- Capabilities: simplicity, cross-platform support, OOP, sessions, extensive libraries, many frameworks, easy integration with HTML.

**A1.2. Variables and dynamic typing.** [S1]
- A variable starts with `$`; its type is determined by the assigned value (dynamic typing). This is also the root of the loose comparison trap (A1.3).

**A1.3. The difference between `==` and `===`.** [S1] 🪤
- `==` compares values **after type coercion**; `===` compares the value **and the type** with no coercion.
```php
var_dump(1 == "1");   // true  - type juggling
var_dump(1 === "1");  // false - different types
```
- 💡 Strict comparison makes code predictable - interviewers value that.
- ⚠️ **PHP 8 changed number-vs-string comparison** (the "Saner string to number comparisons" RFC): a number vs a **non-numeric** string is now compared as strings - `0 == "abc"` → **false** (in PHP 7 it was true), `0 == ""` → false. Numeric strings behave as before: `42 == " 42"` → true. The same bug lived in `in_array(0, ["foo"])`: PHP 7 - true, PHP 8 - false.
- 🧠 **Mechanism:** in PHP 7 a string in a loose comparison was cast to a number - any non-numeric string became 0, hence `0 == "abc"` → true. In PHP 8, with a non-numeric string it is the other way around: the number is cast to a string and the strings are compared; numeric strings are compared as numbers, as before.

**A1.4. `strlen` vs `mb_strlen`.** [S4] 🪤
- `strlen()` counts **bytes**, `mb_strlen()` counts **characters**. For Cyrillic in UTF-8 the difference is twofold.
```php
strlen('Вася');    // 8 - bytes (4 chars × 2 bytes)
mb_strlen('Вася'); // 4 - characters
```
- ⚠️ In S4 the candidate answered "4" - he confused bytes with characters.

**A1.5. Function arguments - by value or by reference?** [S3]
- By default everything **except objects is by value**; **objects are always by reference**.
- Forcing by-reference passing - the `&` sign in the signature.
- Arrays are by value, but under the hood it is **copy-on-write (COW)**: a reference to the same zval is passed, and a copy happens only on modification (memory savings on large arrays).
- 💡 Mentioning COW is a strong point in the answer.

**A1.6. `include` vs `require` (and the `_once` variants).** [S1]
- Both include a file; the difference is the behavior when the file is missing: `include` - a **warning** and the script continues; `require` - a **fatal error** and it stops.
- `include_once`/`require_once` guarantee a **single** inclusion (protection against redeclaring functions/classes).
- `require` - for mandatory files, `include` - for optional ones.

**A1.7. Generators vs iterators.** [S4]
- A generator is a function with `yield` that keeps its state between calls; it returns a `Generator` object implementing the `Iterator` interface (`current`, `next`, `key`, `rewind`, `valid`).
- Use: reading millions of rows from the DB **in batches/chunks** (`chunk()` in Laravel) - memory savings, data is loaded in portions.
- An iterator is a PHP interface and a pattern of the same name; `foreach` works through it; a custom class can implement `Iterator` and be used in `foreach`.
- 💡 `yield` - lazy (one-at-a-time) processing of large result sets without loading everything into memory.

## A2. Strings and arrays

**A2.1. Arrays in PHP.** [S1]
- They hold many values in one variable; three kinds: **indexed, associative, multidimensional**. They simplify loops, searches, sorting.

**A2.2. Single vs double quotes.** [S3]
- PHP **parses the contents** of double quotes (variable interpolation) → slightly slower; single quotes - it does not.
- 💡 A code-review-level detail: use single quotes for strings without substitutions.

**A2.3. The map key in PHP; a canonical key for grouping.** [✍️] 🔧
- A PHP associative array is itself a hash table; keys are **only int | string**, an array cannot be a key (Illegal offset type). A composite signature (analogous to Go's `[26]int`) has to be **serialized into a string by you**.
- Grouping anagrams, option 1 (sort the letters): `$chars = str_split($str); sort($chars); $key = implode('', $chars);` - "eat"/"tea"/"ate" → `aet`, O(n·k log k).
- Option 2 (counting, O(n·k)): `$counts = array_fill(0, 26, 0);` … `$counts[ord($str[$i]) - ord('a')]++;` … `$key = implode(',', $counts);` - commas separate the 26 numbers, the serialization is unambiguous.
- Fill idiom: `$groups[$key][] = $str;` (there is no key yet - the array creates itself), then `array_values($groups)` at the end.
- The closest thing to `[26]int`: key = `pack('C*', ...$counts)` - a fixed 26-byte string (counters ≤ 255). A PHP bonus: `count_chars($str, 1)` returns [byte_code => frequency] right away (then ksort + serialization).
- 🧠 **Mechanism:** in Go, arrays/structs are comparable by value and the runtime computes a hash over them itself → the signature goes into the key as is. PHP hashes only int|string, so the serialization step is up to the programmer. The trick is always the same: normalize the element (sort - order of letters; count - frequency vector) → identical signatures → one bucket.

## A3. OOP

**A3.1. The three pillars of OOP.** [S1, S5]
- **Encapsulation** - hiding internal state behind a public API; **inheritance** - a child class takes over the parent's behavior; **polymorphism** - different implementations behind one interface.
- Polymorphism by example: a remote with a "power on" button, but the implementation differs across TV models (S5).

**A3.2. Interfaces vs abstract classes.** [S4, S5]
- **An interface** is a contract: signatures only (public methods, constants, and since 8.4 - properties). **An abstract class** is an inheritance hierarchy with shared implementation; you cannot instantiate it.
- The choice: an interface - when you need a contract and the implementation is injected via DI; an abstract class - when there is shared logic.
- 💡 The LSP trap: "a penguin does not fly" - if the abstract "Bird" has `fly()` and a penguin does not fly, flying is better moved into an interface.

**A3.3. Traits.** [S5] 🪤
- A mechanism for code reuse (composition) under PHP's single inheritance.
- Pros: horizontal reuse without inheritance; convenient for `final` classes.
- Cons: harder to trace, name conflicts (resolved with `insteadof`/`as`), you cannot type an argument "by trait".
- ⚠️ The candidate called traits "a multithreading solution" - wrong: they are about code reuse, not threads.

**A3.4. `final` on a class.** [S5] 🪤
- `final class` forbids inheritance; a `final` method cannot be overridden.
- Downside: final classes cannot be properly mocked (mocking relies on inheritance) - you have to reach for reflection.
- ⚠️ `final` applies to classes, methods and (since 8.1) constants, but **not to ordinary properties**.

**A3.5. `self` vs `static`.** [S4] 🪤
- This is about **late static binding**.
- `self` - early binding: the call is bound to the class where `self` is written.
- `static` - late binding: it returns/calls the class it was actually invoked from (the child) - this solves the inheritance problem. Introduced in PHP 5.3.

**A3.6. Magic methods.** [S4, S5]
- The full set: `__construct`, `__destruct`, `__clone`, `__toString`, `__invoke`, `__serialize`/`__unserialize` (the modern replacement for `__sleep`/`__wakeup`, 7.4+).
- `__destruct` is usually unnecessary in PHP (it is mandatory in C++); it comes in handy in workers/async code.

**A3.7. `__invoke`.** [S5]
- Lets you call an object as a function (`$obj()`). In Laravel - invokable controllers: the framework uses reflection to call the class as a route handler without an explicit method name.

**A3.8. `clone`.** [S5, S4]
- By default it is a **shallow** copy: the object is cloned, nested objects/references are not.
- A deep copy comes from overriding `__clone` (you create new nested instances) or from `serialize`/`unserialize`.
- Be careful with resources (resource descriptors) - cloning behaves in its own specific way.

**A3.9. Attributes.** [S4]
- Introduced in PHP 8.0, they work through the Reflection API; they replaced docblock annotations (phpdoc).
- Symfony is built on attributes (entity validation); in Laravel - subscribing a listener to an event.
- 💡 A native way to add metadata that reflection can read.

**A3.10. SOLID: LSP and DIP in the PHP context.** [S3, S4, S5]
- **L (Liskov)** - substitutability: a subclass must replace its parent correctly. The violation from S3: the child's `saveUser(name, email, role)` gained a required argument that the parent's `saveUser(name, email)` does not have - code that worked with the parent will break. The fix is a default value for the new argument.
- **D (Dependency Inversion)** - depend on abstractions, not on concrete implementations; solved with **interfaces** + DI. ⚠️ In S5 the candidate first named OCP/LSP - the right emphasis is DIP.
- **S (Single responsibility)** - move `exportUsers`/`sendWelcomeEmail` out of `UserManager` (S3).
- **O (Open/closed)** - a hardcoded dependency that cannot be swapped out (S3).
- **I (Interface segregation)** - put the CSV/JSON exporters behind one interface (S3).
- 💡 A mature stance (and it is valued): you do not have to shove every SOLID principle everywhere - otherwise you get "Hello World in abstract factories" (S3).

## A4. Autoload, Composer, PSR

**A4.1. Autoloading: PSR-4, Composer, `spl_autoload_register`.** [S3, S4]
- PSR-0 was the first autoloading standard; **PSR-4** is the current one (PSR-0 is outdated).
- Composer: in `composer.json` the `autoload` section maps a folder (`src` → namespace `App`); it generates `vendor/autoload.php`, which is `require`d in `index.php`.
- You can implement it yourself via **`spl_autoload_register`** - it registers an autoload function; one declared later overrides Composer's.

**A4.2. What `use` does and how autoloading works.** [S4, S3]
- `use` **imports/aliases** a class name from a namespace; by itself it loads nothing - the actual loading is done by the autoloader.
- Composer builds a "namespace ↔ path" map and loads a class on demand, freeing you from manual `include`/`require`.
- A bonus of namespaces: two classes with the same name from different namespaces + the `use ... as ...` alias.
- If the file is not found: a warning with `include`, a fatal error with `require` (S3).

**A4.3. `composer install` vs `composer update`.** [S4, S5]
- `install` - installs dependencies at the versions pinned in `composer.lock`.
- `update` - upgrades to the latest allowed versions and **rewrites** `composer.lock`.
- ⚠️ `update` in production is dangerous: dependencies can update and break a tested application.

**A4.4. `composer.lock` vs `composer.json`; should the lock be committed.** [S4, S5] ⚠️
- `composer.json` - the description of dependencies and version constraints (`~`, `^`, `*`); `composer.lock` - generated by Composer, pins **exact versions** (including transitive ones) with hashes.
- The lock **is committed to the repository** - it guarantees a reproducible install (in S4 the candidate first got this wrong, answering "no").
- ⚠️ In S5 the candidate claimed "Composer does not work without the lock" - wrong: if the lock is deleted, `install` will resolve dependencies again from `composer.json` and create a new lock; the risk is drifting versions, not a crash.

**A4.5. Is there a `composer.lock` inside packages.** [S5]
- Every package has its own `composer.json`, but `composer.lock` is not used in **library** packages - the lock only makes sense for the root application. The full dependency graph is pinned by the root `composer.lock`.

**A4.6. Private Composer packages.** [S5]
- The candidate had no experience with this; the principle: a package is described in `composer.json`, published, then included as a dependency; a common package can be shared between microservices.

**A4.7. PSR - what it is and the key standards.** [S5]
- PSR = PHP Standards Recommendation; published by **PHP-FIG** (Framework Interop Group).
- The main ones: PSR-1 (basic coding style), **PSR-4** (autoloading), PSR-12 (extended code style), PSR-6/PSR-16 (cache), PSR-7 (HTTP messages), PSR-18 (HTTP client).
- ⚠️ The candidate did not name the exact numbers - the interviewer confirmed: the numbers are not critical, what matters is the purpose.

## A5. PDO and database access

**A5.1. Connecting to MySQL: PDO vs MySQLi.** [S1]
- Both are extensions; PDO is preferable: it supports **multiple databases**, prepared statements, and better security.
- You specify the server, user, password, database name; handle connection errors, close unused connections.

**A5.2. Prepared statements and SQL injection.** [S1, S3] 🔧
- A prepared statement **separates SQL from user input** - an injection like `1; DROP DATABASE users` will not work.
```php
$stmt = $pdo->prepare("SELECT * FROM users WHERE email = :email");
$stmt->execute(['email' => $email]);
```
- The main fix against injections is precisely the prepared statement; type validation (for example, that it is an `int`) is only a complement (S3).
- 💡 A sign of professional style; it also speeds up repeated queries.

**A5.3. `json_encode` and the "silent" null.** [S2, S3] ⚠️
- On error `json_encode` **returns `null`/`false`** (it does not throw an exception) - and that null "successfully prints" into the result.
- The cure: pass the **`JSON_THROW_ON_ERROR`** flag (it throws an exception) or check `json_last_error()` / `JSON_ERROR_*`.
- On top of that: with large data there is a memory risk.

**A5.4. Mapping a result to a class.** [S3]
- Instead of an array - `PDO::FETCH_*`, for example `PDO::FETCH_CLASS`: it maps a row to an object/DTO and improves typing.

## A6. Errors and exceptions

**A6.1. How to handle errors in PHP.** [S1]
- Validate data, use `try-catch`, log meaningfully, give clear messages; **never show sensitive system information to the user**; find the root cause.

**A6.2. `try/catch` and `PDOException`.** [S3]
- When a query fails, PDO throws **`PDOException`** (the candidate said "PDO result exception" - a slip of the tongue).
- ⚠️ The candidate did not remember whether `PDO::query` supports multiple statements; the point is not to interpolate input and to use prepared statements.

**A6.3. Semantics of `get` vs `find`.** [S3]
- The convention: **`find`** returns null or the result, **`get`** returns the result or **throws an exception**. Check the `fetch` result: if it is empty → `UserNotFoundException`.

## A7. PHP security

**A7.1. `unserialize` and object injection.** [S3] 🪤
- `unserialize` triggers magic methods (`__wakeup`, `__unserialize`) - deserializing **untrusted data** can lead to object injection.
- Protection: pass **`allowed_classes`** - a list of allowed classes; only a listed class can be instantiated from the string.

**A7.2. How to secure a PHP application (a general checklist).** [S1]
- Input validation, data sanitization, prepared statements, password hashing, HTTPS, secure sessions.
- Preventing **XSS** and **CSRF**; regularly updating dependencies; restricting file permissions.
- 💡 Security is a cross-cutting concern at every stage: development, testing, deployment, maintenance.

## A8. Sessions and cookies

**A8.1. Sessions vs cookies.** [S1]
- **Sessions** store user data **on the server**; **cookies** store small data **in the browser**.
- Sessions are safer for authentication and sensitive data; cookies are for preferences/login data.
- (Cookie flags - `HttpOnly`/`Secure`/`SameSite` - see E1.2.)

---

# Part B. Laravel and PHP frameworks

**B1. The request lifecycle in Laravel.** [S5] 🔧
- Request → nginx → PHP-FPM → `public/index.php` (the Composer autoloader is included, the application is created from `bootstrap/app.php`).
- The Kernel determines the request type (HTTP/Console); service providers are registered.
- Then: **global middleware → routing → route middleware → controller → response**.
- ⚠️ The candidate first put middleware before routing and mixed up the order; then he corrected himself: the route is resolved first, then middleware runs.

**B2. Middleware: when exactly it is invoked.** [S5] ⚠️
- **Global** - before routing, for all requests. **Route** - after the route is resolved. **Terminable** - after the response is sent (post-processing).
- ⚠️ Middleware ≠ validation: "prepare the data and bail out with a header before the route" is pre-validation, not middleware.

**B3. Laravel facades.** [S5]
- A facade is syntactic sugar: a static proxy to the service container; under the hood the DI container initializes a concrete class.
- Pros: a concise, readable API. Cons: static state, hidden dependencies, harder to test; in **long-running** setups (RoadRunner/Swoole/Octane) the static context can cause **memory leaks**.
- 💡 The candidate's team also prefers constructor DI.

**B4. Anemic vs rich model.** [S5] 🧭
- **Anemic** - "empty": properties and relations, with the logic living in repositories/services. **A rich model** - data + behavior logic.
- Active Record (Eloquent) often leads to an anemic model and an SRP violation; Data Mapper (Doctrine) is closer to separating contexts.

**B5. Service container / DI / reflection.** [S5]
- The container runs on **reflection**: it analyzes the constructor signature and resolves dependencies automatically (autowiring) + caches the resolution.
- Service providers register bindings (singleton vs transient).
- Two ways to get an instance: **DI** (automatic) or a **service locator** (`app()->make()`). A service locator means hardcoded dependencies and is harder to control; clean DI is better.
- The **Reflection API** - introspection of classes/methods/properties at runtime; autowiring and invokable controllers are built on it.

**B6. Swoole / RoadRunner / FrankenPHP; Laravel vs Symfony.** [S5]
- Laravel spends more time initializing the kernel (it pulls in service providers) than Symfony; Symfony caches routing and DI into PHP files more aggressively.
- Long-running/async servers - **Swoole, RoadRunner, FrankenPHP** - keep the initialized application in memory and serve requests fast.
- Swoole - for multithreading/coroutines (async routines, mutexes, semaphores), effectively a "replacement for PHP-FPM". RoadRunner (Go) keeps the kernel up, but in terms of concurrency it is closer to FPM (worker processes).

**B7. Queues, jobs, events/listeners.** [S5]
- Jobs (via `dispatch`) - heavy/background computation; events/listeners - one event with several listeners (more flexible), while a job is a single unit of work.
- Workers (`php artisan queue:work`) pull tasks from the queue (Redis) asynchronously; they are kept running via Supervisor.
- By default a listener runs **synchronously** in the same process as the dispatch; it only goes to the background if it implements `ShouldQueue`.

**B8. Supervisor - what it is for and alternatives.** [S5] 🧭
- A daemon that keeps worker processes alive (workers die - it restarts them).
- Alternatives: Docker (restart policies), systemd, pod replication in Kubernetes.

**B9. Auth: Sanctum / Passport / JWT.** [S5]
- **Sanctum** - token-based (SPA tokens, personal access tokens); **Passport** - a full OAuth2 server; JWT libraries (`tymon/jwt-auth`).

**B10. Why parsing JWT in PHP is bad under high load.** [S5] 💡
- Parsing/verifying the JWT signature on every request loads the CPU → a bottleneck at high RPS.
- Options: move verification to **nginx (Lua/OpenResty)** - faster than PHP; a separate auth microservice just to verify a token is **overengineering** (the interviewer pointed this out).
- 💡 Think about where validation should live (the edge) first, do not build microservices for no reason.

**B11. Debugging.** [S5]
- Tinker, console commands, Postman, Xdebug, `dd()/dump()`, `laravel-debugbar` (inspect the queries being built, catch N+1).

**B12. MVC and frameworks (general).** [S1, S5]
- MVC = Model (data) - View (UI) - Controller (handles the request, coordinates).
- Frameworks (Laravel, Symfony, CodeIgniter, CakePHP) simplify routing, authentication, database access, validation.
- 💡 Name the frameworks you have actually worked with, and emphasize that you follow conventions.

---

# Part C. 🧭 Non-PHP: SQL and databases

## C1. SQL: basic constructs

**C1.1. WHERE vs HAVING.** [S3, S5, S4]
- `WHERE` filters **rows before grouping**; `HAVING` - **after GROUP BY**, over aggregates (`COUNT`, `SUM`).
- 💡 A mnemonic phrasing: "WHERE - by rows, HAVING - by groups".

**C1.2. HAVING without GROUP BY.** [S3] ⚠️
- `HAVING` without grouping is allowed - the whole result is treated as one group; the typical place for it is after GROUP BY.
- ⚠️ The candidate answered "it complains" (the `must appear in GROUP BY or be used in an aggregate function` error refers to columns in SELECT, not to HAVING itself); the interviewer accepted the phrasing about the aggregate context.

**C1.3. JOIN types.** [S3] 🪤
- **INNER** - only matching rows. **LEFT/RIGHT** - all rows from one side, NULLs from the other. **FULL OUTER** - all rows from both. **CROSS** - the **Cartesian product** (every row × every row), not an "intersection" (the candidate got it wrong).
- Users without carts: INNER will not return them; LEFT JOIN → NULL for the cart.

**C1.4. CROSS JOIN - what it is for.** [S3]
- Generating all combinations of rows from two tables. ⚠️ The candidate could not name a use case.

## C2. Indexes

**C2.1. What an index is: types, pros/cons.** [S4, S5]
- A structure that speeds up `SELECT`; stored on disk; **slows down INSERT/UPDATE/DELETE** (recomputation/rebalancing).
- B-tree (a balanced tree; the leaves are linked as a list - speeds up range queries). Without an index - full scan.
- Types: primary key, unique, plain B-tree, composite, hash, full-text, partial, GIN/GiST (Postgres: JSONB, geodata, full-text).

**C2.2. Selectivity; an index on bool.** [S4, S2, S3] 🪤
- The key concept is **selectivity**. On low-selectivity columns (bool 50/50, gender, a `status` with two values) an index is almost useless: "we cut it in half" - the scan is still large.
- It makes sense at high selectivity (true for ~10% - it helps; true for 90%+ - the optimizer will choose a full scan).
- In a composite index **the most selective column goes first (leftmost)**.
- 💡 You need to be able to name the term "selectivity" (candidates in S2/S3/S4 stumbled on it).

**C2.3. When indexes hurt.** [S2, S4]
- Not used - wasted space; it slows down INSERT/UPDATE/DELETE; on small tables a seq scan is often faster; low selectivity - useless.
- On a large table - `CREATE INDEX CONCURRENTLY` (without locking writes). ⚠️ Creating an index on a live table can lock it (S3).

**C2.4. Covering index.** [S3]
- The `INCLUDE (...)` syntax - the columns are not indexed but are placed in the index leaf: having found the value, there is no need to visit the table (Index Only Scan).

**C2.5. How to check that indexes are working.** [S3]
- **`EXPLAIN`** (`EXPLAIN ANALYZE` is also possible): you can see `using index` and the name of the index used; the join algorithms and order.

## C3. Transactions and isolation

**C3.1. What a transaction is.** [S3]
- Several operations as one, "all or nothing" (atomicity, ACID); it is needed for data consistency (debiting one account → crediting another with no intermediate state).

**C3.2. Isolation levels.** [S3]
- `Read uncommitted`, `Read committed`, `Repeatable read`, `Serializable`; the default in PostgreSQL is **Read committed**.
- **Read committed**: only committed data from other transactions is visible.
- **Repeatable read**: a snapshot taken at the start of the transaction; it protects against **non-repeatable reads**; phantom reads remain.
- Anomalies: dirty read, non-repeatable read, phantom read, lost update.

## C4. Replication and scaling

**C4.1. Replication: why and how.** [S5, S4]
- Primary-replica: write to the primary, read from replicas (there are usually more read queries). Horizontal scaling + fault tolerance.
- Synchronous/asynchronous; with async replication there can be **replication lag** → eventual consistency (inconsistency has to be handled).

**C4.2. Failover: can the primary and a replica swap roles.** [S5] ⚠️
- Yes - when the primary goes down we switch to a replica (the new primary), and the service keeps running.
- Risk: with asynchronous replication some data may not have made it over → inconsistency.
- ⚠️ The candidate went off describing lag, while the main scenario is fault tolerance when the DB server goes down.

**C4.3. Scaling a database.** [S4, S5]
- **Partitioning** - a built-in PostgreSQL mechanism: the table is split into partitions and the DB itself knows where to go.
- **Sharding** - data spread across different databases (fintech); the sharding key is chosen based on the business access logic.
- "Read a lot, write little" → replication (reads offload the primary).

## C5. Locks and migrations

**C5.1. Locks in a database.** [S5]
- Levels: database, table, row. At the row level: `FOR UPDATE` (exclusive), `FOR NO KEY UPDATE`, `FOR SHARE` (less exclusive).

**C5.2. A heavy migration (changing a column type).** [S5] ⚠️ 🔧
- A naive `ALTER ... TYPE` (integer → bigint) on millions of rows takes a heavy lock (in Postgres - `ACCESS EXCLUSIVE`) → downtime.
- The lock-free path:
  1. A new column of the required type **with NULL and without DEFAULT** (in Postgres it is instant, metadata only).
  2. In the code - dual writes to both columns.
  3. Transfer/convert the old data in background batches.
  4. Switch the code over, drop the old column.
- Tools: online schema change, gh-ost / pt-online-schema-change (MySQL), pg-osc (Postgres).
- ⚠️ The candidate searched for a long time (batches, dump, partitioning - all missed); the interviewer had to suggest the solution.

## C6. CTEs and JSON

**C6.1. CTE (`WITH`).** [S3]
- It forms a temporary "view" that lives in memory for the duration of the query; you can create as many as you like and use them in joins.
- Performance: creating/storing them in memory affects the database. If a subquery is needed twice, `WITH` is better (computed once); if it is simple/small - a plain subquery will do.

**C6.2. JSON columns.** [S2]
- Trade-off: implementing this via JSON or via separate tables. Via JSON - worse in terms of speed and database size; JSON can be indexed, but "not very fast". The argument for normalized tables is more reasonable.

## C7. Normalization and schema design

**C7.1. Normal forms.** [S3]
- **1NF** - do not stuff several values into one field (split the full name into three fields). **2NF** - avoid duplication (extract `resume_link`/`source`). **3NF** - ⚠️ the candidate could not recall it.

**C7.2. A "companies and ads" schema.** [S2]
- Entities: `users`, `companies`, `ads`, `countries`, `tags` + junction tables for white/black list filtering (many-to-many): `country_rules`, `tag_rules` with an allow/forbid flag (a boolean - there is no third state).
- `status` - `varchar`, not `enum` (a new enum status requires an ALTER); `id` - UUIDv7 or unsigned bigint; soft delete via `deleted_at` (NULL does not participate in a unique index → a composite unique on `name` + `deleted_at`).
- 💡 The JSON in the spec is just a format example, it has nothing to do with the DB.

**C7.3. Indexes for "companies/ads".** [S2]
- On filtered/joined columns; on the country name - possible (small table, small index); on tags - yes (`LIKE` + `lower()` searches); on `deleted_at` - yes (`WHERE deleted_at IS NULL`); on `status` - no (low selectivity); on `text` - no (full-text only).

**C7.4. A recruiting schema (vacancies/candidates).** [S3]
- Tables: `departments`, `vacancies`, `vacancy_links`, `responses` (applications), `interviews`, `notes`, `candidates`, `resume`, `source`.
- Key decisions: **separate "candidate" from "resume"** (one person = several resumes for different vacancies); interviews reference `resume_id`; the source is a separate table; dynamic fields - **EAV** (`candidate_properties` + `property_values`).
- Indexes: foreign keys + `status` (where there are many statuses) + `created`; do not index a status with two values.

**C7.5. A partner program (statistics).** [S2]
- `accounts` (our id ↔ their id + token), `statistics` (date + metrics + `external_account_id` + geo). Store a **specific date** (not a period) - otherwise you cannot aggregate.
- The token must be **encrypted reversibly** (do not hash it: you need the original to call the API).
- Month closing: two date fields - `date` (business date) and `source_date` (the date in the source); after closing, changes are forbidden (enforced at the application level + a `CHECK`/trigger in the DB as a safety net); refunds in a closed month are recorded as a negative number.
- Updates - a background worker/cron with a queue of jobs, not a button click (otherwise the user waits 30+ seconds); do not rely on a webhook (most likely you will not be given one).

**C7.6. Loading 30 million records (a tax database).** [S6]
- A practical task: describe the solution design (what to use, how to build it, and why), the impact of optimizations on performance, and how to profile it.
- What is assessed is your approach to thinking, not a memorized answer. (See also: chunked/streaming processing under memory limits - the analogue of generators in A1.7.)

---

# Part D. 🧭 Non-PHP: architecture, networking, infrastructure

## D1. Architecture

**D1.1. MVC / DDD / hexagonal / CQRS.** [S3]
- They do not contradict each other: MVC is about View/Controller, the model is the domain layer. View = the DTO in the response, controller = the HTTP layer of the infrastructure, model = the domain.
- **CQRS** - separating reads from writes; the motivation is heavy read load (the query side reads directly from the database, bypassing the ORM).
- 💡 "MVC + DDD + hexagonal do not contradict each other" - a competent stance.

**D1.2. Coupling vs Cohesion.** [S5] ⚠️
- **Coupling** - links **between** modules, should be low. **Cohesion** - links **inside** a module, should be high (the module is responsible for its own context).
- An anti-example - god objects ("god services").
- ⚠️ The candidate mixed the two terms up.

**D1.3. Layered architecture.** [S5, S3]
- DDD layers: **domain** (business logic), **application** (use cases), **infrastructure** (DB, integrations), **presentation** (API/controllers).
- Patterns: onion, hexagonal (the core is the business, adapters/ports around it).

**D1.4. DDD aggregate.** [S5]
- An aggregate is a consistency boundary: one entry point (the aggregate root), related objects change only through it.

**D1.5. Microservices: what they solve and the pitfalls.** [S5]
- What they solve: extracting a high-load module from the monolith, independent deploys, horizontal scaling, a stack of your own.
- Pitfalls of extracting a feature (sending data to the tax service, FNS): coupling → refactoring; choosing sync vs async (brokers) - async is simpler but adds complexity; data consistency between databases; DevOps load; different stacks → narrow specialists. "You can do anything, but the question is why".

**D1.6. Switching the DBMS / external services.** [S3]
- Switching the DBMS is an **extremely rare process**; you often do not need flexibility for it. ORMs can work with different databases, but we pay for it in performance.
- External services - via an **interface + inversion**: implementations behind a single interface, interchangeable via feature flags/env.
- Risk mitigation: caching read requests, fallback, a clear error, API versioning, stubs for tests.

**D1.7. SOLID (general).** [S4, S5] - the full breakdown of LSP/DIP and the other principles is in A3.10. A quick cheat sheet: S - single responsibility, O - open/closed, L - Liskov substitution, I - interface segregation, D - dependency inversion.

**D1.8. Design patterns.** [S4]
- The basic set: Singleton, Factory (Abstract/Factory Method/Simple), Decorator, Adapter, Proxy, Chain of Responsibility, Iterator, Strategy, Builder.
- In practice: CQRS (read/write separation), **Decorator** (wrapping a repository with a cache), **Builder** (DTOs), **Singleton** (service provider), **Strategy** (payment provider drivers).
- ⚠️ The candidate confused Strategy / Abstract Factory / Factory Method: Strategy - choosing an algorithm through a common interface; Factory - creating objects.

## D2. Networking and HTTP

**D2.1. TCP vs UDP.** [S4]
- TCP guarantees delivery (connection setup, ordering, retransmission). UDP is a "direct stream" (VoIP/streaming): packets can be lost, but it is faster.

**D2.2. What happens when you type google.com.** [S4]
- DNS → IP; TCP handshake; nginx/apache accepts → static content or a proxy to PHP-FPM → the application (Laravel).
- ⚠️ A basic scheme, missing TLS/HTTPS, the HTTP request/response structure, and status codes (the interviewer: "possibly").

**D2.3. Timeouts.** [S4] 💡
- They protect against "infinite" requests and stuck connections; nginx returns 504 when the threshold is exceeded.
- For PHP-FPM: the worker pool is limited - stuck requests exhaust the pool (self-DoS) → the application becomes unavailable.
- Long operations (document generation) should be moved to the background/queues.

## D3. Queues and brokers

**D3.1. Why queues.** [S4, S5]
- Move long/non-critical operations (email, notifications, logging) into an asynchronous background; the user gets a response "here and now".

**D3.2. At least once / at most once / exactly once.** [S4] ⚠️
- Delivery semantics: at-least-once (duplicates are possible), at-most-once (loss is possible), exactly-once (hard to achieve, requires idempotency).
- ⚠️ The candidate did not know the terminology.

**D3.3. Poison message and DLQ.** [S4]
- A message fails → it goes back to the queue → it fails again: a retry counter + a limit (max attempts, e.g. 3) → dead-letter queue (**DLQ**).

**D3.4. The broker is down - how not to lose messages.** [S3]
- In RabbitMQ - the **durable/persistent** flag when creating the queue: messages are persisted to disk rather than memory. In Symfony it is enabled by default.
- A payment system with polling: retries with delays + a **dead letter queue** - "a delay in payments (bad, but not critical) instead of a loss (critical)".

## D4. Docker, K8s, Git, observability

**D4.1. Docker - what it is for.** [S5]
- A consistent environment for dev/prod, fast startup; a container is an isolated process using the host's resources, lighter than a VM (a VM is a full OS).

**D4.2. Kubernetes - what it solves.** [S5]
- Orchestration: self-healing, load distribution, rolling updates. For fault tolerance - 3+ nodes.

**D4.3. Git merge vs rebase.** [S4]
- `merge` - a regular merge, **preserves history**. `rebase` - **rewrites history** (you can lose commits); acceptable when working alone on a branch, but on a team merge is better.

**D4.4. Atomic file write.** [S4] 🪤 🔧
- To rule out the "wrote half and crashed" state: `fwrite` flushes in chunks → it does not cut it.
- The right way: write to a **temp file next to it**, then **`rename()`** - atomic at the OS level (either the whole thing or nothing).
```php
file_put_contents($tmp = $path . '.tmp', $data);
rename($tmp, $path);
```

**D4.5. Profiling PHP.** [✍️] 🧭
- **Xdebug** (locally/staging): profiler → cachegrind files → Qcachegrind/Webgrind - you can see where the CPU goes. For memory, Xdebug 3 does not provide a profile - use Blackfire/APM agents or take `memory_get_peak_usage()` readings at key points.
- You do not leave Xdebug on in production (it slows things down a lot; besides, it is a developer tool - see E1.1 "disable profilers in production"). Production tooling - APM (Datadog/New Relic) or OTel.
- "Production is slow" in PHP is more often not the CPU: an exhausted **FPM pool** (stuck requests, timeouts), SQL without an index (EXPLAIN), external calls - a trace shows it immediately.

**D4.6. Tracing in a PHP project.** [✍️] 🧭
- **OpenTelemetry PHP**: SDK + auto-instrumentation (Symfony/Laravel integrations), export over OTLP.
- Background jobs: **traceparent in the message headers** when publishing to the queue (RabbitMQ/Symfony Messenger) - the worker continues the parent's trace instead of starting a new one.
- Trace ID in logs: a **Monolog processor** - log ↔ trace correlation (see the [language-agnostic bank](non-language-question-bank.md), D3.7).

**D4.7. Observability tools (a quick answer).** [✍️] 🧭
- Metrics - **Prometheus** (+ Grafana), logs - ELK/Loki, traces - Jaeger/Tempo, the standard - **OpenTelemetry (OTLP)**.
- The canonical version with answers - the [language-agnostic bank](non-language-question-bank.md), D3; I am not duplicating it here.

---

# Part E. 🧭 Non-PHP: security, algorithms, behavioral

## E1. Security (common vulnerabilities)

**E1.1. Which vulnerabilities have you encountered and how do you defend against them.** [S3]
- SQL injection (prepared statements - A5.2). Enumerating vulnerable files (`phpinfo()`, config files): only `index.php` should be publicly accessible, disable profilers in production. **XSS** - someone else's JS executes in the victim's browser. XML/ZIP bombs. Open ports (the nuclei scanner). Privilege escalation. No TLS (sniffing).

**E1.2. Cookie protection: the flags.** [S3] ⚠️
- **`HttpOnly`** - blocks access from JS (`document.cookie`), protection from theft via XSS. **`Secure`** - HTTPS only. **`SameSite`**, **`Domain`**.
- ⚠️ The candidate named only SameSite/Domain and missed HttpOnly and Secure.

**E1.3. Why MD5 must not be used for passwords.** [S3] ⚠️
- Fast collisions + **timing attacks** (characters can be guessed from the response time).
- You need modern algorithms with a **salt**: bcrypt/argon2id. MD5 falls in an hour, a modern one takes ~10 years.

**E1.4. JWT: what it is and why.** [S3]
- JSON Web Token - an authentication method on par with sessions; it is needed in **distributed systems** (several servers, sessions cannot be shared).
- It contains user information (issuer, expiration); three parts separated by dots (base64). It enables token revocation and offload (checking a blocklist instead of session existence).

## E2. Algorithms

**E2.1. Reversing an array (a task with a bug).** [S5] 🪤 🔧
- The function has a bug: the loop goes over **all N**, not over **N/2** → the array reverses twice and comes back to the original order.
```php
// The buggy version from the source: the loop goes over all N → it reverses twice
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
- For `[1,2,3,4,5,6]` the result is `[1,2,3,4,5,6]` again. The fix: `$i < intdiv(count($array), 2)`.
- In practice - **`array_reverse()`**, do not reinvent the wheel (the built-in functions are optimized).
- ⚠️ The candidate got the manual index tracing wrong (he counted from one, not from zero).

## E3. Behavioral and preparation

**E3.1. "Tell me about yourself" and motivation.** [S1, S5]
- Structure: background → experience/stack → relevant projects → achievements; keep it short and tailored to the specific company.

**E3.2. Deadlines and teamwork.** [S1, S5]
- Prioritization, breaking work into stages, transparent communication. A review plus a deadline tomorrow: if everything works, ship the business task and put the refactoring into **tech debt** (the interviewer confirmed this as the right answer).

**E3.3. STAR questions (a hard project, a hard problem).** [S1]
- A real project → the challenge → the approach → a measurable result.

**E3.4. AI/LLM and staying up to date.** [S4]
- Documentation, books; LLMs as a search engine and for broadening your horizons, not in "write the code for me" mode (they can produce nonsense).

**E3.5. Preparation: what actually works.** [S6, S5]
- Know your grade (junior/middle/senior - different questions); learn the standard checklist ("100 PHP questions") - the market runs on it; have some practice you can talk about; set up a GitHub with a pet project (even unfinished, but architecturally sound, it works as a portfolio). Grades are blurry - prepare for the max (S5).

**E3.6. Candidate mistakes (do not do these).** [S6]
- Rote memorization without understanding; ignoring security best practices; dirty code; no real examples; weak DB skills; outdated PHP practices; poor communication; no tests; being unable to admit you do not know (an honest "I don't know, but this is how I would have looked it up" is valued).

**E3.7. Should you prepare algorithms?** [S5]
- Out of ~35 interviews, algorithms came up in ~6; if you do prepare - "Grokking Algorithms" + ~200 easy / 50 medium on LeetCode.

---

*Synthesis completed on 24.08.2026. As new transcripts come in, questions are added here with a source label.*
