#!/usr/bin/env python3
# -*- coding: utf-8 -*-
#
# emailservice — EmailService gRPC server (Python).
#
# Local replica of GoogleCloudPlatform/microservices-demo src/emailservice.
# Implements the hipstershop.EmailService contract from protos/email.proto
# (SendOrderConfirmation) plus the standard grpc.health.v1.Health service.
#
# Delivery is simulated locally by default: the confirmation email is
# rendered from templates/confirmation.html and logged. To send through a
# real SMTP relay instead, set SMTP_HOST (optionally SMTP_PORT, SMTP_USER,
# SMTP_PASSWORD, FROM_ADDRESS) in the environment.

import argparse
import logging
import os
import smtplib
import sys
import time
from concurrent import futures
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from pathlib import Path

import grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc
from jinja2 import Environment, FileSystemLoader, select_autoescape, TemplateError

# The generated protobuf stubs live in the monorepo genproto/python dir;
# add it to sys.path (relative to this file) before importing them.
_GENPROTO_PYTHON = Path(__file__).resolve().parent.parent.parent / "genproto" / "python"
if str(_GENPROTO_PYTHON) not in sys.path:
    sys.path.insert(0, str(_GENPROTO_PYTHON))

import demo_pb2  # noqa: E402  (imported for type references / tests)
import email_pb2  # noqa: E402
import email_pb2_grpc  # noqa: E402

logger = logging.getLogger("emailservice")

SERVICE_NAME = "hipstershop.EmailService"
DEFAULT_PORT = 8080

TEMPLATES_DIR = Path(__file__).resolve().parent / "templates"
env = Environment(
    loader=FileSystemLoader(str(TEMPLATES_DIR)),
    autoescape=select_autoescape(["html", "xml"]),
)
confirmation_template = env.get_template("confirmation.html")


def format_money(money):
    """Format a hipstershop.Money as '<units>.<nanos2> <currency_code>'.

    nanos is rendered with two decimal digits, mirroring upstream behavior
    (nanos // 10_000_000 gives the cent part).
    """
    units = int(money.units)
    nanos = int(money.nanos)
    sign = "-" if (units < 0 or nanos < 0) else ""
    return "{}{}.{:02d} {}".format(sign, abs(units), abs(nanos) // 10_000_000, money.currency_code)


def render_confirmation(order):
    """Render the order confirmation HTML for a hipstershop.OrderResult."""
    try:
        return confirmation_template.render(order=order)
    except TemplateError as exc:
        raise RuntimeError("failed to render confirmation email: {}".format(exc)) from exc


def send_email_smtp(recipient, subject, html_body):
    """Send an HTML email through the SMTP relay configured via env vars."""
    host = os.environ["SMTP_HOST"]
    port = int(os.environ.get("SMTP_PORT", "587"))
    user = os.environ.get("SMTP_USER")
    password = os.environ.get("SMTP_PASSWORD")
    from_address = os.environ.get("FROM_ADDRESS", "no-reply@example.com")

    msg = MIMEMultipart("alternative")
    msg["Subject"] = subject
    msg["From"] = from_address
    msg["To"] = recipient
    msg.attach(MIMEText(html_body, "html", "utf-8"))

    with smtplib.SMTP(host, port) as smtp:
        smtp.starttls()
        if user:
            smtp.login(user, password)
        smtp.sendmail(from_address, [recipient], msg.as_string())


class EmailService(email_pb2_grpc.EmailServiceServicer):
    """hipstershop.EmailService — sends order confirmation emails."""

    def SendOrderConfirmation(self, request, context):
        email = request.email
        order = request.order
        logger.info("Received SendOrderConfirmation request for %s (order %s)", email, order.order_id)

        try:
            html_body = render_confirmation(order)
        except RuntimeError as exc:
            logger.error("SendOrderConfirmation failed while preparing the mail: %s", exc)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details("An error occurred when preparing the confirmation mail.")
            return demo_pb2.Empty()

        subject = "Your Order Confirmation"
        if order.order_id:
            subject = "Your Order Confirmation (#{})".format(order.order_id)

        if os.environ.get("SMTP_HOST"):
            try:
                send_email_smtp(email, subject, html_body)
                logger.info("Order confirmation email sent to %s (order %s)", email, order.order_id)
            except Exception as exc:  # noqa: BLE001 — surface any delivery failure to the caller
                logger.error("Sending email to %s failed: %s", email, exc)
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("An error occurred when sending the email.")
                return demo_pb2.Empty()
        else:
            # Local simulation: no SMTP relay configured, just log the email.
            logger.info("Simulated sending order confirmation to %s (order %s)", email, order.order_id)
            logger.debug("Subject: %s", subject)
            logger.debug("HTML body (%d bytes):\n%s", len(html_body), html_body)

        return demo_pb2.Empty()


class HealthServicer(health_pb2_grpc.HealthServicer):
    """grpc.health.v1.Health — SERVING for this service, NOT_FOUND otherwise."""

    def Check(self, request, context):
        if request.service in ("", SERVICE_NAME):
            return health_pb2.HealthCheckResponse(status=health_pb2.HealthCheckResponse.SERVING)
        context.set_code(grpc.StatusCode.NOT_FOUND)
        context.set_details("unknown service: {}".format(request.service))
        return health_pb2.HealthCheckResponse(status=health_pb2.HealthCheckResponse.NOT_SERVING)


def build_server():
    """Create a fully wired gRPC server (EmailService + Health), not started."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    email_pb2_grpc.add_EmailServiceServicer_to_server(EmailService(), server)
    health_pb2_grpc.add_HealthServicer_to_server(HealthServicer(), server)
    return server


def serve(port=DEFAULT_PORT):
    """Start the gRPC server on the given port and block until interrupted."""
    server = build_server()
    server.add_insecure_port("[::]:{}".format(port))
    server.start()
    logger.info("emailservice listening on port %s", port)
    try:
        while True:
            time.sleep(3600)
    except KeyboardInterrupt:
        server.stop(0)


def main():
    parser = argparse.ArgumentParser(description="emailservice gRPC server")
    parser.add_argument(
        "--port",
        type=int,
        default=int(os.environ.get("PORT", DEFAULT_PORT)),
        help="gRPC listen port (default: $PORT or {})".format(DEFAULT_PORT),
    )
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if os.environ.get("EMAILSERVICE_DEBUG") else logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )
    serve(args.port)


if __name__ == "__main__":
    main()
