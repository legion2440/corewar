[EXECUTION-CORE v2.1]

EX-1 Scope first. Resolve the smallest owning module before deep inspection. Batch independent bounded facts when practical. Expand only for a concrete unanswered question, dependency, contract uncertainty, or verification gap.

EX-2 Measure unknown size when cheap. Read small relevant files whole; inspect large files by symbol/range first. Never truncate a complete input whose whole structure is required for correctness.

EX-3 Prefer a formal schema/specification, then a canonical generator/template, then two examples of the exact construct. Do not sample arbitrary files when a formal contract exists.

EX-4 Probe task-required environment dependencies together. Repair missing tooling only when the change is local, reversible, and in scope; otherwise report the exact blocker and never claim green verification.

EX-5 Run explicit task checks plus affected validators/tests and cross-module/runtime checks where applicable. When required verification is green and no runtime claim remains unresolved, stop unless broader review was requested.

EX-6 If the same check fails twice for materially the same reason under the same approach, stop symptom patching. Name an alternative root-cause hypothesis and take a materially different action before another similar patch.

EX-7 Poll only when polling can produce new information. Do not rapid-poll long-running work; use roughly 30 seconds or longer where appropriate. Blocking execution needs no artificial wait.
