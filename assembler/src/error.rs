use std::fmt::{self, Display};

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct AsmError {
    pub line: Option<usize>,
    pub column: Option<usize>,
    pub message: String,
}

impl AsmError {
    pub fn new(message: impl Into<String>) -> Self {
        Self { line: None, column: None, message: message.into() }
    }

    pub fn line(line: usize, message: impl Into<String>) -> Self {
        Self { line: Some(line), column: None, message: message.into() }
    }

    pub fn at(line: usize, column: usize, message: impl Into<String>) -> Self {
        Self { line: Some(line), column: Some(column), message: message.into() }
    }
}

impl Display for AsmError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match (self.line, self.column) {
            (Some(line), Some(col)) => write!(f, "line {line}:{col}: {}", self.message),
            (Some(line), None) => write!(f, "line {line}: {}", self.message),
            _ => f.write_str(&self.message),
        }
    }
}

impl std::error::Error for AsmError {}
