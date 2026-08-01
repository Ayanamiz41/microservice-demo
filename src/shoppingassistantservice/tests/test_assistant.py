"""Unit tests for the rule-based assistant engine.

These tests run with a fake in-memory catalog, so they need no running
productcatalogservice (the Go service is a separate deployment).
"""

import os
import sys

# Make the service dir importable (assistant.py, paths.py live there).
SERVICE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if SERVICE_DIR not in sys.path:
    sys.path.insert(0, SERVICE_DIR)

import paths  # noqa: F401,E402  (sys.path bootstrap for genproto/python)

import demo_pb2  # noqa: E402  (Product / Money live in demo.proto)

from assistant import Assistant, format_price  # noqa: E402


def make_product(pid, name, units=10, nanos=0, categories=("general",), description="desc"):
    product = demo_pb2.Product(id=pid, name=name, description=description)
    product.price_usd.units = units
    product.price_usd.nanos = nanos
    product.categories[:] = categories
    return product


class FakeCatalog:
    """In-memory stand-in for ProductCatalogClient."""

    def __init__(self, products):
        self._products = products

    def list_products(self):
        return list(self._products)

    def search_products(self, query):
        query = query.lower()
        return [
            p
            for p in self._products
            if query in p.name.lower()
            or query in p.description.lower()
            or any(query in c.lower() for c in p.categories)
        ]

    def get_product(self, product_id):
        return next((p for p in self._products if p.id == product_id), None)


PRODUCTS = [
    make_product("1", "Sunglasses", units=22, nanos=0, categories=("accessories",)),
    make_product("2", "Running Shoes", units=95, nanos=500000000, categories=("footwear",)),
    make_product("3", "T-Shirt", units=12, nanos=990000000, categories=("clothing",)),
    make_product("4", "Sneakers", units=74, nanos=0, categories=("footwear",)),
]


def reply_for(message, assistant=None, context_ids=(), last_context=None):
    assistant = assistant or Assistant(catalog=FakeCatalog(PRODUCTS))
    reply, _ctx = assistant.get_completion(message, context_ids, last_context)
    return reply


# ---------------------------------------------------------------- greetings

def test_greeting_returns_welcome():
    reply = reply_for("Hello!")
    assert "Online Boutique" in reply
    assert "assistant" in reply.lower()


def test_greeting_chinese():
    reply = reply_for("你好")
    assert "Online Boutique" in reply


def test_empty_message_falls_back_to_capabilities():
    reply = reply_for("")
    assert "recommend" in reply.lower()


# ------------------------------------------------------------------ thanks

def test_thanks_reply():
    reply = reply_for("Thanks a lot!")
    assert "welcome" in reply.lower()


# ------------------------------------------------------------ recommendation

def test_recommendation_lists_products_with_prices():
    assistant = Assistant(catalog=FakeCatalog(PRODUCTS))
    reply, ctx = assistant.get_completion("Can you recommend a gift?")
    assert "Sunglasses" in reply
    assert "$22.00" in reply
    assert "recommended_ids" in ctx
    assert ctx["recommended_ids"] == ["1", "2", "3"]


def test_recommendation_prefers_related_to_context_products():
    assistant = Assistant(catalog=FakeCatalog(PRODUCTS))
    reply, _ctx = assistant.get_completion(
        "What should I buy?", context_product_ids=("2",)  # Running Shoes
    )
    # Related items (same "footwear" category) come first.
    assert "Sneakers" in reply


def test_recommendation_without_catalog_falls_back_gracefully():
    reply = reply_for("recommend something", assistant=Assistant())
    assert "unreachable" in reply.lower()


# ------------------------------------------------------------------ details

def test_details_search_returns_product_info():
    assistant = Assistant(catalog=FakeCatalog(PRODUCTS))
    reply, ctx = assistant.get_completion("Tell me about the shoes")
    assert "Running Shoes" in reply or "Sneakers" in reply
    assert "detailed_ids" in ctx


def test_details_without_catalog_falls_back_gracefully():
    reply = reply_for("how much is the t-shirt?", assistant=Assistant())
    assert "couldn't find" in reply.lower() or "unreachable" in reply.lower()


# ------------------------------------------------------------ order tracking

def test_order_tracking_reply_points_to_support():
    reply = reply_for("Where is my order?")
    assert "support" in reply.lower()
    assert "order number" in reply.lower()


# ------------------------------------------------------------------ follow-up

def test_followup_expands_on_last_recommendations():
    assistant = Assistant(catalog=FakeCatalog(PRODUCTS))
    _reply, ctx = assistant.get_completion("recommend something for me")
    assert ctx["recommended_ids"]
    reply, _ctx2 = assistant.get_completion("tell me more", last_context=ctx)
    assert "Running Shoes" in reply or "Sunglasses" in reply


def test_followup_without_context_falls_back_to_generic():
    reply = reply_for("tell me more")
    assert "recommend" in reply.lower()


# ------------------------------------------------------------------ fallback

def test_unrecognized_message_lists_capabilities():
    reply = reply_for("asdfgh")
    assert "recommend" in reply.lower()


# ------------------------------------------------------------------ money

def test_format_price():
    money = demo_pb2.Money(currency_code="USD", units=22, nanos=0)
    assert format_price(money) == "$22.00"
    money.units = 12
    money.nanos = 990000000
    assert format_price(money) == "$12.99"
    assert format_price(None) == "n/a"
