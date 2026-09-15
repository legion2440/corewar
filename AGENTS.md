# Agent navigation

0. If `[EXECUTION-CORE v2.1]` is not already present in current instructions, read `agent/execution-core.md`. Do not search for it.
1. Read `agent/module-index.json`.
2. Identify the owning module from roots and entrypoints.
3. Read only `agent/modules/<module-id>.json` for the owning module.
4. Inspect relevant production code/configuration; metadata is navigation, not implementation evidence.
5. Read `docs/architecture.md` only when a concrete architectural question requires it.
6. Inspect `agent/dependency-graph.json` before changing a cross-module contract.
7. Run explicit task checks and affected validators/tests after changes.
8. Run `python3 scripts/validate_agent_contracts.py` when agent metadata changes.
9. Expand context only when the current step reveals a concrete unanswered question, dependency, contract uncertainty, or verification gap.
