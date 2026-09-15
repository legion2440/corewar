use crate::{
    error::AsmError,
    parser::{Argument, Program},
    spec::{pcode_bits, COREWAR_EXEC_SIGNATURE, DESCRIPTION_LENGTH, HEADER_SIZE, PROG_NAME_LENGTH},
};

pub fn encode(program: &Program) -> Result<Vec<u8>, AsmError> {
    let mut out = Vec::with_capacity(HEADER_SIZE + program.code_size);
    out.extend_from_slice(&COREWAR_EXEC_SIGNATURE.to_be_bytes());
    write_padded(&mut out, program.name.as_bytes(), PROG_NAME_LENGTH);
    out.extend_from_slice(&[0; 4]);
    out.extend_from_slice(&(program.code_size as u32).to_be_bytes());
    write_padded(&mut out, program.description.as_bytes(), DESCRIPTION_LENGTH);
    out.extend_from_slice(&[0; 4]);

    for ins in &program.instructions {
        out.push(ins.op.opcode);
        if ins.op.has_pcode {
            let mut pcode = 0u8;
            for (idx, arg) in ins.args.iter().enumerate() {
                pcode |= pcode_bits(arg.kind()) << (6 - 2 * idx);
            }
            out.push(pcode);
        }
        for arg in &ins.args {
            match arg {
                Argument::Register(reg) => out.push(*reg),
                Argument::Direct(expr) => {
                    let value = expr.eval(&program.labels, ins.offset).map_err(|e| AsmError::line(ins.line, e.message))?;
                    if ins.op.has_idx {
                        out.extend_from_slice(&(value as i16).to_be_bytes());
                    } else {
                        out.extend_from_slice(&(value as i32).to_be_bytes());
                    }
                }
                Argument::Indirect(expr) => {
                    let value = expr.eval(&program.labels, ins.offset).map_err(|e| AsmError::line(ins.line, e.message))?;
                    out.extend_from_slice(&(value as i16).to_be_bytes());
                }
            }
        }
    }
    Ok(out)
}

fn write_padded(out: &mut Vec<u8>, bytes: &[u8], len: usize) {
    out.extend_from_slice(bytes);
    out.resize(out.len() + (len - bytes.len()), 0);
}
