"""loadgenerator — Locust load test for the Online Boutique replica.

Client-only service: it drives the frontend's HTTP endpoints (which in turn
fan out to the gRPC backend) and defines no server contract of its own.
"""

__version__ = "1.0.0"
