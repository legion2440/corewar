use std::{env, fs, path::{Path, PathBuf}, process};

fn main() {
    if let Err(err) = run() {
        eprintln!("asm: {err}");
        process::exit(1);
    }
}

fn run() -> Result<(), String> {
    let args: Vec<String> = env::args().skip(1).collect();
    if args.is_empty() || args.iter().any(|a| a == "-h" || a == "--help") {
        print_help();
        return Ok(());
    }

    let mut disassemble = false;
    let mut output: Option<PathBuf> = None;
    let mut input: Option<PathBuf> = None;
    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "-d" | "--disassemble" => disassemble = true,
            "-o" | "--output" => {
                i += 1;
                let Some(path) = args.get(i) else { return Err("-o/--output requires a path".into()); };
                output = Some(PathBuf::from(path));
            }
            arg if arg.starts_with('-') => return Err(format!("unknown option '{arg}'")),
            arg => {
                if input.replace(PathBuf::from(arg)).is_some() {
                    return Err("only one input file may be provided".into());
                }
            }
        }
        i += 1;
    }
    let input = input.ok_or_else(|| "input file is required".to_string())?;

    if disassemble {
        let bytes = fs::read(&input).map_err(|e| format!("{}: {e}", input.display()))?;
        let source = corewar_asm::disassemble(&bytes).map_err(|e| e.to_string())?;
        let out = output.unwrap_or_else(|| disassembly_path(&input));
        atomic_write(&out, source.as_bytes())?;
        println!("Wrote {}", out.display());
    } else {
        if input.extension().and_then(|s| s.to_str()) != Some("s") {
            return Err("assembler input must have .s extension".into());
        }
        let source = fs::read_to_string(&input).map_err(|e| format!("{}: {e}", input.display()))?;
        let bytes = corewar_asm::assemble(&source).map_err(|e| e.to_string())?;
        let out = output.unwrap_or_else(|| input.with_extension("cor"));
        atomic_write(&out, &bytes)?;
        println!("Wrote {} ({} bytes of code)", out.display(), bytes.len() - corewar_asm::spec::HEADER_SIZE);
    }
    Ok(())
}

fn disassembly_path(input: &Path) -> PathBuf {
    let stem = input.file_stem().and_then(|s| s.to_str()).unwrap_or("output");
    input.with_file_name(format!("{stem}.dis.s"))
}

fn atomic_write(path: &Path, data: &[u8]) -> Result<(), String> {
    let parent = path.parent().unwrap_or_else(|| Path::new("."));
    let file_name = path.file_name().and_then(|s| s.to_str()).unwrap_or("output");
    let tmp = parent.join(format!(".{file_name}.tmp-{}", process::id()));
    fs::write(&tmp, data).map_err(|e| format!("{}: {e}", tmp.display()))?;
    if let Err(e) = fs::rename(&tmp, path) {
        let _ = fs::remove_file(&tmp);
        return Err(format!("{}: {e}", path.display()));
    }
    Ok(())
}

fn print_help() {
    println!("Corewar assembler\n\nUSAGE:\n  asm [OPTIONS] <file.s>\n  asm --disassemble [OPTIONS] <file.cor>\n\nOPTIONS:\n  -d, --disassemble   Convert .cor bytecode back to assembly\n  -o, --output PATH   Select output path\n  -h, --help          Show this help\n\nBONUS SOURCE EXTENSIONS:\n  .define NAME VALUE  Define a simple $NAME textual macro\n  arithmetic          Numeric operands accept + - * / % and parentheses");
}
