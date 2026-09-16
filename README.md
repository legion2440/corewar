# Corewar — Assembler, VM & Visualizer

A complete implementation of the 01-edu **Corewar** project with a deliberately polyglot architecture:

- **Rust** — assembler, parser, validation, binary encoder, disassembler, arithmetic expressions, and macros.
- **Go** — deterministic virtual machine, scheduler, arena, instruction execution, lifecycle rules, and the visualization boundary.
- **Ebitengine** — optional native graphical visualizer.
- **Corewar assembly** — the submitted champion and demo fixtures.

The components communicate through the standard `.cor` binary format. There is no FFI or RPC layer between the Rust assembler and the Go VM.

· [Русская версия](README_RU.md)

## 📋 TOC

- [🚀 Quick start](#-quick-start)
  - [WSL / Linux](#wsl--linux)
  - [Windows](#windows)
- [📝 About](#-about)
- [🏗️ Architecture](#️-architecture)
- [🧰 Assembler](#-assembler)
- [⚙️ Virtual machine](#️-virtual-machine)
- [🎮 Visualizer bonus](#-visualizer-bonus)
- [🤖 Champion](#-champion)
- [🧪 Verification and audit](#-verification-and-audit)
- [📁 Project structure](#-project-structure)
- [⚠️ Notes](#️-notes)
- [🧑‍💻 Authors](#-authors)

## 🚀 Quick start

### WSL / Linux

This is the recommended workflow and the **required environment for the full reference audit**.

Requirements:

- Rust stable + Cargo
- Go **1.23+** for the mandatory VM
- Go **1.25+** for the optional Ebitengine visualizer
- Python 3 + `jsonschema`
- GNU Make
- Bash
- `curl` and `unzip`

Build the mandatory binaries:

```bash
make build
```

Run unit tests:

```bash
make test
```

Build the native visualizer:

```bash
make visual
```

Run the full evaluation-style audit:

```bash
make audit
```

> **Important:** `make audit` is intended for **Linux / WSL only**. The official `asm_ref` and `vm_ref` supplied by 01-edu are Linux ELF binaries and do not run natively on Windows.

The audit downloads the official playground automatically. A persistent copy can also be supplied:

```bash
PLAYGROUND_DIR=$HOME/corewar-playground/playground make audit
```

### Windows

The assembler, VM, and graphical visualizer can be built and used natively on Windows.

#### Git Bash + GNU Make

If Rust, Go, Python, GNU Make, and Git Bash are available:

```bash
make build
make test
make visual
```

Run:

```bash
./asm.exe testdata/ameba.s
./corewar.exe testdata/ameba.cor
./corewar-visual.exe testdata/ameba.cor champions/terminator.cor
```

#### PowerShell without Make

Build the assembler:

```powershell
Push-Location assembler
cargo build --release
Copy-Item target\release\asm.exe ..\asm.exe
Pop-Location
```

Build the VM:

```powershell
Push-Location vm
go build -o ..\corewar.exe .\cmd\corewar
Pop-Location
```

Build the visualizer:

```powershell
Push-Location vm\visualizer
go build -o ..\..\corewar-visual.exe .
Pop-Location
```

Run tests:

```powershell
Push-Location assembler
cargo test
Pop-Location

Push-Location vm
go test ./...
Pop-Location

python scripts\validate_agent_contracts.py
```

For the reference audit, switch to WSL:

```bash
wsl
cd /mnt/d/path/to/corewar
make audit
```

## 📝 About

Corewar runs small programs inside a shared circular **4096-byte arena**. Each player starts with one process, 16 registers, a program counter, and a carry flag.

Processes execute the 16 Corewar instructions, including:

- `live` — report a player as alive;
- `ld`, `st`, `ldi`, `sti` — move data between registers and arena memory;
- `add`, `sub`, `and`, `or`, `xor` — arithmetic and logic;
- `zjmp` — conditional relative jump;
- `fork`, `lfork` — create new processes;
- `lld`, `lldi` — long load variants;
- `nop`.

The VM periodically performs life checks using `CYCLE_TO_DIE`, `CYCLE_DELTA`, `NBR_LIVE`, and `MAX_CHECKS`. Processes that do not execute `live` in time are removed. When no processes remain, the last player credited by a valid `live` wins.

## 🏗️ Architecture

```text
.s source
   │
   ▼
Rust assembler
   │
   ├── parser / validation
   ├── labels
   ├── arithmetic expressions
   ├── macros
   ├── disassembler
   └── big-endian .cor encoder
   │
   ▼
.cor binary
   │
   ▼
Go VM
   │
   ├── 4096-byte circular arena
   ├── deterministic scheduler
   ├── 16-opcode execution engine
   ├── process lifecycle
   └── Snapshot + []Event
              │
              ├── terminal diagnostics
              └── Ebitengine visualizer
```

The mandatory assembler and VM do not depend on external runtime libraries. The graphical bonus lives in a separate nested Go module.

Detailed architecture: [docs/architecture.md](docs/architecture.md).

## 🧰 Assembler

Compile a champion:

```bash
./asm player.s
```

The result is written to `player.cor`.

Invalid source files fail with an error and do not leave a generated `.cor` file.

### Disassembler bonus

```bash
./asm --disassemble player.cor
```

Custom output path:

```bash
./asm --disassemble player.cor -o restored.s
```

The emitted assembly can be assembled back to the same binary.

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

### Macro bonus

```asm
.define REG r3
.define VALUE 21

ld %($VALUE * 2), $REG
```

Undefined and recursive macros are rejected.

## ⚙️ Virtual machine

Run one to four champions:

```bash
./corewar player1.cor player2.cor
```

Dump the arena at a selected cycle:

```bash
./corewar -d 100 player1.cor player2.cor
```

The dump contains 32 bytes per row.

Deterministic execution events can be written to `stderr`:

```bash
./corewar -v player1.cor player2.cor
```

Flags may appear before, between, or after champion paths.

## 🎮 Visualizer bonus

Build:

```bash
make visual
```

Run:

```bash
./corewar-visual player1.cor player2.cor
```

The visualizer uses **Ebitengine 2.10.1** and renders the real VM through the same immutable `Snapshot + []Event` boundary.

Controls:

- `Space` — play / pause
- `S` — execute one cycle while paused
- `R` — reset
- `1..4` — 1x / 10x / 50x / 100x
- mouse hover — inspect an arena byte
- mouse click — pin an address

The 1600×960 interface contains:

- a 64×64 arena;
- player-colored memory ownership;
- active process cursors;
- write flashes;
- cycle and `CYCLE_TO_DIE` telemetry;
- process/player status;
- winner state;
- byte/process inspector.

### Bomber demo

The ordinary playground fixtures are intentionally small and often perform almost no memory writes. For a visibly active match, use the dedicated demo bomber:

```bash
./asm testdata/bomber.s
./corewar-visual testdata/bomber.cor testdata/bomber.cor
```

Each copy repeatedly writes across the arena, forks a local bomber, and uses `lfork` to start another process in the opponent's half. The demo exists only for visualization; the submitted champion is still `champions/terminator.s`.

Visualizer checks:

```bash
make visual-check
```

## 🤖 Champion

The submitted champion is:

```text
champions/terminator.s
```

Its strategy uses `lfork` to exploit the behavior of the provided `ameba` champion and redirect future `live` credits.

The audit assembles the champion with the official reference assembler and verifies full-match behavior against the official reference VM in both player orders.

## 🧪 Verification and audit

Local checks:

```bash
make build
make test
make contracts
make visual-check
```

Full reference audit — **Linux / WSL only**:

```bash
make audit
```

The audit verifies:

1. CLI/help and failure behavior.
2. Official invalid assembler inputs.
3. Byte-for-byte assembler compatibility with `asm_ref`, including header boundaries and the bomber demo.
4. Disassembler round trips.
5. Arithmetic and macro bonuses.
6. Four-player VM dump formatting.
7. Corrupted `.cor` rejection.
8. Differential memory dumps against `vm_ref`, including cycle 0 and carry probes.
9. Full-match compatibility: `terminator` vs `ameba` in both orders plus four-player reference matches.

Current audit coverage includes **41 official assembler fixtures** and **41 disassembler round trips**.

CI runs the same mandatory audit gate.

## 📁 Project structure

```text
corewar/
├── assembler/               # Rust assembler + disassembler
│   ├── src/
│   └── tests/
├── vm/
│   ├── cmd/corewar/         # VM CLI
│   ├── internal/
│   │   ├── arena/
│   │   ├── champion/
│   │   ├── corewar/
│   │   └── visual/
│   └── visualizer/          # Ebitengine nested Go module
├── champions/
│   └── terminator.s         # submitted champion
├── testdata/
│   ├── ameba.s
│   ├── bomber.s             # visual activity demo
│   └── probes/
├── scripts/
│   ├── audit.sh
│   ├── compare_dumps.py
│   └── validate_agent_contracts.py
├── spec/
├── docs/
├── agent/
├── Makefile
├── README.md
└── README_RU.md
```

## ⚠️ Notes

- The mandatory assembler and VM are intentionally dependency-light and deterministic.
- The Rust assembler and Go VM communicate only through the `.cor` format.
- The graphical visualizer is isolated from mandatory execution and cannot change VM semantics.
- `make audit` requires Linux / WSL because the official reference executables are Linux binaries.
- Generated `.cor` files are build artifacts and can be recreated from their `.s` sources.

## 🧑‍💻 Authors

- Nazar Yestayev [**@nyestaye**](https://01.tomorrow-school.ai/intra/astanahub/users/4468)
- Sultan Yersultan [**@syersult**](https://01.tomorrow-school.ai/intra/astanahub/users/4423)
- Daniyar Shadykhanov [**@dshadykh**](https://01.tomorrow-school.ai/intra/astanahub/users/2418)
- Maksat Kapan [**@mkapan**](https://01.tomorrow-school.ai/intra/astanahub/users/3597)
