# `.cor` binary format

All integer fields are big-endian.

| Offset | Size | Field |
|---:|---:|---|
| 0 | 4 | magic `0x00EA83F3` |
| 4 | 128 | NUL-padded champion name |
| 132 | 4 | zero padding |
| 136 | 4 | bytecode size |
| 140 | 2048 | NUL-padded description |
| 2188 | 4 | zero padding |
| 2192 | variable | bytecode, maximum 682 bytes |

The VM rejects a wrong magic, non-zero header padding, a declared bytecode size that does not equal the remaining file size, bytecode larger than 682 bytes, or a file smaller than the 2192-byte header.
