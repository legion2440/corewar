# Corewar — Assembler, VM & Visualizer

Полная реализация задания 01-edu **Corewar** с разделением компонентов по подходящим языкам:

- **Rust** — assembler, parser, validation, binary encoder, disassembler, arithmetic expressions и macros.
- **Go** — детерминированная VM, scheduler, arena, исполнение инструкций, lifecycle и граница визуализации.
- **Ebitengine** — опциональный нативный графический visualizer.
- **Corewar assembly** — сдаваемый champion и demo-fixtures.

Компоненты взаимодействуют через стандартный бинарный формат `.cor`. Между Rust assembler и Go VM нет FFI или RPC.

· [English version](README.md)

## 📋 Содержание

- [🚀 Быстрый старт](#-быстрый-старт)
  - [WSL / Linux](#wsl--linux)
  - [Windows](#windows)
- [📝 О проекте](#-о-проекте)
- [🏗️ Архитектура](#️-архитектура)
- [🧰 Ассемблер](#-ассемблер)
- [⚙️ Виртуальная машина](#️-виртуальная-машина)
- [🎮 Бонусный visualizer](#-бонусный-visualizer)
- [🤖 Champion](#-champion)
- [🧪 Проверка и audit](#-проверка-и-audit)
- [📁 Структура проекта](#-структура-проекта)
- [⚠️ Примечания](#️-примечания)
- [🧑‍💻 Авторы](#-авторы)

## 🚀 Быстрый старт

### WSL / Linux

Это рекомендуемый вариант работы и **обязательная среда для полного reference audit**.

Требования:

- Rust stable + Cargo
- Go **1.23+** для обязательной VM
- Go **1.25+** для Ebitengine visualizer
- Python 3 + `jsonschema`
- GNU Make
- Bash
- `curl` и `unzip`

Сборка обязательной части:

```bash
make build
```

Unit-тесты:

```bash
make test
```

Сборка графического visualizer:

```bash
make visual
```

Полный audit в стиле оценки:

```bash
make audit
```

> **Важно:** `make audit` запускается только в **Linux / WSL**. Официальные `asm_ref` и `vm_ref` из playground 01-edu — Linux ELF binaries и нативно в Windows не запускаются.

По умолчанию audit сам скачивает официальный playground. Можно использовать постоянную копию:

```bash
PLAYGROUND_DIR=$HOME/corewar-playground/playground make audit
```

### Windows

Assembler, VM и графический visualizer можно собирать и запускать нативно в Windows.

Корневой `Makefile` рассчитан на Unix shell, поэтому в native Windows используй прямые PowerShell-команды ниже. Git Bash можно использовать для запуска получившихся `.exe`, но полный `make audit` всё равно запускается в WSL/Linux.

#### PowerShell

Assembler:

```powershell
Push-Location assembler
cargo build --release
Copy-Item target\release\asm.exe ..\asm.exe
Pop-Location
```

VM:

```powershell
Push-Location vm
go build -o ..\corewar.exe .\cmd\corewar
Pop-Location
```

Visualizer:

```powershell
Push-Location vm\visualizer
go build -o ..\..\corewar-visual.exe .
Pop-Location
```

Тесты:

```powershell
Push-Location assembler
cargo test
Pop-Location

Push-Location vm
go test ./...
Pop-Location

python scripts\validate_agent_contracts.py
```

Для полного reference audit переключись в WSL:

```bash
wsl
cd /mnt/d/path/to/corewar
make audit
```

## 📝 О проекте

Corewar запускает небольшие программы в общей кольцевой **arena на 4096 байт**. Каждый игрок начинает с одного процесса, 16 регистров, program counter и carry flag.

Процессы исполняют 16 инструкций Corewar, включая:

- `live` — отметить игрока как живого;
- `ld`, `st`, `ldi`, `sti` — работа с регистрами и памятью;
- `add`, `sub`, `and`, `or`, `xor` — арифметика и логика;
- `zjmp` — условный относительный jump;
- `fork`, `lfork` — создание процессов;
- `lld`, `lldi` — long load варианты;
- `nop`.

VM периодически выполняет life-check по правилам `CYCLE_TO_DIE`, `CYCLE_DELTA`, `NBR_LIVE` и `MAX_CHECKS`. Процессы, которые вовремя не выполнили `live`, удаляются. Когда процессов не остаётся, побеждает игрок, которому последним был засчитан валидный `live`.

## 🏗️ Архитектура

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
   ├── кольцевая arena 4096 bytes
   ├── deterministic scheduler
   ├── 16-opcode execution engine
   ├── process lifecycle
   └── Snapshot + []Event
              │
              ├── terminal diagnostics
              └── Ebitengine visualizer
```

Обязательные assembler и VM не зависят от внешних runtime libraries. Графический bonus вынесен в отдельный nested Go module.

Подробнее: [docs/architecture.md](docs/architecture.md).

## 🧰 Ассемблер

Сборка champion:

```bash
./asm player.s
```

Результат создаётся как `player.cor`.

Невалидный исходник завершается ошибкой и не оставляет `.cor`.

### Disassembler bonus

```bash
./asm --disassemble player.cor
```

С указанием выходного файла:

```bash
./asm --disassemble player.cor -o restored.s
```

Полученный assembly можно собрать обратно в тот же бинарник.

### Arithmetic bonus

Числовые выражения поддерживают:

```text
+  -  *  /  %  ( )
```

Пример:

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

Undefined и recursive macros отклоняются.

## ⚙️ Виртуальная машина

Запуск от одного до четырёх champions:

```bash
./corewar player1.cor player2.cor
```

Dump arena на выбранном цикле:

```bash
./corewar -d 100 player1.cor player2.cor
```

Dump содержит 32 байта на строку.

Детерминированные execution events можно вывести в `stderr`:

```bash
./corewar -v player1.cor player2.cor
```

Flags можно ставить до, между или после путей к champion-файлам.

## 🎮 Бонусный visualizer

Сборка:

```bash
make visual
```

Запуск:

```bash
./corewar-visual player1.cor player2.cor
```

Visualizer использует **Ebitengine 2.10.1** и отображает настоящую VM через ту же immutable-границу `Snapshot + []Event`.

Управление:

- `Space` — play / pause
- `S` — один цикл в pause
- `R` — reset
- `1..4` — 1x / 10x / 50x / 100x
- mouse hover — inspect arena byte
- mouse click — закрепить address

Интерфейс 1600×960 показывает:

- arena 64×64;
- memory ownership цветами игроков;
- активные process cursors;
- write flashes;
- cycle и `CYCLE_TO_DIE` telemetry;
- состояние процессов и игроков;
- winner;
- byte/process inspector.

### Bomber demo

Обычные playground-fixtures маленькие и часто почти не пишут в память. Для наглядного матча есть отдельный bomber:

```bash
./asm testdata/bomber.s
./corewar-visual testdata/bomber.cor testdata/bomber.cor
```

Каждая копия постоянно пишет по arena, создаёт локальный процесс через `fork` и запускает ещё один процесс в половине противника через `lfork`. Это только demo для visualizer; сдаваемый champion остаётся `champions/terminator.s`.

Проверка visualizer:

```bash
make visual-check
```

## 🤖 Champion

Сдаваемый champion:

```text
champions/terminator.s
```

Стратегия использует `lfork`, чтобы эксплуатировать поведение предоставленного `ameba` и перенаправлять будущие `live` credits.

Audit собирает champion официальным reference assembler и проверяет полный матч через reference VM в обоих порядках игроков.

## 🧪 Проверка и audit

Локальные проверки:

```bash
make build
make test
make contracts
make visual-check
```

Полный reference audit — **только Linux / WSL**:

```bash
make audit
```

Audit проверяет:

1. CLI/help и error behavior.
2. Официальные invalid assembler inputs.
3. Byte-for-byte совместимость с `asm_ref`, включая границы header и bomber demo.
4. Disassembler round trips.
5. Arithmetic и macro bonuses.
6. Four-player VM dump format.
7. Отклонение повреждённых `.cor`.
8. Differential memory dumps против `vm_ref`, включая cycle 0 и carry probes.
9. Полные матчи: `terminator` против `ameba` в обоих порядках плюс four-player reference matches.

Текущий audit покрывает **41 официальный assembler fixture** и **41 disassembler round trip**.

CI запускает тот же mandatory audit gate.

## 📁 Структура проекта

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
│   └── terminator.s         # сдаваемый champion
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

## ⚠️ Примечания

- Обязательные assembler и VM сделаны детерминированными и с минимальными зависимостями.
- Rust assembler и Go VM взаимодействуют только через формат `.cor`.
- Графический visualizer изолирован от mandatory execution и не меняет семантику VM.
- `make audit` требует Linux / WSL, потому что официальные reference executables — Linux binaries.
- Сгенерированные `.cor` — build artifacts; их можно заново получить из `.s`.

## 🧑‍💻 Авторы

- Nazar Yestayev ([**@nyestaye**](https://01.tomorrow-school.ai/intra/astanahub/users/4468))
- Sultan Yersultan ([**@syersult**](https://01.tomorrow-school.ai/intra/astanahub/users/4423))
- Daniyar Shadykhanov ([**@dshadykh**](https://01.tomorrow-school.ai/intra/astanahub/users/2418))
- Maksat Kapan ([**@mkapan**](https://01.tomorrow-school.ai/intra/astanahub/users/3597))
