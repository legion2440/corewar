# Architecture

## Boundaries

Corewar is split at natural protocol boundaries rather than by implementation convenience.

```text
assembler (Rust) ── .cor ──> VM (Go) ── snapshot/events ──> renderer
```

The assembler and VM share no language-level code. The `.cor` format and instruction-set specification are the contract.

## Rust assembler

The assembler is a two-pass compiler.

```text
source
  -> macro preprocessing
  -> parser
  -> typed IR
  -> semantic validation / instruction sizing
  -> label table
  -> expression resolution
  -> binary encoding
```

Labels resolve to byte offsets relative to the start of the instruction that references them. The encoder creates the complete binary in memory before the CLI atomically writes the result, so a failed compile cannot leave a partial `.cor`.

The bonus disassembler uses the same opcode metadata and emits numeric relative operands. This avoids inventing synthetic labels and makes binary round trips deterministic.

## Go virtual machine

The VM is a single-threaded deterministic state machine. Goroutines are intentionally not used for Corewar processes: game processes are simulated CPU contexts, not operating-system threads.

A process contains:

- program counter;
- sixteen 32-bit registers;
- carry;
- last successful `live` cycle;
- currently latched opcode;
- remaining instruction wait cycles.

At instruction fetch time only the opcode and its cycle cost are latched. The pcode and operands are decoded from arena memory only when the instruction executes. This preserves self-modifying-code semantics: another process may overwrite an instruction's arguments while that instruction is waiting.

Forked processes are appended to the process list but excluded from the current cycle. Reverse iteration means the newly appended process executes first on the following cycle, matching the project specification.

## Arena

The arena owns exactly 4096 bytes and centralizes circular addressing and big-endian reads/writes. VM code does not manually wrap raw indexes.

An auxiliary owner map records the player responsible for the latest write to each byte. It is presentation metadata only and does not participate in VM execution.

## Visualization boundary

Rendering is downstream from VM semantics.

```text
VM.Step()
  -> immutable Snapshot
  -> []Event
  -> renderer
```

Snapshots contain memory, byte ownership, player state, and process state. Events describe instruction execution, writes, forks, live calls, deaths, and lifecycle checks.

Two renderers consume this boundary:

- `vm/internal/visual` is the dependency-free terminal diagnostic renderer used by the core module.
- `vm/visualizer` is a separate Go 1.25 nested module using Ebitengine. It renders the same real VM snapshots/events and therefore does not duplicate or mock VM execution.

The nested-module boundary is intentional: the mandatory VM retains zero graphical dependencies and its Go 1.23 toolchain contract, while the bonus renderer can use the current Ebitengine release independently.

## Verification

Correctness is checked at three levels:

1. **Unit tests** for parsing, expression evaluation, binary layout, arena wrapping, scheduling, instructions, and lifecycle.
2. **Contract/golden tests** for known bytecode such as the subject's `ameba`.
3. **Reference differential tests** against the official 01-edu `asm_ref` and `vm_ref`.

The strongest assembler condition is byte-for-byte equality with `asm_ref` over the official valid fixtures. VM verification compares arena dumps with `vm_ref` at multiple execution cycles and tests the submitted champion using the reference toolchain.
