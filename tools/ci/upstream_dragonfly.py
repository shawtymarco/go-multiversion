"""Validate the pending Dragonfly adapter against the exact unmerged 1.26.60 PR.

The replacement exists only in a temporary modfile. It is not the adapter's
published dependency policy and does not claim that Dragonfly master is ready.
"""
from pathlib import Path
import json
import os
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
ADAPTER = ROOT / "dragonfly"
PR_REVISION = "d86aefd9ff9b6a357d6dfadab2093af811545256"


def main():
    env = dict(os.environ, GOWORK="off")
    subprocess.run(["go", "mod", "tidy", "-diff"], cwd=ADAPTER, env=env, check=True)
    with tempfile.TemporaryDirectory(prefix="gomv-dragonfly-pr-") as temporary:
        modfile = Path(temporary) / "validation.mod"
        modfile.write_text((ADAPTER / "go.mod").read_text())
        modfile.with_suffix(".sum").write_text((ADAPTER / "go.sum").read_text())

        def run(*args, capture=False):
            return subprocess.run(["go", *args], cwd=ADAPTER, env=env, check=True,
                                  text=True, stdout=subprocess.PIPE if capture else None)

        # Relative replacements are made absolute only in this disposable file.
        run("mod", "edit", f"-modfile={modfile}",
            f"-replace=github.com/shawtymarco/go-multiversion={ROOT}")
        resolved = json.loads(run("list", "-m", "-json",
                                 "github.com/HashimTheArab/dragonfly@" + PR_REVISION,
                                 capture=True).stdout)
        version = resolved["Version"]
        if not version.endswith("-" + PR_REVISION[:12]):
            raise ValueError("PR resolved to an unexpected revision: " + version)
        print("Testing unmerged Dragonfly PR 1481 at " + PR_REVISION, flush=True)
        run("mod", "edit", f"-modfile={modfile}",
            "-replace=github.com/df-mc/dragonfly=github.com/HashimTheArab/dragonfly@" + version)
        run("mod", "tidy", f"-modfile={modfile}")
        run("test", f"-modfile={modfile}", "-count=1", "-timeout=3m", "./...")
        run("vet", f"-modfile={modfile}", "./...")
        run("build", f"-modfile={modfile}", "./...")


if __name__ == "__main__":
    main()
