use corewar_asm::{assemble, disassemble, spec::HEADER_SIZE};

const AMEBA: &str = r#"
.name "ameba"
.description "not doing much"

        sti r1,%:hello,%1
        and r1,%0,r1

hello:  live %1
        zjmp %:hello
"#;

#[test]
fn ameba_matches_subject_bytecode() {
    let cor = assemble(AMEBA).expect("assemble");
    assert_eq!(cor.len(), HEADER_SIZE + 23);
    assert_eq!(
        &cor[HEADER_SIZE..],
        &[
            0x0b, 0x68, 0x01, 0x00, 0x0f, 0x00, 0x01, 0x06, 0x64, 0x01, 0x00, 0x00, 0x00, 0x00,
            0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x09, 0xff, 0xfb
        ]
    );
    assert_eq!(&cor[0..4], &[0x00, 0xea, 0x83, 0xf3]);
    assert_eq!(&cor[132..136], &[0, 0, 0, 0]);
    assert_eq!(&cor[136..140], &[0, 0, 0, 23]);
    assert_eq!(&cor[2188..2192], &[0, 0, 0, 0]);
}

#[test]
fn forward_and_backward_labels_are_relative_to_instruction_start() {
    let src = r#"
.name "labels"
.description "labels"
start:  zjmp %:end
        live %:start
end:    live %1
"#;
    let cor = assemble(src).unwrap();
    let code = &cor[HEADER_SIZE..];
    assert_eq!(&code[0..3], &[9, 0, 8]);
    assert_eq!(&code[3..8], &[1, 0xff, 0xff, 0xff, 0xfd]);
}

#[test]
fn arithmetic_and_macros_work() {
    let src = r#"
.define REG r3
.define BASE 4
.name "bonus"
.description "bonus"
ld %($BASE * (2 + 3)), $REG
add r3, r3, r3
live %(-1 + 2)
"#;
    let cor = assemble(src).unwrap();
    let code = &cor[HEADER_SIZE..];
    assert_eq!(&code[0..7], &[2, 0x90, 0, 0, 0, 20, 3]);
    assert_eq!(&code[7..12], &[4, 0x54, 3, 3, 3]);
    assert_eq!(&code[12..17], &[1, 0, 0, 0, 1]);
}

#[test]
fn disassembly_round_trip_preserves_binary_without_labels() {
    let src = r#"
.name "roundtrip"
.description "roundtrip"
ld %42, r2
st r2, 12
xor r2, %7, r3
zjmp %-5
nop r1
"#;
    let first = assemble(src).unwrap();
    let text = disassemble(&first).unwrap();
    let second = assemble(&text).unwrap();
    assert_eq!(first, second);
}

#[test]
fn utf8_headers_and_backslashes_round_trip() {
    let src = ".name \"тест\\champ\"\n.description \"описание\\path\"\nlive %1\n";
    let first = assemble(src).unwrap();
    let text = disassemble(&first).unwrap();
    let second = assemble(&text).unwrap();
    assert_eq!(first, second);
}

#[test]
fn invalid_programs_are_rejected() {
    let cases = [
        ".description \"x\"\nlive %1\n",
        ".name \"x\"\n.description \"x\"\nadd r1,r2,r17\n",
        ".name \"x\"\n.description \"x\"\nld r1,r2\n",
        ".name \"x\"\n.description \"x\"\nzjmp %:missing\n",
        ".name \"x\"\n.description \"x\"\nlive %1, %2\n",
        ".name \"x\"\n.description \"x\"\nlive %2147483648\n",
        ".name \"x\"\n.description \"x\"\nzjmp %32768\n",
        ".name \"x\"\n.description \"x\"\nld 32768, r1\n",
    ];
    for src in cases {
        assert!(assemble(src).is_err(), "expected error for {src:?}");
    }
}
