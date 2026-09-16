#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail() {
  printf 'audit: FAIL: %s\n' "$*" >&2
  exit 1
}

expect_failure() {
  local stdout stderr
  stdout="$(mktemp)"
  stderr="$(mktemp)"
  if "$@" >"$stdout" 2>"$stderr"; then
    rm -f "$stdout" "$stderr"
    fail "command unexpectedly succeeded: $*"
  fi
  [[ -s "$stderr" ]] || {
    rm -f "$stdout" "$stderr"
    fail "failing command produced no stderr: $*"
  }
  rm -f "$stdout" "$stderr"
}

PLAYGROUND_DIR="${PLAYGROUND_DIR:-}"
TMP_ROOT=""
if [[ -z "$PLAYGROUND_DIR" ]]; then
  command -v curl >/dev/null || fail "curl is required to fetch the official playground"
  command -v unzip >/dev/null || fail "unzip is required to extract the official playground"
  TMP_ROOT="$(mktemp -d)"
  trap 'rm -rf "$TMP_ROOT"' EXIT
  curl -fsSL     https://raw.githubusercontent.com/01-edu/public/master/subjects/corewar/data/playground.zip     -o "$TMP_ROOT/playground.zip"
  unzip -q "$TMP_ROOT/playground.zip" -d "$TMP_ROOT"
  PLAYGROUND_DIR="$TMP_ROOT/playground"
fi

ASM_REF="$PLAYGROUND_DIR/asm_ref"
VM_REF="$PLAYGROUND_DIR/vm_ref"
PLAYERS="$PLAYGROUND_DIR/players_src"
[[ -f "$ASM_REF" && -f "$VM_REF" && -d "$PLAYERS" ]] || fail "invalid playground path: $PLAYGROUND_DIR"
chmod +x "$ASM_REF" "$VM_REF"

[[ -x ./asm ]] || fail "./asm is missing; run make build"
[[ -x ./corewar ]] || fail "./corewar is missing; run make build"

printf '%s\n' '[1/9] CLI and error contract'
./asm | grep -qi 'usage' || fail "assembler without args did not print usage"
./corewar | grep -qi 'usage' || fail "VM without args did not print usage"
expect_failure ./asm "$ROOT/does-not-exist.s"
expect_failure ./corewar "$ROOT/does-not-exist.cor"

printf '%s\n' '[2/9] Reject official invalid assembler inputs'
invalid_work="$(mktemp -d)"
for src in "$PLAYERS"/invalid/*; do
  base="$(basename "$src")"
  cp "$src" "$invalid_work/$base"
  out="$invalid_work/${base%.*}.cor"
  rm -f "$out"
  expect_failure ./asm "$invalid_work/$base"
  [[ ! -e "$out" ]] || fail "invalid input created $out"
done

special_src="$PLAYERS/specials/pierino_r_out_of_range.s"
[[ -f "$special_src" ]] || fail "missing official special fixture: $special_src"
cp "$special_src" "$invalid_work/pierino_r_out_of_range.s"
special_out="$invalid_work/pierino_r_out_of_range.cor"
rm -f "$special_out"
expect_failure ./asm "$invalid_work/pierino_r_out_of_range.s"
[[ ! -e "$special_out" ]] || fail "special invalid input created $special_out"
expect_failure "$ASM_REF" "$invalid_work/pierino_r_out_of_range.s"
[[ ! -e "$special_out" ]] || fail "reference assembler unexpectedly created $special_out"
rm -rf "$invalid_work"

printf '%s\n' '[3/9] Byte-for-byte assembler compatibility'
ours_dir="$(mktemp -d)"
ref_dir="$(mktemp -d)"
valid_count=0
for src in "$PLAYERS"/*.s; do
  [[ -e "$src" ]] || continue
  base="$(basename "$src")"
  cp "$src" "$ours_dir/$base"
  cp "$src" "$ref_dir/$base"
  ./asm "$ours_dir/$base" >/dev/null
  "$ASM_REF" "$ref_dir/$base" >/dev/null
  ours_cor="$ours_dir/${base%.s}.cor"
  ref_cor="$ref_dir/${base%.s}.cor"
  cmp -s "$ours_cor" "$ref_cor" || {
    cmp -l "$ours_cor" "$ref_cor" | head -20 >&2 || true
    fail "assembler output differs for $base"
  }
  valid_count=$((valid_count + 1))
done
(( valid_count > 0 )) || fail "no official valid players found"
rm -rf "$ours_dir" "$ref_dir"

header_dir="$(mktemp -d)"
mkdir -p "$header_dir/src" "$header_dir/ours" "$header_dir/ref"
python3 - "$header_dir/src" <<'PY'
from pathlib import Path
import sys

root = Path(sys.argv[1])
for n in (128, 129):
    (root / f"name-{n}.s").write_text(
        f'.name "{"N" * n}"\n'
        '.description "test"\n'
        'live %1\n',
        encoding="utf-8",
    )
for n in (2048, 2049):
    (root / f"desc-{n}.s").write_text(
        '.name "test"\n'
        f'.description "{"D" * n}"\n'
        'live %1\n',
        encoding="utf-8",
    )
PY

for stem in name-128 desc-2048; do
  cp "$header_dir/src/$stem.s" "$header_dir/ours/$stem.s"
  cp "$header_dir/src/$stem.s" "$header_dir/ref/$stem.s"
  ./asm "$header_dir/ours/$stem.s" >/dev/null
  "$ASM_REF" "$header_dir/ref/$stem.s" >/dev/null
  cmp -s "$header_dir/ours/$stem.cor" "$header_dir/ref/$stem.cor" ||
    fail "header boundary output differs for $stem"
done

for stem in name-129 desc-2049; do
  cp "$header_dir/src/$stem.s" "$header_dir/ours/$stem.s"
  cp "$header_dir/src/$stem.s" "$header_dir/ref/$stem.s"
  expect_failure ./asm "$header_dir/ours/$stem.s"
  expect_failure "$ASM_REF" "$header_dir/ref/$stem.s"
  [[ ! -e "$header_dir/ours/$stem.cor" ]] || fail "invalid header boundary created learner $stem.cor"
  [[ ! -e "$header_dir/ref/$stem.cor" ]] || fail "invalid header boundary created reference $stem.cor"
done
rm -rf "$header_dir"

printf '%s\n' '[4/9] Disassembler round-trip bonus'
round_dir="$(mktemp -d)"
round_count=0
for src in "$PLAYERS"/*.s; do
  [[ -e "$src" ]] || continue
  base="$(basename "$src")"
  cp "$src" "$round_dir/$base"
  ./asm "$round_dir/$base" >/dev/null
  original="$round_dir/${base%.s}.cor"
  ./asm --disassemble "$original" -o "$round_dir/${base%.s}.dis.s" >/dev/null
  cp "$original" "$round_dir/original.cor"
  ./asm "$round_dir/${base%.s}.dis.s" -o "$round_dir/rebuilt.cor" >/dev/null
  cmp -s "$round_dir/original.cor" "$round_dir/rebuilt.cor" || fail "disassembler round-trip differs for $base"
  round_count=$((round_count + 1))
done
rm -rf "$round_dir"

printf '%s\n' '[5/9] Arithmetic and macro bonuses'
bonus_dir="$(mktemp -d)"
cat >"$bonus_dir/bonus.s" <<'EOF'
.define REG r3
.define BASE 4
.name "bonus"
.description "macro and arithmetic bonus"
ld %($BASE * (2 + 3)), $REG
add r3, r3, r3
live %(-1 + 2)
EOF
./asm "$bonus_dir/bonus.s" >/dev/null
./asm --disassemble "$bonus_dir/bonus.cor" -o "$bonus_dir/bonus.dis.s" >/dev/null
./asm "$bonus_dir/bonus.dis.s" -o "$bonus_dir/bonus.rebuilt.cor" >/dev/null
cmp -s "$bonus_dir/bonus.cor" "$bonus_dir/bonus.rebuilt.cor" || fail "bonus program does not round-trip"
rm -rf "$bonus_dir"

printf '%s\n' '[6/9] VM four-player dump contract'
vm_dir="$(mktemp -d)"
mapfile -t vm_sources < <(find "$PLAYERS" -maxdepth 1 -type f -name '*.s' | sort | head -4)
(( ${#vm_sources[@]} == 4 )) || fail "need four official players"
vm_players=()
for src in "${vm_sources[@]}"; do
  base="$(basename "$src")"
  cp "$src" "$vm_dir/$base"
  "$ASM_REF" "$vm_dir/$base" >/dev/null
  vm_players+=("$vm_dir/${base%.s}.cor")
done
./corewar -d 10 "${vm_players[@]}" >"$vm_dir/dump.txt" 2>"$vm_dir/dump.err"
[[ "$(grep -c '^Player [1-4] (' "$vm_dir/dump.txt")" -eq 4 ]] || fail "VM did not introduce four players"
[[ "$(grep -c '^[0-9a-fA-F]\{8\}  ' "$vm_dir/dump.txt")" -eq 128 ]] || fail "VM dump is not 128 rows of 32 bytes"
rm -rf "$vm_dir"

printf '%s\n' '[7/9] VM corrupted binary rejection'
corrupt_dir="$(mktemp -d)"
cp "$PLAYERS/crab.cor" "$corrupt_dir/base.cor"
cp "$corrupt_dir/base.cor" "$corrupt_dir/bad-magic.cor"
printf '\xff' | dd of="$corrupt_dir/bad-magic.cor" bs=1 seek=0 conv=notrunc status=none
expect_failure ./corewar "$corrupt_dir/bad-magic.cor"
head -c 100 "$corrupt_dir/base.cor" >"$corrupt_dir/truncated.cor"
expect_failure ./corewar "$corrupt_dir/truncated.cor"
python3 - "$corrupt_dir/base.cor" "$corrupt_dir/bad-size.cor" <<'PY'
import pathlib, struct, sys
data = bytearray(pathlib.Path(sys.argv[1]).read_bytes())
data[136:140] = struct.pack(">I", (struct.unpack(">I", data[136:140])[0] + 1))
pathlib.Path(sys.argv[2]).write_bytes(data)
PY
expect_failure ./corewar "$corrupt_dir/bad-size.cor"
rm -rf "$corrupt_dir"

printf '%s\n' '[8/9] Reference VM differential dumps'
diff_dir="$(mktemp -d)"
diff_sources=(
  "$PLAYERS/pierino_add.s"
  "$PLAYERS/pierino_ld.s"
  "$PLAYERS/pierino_ldi_reg_reg.s"
  "$PLAYERS/pierino_sti_reg_reg_reg.s"
)
diff_players=()
for src in "${diff_sources[@]}"; do
  base="$(basename "$src")"
  cp "$src" "$diff_dir/$base"
  "$ASM_REF" "$diff_dir/$base" >/dev/null
  diff_players+=("$diff_dir/${base%.s}.cor")
done
for cycle in 10 100 500 1600; do
  if ! "$VM_REF" -d "$cycle" "${diff_players[@]}" >"$diff_dir/ref-$cycle.txt" 2>"$diff_dir/ref-$cycle.err"; then
    cat "$diff_dir/ref-$cycle.err" >&2
    fail "reference VM failed at cycle $cycle"
  fi
  if ! ./corewar -d "$cycle" "${diff_players[@]}" >"$diff_dir/ours-$cycle.txt" 2>"$diff_dir/ours-$cycle.err"; then
    cat "$diff_dir/ours-$cycle.err" >&2
    fail "learner VM failed at cycle $cycle"
  fi
  python3 scripts/compare_dumps.py "$diff_dir/ref-$cycle.txt" "$diff_dir/ours-$cycle.txt" ||
    fail "VM memory diverges from reference at cycle $cycle"
done

for probe in testdata/probes/lld-carry.s testdata/probes/lldi-carry.s; do
  base="$(basename "$probe" .s)"
  cp "$probe" "$diff_dir/$base.s"
  "$ASM_REF" "$diff_dir/$base.s" >/dev/null
  if ! "$VM_REF" -d 100 "$diff_dir/$base.cor" >"$diff_dir/ref-$base.txt" 2>"$diff_dir/ref-$base.err"; then
    cat "$diff_dir/ref-$base.err" >&2
    fail "reference VM failed for $base probe"
  fi
  if ! ./corewar -d 100 "$diff_dir/$base.cor" >"$diff_dir/ours-$base.txt" 2>"$diff_dir/ours-$base.err"; then
    cat "$diff_dir/ours-$base.err" >&2
    fail "learner VM failed for $base probe"
  fi
  python3 scripts/compare_dumps.py "$diff_dir/ref-$base.txt" "$diff_dir/ours-$base.txt" ||
    fail "VM differs from reference for $base probe"
done

rm -rf "$diff_dir"

printf '%s\n' '[9/9] Champion and full-match VM compatibility'
fight_dir="$(mktemp -d)"
cp champions/terminator.s "$fight_dir/terminator.s"
cp testdata/ameba.s "$fight_dir/ameba.s"
"$ASM_REF" "$fight_dir/terminator.s" >/dev/null
"$ASM_REF" "$fight_dir/ameba.s" >/dev/null

"$VM_REF" "$fight_dir/terminator.cor" "$fight_dir/ameba.cor" >"$fight_dir/ref-a.txt" 2>"$fight_dir/ref-a.err"
"$VM_REF" "$fight_dir/ameba.cor" "$fight_dir/terminator.cor" >"$fight_dir/ref-b.txt" 2>"$fight_dir/ref-b.err"
./corewar "$fight_dir/terminator.cor" "$fight_dir/ameba.cor" >"$fight_dir/ours-a.txt" 2>"$fight_dir/ours-a.err"
./corewar "$fight_dir/ameba.cor" "$fight_dir/terminator.cor" >"$fight_dir/ours-b.txt" 2>"$fight_dir/ours-b.err"

grep -qi 'winner.*terminator' "$fight_dir/ref-a.txt" || {
  cat "$fight_dir/ref-a.txt" >&2
  fail "terminator did not beat ameba from first position on reference VM"
}
grep -qi 'winner.*terminator' "$fight_dir/ref-b.txt" || {
  cat "$fight_dir/ref-b.txt" >&2
  fail "terminator did not beat ameba from second position on reference VM"
}

ref_a="$(tail -n 1 "$fight_dir/ref-a.txt")"
ref_b="$(tail -n 1 "$fight_dir/ref-b.txt")"
ours_a="$(tail -n 1 "$fight_dir/ours-a.txt")"
ours_b="$(tail -n 1 "$fight_dir/ours-b.txt")"
[[ "$ours_a" == "$ref_a" ]] || {
  printf 'reference: %s\nlearner:   %s\n' "$ref_a" "$ours_a" >&2
  fail "full-match result differs from reference in terminator/ameba order"
}
[[ "$ours_b" == "$ref_b" ]] || {
  printf 'reference: %s\nlearner:   %s\n' "$ref_b" "$ours_b" >&2
  fail "full-match result differs from reference in ameba/terminator order"
}

rm -rf "$fight_dir"

printf 'audit: PASS (%d official assembler fixtures, %d disassembler round-trips)\n' "$valid_count" "$round_count"
