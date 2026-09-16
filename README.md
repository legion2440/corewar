# Corewar

A complete implementation of the 01-edu Corewar project with a deliberately polyglot architecture:

- **Rust** — assembler, parser, validation, binary encoder, disassembler, arithmetic expressions, and macros.
- **Go** — deterministic virtual machine, process scheduler, arena, instruction execution, lifecycle rules, and visualization boundary.
- **Corewar assembly** — the submitted champion.

The components are independent. Their integration contract is the standard `.cor` binary format rather than FFI, RPC, or language-specific bindings.

## Architecture

```text
.s source
   │
   ▼
Rust assembler
   │
   ├── lexer/parser
   ├── semantic validation
   ├── labels / expressions / macros
   └── big-endian encoder
   │
   ▼
.cor binary
   │
   ▼
Go VM
   │
   ├── circular 4096-byte arena
   ├── deterministic scheduler
   ├── 16-opcode execution engine
   ├── process lifecycle / CYCLE_TO_DIE
   └── snapshots + domain events
          │
          └── visualizer
```

More detail is available in [docs/architecture.md](docs/architecture.md).

## Requirements

- Rust stable
- Go 1.23+ for the VM
- Go 1.25+ for the optional Ebitengine visualizer
- Python 3 with `jsonschema` for repository-contract validation
- `make`
- `curl` and `unzip` for the official reference audit

No external runtime or parsing libraries are used by the assembler or mandatory VM. The graphical bonus is isolated in its own nested Go module and depends only on Ebitengine.

## Build

```bash
make build
```

This produces:

```text
./asm
./corewar
```

## Assembler

Compile a champion:

```bash
./asm player.s
```

The output is written to `player.cor`.

Help:

```bash
./asm
```

Errors are written to stderr and invalid programs do not leave a `.cor` output.

### Disassembler bonus

```bash
./asm --disassemble player.cor
```

or select an output path:

```bash
./asm --disassemble player.cor -o restored.s
```

The disassembler emits valid assembly that can be assembled back to the same binary.

### Arithmetic bonus

Numeric operands support:

```text
+  -  *  /  %  ( )
```

Example:

```asm
ld %(4 * (2 + 3)), r2
zjmp %(:loop - 2)
```

Normal operator precedence applies.

### Macro bonus

Simple textual macros use `$NAME`:

```asm
.define REG r3
.define VALUE 21

ld %($VALUE * 2), $REG
```

Macros may expand into operands or expression fragments. Recursive or undefined macros are rejected.

## Virtual machine

Run one to four champions:

```bash
./corewar player1.cor player2.cor
```

Dump memory at a specific cycle:

```bash
./corewar -d 100 player1.cor player2.cor
```

The dump contains 32 bytes per row.

Deterministic execution events can be written to stderr:

```bash
./corewar -v player1.cor player2.cor
```

### Visualizer bonus

The native graphical visualizer is built separately so the mandatory VM stays dependency-free:

```bash
make visual
./corewar-visual player1.cor player2.cor
```

It uses **Ebitengine 2.10.1** and renders the real VM through the existing immutable `Snapshot + []Event` boundary.

Controls:

- `Space` — play / pause.
- `S` — execute exactly one cycle while paused.
- `R` — reset the match.
- `1..4` — 1x / 10x / 50x / 100x simulation speed.
- Mouse hover — inspect an arena byte.
- Mouse click — pin the selected address.

The 1600×960 interface contains a 64×64 arena, player-colored memory ownership, active PC cursors, write-flash animations, cycle/CYCLE_TO_DIE telemetry, player status cards, winner state, and a byte/process inspector.

For a visual activity demo, assemble the dedicated bomber and run two copies:

```bash
./asm testdata/bomber.s
./corewar-visual testdata/bomber.cor testdata/bomber.cor
```

The demo is intentionally written for a two-player arena: each copy bombs two 376-byte bands around its code, forks a local bomber, and uses `lfork` to start an invading bomber in the opponent's copy. It is demonstration data only; the submitted champion remains `champions/terminator.s`.

A dependency-free terminal renderer remains available for diagnostics:

```bash
./corewar --visual player1.cor player2.cor
```

The Ebitengine code lives in `vm/visualizer/` as a nested module. It imports the VM read model but the VM never imports the renderer, so UI dependencies cannot affect mandatory execution or the reference audit.

Visualizer verification:

```bash
make visual-check
```

## Champion

The submitted champion is:

```text
champions/terminator.s
```

The audit compiles it with the official reference assembler and verifies it against `ameba` in the official reference VM in both player orders.

## Testing

Unit and contract tests:

```bash
make test
make contracts
```

Full evaluation-style audit:

```bash
make audit
```

The audit downloads the official 01-edu playground and checks:

1. CLI/help and failure behavior.
2. All official invalid assembler inputs.
3. Byte-for-byte assembler compatibility with `asm_ref`.
4. Disassembler round trips.
5. Arithmetic and macro bonuses.
6. Four-player `-d 10` VM behavior.
7. Corrupted `.cor` rejection.
8. VM memory dumps against `vm_ref` at multiple cycles.
9. The submitted champion against `ameba` in both orders using `vm_ref`.

CI runs the same `make audit` gate.

## Repository contract

This repository uses Agent Project Methodology 2.1. The navigation contract starts in [AGENTS.md](AGENTS.md), with module ownership in `agent/module-index.json`. Production code and runtime/reference verification remain the source of truth.

## Author

Nazar Yestayev
