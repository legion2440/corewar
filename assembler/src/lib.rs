pub mod disasm;
pub mod encode;
pub mod error;
pub mod expr;
pub mod parser;
pub mod preprocess;
pub mod spec;

use error::AsmError;

pub fn assemble(source: &str) -> Result<Vec<u8>, AsmError> {
    let program = parser::parse(source)?;
    encode::encode(&program)
}

pub fn disassemble(bytes: &[u8]) -> Result<String, AsmError> {
    disasm::disassemble(bytes)
}
