#!/usr/bin/env python3
from __future__ import annotations

import re
import sys
from pathlib import Path

ROW = re.compile(r"^([0-9a-fA-F]{8})\s{2}((?:[0-9a-fA-F]{2}(?:\s+|$)){32})$")


def parse(path: Path) -> dict[int, bytes]:
    rows: dict[int, bytes] = {}
    for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        match = ROW.match(line.strip("\r"))
        if not match:
            continue
        address = int(match.group(1), 16)
        data = bytes(int(token, 16) for token in match.group(2).split())
        rows[address] = data
    return rows


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: compare_dumps.py REFERENCE LEARNER", file=sys.stderr)
        return 2

    ref_path = Path(sys.argv[1])
    learner_path = Path(sys.argv[2])
    ref = parse(ref_path)
    learner = parse(learner_path)

    if not ref:
        print(f"no memory rows found in reference dump: {ref_path}", file=sys.stderr)
        return 1
    if not learner:
        print(f"no memory rows found in learner dump: {learner_path}", file=sys.stderr)
        return 1

    for address in sorted(ref):
        expected = ref[address]
        actual = learner.get(address)
        if actual is None:
            print(f"learner dump is missing address 0x{address:08x}", file=sys.stderr)
            return 1
        if actual != expected:
            for offset, (a, b) in enumerate(zip(actual, expected)):
                if a != b:
                    print(
                        f"memory mismatch at 0x{address + offset:08x}: "
                        f"learner={a:02x} reference={b:02x}",
                        file=sys.stderr,
                    )
                    break
            return 1

    print(f"matched {len(ref)} reference rows ({len(ref) * 32} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
