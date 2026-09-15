pub const IND_SIZE: usize = 2;
pub const REG_SIZE: usize = 4;
pub const DIR_SIZE: usize = REG_SIZE;
pub const MAX_PLAYERS: usize = 4;
pub const MEM_SIZE: usize = 4096;
pub const IDX_MOD: usize = MEM_SIZE / 8;
pub const PLAYER_MAX_SIZE: usize = MEM_SIZE / 6;
pub const REG_NUMBER: u8 = 16;
pub const PROG_NAME_LENGTH: usize = 128;
pub const DESCRIPTION_LENGTH: usize = 2048;
pub const COREWAR_EXEC_SIGNATURE: u32 = 0x00ea83f3;
pub const HEADER_SIZE: usize = 4 + PROG_NAME_LENGTH + 4 + 4 + DESCRIPTION_LENGTH + 4;

pub const ARG_REG: u8 = 0b001;
pub const ARG_DIR: u8 = 0b010;
pub const ARG_IND: u8 = 0b100;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct OpSpec {
    pub name: &'static str,
    pub opcode: u8,
    pub cycles: u16,
    pub has_pcode: bool,
    pub has_idx: bool,
    pub allowed: &'static [u8],
}

const LIVE: &[u8] = &[ARG_DIR];
const LD: &[u8] = &[ARG_IND | ARG_DIR, ARG_REG];
const ST: &[u8] = &[ARG_REG, ARG_REG | ARG_IND];
const RRR: &[u8] = &[ARG_REG, ARG_REG, ARG_REG];
const LOGIC: &[u8] = &[
    ARG_REG | ARG_IND | ARG_DIR,
    ARG_REG | ARG_IND | ARG_DIR,
    ARG_REG,
];
const ZJMP: &[u8] = &[ARG_DIR];
const LDI: &[u8] = &[ARG_REG | ARG_IND | ARG_DIR, ARG_REG | ARG_DIR, ARG_REG];
const STI: &[u8] = &[ARG_REG, ARG_REG | ARG_IND | ARG_DIR, ARG_REG | ARG_DIR];
const FORK: &[u8] = &[ARG_DIR];
const LLD: &[u8] = &[ARG_IND | ARG_DIR, ARG_REG];
const LLDI: &[u8] = &[ARG_REG | ARG_IND | ARG_DIR, ARG_REG | ARG_DIR, ARG_REG];
const NOP: &[u8] = &[ARG_REG];

pub const OPS: [OpSpec; 16] = [
    OpSpec {
        name: "live",
        opcode: 1,
        cycles: 10,
        has_pcode: false,
        has_idx: false,
        allowed: LIVE,
    },
    OpSpec {
        name: "ld",
        opcode: 2,
        cycles: 5,
        has_pcode: true,
        has_idx: false,
        allowed: LD,
    },
    OpSpec {
        name: "st",
        opcode: 3,
        cycles: 5,
        has_pcode: true,
        has_idx: false,
        allowed: ST,
    },
    OpSpec {
        name: "add",
        opcode: 4,
        cycles: 10,
        has_pcode: true,
        has_idx: false,
        allowed: RRR,
    },
    OpSpec {
        name: "sub",
        opcode: 5,
        cycles: 10,
        has_pcode: true,
        has_idx: false,
        allowed: RRR,
    },
    OpSpec {
        name: "and",
        opcode: 6,
        cycles: 6,
        has_pcode: true,
        has_idx: false,
        allowed: LOGIC,
    },
    OpSpec {
        name: "or",
        opcode: 7,
        cycles: 6,
        has_pcode: true,
        has_idx: false,
        allowed: LOGIC,
    },
    OpSpec {
        name: "xor",
        opcode: 8,
        cycles: 6,
        has_pcode: true,
        has_idx: false,
        allowed: LOGIC,
    },
    OpSpec {
        name: "zjmp",
        opcode: 9,
        cycles: 20,
        has_pcode: false,
        has_idx: true,
        allowed: ZJMP,
    },
    OpSpec {
        name: "ldi",
        opcode: 10,
        cycles: 25,
        has_pcode: true,
        has_idx: true,
        allowed: LDI,
    },
    OpSpec {
        name: "sti",
        opcode: 11,
        cycles: 25,
        has_pcode: true,
        has_idx: true,
        allowed: STI,
    },
    OpSpec {
        name: "fork",
        opcode: 12,
        cycles: 800,
        has_pcode: false,
        has_idx: true,
        allowed: FORK,
    },
    OpSpec {
        name: "lld",
        opcode: 13,
        cycles: 10,
        has_pcode: true,
        has_idx: false,
        allowed: LLD,
    },
    OpSpec {
        name: "lldi",
        opcode: 14,
        cycles: 50,
        has_pcode: true,
        has_idx: true,
        allowed: LLDI,
    },
    OpSpec {
        name: "lfork",
        opcode: 15,
        cycles: 1000,
        has_pcode: false,
        has_idx: true,
        allowed: FORK,
    },
    OpSpec {
        name: "nop",
        opcode: 16,
        cycles: 2,
        has_pcode: true,
        has_idx: false,
        allowed: NOP,
    },
];

pub fn by_name(name: &str) -> Option<&'static OpSpec> {
    OPS.iter().find(|op| op.name == name)
}

pub fn by_opcode(opcode: u8) -> Option<&'static OpSpec> {
    if opcode == 0 {
        None
    } else {
        OPS.get((opcode - 1) as usize)
    }
}

pub fn arg_size(kind: u8, has_idx: bool) -> usize {
    match kind {
        ARG_REG => 1,
        ARG_IND => IND_SIZE,
        ARG_DIR => {
            if has_idx {
                IND_SIZE
            } else {
                DIR_SIZE
            }
        }
        _ => 0,
    }
}

pub fn pcode_bits(kind: u8) -> u8 {
    match kind {
        ARG_REG => 0b01,
        ARG_DIR => 0b10,
        ARG_IND => 0b11,
        _ => 0,
    }
}
