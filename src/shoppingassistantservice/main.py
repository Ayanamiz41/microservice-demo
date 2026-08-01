"""Entrypoint for the local shoppingassistantservice (Python).

Usage:
    python main.py [--port 8082]

Port defaults to the ``PORT`` env var, then 8082 (upstream default).
The service registers ShoppingAssistantService.GetCompletion plus the
standard grpc.health.v1.Health service.
"""

import argparse
import logging
import os
import signal

import paths  # noqa: F401  (sys.path bootstrap for genproto/python)

from assistant import Assistant
from catalog import ProductCatalogClient
from server import SERVICE_NAME, build_server

LOG = logging.getLogger(__name__)


def serve(port):
    """Starts the gRPC server on the given port and returns it."""
    catalog = ProductCatalogClient()  # optional; fails soft when unreachable
    server, _health = build_server(assistant=Assistant(catalog=catalog))
    server.add_insecure_port(f"0.0.0.0:{port}")
    server.start()
    LOG.info(
        "shoppingassistantservice listening on 0.0.0.0:%s (health service=%s)",
        port,
        SERVICE_NAME,
    )
    return server


def main():
    parser = argparse.ArgumentParser(description="Local ShoppingAssistantService (gRPC).")
    parser.add_argument(
        "--port", type=int, default=int(os.environ.get("PORT", "8082"))
    )
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s"
    )

    server = serve(args.port)

    def _stop(_signum, _frame):
        LOG.info("shutting down...")
        server.stop(grace=2)

    signal.signal(signal.SIGINT, _stop)
    signal.signal(signal.SIGTERM, _stop)
    server.wait_for_termination()


if __name__ == "__main__":
    main()
