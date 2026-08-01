"""``python -m loadgenerator`` — Python entry point for the load generator.

Thin wrapper around the Locust CLI so the service can be started with a
plain Python command and the same flags as ``locust`` itself. It pins
``-f`` to this package's own ``locustfile.py`` so the working directory
does not matter.

Run from the repository root::

    # Git Bash / Linux / macOS
    PYTHONPATH=src python -m loadgenerator --host http://localhost:8080 --headless -u 10 -r 1

    # Windows PowerShell
    $env:PYTHONPATH = "src"
    python -m loadgenerator --host http://localhost:8080 --headless -u 10 -r 1

or simply ``cd src`` first (then ``src`` is on the module path) and drop the
``PYTHONPATH`` prefix. The plain ``locust`` command works too::

    cd src/loadgenerator
    locust --host http://localhost:8080 --headless -u 10 -r 1
"""

import os
import sys

from locust import main as locust_main

LOCUSTFILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "locustfile.py")


def main() -> int:
    """Run the Locust CLI (parses ``sys.argv`` like the ``locust`` command)."""
    args = list(sys.argv[1:])
    has_locustfile = ("-f" in args) or ("--locustfile" in args)
    has_help = ("-h" in args) or ("--help" in args)
    if not has_locustfile and not has_help:
        args[:0] = ["-f", LOCUSTFILE]
    sys.argv[1:] = args
    return locust_main.main()


if __name__ == "__main__":
    sys.exit(main())
