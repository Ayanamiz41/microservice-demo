"""Pytest bootstrap for recommendationservice.

Makes the service directory (for `import recommendation_server`, `logger`)
and the committed genproto/python stubs importable from any CWD.
"""

import os
import sys

_SERVICE_DIR = os.path.dirname(os.path.abspath(__file__))
_REPO_ROOT = os.path.dirname(os.path.dirname(_SERVICE_DIR))
_GENPROTO_PY = os.path.join(_REPO_ROOT, "genproto", "python")

for _path in (_SERVICE_DIR, _GENPROTO_PY):
    if _path not in sys.path:
        sys.path.insert(0, _path)
