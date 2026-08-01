#!/usr/bin/env python
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""recommendation_server.py — gRPC server for the RecommendationService.

Replicates the upstream GoogleCloudPlatform/microservices-demo
recommendationservice:

  * ListRecommendations fetches the full product catalog from the
    productcatalogservice, drops the products that are already in the
    user's cart and returns up to MAX_RESPONSES randomly sampled
    product ids.
  * A standard grpc.health.v1.Health service answers health checks
    (Check -> SERVING, Watch -> UNIMPLEMENTED).

Runtime configuration (environment variables):

  PORT                         listen port (default: 8080)
  PRODUCT_CATALOG_SERVICE_ADDR productcatalogservice gRPC address
                               (default: localhost:3550)
"""

import os
import random
import sys
import time
from concurrent import futures

import grpc
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc

# Make the committed genproto/python stubs importable no matter where this
# module is run from (see src/README.md "使用生成代码").
_REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
_GENPROTO_PY = os.path.join(_REPO_ROOT, "genproto", "python")
if _GENPROTO_PY not in sys.path:
    sys.path.insert(0, _GENPROTO_PY)

import demo_pb2  # noqa: E402
import product_catalog_pb2_grpc  # noqa: E402
import recommendation_pb2  # noqa: E402
import recommendation_pb2_grpc  # noqa: E402

from logger import getJSONLogger  # noqa: E402

logger = getJSONLogger("recommendationservice-server")

DEFAULT_PORT = "8080"
DEFAULT_CATALOG_ADDR = "localhost:3550"
MAX_RESPONSES = 5
HEALTH_SERVICE_NAME = "hipstershop.RecommendationService"


class RecommendationService(recommendation_pb2_grpc.RecommendationServiceServicer):
    """Servicer implementing hipstershop.RecommendationService.ListRecommendations."""

    def __init__(self, product_catalog_stub):
        self._product_catalog_stub = product_catalog_stub

    def ListRecommendations(self, request, context):
        # Fetch the full product catalog from the productcatalogservice.
        cat_response = self._product_catalog_stub.ListProducts(demo_pb2.Empty())
        all_ids = [product.id for product in cat_response.products]

        # Never recommend a product that is already in the user's cart.
        cart_ids = set(request.product_ids)
        candidates = [pid for pid in all_ids if pid not in cart_ids]

        num_return = min(MAX_RESPONSES, len(candidates))
        selected = random.sample(candidates, num_return) if candidates else []

        logger.info("[Recv ListRecommendations] product_ids=%s", selected)

        response = recommendation_pb2.ListRecommendationsResponse()
        response.product_ids.extend(selected)
        return response


class HealthService(health_pb2_grpc.HealthServicer):
    """grpc.health.v1.Health implementation for this service."""

    def Check(self, request, context):
        return health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.SERVING
        )

    def Watch(self, request, context):
        yield health_pb2.HealthCheckResponse(
            status=health_pb2.HealthCheckResponse.UNIMPLEMENTED
        )


def create_server(product_catalog_stub, server=None):
    """Registers this service's servicers on a gRPC server and returns it.

    If `server` is None a fresh server is created (the caller decides the
    listen port, e.g. via server.add_insecure_port('localhost:0')). Passing
    an existing server lets callers host multiple services on one port
    (used by the smoke tests to run a fake productcatalogservice alongside).
    """
    if server is None:
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    recommendation_pb2_grpc.add_RecommendationServiceServicer_to_server(
        RecommendationService(product_catalog_stub), server
    )
    health_pb2_grpc.add_HealthServicer_to_server(HealthService(), server)
    return server


def main():
    port = os.environ.get("PORT", DEFAULT_PORT)
    catalog_addr = os.environ.get(
        "PRODUCT_CATALOG_SERVICE_ADDR", DEFAULT_CATALOG_ADDR
    )
    logger.info("product catalog address: %s", catalog_addr)

    channel = grpc.insecure_channel(catalog_addr)
    product_catalog_stub = product_catalog_pb2_grpc.ProductCatalogServiceStub(channel)
    server = create_server(product_catalog_stub)
    logger.info("listening on port: %s", port)
    server.add_insecure_port("[::]:" + port)
    server.start()

    try:
        while True:
            time.sleep(10000)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == "__main__":
    main()
