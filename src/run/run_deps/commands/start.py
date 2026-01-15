"""
Start the project in the current directory.
Automatically detects the project type and runs the correct command.
"""

import os
import subprocess

def register(sub):
    p = sub.add_parser(
        "start",
        help="Smart project starter (auto-detects project type)"
    )
    p.set_defaults(func=run)

def run(args):
    cwd = os.getcwd()

    if os.path.exists("docker-compose.yml"):
        run_cmd(["docker", "compose", "up"])
        return

    if os.path.exists("Makefile"):
        run_cmd(["make"])
        return

    if os.path.exists("package.json"):
        run_cmd(["npm", "run", "dev"])
        return

    if os.path.exists("pyproject.toml") or os.path.exists("requirements.txt"):
        run_cmd(["python", "main.py"])
        return

    print("❌ No known project type detected.")
    print("Suggestions:")
    print("  - docker-compose.yml")
    print("  - Makefile")
    print("  - package.json")
    print("  - Python project")

def run_cmd(cmd):
    print("▶ Running:", " ".join(cmd))
    subprocess.run(cmd)
