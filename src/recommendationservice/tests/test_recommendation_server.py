"""Smoke tests for the recommendationservice.

Covers:
  * unit-level behaviour of ListRecommendations (filtering, cap, empty
    catalog) against a fake product catalog stub;
  * an in-process gRPC round trip of the full server (start server on an
    ephemeral port -> real ListRecommendations call -> real
    grpc.health.v1.Health/Check call).
"""

from concurrent import futures

import grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc

import demo_pb2
import product_catalog_pb2
import product_catalog_pb2_grpc
import recommendation_pb2
import recommendation_pb2_grpc

import recommendation_server


class FakeProductCatalog:
    """In-process stand-in for productcatalogservice (duck-typed stub)."""

    def __init__(self, product_ids):
        self.product_ids = list(product_ids)

    def ListProducts(self, request):
        response = product_catalog_pb2.ListProductsResponse()
        for pid in self.product_ids:
            response.products.add(id=pid, name="product-" + pid)
        return response


class FakeProductCatalogServicer(product_catalog_pb2_grpc.ProductCatalogServiceServicer):
    """gRPC servicer for productcatalogservice used in the E2E smoke test."""

    def __init__(self, product_ids):
        self.product_ids = list(product_ids)

    def ListProducts(self, request, context):
        response = product_catalog_pb2.ListProductsResponse()
        for pid in self.product_ids:
            response.products.add(id=pid, name="product-" + pid)
        return response


def _request(user_id="user-1", product_ids=()):
    req = recommendation_pb2.ListRecommendationsRequest()
    req.user_id = user_id
    req.product_ids.extend(product_ids)
    return req


def test_list_recommendations_filters_cart_items():
    ids = ["p1", "p2", "p3", "p4", "p5", "p6", "p7"]
    service = recommendation_server.RecommendationService(FakeProductCatalog(ids))

    result = service.ListRecommendations(_request(product_ids=["p1", "p2"]), None)

    returned = list(result.product_ids)
    assert len(returned) <= recommendation_server.MAX_RESPONSES
    assert set(returned).issubset(set(ids))
    assert "p1" not in returned and "p2" not in returned
    # All candidates are available (7 > 5), so a full batch is returned.
    assert len(returned) == recommendation_server.MAX_RESPONSES


def test_list_recommendations_empty_cart_returns_full_batch():
    ids = ["p1", "p2", "p3", "p4", "p5", "p6"]
    service = recommendation_server.RecommendationService(FakeProductCatalog(ids))

    result = service.ListRecommendations(_request(product_ids=[]), None)

    returned = list(result.product_ids)
    assert len(returned) == recommendation_server.MAX_RESPONSES
    assert set(returned).issubset(set(ids))


def test_list_recommendations_less_than_cap():
    ids = ["p1", "p2", "p3"]
    service = recommendation_server.RecommendationService(FakeProductCatalog(ids))

    result = service.ListRecommendations(_request(), None)

    returned = list(result.product_ids)
    assert sorted(returned) == sorted(ids)


def test_list_recommendations_empty_catalog():
    service = recommendation_server.RecommendationService(FakeProductCatalog([]))

    result = service.ListRecommendations(_request(), None)

    assert list(result.product_ids) == []


def test_list_recommendations_empty_cart_does_not_die_on_single_product():
    service = recommendation_server.RecommendationService(FakeProductCatalog(["only"]))

    result = service.ListRecommendations(_request(), None)

    assert list(result.product_ids) == ["only"]


def test_end_to_end_smoke_grpc_and_health_check():
    ids = ["a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a10"]

    # One real server hosting the fake ProductCatalogService AND the real
    # RecommendationService (+ health service); the recommendation servicer
    # talks to the fake catalog over a real gRPC channel.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    port = server.add_insecure_port("localhost:0")
    catalog_channel = grpc.insecure_channel("localhost:%d" % port)
    product_catalog_pb2_grpc.add_ProductCatalogServiceServicer_to_server(
        FakeProductCatalogServicer(ids), server
    )
    recommendation_server.create_server(
        product_catalog_pb2_grpc.ProductCatalogServiceStub(catalog_channel),
        server=server,
    )
    server.start()

    client_channel = grpc.insecure_channel("localhost:%d" % port)
    try:
        stub = recommendation_pb2_grpc.RecommendationServiceStub(client_channel)
        response = stub.ListRecommendations(_request(product_ids=["a1"]))

        returned = list(response.product_ids)
        assert len(returned) <= recommendation_server.MAX_RESPONSES
        assert "a1" not in returned
        assert set(returned).issubset(set(ids))

        health_stub = health_pb2_grpc.HealthStub(client_channel)
        status = health_stub.Check(
            health_pb2.HealthCheckRequest(
                service=recommendation_server.HEALTH_SERVICE_NAME
            )
        ).status
        assert status == health_pb2.HealthCheckResponse.SERVING
    finally:
        server.stop(0)
        catalog_channel.close()
        client_channel.close()


def test_health_check_serving_for_empty_service_name():
    server = recommendation_server.create_server(
        product_catalog_pb2_grpc.ProductCatalogServiceStub(
            grpc.insecure_channel("localhost:0")
        )
    )
    port = server.add_insecure_port("localhost:0")
    server.start()
    channel = grpc.insecure_channel("localhost:%d" % port)
    try:
        health_stub = health_pb2_grpc.HealthStub(channel)
        status = health_stub.Check(health_pb2.HealthCheckRequest(service="")).status
        assert status == health_pb2.HealthCheckResponse.SERVING
    finally:
        server.stop(0)
        channel.close()
