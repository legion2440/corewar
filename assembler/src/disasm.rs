use crate::{
    error::AsmError,
    spec::{arg_size, by_opcode, COREWAR_EXEC_SIGNATURE, DESCRIPTION_LENGTH, HEADER_SIZE, PLAYER_MAX_SIZE, PROG_NAME_LENGTH, ARG_DIR, ARG_IND, ARG_REG, REG_NUMBER},
};

pub fn disassemble(bytes: &[u8]) -> Result<String, AsmError> {
    let (name, description, code) = decode_header(bytes)?;
    let mut out = String::new();
    out.push_str(&format!(".name \"{}\"\n", escape(&name)));
    out.push_str(&format!(".description \"{}\"\n\n", escape(&description)));

    let mut pc = 0usize;
    while pc < code.len() {
        let start = pc;
        let opcode = code[pc];
        let op = by_opcode(opcode).ok_or_else(|| AsmError::new(format!("unknown opcode {opcode} at code offset {start}")))?;
        pc += 1;
        let mut kinds = Vec::with_capacity(op.allowed.len());
        if op.has_pcode {
            let pcode = *code.get(pc).ok_or_else(|| AsmError::new(format!("truncated pcode at code offset {start}")))?;
            pc += 1;
            for idx in 0..op.allowed.len() {
                let bits = (pcode >> (6 - idx * 2)) & 0b11;
                let kind = match bits {
                    0b01 => ARG_REG,
                    0b10 => ARG_DIR,
                    0b11 => ARG_IND,
                    _ => return Err(AsmError::new(format!("invalid pcode for {} at code offset {start}", op.name))),
                };
                if op.allowed[idx] & kind == 0 {
                    return Err(AsmError::new(format!("invalid argument type in pcode for {} at code offset {start}", op.name)));
                }
                kinds.push(kind);
            }
            let unused_pairs = 4usize.saturating_sub(op.allowed.len());
            if unused_pairs > 0 {
                let mask = (1u16 << (unused_pairs * 2)) - 1;
                if (pcode as u16) & mask != 0 {
                    return Err(AsmError::new(format!("non-zero unused pcode bits for {} at code offset {start}", op.name)));
                }
            }
        } else {
            for mask in op.allowed {
                if *mask == ARG_DIR || *mask == ARG_REG || *mask == ARG_IND {
                    kinds.push(*mask);
                } else {
                    return Err(AsmError::new(format!("cannot infer argument type for {}", op.name)));
                }
            }
        }

        let mut rendered = Vec::with_capacity(kinds.len());
        for kind in kinds {
            let size = arg_size(kind, op.has_idx);
            if pc + size > code.len() {
                return Err(AsmError::new(format!("truncated {} at code offset {start}", op.name)));
            }
            let text = match kind {
                ARG_REG => {
                    let reg = code[pc];
                    if reg == 0 || reg > REG_NUMBER {
                        return Err(AsmError::new(format!("invalid register r{reg} at code offset {start}")));
                    }
                    format!("r{reg}")
                }
                ARG_DIR => {
                    let value = if size == 2 { i16::from_be_bytes([code[pc], code[pc + 1]]) as i32 }
                                else { i32::from_be_bytes([code[pc], code[pc + 1], code[pc + 2], code[pc + 3]]) };
                    format!("%{value}")
                }
                ARG_IND => {
                    let value = i16::from_be_bytes([code[pc], code[pc + 1]]) as i32;
                    value.to_string()
                }
                _ => unreachable!(),
            };
            pc += size;
            rendered.push(text);
        }
        out.push_str(op.name);
        if !rendered.is_empty() {
            out.push(' ');
            out.push_str(&rendered.join(", "));
        }
        out.push('\n');
    }
    Ok(out)
}

fn decode_header(bytes: &[u8]) -> Result<(String, String, &[u8]), AsmError> {
    if bytes.len() < HEADER_SIZE {
        return Err(AsmError::new(format!("file is too small: {} bytes, minimum is {HEADER_SIZE}", bytes.len())));
    }
    let magic = u32::from_be_bytes(bytes[0..4].try_into().unwrap());
    if magic != COREWAR_EXEC_SIGNATURE {
        return Err(AsmError::new(format!("wrong signature: 0x{magic:08x}")));
    }
    if bytes[4 + PROG_NAME_LENGTH..4 + PROG_NAME_LENGTH + 4] != [0; 4] {
        return Err(AsmError::new("non-zero padding after name"));
    }
    let size_offset = 4 + PROG_NAME_LENGTH + 4;
    let code_size = u32::from_be_bytes(bytes[size_offset..size_offset + 4].try_into().unwrap()) as usize;
    if code_size > PLAYER_MAX_SIZE {
        return Err(AsmError::new(format!("declared program size {code_size} exceeds {PLAYER_MAX_SIZE}")));
    }
    let desc_offset = size_offset + 4;
    let pad_offset = desc_offset + DESCRIPTION_LENGTH;
    if bytes[pad_offset..pad_offset + 4] != [0; 4] {
        return Err(AsmError::new("non-zero padding after description"));
    }
    if bytes.len() != HEADER_SIZE + code_size {
        return Err(AsmError::new(format!("declared program size {code_size} does not match file size")));
    }
    let name = nul_string(&bytes[4..4 + PROG_NAME_LENGTH], "name")?;
    let description = nul_string(&bytes[desc_offset..desc_offset + DESCRIPTION_LENGTH], "description")?;
    Ok((name, description, &bytes[HEADER_SIZE..]))
}

fn nul_string(bytes: &[u8], field: &str) -> Result<String, AsmError> {
    let end = bytes.iter().position(|b| *b == 0).unwrap_or(bytes.len());
    if bytes[end..].iter().any(|b| *b != 0) {
        return Err(AsmError::new(format!("non-zero bytes after NUL in {field}")));
    }
    String::from_utf8(bytes[..end].to_vec()).map_err(|_| AsmError::new(format!("{field} is not valid UTF-8")))
}

fn escape(input: &str) -> String {
    input.to_owned()
}
