"""Adds the committed Python genproto directory to ``sys.path``.

The generated protobuf/gRPC modules live in ``genproto/python`` (repo root)
and are imported by the service modules below.  This tiny bootstrap keeps
those imports working whether the service is launched from the repo root or
from ``src/shoppingassistantservice`` directly.
"""

import os
import sys

GENPROTO_PYTHON = os.path.normpath(
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "genproto", "python")
)

if GENPROTO_PYTHON not in sys.path:
    sys.path.insert(0, GENPROTO_PYTHON)
