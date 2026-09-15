use std::collections::HashMap;

use crate::error::AsmError;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Expr {
    Number(i64),
    Label(String),
    Unary(char, Box<Expr>),
    Binary(char, Box<Expr>, Box<Expr>),
}

impl Expr {
    pub fn eval(&self, labels: &HashMap<String, usize>, origin: usize) -> Result<i64, AsmError> {
        match self {
            Expr::Number(v) => Ok(*v),
            Expr::Label(name) => labels
                .get(name)
                .map(|v| *v as i64 - origin as i64)
                .ok_or_else(|| AsmError::new(format!("unknown label '{name}'"))),
            Expr::Unary('+', inner) => inner.eval(labels, origin),
            Expr::Unary('-', inner) => inner
                .eval(labels, origin)?
                .checked_neg()
                .ok_or_else(|| AsmError::new("integer overflow in unary '-'")),
            Expr::Unary(op, _) => Err(AsmError::new(format!("unsupported unary operator '{op}'"))),
            Expr::Binary(op, lhs, rhs) => {
                let a = lhs.eval(labels, origin)?;
                let b = rhs.eval(labels, origin)?;
                match op {
                    '+' => a.checked_add(b),
                    '-' => a.checked_sub(b),
                    '*' => a.checked_mul(b),
                    '/' => if b == 0 { None } else { a.checked_div(b) },
                    '%' => if b == 0 { None } else { a.checked_rem(b) },
                    _ => return Err(AsmError::new(format!("unsupported binary operator '{op}'"))),
                }
                .ok_or_else(|| {
                    if (*op == '/' || *op == '%') && b == 0 {
                        AsmError::new("division by zero in expression")
                    } else {
                        AsmError::new("integer overflow in expression")
                    }
                })
            }
        }
    }
}

pub fn parse_expression(input: &str) -> Result<Expr, AsmError> {
    let mut parser = Parser { bytes: input.as_bytes(), pos: 0 };
    let expr = parser.parse_add_sub()?;
    parser.skip_ws();
    if parser.pos != parser.bytes.len() {
        return Err(AsmError::new(format!("unexpected token in expression near '{}'", &input[parser.pos..])));
    }
    Ok(expr)
}

struct Parser<'a> {
    bytes: &'a [u8],
    pos: usize,
}

impl Parser<'_> {
    fn skip_ws(&mut self) {
        while self.pos < self.bytes.len() && self.bytes[self.pos].is_ascii_whitespace() {
            self.pos += 1;
        }
    }

    fn peek(&mut self) -> Option<u8> {
        self.skip_ws();
        self.bytes.get(self.pos).copied()
    }

    fn parse_add_sub(&mut self) -> Result<Expr, AsmError> {
        let mut lhs = self.parse_mul_div()?;
        loop {
            let op = match self.peek() {
                Some(b'+') => '+',
                Some(b'-') => '-',
                _ => break,
            };
            self.pos += 1;
            let rhs = self.parse_mul_div()?;
            lhs = Expr::Binary(op, Box::new(lhs), Box::new(rhs));
        }
        Ok(lhs)
    }

    fn parse_mul_div(&mut self) -> Result<Expr, AsmError> {
        let mut lhs = self.parse_unary()?;
        loop {
            let op = match self.peek() {
                Some(b'*') => '*',
                Some(b'/') => '/',
                Some(b'%') => '%',
                _ => break,
            };
            self.pos += 1;
            let rhs = self.parse_unary()?;
            lhs = Expr::Binary(op, Box::new(lhs), Box::new(rhs));
        }
        Ok(lhs)
    }

    fn parse_unary(&mut self) -> Result<Expr, AsmError> {
        match self.peek() {
            Some(b'+') | Some(b'-') => {
                let op = self.bytes[self.pos] as char;
                self.pos += 1;
                Ok(Expr::Unary(op, Box::new(self.parse_unary()?)))
            }
            _ => self.parse_primary(),
        }
    }

    fn parse_primary(&mut self) -> Result<Expr, AsmError> {
        self.skip_ws();
        match self.bytes.get(self.pos).copied() {
            Some(b'(') => {
                self.pos += 1;
                let expr = self.parse_add_sub()?;
                if self.peek() != Some(b')') {
                    return Err(AsmError::new("missing ')' in expression"));
                }
                self.pos += 1;
                Ok(expr)
            }
            Some(b':') => self.parse_label(),
            Some(c) if c.is_ascii_digit() => self.parse_number(),
            Some(c) => Err(AsmError::new(format!("unexpected '{}' in expression", c as char))),
            None => Err(AsmError::new("empty expression")),
        }
    }

    fn parse_label(&mut self) -> Result<Expr, AsmError> {
        self.pos += 1;
        let start = self.pos;
        while self.pos < self.bytes.len() && is_label_char(self.bytes[self.pos]) {
            self.pos += 1;
        }
        if self.pos == start {
            return Err(AsmError::new("expected label name after ':'"));
        }
        let name = std::str::from_utf8(&self.bytes[start..self.pos]).unwrap().to_owned();
        Ok(Expr::Label(name))
    }

    fn parse_number(&mut self) -> Result<Expr, AsmError> {
        let start = self.pos;
        while self.pos < self.bytes.len() && self.bytes[self.pos].is_ascii_digit() {
            self.pos += 1;
        }
        let s = std::str::from_utf8(&self.bytes[start..self.pos]).unwrap();
        let value = s.parse::<i64>().map_err(|_| AsmError::new(format!("invalid integer '{s}'")))?;
        Ok(Expr::Number(value))
    }
}

pub fn is_label_char(c: u8) -> bool {
    c.is_ascii_lowercase() || c.is_ascii_digit() || c == b'_'
}
