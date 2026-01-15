import argparse
import pkgutil
import sys
import commands

def main():
    parser = argparse.ArgumentParser(
        prog="kron run",
        description="Run kron modules and scripts",
        add_help=False  # 🔥 IMPORTANT
    )

    # Manually add help
    parser.add_argument(
        "-h", "--help",
        action="store_true",
        help="Show this help message and exit"
    )

    sub = parser.add_subparsers(
        dest="command",
        metavar="<command>"
    )

    # 🔹 Dynamically load commands
    for module in pkgutil.iter_modules(commands.__path__):
        m = __import__(f"commands.{module.name}", fromlist=["register"])
        if hasattr(m, "register"):
            m.register(sub)

    args, unknown = parser.parse_known_args()

    # 🔹 Explicit help only
    if args.help:
        parser.print_help()
        return

    # 🔹 If a command matched
    if hasattr(args, "func"):
        args.func(args)
        return

    # 🔹 Default behavior (NO help spam)
    # This is where your actual runner logic goes
    print("Running default kron run behavior with:", unknown)

if __name__ == "__main__":
    main()
