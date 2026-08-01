"""gRPC server wiring for ShoppingAssistantService (package hipstershop)."""

import logging
import threading
import uuid
from concurrent import futures

import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
from grpc_health.v1.health import HealthServicer

import paths  # noqa: F401  (sys.path bootstrap for genproto/python)

import shopping_assistant_pb2
import shopping_assistant_pb2_grpc

from assistant import Assistant

SERVICE_NAME = "hipstershop.ShoppingAssistantService"

LOG = logging.getLogger(__name__)


class ShoppingAssistantServicer(
    shopping_assistant_pb2_grpc.ShoppingAssistantServiceServicer
):
    """Implements ``GetCompletion`` plus per-conversation continuity.

    ``conversation_id`` from the request is echoed back (or allocated on the
    first message) and used to keep a small in-memory context so follow-up
    messages like "tell me more" can expand on earlier recommendations.
    """

    def __init__(self, assistant=None):
        self._assistant = assistant or Assistant()
        self._sessions = {}
        self._lock = threading.Lock()

    def GetCompletion(self, request, context):
        conversation_id = request.conversation_id or uuid.uuid4().hex
        with self._lock:
            last_context = self._sessions.get(conversation_id)
            reply, new_context = self._assistant.get_completion(
                request.message,
                context_product_ids=list(request.context_product_ids),
                last_context=last_context,
            )
            self._sessions[conversation_id] = new_context
        return shopping_assistant_pb2.ShoppingAssistantResponse(
            message=reply, conversation_id=conversation_id
        )


def build_server(assistant=None):
    """Builds a gRPC server with the assistant + the standard health service.

    Health reports SERVING for both the overall server ("") and
    ``hipstershop.ShoppingAssistantService``, which is the service name used
    by the repo-wide health-check convention.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    shopping_assistant_pb2_grpc.add_ShoppingAssistantServiceServicer_to_server(
        ShoppingAssistantServicer(assistant), server
    )
    health = HealthServicer()
    health.set("", health_pb2.HealthCheckResponse.SERVING)
    health.set(SERVICE_NAME, health_pb2.HealthCheckResponse.SERVING)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    return server, health
