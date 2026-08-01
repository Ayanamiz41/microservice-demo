#!/usr/bin/env python3
# -*- coding: utf-8 -*-
#
# email_client.py — minimal command-line client for hipstershop.EmailService.
# Useful for manually smoke-testing a locally running emailservice:
#
#   python email_client.py --email alice@example.com --order-id 1234
#
# When the server is running in simulated mode (no SMTP_HOST) the request
# is logged by the server; the client reports the gRPC status it received.

import argparse
import os
import sys
from pathlib import Path

import grpc

# genproto/python stubs (same wiring as email_server.py).
_GENPROTO_PYTHON = Path(__file__).resolve().parent.parent.parent / "genproto" / "python"
if str(_GENPROTO_PYTHON) not in sys.path:
    sys.path.insert(0, str(_GENPROTO_PYTHON))

import demo_pb2  # noqa: E402
import email_pb2  # noqa: E402
import email_pb2_grpc  # noqa: E402


def build_request(email, order_id):
    request = email_pb2.SendOrderConfirmationRequest()
    request.email = email
    request.order.order_id = order_id
    request.order.shipping_tracking_id = "TRACK-{}".format(order_id)
    request.order.shipping_cost.currency_code = "USD"
    request.order.shipping_cost.units = 5
    request.order.shipping_cost.nanos = 990000000
    request.order.shipping_address.street_address = "1 Microservices Way"
    request.order.shipping_address.city = "Mountain View"
    request.order.shipping_address.state = "CA"
    request.order.shipping_address.country = "USA"
    request.order.shipping_address.zip_code = 94040

    item = request.order.items.add()
    item.item.product_id = "OLJCESPC7Z"
    item.item.quantity = 2
    item.cost.currency_code = "USD"
    item.cost.units = 10
    item.cost.nanos = 0
    return request


def main():
    parser = argparse.ArgumentParser(description="hipstershop.EmailService test client")
    parser.add_argument("--addr", default=os.environ.get("EMAILSERVICE_ADDR", "localhost:8080"), help="server address (default localhost:8080)")
    parser.add_argument("--email", default="test@example.com", help="recipient email address")
    parser.add_argument("--order-id", default="12345", help="order id to include in the request")
    args = parser.parse_args()

    request = build_request(args.email, args.order_id)
    with grpc.insecure_channel(args.addr) as channel:
        stub = email_pb2_grpc.EmailServiceStub(channel)
        try:
            stub.SendOrderConfirmation(request, timeout=10)
            print("SendOrderConfirmation OK -> Empty (email to {} for order {})".format(args.email, args.order_id))
        except grpc.RpcError as exc:
            print("SendOrderConfirmation FAILED: code={} details={!r}".format(exc.code(), exc.details()))
            sys.exit(1)


if __name__ == "__main__":
    main()
