"""Thin gRPC client for the local ProductCatalogService (Go service).

The assistant uses the catalog to answer recommendation / product-details
questions.  The catalog is an independent deployment (productcatalogservice,
port 3550) and may be down while the assistant itself is up, so every call
fails soft: methods return ``None`` / ``[]`` on error and never raise toward
the caller.
"""

import os

import grpc

import paths  # noqa: F401  (sys.path bootstrap for genproto/python)

import demo_pb2
import product_catalog_pb2
import product_catalog_pb2_grpc

DEFAULT_CATALOG_ADDR = "localhost:3550"


class ProductCatalogClient:
    """Graceful client for ProductCatalogService.

    ``target`` defaults to ``PRODUCT_CATALOG_SERVICE_ADDR`` (or
    ``localhost:3550``); ``timeout`` bounds every catalog call so a slow or
    unreachable catalog never blocks an assistant reply for long.
    """

    def __init__(self, target=None, timeout=2.0):
        self._target = target or os.environ.get(
            "PRODUCT_CATALOG_SERVICE_ADDR", DEFAULT_CATALOG_ADDR
        )
        self._timeout = timeout
        self._channel = grpc.insecure_channel(self._target)
        self._stub = product_catalog_pb2_grpc.ProductCatalogServiceStub(self._channel)

    def close(self):
        self._channel.close()

    def list_products(self):
        try:
            response = self._stub.ListProducts(
                demo_pb2.Empty(), timeout=self._timeout
            )
            return list(response.products)
        except grpc.RpcError:
            return []

    def get_product(self, product_id):
        try:
            return self._stub.GetProduct(
                product_catalog_pb2.GetProductRequest(id=product_id),
                timeout=self._timeout,
            )
        except grpc.RpcError:
            return None

    def search_products(self, query):
        try:
            response = self._stub.SearchProducts(
                product_catalog_pb2.SearchProductsRequest(query=query),
                timeout=self._timeout,
            )
            return list(response.results)
        except grpc.RpcError:
            return []
