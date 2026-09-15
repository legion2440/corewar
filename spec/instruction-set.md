# Corewar instruction set

This document mirrors the 01-edu Corewar subject and is the interoperability contract between the Rust assembler and Go VM.

Constants: `MEM_SIZE=4096`, `IDX_MOD=512`, `PLAYER_MAX_SIZE=682`, `REG_NUMBER=16`, `CYCLE_TO_DIE=1536`, `CYCLE_DELTA=50`, `NBR_LIVE=21`, `MAX_CHECKS=10`.

| Name | Opcode | Cycles | Pcode | IDX | Arguments |
|---|---:|---:|:---:|:---:|---|
| live | 1 | 10 | no | no | D |
| ld | 2 | 5 | yes | no | I/D, R |
| st | 3 | 5 | yes | no | R, R/I |
| add | 4 | 10 | yes | no | R, R, R |
| sub | 5 | 10 | yes | no | R, R, R |
| and | 6 | 6 | yes | no | R/I/D, R/I/D, R |
| or | 7 | 6 | yes | no | R/I/D, R/I/D, R |
| xor | 8 | 6 | yes | no | R/I/D, R/I/D, R |
| zjmp | 9 | 20 | no | yes | D |
| ldi | 10 | 25 | yes | yes | R/I/D, R/D, R |
| sti | 11 | 25 | yes | yes | R, R/I/D, R/D |
| fork | 12 | 800 | no | yes | D |
| lld | 13 | 10 | yes | no | I/D, R |
| lldi | 14 | 50 | yes | yes | R/I/D, R/D, R |
| lfork | 15 | 1000 | no | yes | D |
| nop | 16 | 2 | yes | no | R |

Argument coding byte uses `01=register`, `10=direct`, `11=indirect`, two bits per argument from the most-significant pair. Registers occupy 1 byte, indirect values 2 bytes, direct values 4 bytes unless the instruction has IDX, in which case direct values occupy 2 bytes.

All multi-byte integers are signed two's-complement values encoded big-endian. All addresses in the VM are relative to the current process PC. `ld`, `st`, `and`, `or`, `xor`, `ldi`, `sti`, `zjmp`, and `fork` apply `IDX_MOD` exactly as defined by the subject; long variants do not.


## Carry behavior verified against the reference VM

The reference VM has one detail that is easy to miss from the subject prose:

- `ld`, `lld`, `add`, `sub`, `and`, `or`, and `xor` update carry from the written result (`true` when the result is zero).
- `lldi` does **not** update carry.
- `zjmp` reads carry and does not modify it.

The local reference probes in `testdata/probes/` make the `lld` / `lldi` distinction observable through arena writes, and `scripts/audit.sh` checks those probes against the official `vm_ref`.
