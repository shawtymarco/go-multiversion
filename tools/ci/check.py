"""Run portable repository CI checks without changing tracked fixtures."""

from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[2]


def run(args: list[str], cwd: Path = ROOT, *, capture: bool = False, env=None) -> str:
    display = ' '.join(args[:2]) + f" ({len(args) - 2} files)" if args[0] == "gofmt" else ' '.join(args)
    print(f"[{cwd.relative_to(ROOT) if cwd.is_relative_to(ROOT) else 'dependency'}] {display}", flush=True)
    result = subprocess.run(args, cwd=cwd, env=env, check=True, text=True,
                            stdout=subprocess.PIPE if capture else None)
    return result.stdout if capture else ""


def tracked(pattern: str) -> list[str]:
    output = subprocess.check_output(["git", "ls-files", "-z", "--", pattern], cwd=ROOT)
    return [name.decode("utf-8") for name in output.split(b"\0") if name]


def modules() -> list[Path]:
    paths = {ROOT}
    paths.update((ROOT / name).parent for name in tracked("**/go.mod"))
    return sorted(paths, key=lambda path: path.as_posix())


def formatting() -> None:
    files = tracked("*.go")
    dirty = run(["gofmt", "-l", *files], capture=True).splitlines()
    if dirty:
        raise ValueError("Unformatted Go files:\n" + "\n".join(dirty))


def module_check() -> None:
    for path in modules():
        run(["go", "mod", "verify"], path)
        run(["go", "mod", "tidy", "-diff"], path)


def test(race: bool = False) -> None:
    for path in modules():
        args = ["go", "test", "-mod=readonly", "-count=1", "-timeout=5m"]
        if race:
            args += ["-race"]
            if path == ROOT:
                args += ["-covermode=atomic", "-coverprofile=" + str(ROOT / "coverage.out")]
        run(args + ["./..."], path)
        if not race:
            run(["go", "vet", "-mod=readonly", "./..."], path)
            with tempfile.TemporaryDirectory(prefix="gomv-build-") as output:
                run(["go", "build", "-mod=readonly", "-o", output + os.sep, "./..."], path)


def lint() -> None:
    for path in modules():
        run(["golangci-lint", "run", "--config", str(ROOT / ".golangci.yml"), "./..."], path)


def oracle_check() -> None:
    catalogue = json.loads((ROOT / "tools/ci/oracles.json").read_text())
    discovered = {path.relative_to(ROOT).as_posix() for path in modules()
                  if path.name.startswith("oracle_")}
    registered = {entry["module"] for entry in catalogue}
    if discovered != registered:
        raise ValueError(f"Oracle catalogue mismatch: modules={discovered}, registered={registered}")
    for entry in catalogue:
        path = ROOT / entry["module"]
        fixture = ROOT / entry["fixture"]
        if not path.resolve().is_relative_to(ROOT) or not fixture.resolve().is_relative_to(ROOT):
            raise ValueError("Oracle paths must stay inside the repository")
        with tempfile.TemporaryDirectory(prefix="gomv-oracle-") as temporary:
            output = Path(temporary) / "reference.json"
            args = ["go", "run", "-mod=readonly", "."]
            if entry["mode"] == "file":
                run(args + ["-out", str(output)], path)
                actual = json.loads(output.read_text())
            else:
                actual = json.loads(run(args, path, capture=True))
            expected = json.loads(fixture.read_text())
            for name in entry.get("unordered_fixtures", []):
                old = expected["fixtures"][name]
                new = actual["fixtures"][name]
                if old["id"] != new["id"]:
                    raise ValueError(f"Oracle packet ID changed: {name}")
                # The pinned v419 writer iterates AddActor's metadata map in
                # arbitrary order. Compare the complete decoded packet through
                # that independent pinned reader, which also rejects unread bytes.
                decode = ["go", "run", "-mod=readonly", ".", "-inspect", "server", str(old["id"])]
                if json.loads(run(decode + [old["hex"]], path, capture=True)) != json.loads(run(decode + [new["hex"]], path, capture=True)):
                    raise ValueError(f"Unordered oracle fixture changed: {name}")
                new["hex"] = old["hex"]
            # Panic text in skipped zero-value constructors can vary by toolchain;
            # compare their key set, and compare remaining encoded values exactly.
            if isinstance(expected, dict):
                if set(expected.get("skipped", {})) != set(actual.get("skipped", {})):
                    raise ValueError(f"Skipped oracle packets changed: {fixture}")
                actual = {key: value for key, value in actual.items() if key != "skipped"}
                expected = {key: value for key, value in expected.items() if key != "skipped"}
            if actual != expected:
                raise ValueError(f"Oracle fixture differs from locked source: {fixture.relative_to(ROOT)}")


def portability() -> None:
    for target in ("linux", "windows", "darwin"):
        env = dict(os.environ, GOOS=target, GOARCH="arm64", CGO_ENABLED="0")
        for path in modules():
            # Build commands produce executables only in the temporary output directory.
            with tempfile.TemporaryDirectory(prefix="gomv-build-") as output:
                run(["go", "build", "-mod=readonly", "-o", output + os.sep, "./..."], path, env=env)


def transport() -> None:
    directory = run(["go", "list", "-m", "-f", "{{.Dir}}", "github.com/sandertv/gophertunnel"], capture=True).strip()
    run(["go", "test", "-race", "-mod=readonly", "-count=1", "-timeout=5m", "./minecraft/..."], Path(directory))


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("check", choices=("format", "modules", "test", "race", "lint", "oracles", "portability", "transport"))
    args = parser.parse_args()
    checks = {"format": formatting, "modules": module_check, "test": test,
              "race": lambda: test(True), "lint": lint, "oracles": oracle_check,
              "portability": portability, "transport": transport}
    checks[args.check]()


if __name__ == "__main__":
    try:
        main()
    except (ValueError, subprocess.CalledProcessError) as error:
        print(error, file=sys.stderr)
        sys.exit(1)
