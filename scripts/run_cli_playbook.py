#!/usr/bin/env python3
"""
Execute the zine-layout CLI verification playbook end-to-end.

This is the same sequence of commands documented in
ttmp/2025-10-10/11-playbook-for-cli-testing.md, bundled into a single script.

Usage:
    python scripts/run_cli_playbook.py
    python scripts/run_cli_playbook.py --server http://localhost:8090
"""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
import time
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
CLI_ENTRYPOINT = ["go", "run", "./cmd/zine-layout"]
DEFAULT_DATA_ROOT = REPO_ROOT / "tmp-playbook-data"
DEFAULT_SERVER = "http://localhost:8090"
DEFAULT_ADDR = ":8090"
EXAMPLE_ASSETS = [
    REPO_ROOT / "data/projects/prj-20251010T232447Z-9ti942/images/0001.png",
    REPO_ROOT / "data/projects/prj-20251010T232447Z-9ti942/images/0002.png",
    REPO_ROOT / "data/projects/prj-20251010T232447Z-9ti942/images/0003.png",
]


class PlaybookError(Exception):
    """Raised when a CLI command fails inside the playbook."""


def check_assets(paths: list[Path]) -> None:
    missing = [p for p in paths if not p.exists()]
    if missing:
        msg = "\n".join(f"  - {p}" for p in missing)
        raise PlaybookError(f"Required sample assets are missing:\n{msg}")


def run_cli(args: list[str], *, expect_json: bool = False) -> object | str:
    cmd = CLI_ENTRYPOINT + args
    print(f"\n>>> zine-layout {' '.join(args)}")
    proc = subprocess.run(cmd, capture_output=True, text=True)
    if proc.stdout.strip():
        print(proc.stdout.strip())
    if proc.stderr.strip():
        print(proc.stderr.strip(), file=sys.stderr)
    if proc.returncode != 0:
        raise PlaybookError(
            f"Command failed ({' '.join(args)}), exit code {proc.returncode}"
        )
    if expect_json:
        return json.loads(proc.stdout)
    return proc.stdout


def run_playbook(server: str, data_root: Path) -> None:
    check_assets(EXAMPLE_ASSETS)

    if data_root.exists():
        shutil.rmtree(data_root)
    data_root.mkdir(parents=True, exist_ok=True)

    server_proc = subprocess.Popen(
        CLI_ENTRYPOINT
        + [
            "serve",
            "--root",
            "cmd/zine-layout/dist",
            "--data-root",
            str(data_root),
            "--addr",
            DEFAULT_ADDR,
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        cwd=REPO_ROOT,
    )

    try:
        time.sleep(3)

        project_json = run_cli(
            [
                "api",
                "projects-create",
                "--server",
                server,
                "--name",
                "Playbook Project",
                "--output",
                "json",
            ],
            expect_json=True,
        )
        project_id = project_json[0]["id"]

        run_cli(["api", "projects-list", "--server", server])
        run_cli(
            ["api", "projects-list", "--server", server, "--output", "json"],
            expect_json=True,
        )
        run_cli(["api", "projects-get", "--server", server, "--id", project_id])

        upload_args = ["api", "images-upload", "--server", server, "--project-id", project_id]
        for path in EXAMPLE_ASSETS:
            upload_args.extend(["--files", str(path)])
        run_cli(upload_args)

        assets_json = run_cli(
            [
                "api",
                "images-list",
                "--server",
                server,
                "--project-id",
                project_id,
                "--output",
                "json",
            ],
            expect_json=True,
        )
        asset_ids = [asset["asset_id"] for asset in assets_json]

        seq_json = run_cli(
            [
                "api",
                "image-sequences",
                "create",
                "--server",
                server,
                "--project-id",
                project_id,
                "--name",
                "Playbook Sequence",
                "--output",
                "json",
            ],
            expect_json=True,
        )
        sequence_id = seq_json[0]["sequence_id"]

        for asset_id in asset_ids:
            run_cli(
                [
                    "api",
                    "image-sequences",
                    "add-item",
                    "--server",
                    server,
                    "--sequence-id",
                    sequence_id,
                    "--asset-id",
                    asset_id,
                ]
            )

        run_cli(["api", "image-sequences", "list", "--server", server, "--project-id", project_id])
        run_cli(["api", "image-sequences", "get", "--server", server, "--sequence-id", sequence_id])

        reversed_items = ",".join(reversed(asset_ids))
        run_cli(
            [
                "api",
                "image-sequences",
                "reorder",
                "--server",
                server,
                "--sequence-id",
                sequence_id,
                "--items",
                reversed_items,
            ]
        )

        run_cli(
            [
                "api",
                "image-sequences",
                "delete-item",
                "--server",
                server,
                "--sequence-id",
                sequence_id,
                "--position",
                "0",
            ]
        )

        run_cli(["api", "image-sequences", "list", "--server", server, "--project-id", project_id])

        template_settings = json.dumps(
            {
                "mode": "fit",
                "paper_width_in": 8.5,
                "paper_height_in": 11,
                "dpi": 300,
                "margin_top_in": 0.25,
                "margin_right_in": 0.25,
                "margin_bottom_in": 0.25,
                "margin_left_in": 0.25,
            }
        )
        template_json = run_cli(
            [
                "api",
                "image-layout-templates",
                "create",
                "--server",
                server,
                "--project-id",
                project_id,
                "--name",
                "Playbook Template",
                "--settings-json",
                template_settings,
                "--output",
                "json",
            ],
            expect_json=True,
        )
        template_id = template_json[0]["template_id"]

        run_cli(
            [
                "api",
                "image-layout-templates",
                "list",
                "--server",
                server,
                "--project-id",
                project_id,
            ]
        )
        run_cli(
            [
                "api",
                "image-layout-templates",
                "get",
                "--server",
                server,
                "--template-id",
                template_id,
            ]
        )

        laid_json = run_cli(
            [
                "api",
                "laid-out-images",
                "create",
                "--server",
                server,
                "--project-id",
                project_id,
                "--asset-id",
                asset_ids[1],
                "--template-id",
                template_id,
                "--output",
                "json",
            ],
            expect_json=True,
        )
        laid_out_image_id = laid_json[0]["laid_out_image_id"]

        run_cli(["api", "laid-out-images", "list", "--server", server, "--project-id", project_id])
        run_cli(
            ["api", "laid-out-images", "get", "--server", server, "--id", laid_out_image_id]
        )

        layout_seq_json = run_cli(
            [
                "api",
                "layout-sequences",
                "create",
                "--server",
                server,
                "--project-id",
                project_id,
                "--name",
                "Playbook Layout Seq",
                "--output",
                "json",
            ],
            expect_json=True,
        )
        layout_sequence_id = layout_seq_json[0]["layout_sequence_id"]

        run_cli(
            [
                "api",
                "layout-sequences",
                "add-item",
                "--server",
                server,
                "--sequence-id",
                layout_sequence_id,
                "--laid-out-image-id",
                laid_out_image_id,
            ]
        )

        run_cli(
            ["api", "layout-sequences", "list", "--server", server, "--project-id", project_id]
        )
        run_cli(
            ["api", "layout-sequences", "get", "--server", server, "--sequence-id", layout_sequence_id]
        )

        run_cli(
            [
                "api",
                "layout-sequences",
                "delete-item",
                "--server",
                server,
                "--sequence-id",
                layout_sequence_id,
                "--position",
                "0",
            ]
        )
        run_cli(
            [
                "api",
                "layout-sequences",
                "delete",
                "--server",
                server,
                "--sequence-id",
                layout_sequence_id,
            ]
        )

        run_cli(
            [
                "api",
                "image-sequences",
                "delete",
                "--server",
                server,
                "--sequence-id",
                sequence_id,
            ]
        )

        print("\nPlaybook completed successfully.")

    finally:
        server_proc.terminate()
        try:
            server_proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            server_proc.kill()


def main() -> int:
    parser = argparse.ArgumentParser(description="Run the zine-layout CLI playbook.")
    parser.add_argument(
        "--server",
        default=DEFAULT_SERVER,
        help=f"Server base URL (default: {DEFAULT_SERVER})",
    )
    parser.add_argument(
        "--data-root",
        default=str(DEFAULT_DATA_ROOT),
        help=f"Filesystem root for temporary data (default: {DEFAULT_DATA_ROOT})",
    )
    args = parser.parse_args()

    try:
        run_playbook(args.server, Path(args.data_root))
    except PlaybookError as exc:
        print(f"\nPlaybook failed: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
