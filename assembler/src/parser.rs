use std::collections::HashMap;

use crate::{
    error::AsmError,
    expr::{is_label_char, parse_expression, Expr},
    preprocess::{preprocess, strip_comment},
    spec::{
        arg_size, by_name, OpSpec, ARG_DIR, ARG_IND, ARG_REG, DESCRIPTION_LENGTH, PLAYER_MAX_SIZE,
        PROG_NAME_LENGTH, REG_NUMBER,
    },
};

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Argument {
    Register(u8),
    Direct(Expr),
    Indirect(Expr),
}

impl Argument {
    pub fn kind(&self) -> u8 {
        match self {
            Argument::Register(_) => ARG_REG,
            Argument::Direct(_) => ARG_DIR,
            Argument::Indirect(_) => ARG_IND,
        }
    }
}

#[derive(Debug, Clone)]
pub struct Instruction {
    pub op: &'static OpSpec,
    pub args: Vec<Argument>,
    pub offset: usize,
    pub size: usize,
    pub line: usize,
}

#[derive(Debug, Clone)]
pub struct Program {
    pub name: String,
    pub description: String,
    pub instructions: Vec<Instruction>,
    pub labels: HashMap<String, usize>,
    pub code_size: usize,
}

pub fn parse(source: &str) -> Result<Program, AsmError> {
    let lines = preprocess(source)?;
    let mut name: Option<String> = None;
    let mut description: Option<String> = None;
    let mut instructions = Vec::new();
    let mut labels = HashMap::new();
    let mut offset = 0usize;
    let mut code_started = false;

    for source_line in lines {
        let line_no = source_line.number;
        let code = strip_comment(&source_line.text).trim();
        if code.is_empty() {
            continue;
        }

        if code.starts_with('.') {
            if code_started {
                return Err(AsmError::line(
                    line_no,
                    "header directives must appear before labels and instructions",
                ));
            }
            if code.starts_with(".name") {
                if name.is_some() {
                    return Err(AsmError::line(line_no, "duplicate .name directive"));
                }
                let value = parse_quoted_directive(code, ".name", line_no)?;
                if value.as_bytes().len() > PROG_NAME_LENGTH {
                    return Err(AsmError::line(
                        line_no,
                        format!("name exceeds {PROG_NAME_LENGTH} bytes"),
                    ));
                }
                name = Some(value);
                continue;
            }
            if code.starts_with(".description") {
                if description.is_some() {
                    return Err(AsmError::line(line_no, "duplicate .description directive"));
                }
                let value = parse_quoted_directive(code, ".description", line_no)?;
                if value.as_bytes().len() > DESCRIPTION_LENGTH {
                    return Err(AsmError::line(
                        line_no,
                        format!("description exceeds {DESCRIPTION_LENGTH} bytes"),
                    ));
                }
                description = Some(value);
                continue;
            }
            return Err(AsmError::line(
                line_no,
                format!("unknown directive in '{code}'"),
            ));
        }

        if name.is_none() || description.is_none() {
            return Err(AsmError::line(
                line_no,
                ".name and .description must appear before code",
            ));
        }
        code_started = true;

        let mut rest = code;
        loop {
            let Some((label, tail)) = take_leading_label(rest) else {
                break;
            };
            if labels.insert(label.to_owned(), offset).is_some() {
                return Err(AsmError::line(
                    line_no,
                    format!("duplicate label '{label}'"),
                ));
            }
            rest = tail.trim_start();
            if rest.is_empty() {
                break;
            }
        }
        if rest.is_empty() {
            continue;
        }

        let (mnemonic, arg_text) = split_mnemonic(rest);
        let op = by_name(mnemonic)
            .ok_or_else(|| AsmError::line(line_no, format!("unknown instruction '{mnemonic}'")))?;
        let raw_args = split_arguments(arg_text, line_no)?;
        if raw_args.len() != op.allowed.len() {
            return Err(AsmError::line(
                line_no,
                format!(
                    "{} expects {} argument(s), got {}",
                    op.name,
                    op.allowed.len(),
                    raw_args.len()
                ),
            ));
        }

        let mut args = Vec::with_capacity(raw_args.len());
        let mut size = 1 + usize::from(op.has_pcode);
        for (idx, raw) in raw_args.iter().enumerate() {
            let arg = parse_argument(raw, line_no)?;
            let kind = arg.kind();
            if op.allowed[idx] & kind == 0 {
                return Err(AsmError::line(
                    line_no,
                    format!("invalid argument {} type for {}", idx + 1, op.name),
                ));
            }
            size += arg_size(kind, op.has_idx);
            args.push(arg);
        }

        instructions.push(Instruction {
            op,
            args,
            offset,
            size,
            line: line_no,
        });
        offset = offset
            .checked_add(size)
            .ok_or_else(|| AsmError::line(line_no, "program size overflow"))?;
        if offset > PLAYER_MAX_SIZE {
            return Err(AsmError::line(
                line_no,
                format!("program exceeds maximum size of {PLAYER_MAX_SIZE} bytes"),
            ));
        }
    }

    let name = name.ok_or_else(|| AsmError::new("missing .name directive"))?;
    let description = description.ok_or_else(|| AsmError::new("missing .description directive"))?;
    Ok(Program {
        name,
        description,
        instructions,
        labels,
        code_size: offset,
    })
}

fn parse_quoted_directive(code: &str, directive: &str, line: usize) -> Result<String, AsmError> {
    let rest = code.strip_prefix(directive).unwrap();
    if !rest.is_empty() && !rest.as_bytes()[0].is_ascii_whitespace() {
        return Err(AsmError::line(line, format!("invalid directive '{code}'")));
    }
    let rest = rest.trim_start();
    if !rest.starts_with('"') {
        return Err(AsmError::line(
            line,
            format!("{directive} requires a quoted string"),
        ));
    }
    let tail = &rest[1..];
    let Some(end) = tail.find('"') else {
        return Err(AsmError::line(
            line,
            format!("unterminated string in {directive}"),
        ));
    };
    if !tail[end + 1..].trim().is_empty() {
        return Err(AsmError::line(
            line,
            format!("unexpected text after {directive} string"),
        ));
    }
    Ok(tail[..end].to_owned())
}

fn take_leading_label(input: &str) -> Option<(&str, &str)> {
    let bytes = input.as_bytes();
    let mut i = 0;
    while i < bytes.len() && is_label_char(bytes[i]) {
        i += 1;
    }
    if i == 0 || bytes.get(i) != Some(&b':') {
        return None;
    }
    Some((&input[..i], &input[i + 1..]))
}

fn split_mnemonic(input: &str) -> (&str, &str) {
    match input.find(char::is_whitespace) {
        Some(i) => (&input[..i], input[i..].trim()),
        None => (input, ""),
    }
}

fn split_arguments(input: &str, line: usize) -> Result<Vec<String>, AsmError> {
    if input.trim().is_empty() {
        return Ok(Vec::new());
    }
    let mut out = Vec::new();
    let mut depth = 0i32;
    let mut start = 0usize;
    for (i, c) in input.char_indices() {
        match c {
            '(' => depth += 1,
            ')' => {
                depth -= 1;
                if depth < 0 {
                    return Err(AsmError::line(line, "unmatched ')' in arguments"));
                }
            }
            ',' if depth == 0 => {
                let arg = input[start..i].trim();
                if arg.is_empty() {
                    return Err(AsmError::line(line, "empty argument"));
                }
                out.push(arg.to_owned());
                start = i + 1;
            }
            _ => {}
        }
    }
    if depth != 0 {
        return Err(AsmError::line(line, "unmatched '(' in arguments"));
    }
    let arg = input[start..].trim();
    if arg.is_empty() {
        return Err(AsmError::line(line, "empty trailing argument"));
    }
    out.push(arg.to_owned());
    Ok(out)
}

fn parse_argument(raw: &str, line: usize) -> Result<Argument, AsmError> {
    let raw = raw.trim();
    if let Some(rest) = raw.strip_prefix('r') {
        if !rest.is_empty() && rest.bytes().all(|c| c.is_ascii_digit()) {
            let reg = rest
                .parse::<u16>()
                .map_err(|_| AsmError::line(line, format!("invalid register '{raw}'")))?;
            if reg == 0 || reg > REG_NUMBER as u16 {
                return Err(AsmError::line(
                    line,
                    format!("register out of range: '{raw}'"),
                ));
            }
            return Ok(Argument::Register(reg as u8));
        }
    }
    if let Some(rest) = raw.strip_prefix('%') {
        let expr = parse_expression(rest).map_err(|e| AsmError::line(line, e.message))?;
        return Ok(Argument::Direct(expr));
    }
    let expr = parse_expression(raw).map_err(|e| AsmError::line(line, e.message))?;
    Ok(Argument::Indirect(expr))
}
