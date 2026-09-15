# Interview question banks: Go, PHP, language-agnostic

Questions that actually get asked in interviews, synthesized from videos of real and mock interviews. Each bank is a trap map ("what they try to trip you with") plus the questions with answers.

English versions are in `en/`, Russian originals in `ru/`.

## Files

- `go-question-bank.md` - Go: scheduler and goroutines, channels, synchronization, slices and maps, errors and defer, context, GC. Plus live coding and databases (synthesis of 7 transcripts)
- `php-question-bank.md` - PHP: language core, OOP, Composer and autoloading, PDO, Laravel, SQL, security (synthesis of 6 transcripts)
- `non-language-question-bank.md` - what gets asked regardless of language: databases, architecture, networking, infrastructure, observability, algorithms, interview behavior

## How to use it

Cover the answer with your hand and say it out loud using the skeleton: definition -> mechanism -> why and what the trade-off is -> the catch. Reading an answer creates the illusion of knowing it; saying it out loud does not.

Markers used in the text:

- 🪤 - trap question: the phrasing is designed to mislead
- ⚠️ - sources disagree or the candidate in the source was wrong; here is the corrected version
- 🧠 - the mechanism, "why it works this way". Read it when the answer doesn't come together
- 🔧 - code worth typing by hand instead of reading
- 💡 - a nuance
- ✍️ - a question that wasn't in the recordings

Get the terminology to the point of reflex (JOIN types, isolation levels, HAVING vs WHERE, selectivity, at-least-once): vague phrasing is noticeable and sinks candidates even when the coding part is solved.

## How it was built

1. Videos were picked by hand: real interviews, mocks, mistake breakdowns. Views, date and content - so that no junk ended up in the bank.
2. An AI agent pulled the transcripts and split questions with answers into files.
3. Duplicates were merged into one canonical answer; candidate answers were checked - where a source was wrong, the corrected version stands; traps were moved into separate maps, non-language questions into a shared bank.

Labels `S1…S7`, `G#`, `P#` in the text point back to the sources; they are described in the header of each file. The videos themselves are not attached.

## Caveats

- The sources are public videos and the answers are synthesized. Disputed points are marked ⚠️, but verifying against the official docs is still worth it.
- Coding tasks were reconstructed from transcripts only partially: the approach is visible from the dialogue, but the full problem statement sometimes is not.
- This is a snapshot from late August 2026: Go synthesis - 2026-08-23, PHP and language-agnostic - 2026-08-24.

---

## По-русски

Русские версии банков лежат в `ru/` (файлы те же, состав вопросов и ответов идентичен).

Вопросы, которые реально задают на собеседованиях, собранные из видео реальных и мок-интервью: карта ловушек плюс сами вопросы с ответами. Как пользоваться - закрывай ответ рукой и пересказывай вслух по скелету: определение -> механизм -> зачем и trade-off -> подвох.

Источники - публичные видео, ответы синтезированы; спорные места помечены ⚠️, но сверяться с документацией всё равно полезно.
