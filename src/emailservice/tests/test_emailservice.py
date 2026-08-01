#!/usr/bin/env python3
# -*- coding: utf-8 -*-
#
# Smoke-level unit tests for src/emailservice.
#
# The tests boot the real gRPC server (EmailService + Health) on a random
# local port and exercise it over an in-process channel, so they cover the
# generated stubs, template rendering and the health contract end to end.

import logging
import os
import sys
from pathlib import Path

import grpc
import pytest
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc

# Make the service module importable (it wires genproto/python itself).
SERVICE_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(SERVICE_DIR))

import email_server  # noqa: E402
from email_server import demo_pb2, email_pb2, email_pb2_grpc  # noqa: E402

SERVICE_NAME = "hipstershop.EmailService"


def build_request(email="alice@example.com", order_id="ABC-123"):
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


@pytest.fixture(scope="module")
def channel():
    server = email_server.build_server()
    port = server.add_insecure_port("127.0.0.1:0")
    server.start()
    chan = grpc.insecure_channel("127.0.0.1:{}".format(port))
    yield chan
    chan.close()
    server.stop(0)


@pytest.fixture(scope="module")
def email_stub(channel):
    return email_pb2_grpc.EmailServiceStub(channel)


@pytest.fixture(scope="module")
def health_stub(channel):
    return health_pb2_grpc.HealthStub(channel)


# --- grpc.health.v1.Health -------------------------------------------------

def test_health_check_serving_for_service(health_stub):
    resp = health_stub.Check(health_pb2.HealthCheckRequest(service=SERVICE_NAME))
    assert resp.status == health_pb2.HealthCheckResponse.SERVING


def test_health_check_serving_for_empty_service(health_stub):
    resp = health_stub.Check(health_pb2.HealthCheckRequest(service=""))
    assert resp.status == health_pb2.HealthCheckResponse.SERVING


def test_health_check_unknown_service(health_stub):
    with pytest.raises(grpc.RpcError) as exc_info:
        health_stub.Check(health_pb2.HealthCheckRequest(service="no.such.Service"))
    assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND


# --- hipstershop.EmailService.SendOrderConfirmation ------------------------

def test_send_order_confirmation_returns_empty(email_stub):
    resp = email_stub.SendOrderConfirmation(build_request(), timeout=10)
    assert isinstance(resp, demo_pb2.Empty)


def test_send_order_confirmation_logs_recipient(email_stub, caplog):
    with caplog.at_level(logging.INFO, logger="emailservice"):
        email_stub.SendOrderConfirmation(build_request(email="carol@example.com", order_id="ORD-42"), timeout=10)
    assert "carol@example.com" in caplog.text
    assert "ORD-42" in caplog.text


def test_send_order_confirmation_without_order_id(email_stub):
    request = build_request(order_id="")
    request.order.order_id = ""
    resp = email_stub.SendOrderConfirmation(request, timeout=10)
    assert isinstance(resp, demo_pb2.Empty)


# --- template rendering -----------------------------------------------------

def test_render_confirmation_contains_order_details():
    request = build_request(order_id="RND-7")
    html = email_server.render_confirmation(request.order)
    assert "RND-7" in html
    assert "OLJCESPC7Z" in html
    assert "Mountain View" in html
    assert "USD" in html


def test_render_confirmation_handles_empty_order():
    html = email_server.render_confirmation(demo_pb2.OrderResult())
    assert isinstance(html, str)
    assert "Your Order Confirmation" in html


def test_format_money():
    money = demo_pb2.Money(currency_code="USD", units=5, nanos=990000000)
    assert email_server.format_money(money) == "5.99 USD"
    money = demo_pb2.Money(currency_code="USD", units=1, nanos=750000000)
    assert email_server.format_money(money) == "1.75 USD"
