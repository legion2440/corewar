use std::collections::{HashMap, HashSet};

use crate::error::AsmError;

#[derive(Debug, Clone)]
pub struct SourceLine {
    pub number: usize,
    pub text: String,
}

pub fn preprocess(source: &str) -> Result<Vec<SourceLine>, AsmError> {
    let mut macros: HashMap<String, String> = HashMap::new();
    let mut out = Vec::new();

    for (idx, raw) in source.lines().enumerate() {
        let line_no = idx + 1;
        let code = strip_comment(raw);
        let trimmed = code.trim();
        if trimmed.starts_with(".define ") || trimmed.starts_with(".macro ") {
            let mut parts = trimmed.splitn(3, char::is_whitespace).filter(|p| !p.is_empty());
            let _directive = parts.next();
            let name = parts.next().ok_or_else(|| AsmError::line(line_no, "macro name is missing"))?;
            let value = parts.next().ok_or_else(|| AsmError::line(line_no, "macro value is missing"))?;
            if !valid_macro_name(name) {
                return Err(AsmError::line(line_no, format!("invalid macro name '{name}'")));
            }
            if macros.insert(name.to_owned(), value.trim().to_owned()).is_some() {
                return Err(AsmError::line(line_no, format!("macro '{name}' already defined")));
            }
            continue;
        }
        let expanded = expand_line(raw, &macros, line_no)?;
        out.push(SourceLine { number: line_no, text: expanded });
    }

    Ok(out)
}

fn valid_macro_name(name: &str) -> bool {
    let mut bytes = name.bytes();
    let Some(first) = bytes.next() else { return false; };
    (first.is_ascii_uppercase() || first == b'_')
        && bytes.all(|c| c.is_ascii_uppercase() || c.is_ascii_digit() || c == b'_')
}

fn expand_line(line: &str, macros: &HashMap<String, String>, line_no: usize) -> Result<String, AsmError> {
    let mut current = line.to_owned();
    let mut seen = HashSet::new();
    for _ in 0..32 {
        let (next, changed, names) = expand_once(&current, macros, line_no)?;
        if !changed {
            return Ok(current);
        }
        for name in names {
            if !seen.insert(name.clone()) && next.contains(&format!("${}", name)) {
                return Err(AsmError::line(line_no, format!("recursive macro expansion involving '{name}'")));
            }
        }
        current = next;
    }
    Err(AsmError::line(line_no, "macro expansion depth exceeded"))
}

fn expand_once(line: &str, macros: &HashMap<String, String>, line_no: usize) -> Result<(String, bool, Vec<String>), AsmError> {
    let bytes = line.as_bytes();
    let mut out = String::with_capacity(line.len());
    let mut i = 0;
    let mut quoted = false;
    let mut changed = false;
    let mut names = Vec::new();
    while i < bytes.len() {
        let c = bytes[i];
        if c == b'"' {
            quoted = !quoted;
            out.push('"');
            i += 1;
            continue;
        }
        if !quoted && c == b'#' {
            out.push_str(&line[i..]);
            break;
        }
        if !quoted && c == b'$' {
            let start = i + 1;
            let mut end = start;
            while end < bytes.len() && (bytes[end].is_ascii_uppercase() || bytes[end].is_ascii_digit() || bytes[end] == b'_') {
                end += 1;
            }
            if end == start {
                return Err(AsmError::line(line_no, "'$' must be followed by a macro name"));
            }
            let name = &line[start..end];
            let value = macros.get(name).ok_or_else(|| AsmError::line(line_no, format!("undefined macro '{name}'")))?;
            out.push_str(value);
            names.push(name.to_owned());
            changed = true;
            i = end;
            continue;
        }
        out.push(c as char);
        i += 1;
    }
    Ok((out, changed, names))
}

pub fn strip_comment(line: &str) -> &str {
    let mut quoted = false;
    for (idx, c) in line.char_indices() {
        match c {
            '"' => quoted = !quoted,
            '#' if !quoted => return &line[..idx],
            _ => {}
        }
    }
    line
}
