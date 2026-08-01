"""Smoke tests for the gRPC server: health check + GetCompletion round-trip.

The server is started in-process on an ephemeral port, so these tests need
no external services running.
"""

import os
import sys

SERVICE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if SERVICE_DIR not in sys.path:
    sys.path.insert(0, SERVICE_DIR)

import pytest  # noqa: E402

import grpc  # noqa: E402
from grpc_health.v1 import health_pb2, health_pb2_grpc  # noqa: E402

import paths  # noqa: F401,E402

import shopping_assistant_pb2  # noqa: E402
import shopping_assistant_pb2_grpc  # noqa: E402

import server as server_mod  # noqa: E402


@pytest.fixture
def running_server():
    server, _health = server_mod.build_server()
    port = server.add_insecure_port("127.0.0.1:0")
    server.start()
    target = f"127.0.0.1:{port}"
    yield target
    server.stop(grace=0)


def test_health_check_returns_serving(running_server):
    with grpc.insecure_channel(running_server) as channel:
        stub = health_pb2_grpc.HealthStub(channel)
        response = stub.Check(
            health_pb2.HealthCheckRequest(service=server_mod.SERVICE_NAME)
        )
        assert response.status == health_pb2.HealthCheckResponse.SERVING


def test_overall_health_check_returns_serving(running_server):
    with grpc.insecure_channel(running_server) as channel:
        stub = health_pb2_grpc.HealthStub(channel)
        response = stub.Check(health_pb2.HealthCheckRequest(service=""))
        assert response.status == health_pb2.HealthCheckResponse.SERVING


def test_health_check_unknown_service_not_found(running_server):
    with grpc.insecure_channel(running_server) as channel:
        stub = health_pb2_grpc.HealthStub(channel)
        # Unknown services are reported as RPC status NOT_FOUND (per the
        # gRPC health protocol) with SERVICE_UNKNOWN in the body.
        with pytest.raises(grpc.RpcError) as excinfo:
            stub.Check(
                health_pb2.HealthCheckRequest(service="hipstershop.NoSuchService")
            )
        assert excinfo.value.code() == grpc.StatusCode.NOT_FOUND


def test_get_completion_round_trip(running_server):
    with grpc.insecure_channel(running_server) as channel:
        stub = shopping_assistant_pb2_grpc.ShoppingAssistantServiceStub(channel)
        response = stub.GetCompletion(
            shopping_assistant_pb2.ShoppingAssistantRequest(message="Hi there!")
        )
        assert response.message
        assert response.conversation_id


def test_conversation_id_echoed_for_follow_up(running_server):
    with grpc.insecure_channel(running_server) as channel:
        stub = shopping_assistant_pb2_grpc.ShoppingAssistantServiceStub(channel)
        first = stub.GetCompletion(
            shopping_assistant_pb2.ShoppingAssistantRequest(message="recommend a gift")
        )
        assert first.conversation_id
        second = stub.GetCompletion(
            shopping_assistant_pb2.ShoppingAssistantRequest(
                message="tell me more", conversation_id=first.conversation_id
            )
        )
        assert second.conversation_id == first.conversation_id
        assert second.message
