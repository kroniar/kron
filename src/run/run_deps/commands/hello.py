def register(sub):
    p = sub.add_parser(
        "hello",
        help="Print hello message",
        description="Prints a hello message from kron"
    )
    p.set_defaults(func=run)

def run(args):
    print("Hello dynamically loaded")
