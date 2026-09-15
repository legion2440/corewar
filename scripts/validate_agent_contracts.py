#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

try:
    import jsonschema
except ImportError as exc:
    print("error: python package 'jsonschema' is required for agent-contract validation", file=sys.stderr)
    raise SystemExit(2) from exc

ROOT = Path(__file__).resolve().parents[1]
AGENT = ROOT / "agent"
SCHEMAS = AGENT / "schemas"


def load(path: Path):
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def validate_json(path: Path, schema_name: str) -> None:
    jsonschema.Draft202012Validator(load(SCHEMAS / schema_name)).validate(load(path))


def size_signal(path: Path, target: int, warning: int) -> str:
    n = path.stat().st_size
    state = "warning" if n > warning else "ok"
    return f"{path.relative_to(ROOT)}: {n} bytes (target <= {target}, warning > {warning}) [{state}]"


def main() -> int:
    errors: list[str] = []

    docs = [
        (AGENT / "methodology.json", "methodology.schema.json"),
        (AGENT / "module-index.json", "module-index.schema.json"),
        (AGENT / "dependency-graph.json", "dependency-graph.schema.json"),
    ]
    try:
        for path, schema in docs:
            validate_json(path, schema)
    except Exception as exc:
        errors.append(f"schema validation: {exc}")

    index = load(AGENT / "module-index.json")
    graph = load(AGENT / "dependency-graph.json")
    modules = index["modules"]

    manifests: dict[str, dict] = {}
    for module_id, meta in modules.items():
        manifest_path = ROOT / meta["manifest"]
        if not manifest_path.exists():
            errors.append(f"missing manifest: {meta['manifest']}")
            continue
        try:
            validate_json(manifest_path, "module-manifest.schema.json")
        except Exception as exc:
            errors.append(f"{meta['manifest']}: schema validation: {exc}")
            continue
        manifest = load(manifest_path)
        manifests[module_id] = manifest
        if manifest["module_id"] != module_id:
            errors.append(f"manifest id mismatch: {meta['manifest']} -> {manifest['module_id']} != {module_id}")

        if meta["status"] == "implemented":
            for rel in meta["roots"] + meta["entrypoints"]:
                if not (ROOT / rel).exists():
                    errors.append(f"implemented module {module_id}: missing {rel}")

        for dep in meta["dependencies"]:
            if dep not in modules:
                errors.append(f"unknown dependency in index: {module_id} -> {dep}")

    manifest_dir = AGENT / "modules"
    for manifest_path in manifest_dir.glob("*.json"):
        data = load(manifest_path)
        if data.get("module_id") not in modules:
            errors.append(f"orphan manifest: {manifest_path.relative_to(ROOT)}")

    edges = {tuple(edge) for edge in graph["edges"]}
    expected_edges = {(module_id, dep) for module_id, meta in modules.items() for dep in meta["dependencies"]}
    if edges != expected_edges:
        errors.append(f"dependency graph drift: graph={sorted(edges)} index={sorted(expected_edges)}")
    for src, dst in edges:
        if src not in modules or dst not in modules:
            errors.append(f"unknown graph node: {src} -> {dst}")

    for module_id, manifest in manifests.items():
        for item in manifest["paths"]:
            path = ROOT / item["path"]
            lifecycle = item["lifecycle"]
            provenance = item["provenance"]
            if lifecycle == "planned" and path.exists():
                errors.append(f"planned path exists: {item['path']}")
            elif lifecycle in {"implemented", "deprecated"} and not path.exists():
                errors.append(f"{lifecycle} path missing: {item['path']}")
            if provenance == "generated" and lifecycle in {"implemented", "deprecated"}:
                generator = ROOT / item["generator"]
                if not generator.exists():
                    errors.append(f"generated path missing generator: {item['generator']}")
                else:
                    proc = subprocess.run(item["check"], cwd=ROOT, shell=True, text=True, capture_output=True)
                    if proc.returncode != 0:
                        errors.append(f"generated check failed for {item['path']}: {proc.stderr.strip() or proc.stdout.strip()}")

    core = AGENT / "execution-core.md"
    agents = ROOT / "AGENTS.md"
    if not core.exists():
        errors.append("missing agent/execution-core.md")
    else:
        first = next((ln.strip() for ln in core.read_text(encoding="utf-8").splitlines() if ln.strip()), "")
        if first != "[EXECUTION-CORE v2.1]":
            errors.append("execution-core sentinel mismatch")
    if not agents.exists():
        errors.append("missing AGENTS.md")
    else:
        text_value = agents.read_text(encoding="utf-8")
        if "[EXECUTION-CORE v2.1]" not in text_value or "agent/execution-core.md" not in text_value:
            errors.append("AGENTS.md missing deterministic execution-core pointer")
        if "EX-1" in text_value or "EX-7" in text_value:
            errors.append("AGENTS.md embeds execution-core rules instead of pointer")

    print(size_signal(AGENT / "module-index.json", 8 * 1024, 12 * 1024))
    for path in sorted(manifest_dir.glob("*.json")):
        print(size_signal(path, 6 * 1024, 10 * 1024))
    print(size_signal(ROOT / "AGENTS.md", 4 * 1024, 6 * 1024))
    print(size_signal(core, 4 * 1024, 6 * 1024))
    total = agents.stat().st_size + core.stat().st_size
    print(f"mandatory instruction surface: {total} bytes (target <= 8192, warning > 12288) [{'warning' if total > 12288 else 'ok'}]")

    if errors:
        for err in errors:
            print(f"error: {err}", file=sys.stderr)
        return 1
    print("agent contracts: ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
