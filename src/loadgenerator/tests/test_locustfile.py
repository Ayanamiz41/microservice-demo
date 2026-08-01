"""Smoke tests for the loadgenerator service.

Covers:
  * the locustfile importing cleanly and exposing the expected task graph;
  * the pure form payload builders (add-to-cart / checkout) producing
    well-formed requests;
  * the gRPC consumption path used by the stack (genproto/python clients
    import cleanly) — the load generator itself is HTTP-only against the
    frontend, matching upstream;
  * the ``python -m loadgenerator`` entry point being wired up.

Run from the repo root (or anywhere) with:
    python -m pytest src/loadgenerator/tests -q
"""

import re
import sys
from pathlib import Path

import pytest

SERVICE_DIR = Path(__file__).resolve().parents[1]
if str(SERVICE_DIR) not in sys.path:
    sys.path.insert(0, str(SERVICE_DIR))

import locustfile  # noqa: E402  (after sys.path setup)

REPO_ROOT = SERVICE_DIR.parents[1]

EXPECTED_PRODUCT_IDS = [
    '0PUK6V6EV0',
    '1YMWWN1N4O',
    '2ZYFJ3GM2N',
    '66VCHSJNUP',
    '6E92ZMYYFZ',
    '9SIQT8TOJO',
    'L9ECAV7KIM',
    'LS4PSXUNUM',
    'OLJCESPC7Z',
]

EXPECTED_TASKS = {
    locustfile.index,
    locustfile.setCurrency,
    locustfile.browseProduct,
    locustfile.addToCart,
    locustfile.viewCart,
    locustfile.checkout,
}


# --------------------------------------------------------------------------
# locustfile structure
# --------------------------------------------------------------------------

def test_products_match_upstream_catalog():
    assert locustfile.products == EXPECTED_PRODUCT_IDS
    assert len(locustfile.products) == len(set(locustfile.products))  # unique


def test_user_behavior_task_weights():
    # Locust 2.x flattens the weighted dict into a list at class creation.
    tasks = locustfile.UserBehavior.tasks
    assert set(tasks) == EXPECTED_TASKS
    # relative weights match upstream (browse dominates, checkout is rare)
    assert tasks.count(locustfile.browseProduct) == 10
    assert tasks.count(locustfile.checkout) == 1
    assert tasks.count(locustfile.index) == 1
    assert tasks.count(locustfile.setCurrency) == 2
    assert tasks.count(locustfile.addToCart) == 2
    assert tasks.count(locustfile.viewCart) == 3


def test_users_are_http_users():
    from locust import FastHttpUser
    assert issubclass(locustfile.WebsiteUser, FastHttpUser)
    assert locustfile.WebsiteUser.tasks == [locustfile.UserBehavior]
    # wait_time sampled between 1 and 10s (upstream `between(1, 10)`)
    for _ in range(50):
        t = locustfile.WebsiteUser.wait_time(None)
        assert 1 <= t <= 10


# --------------------------------------------------------------------------
# form payload builders
# --------------------------------------------------------------------------

def test_add_to_cart_form():
    for _ in range(50):
        form = locustfile.add_to_cart_form()
        assert set(form) == {'product_id', 'quantity'}
        assert form['product_id'] in locustfile.products
        assert isinstance(form['quantity'], int)
        assert 1 <= form['quantity'] <= 10


def test_checkout_form_data():
    import datetime
    year_now = datetime.datetime.now().year
    for _ in range(50):
        form = locustfile.checkout_form_data()
        assert set(form) == {
            'email', 'street_address', 'zip_code', 'city', 'state',
            'country', 'credit_card_number', 'credit_card_expiration_month',
            'credit_card_expiration_year', 'credit_card_cvv',
        }
        assert re.match(r'^[^@\s]+@[^@\s]+\.[^@\s]+$', form['email'])
        assert isinstance(form['zip_code'], str) and len(form['zip_code']) > 0
        assert form['credit_card_expiration_year'] > year_now
        assert 1 <= form['credit_card_expiration_month'] <= 12
        assert re.fullmatch(r'\d{3}', form['credit_card_cvv'])


# --------------------------------------------------------------------------
# gRPC consumption path (genproto/python clients used across the stack)
# --------------------------------------------------------------------------

def test_genproto_python_clients_import():
    genproto_python = REPO_ROOT / 'genproto' / 'python'
    assert genproto_python.is_dir(), 'genproto/python missing'
    sys.path.insert(0, str(genproto_python))
    try:
        import demo_pb2
        import product_catalog_pb2
        import product_catalog_pb2_grpc
        import cart_pb2_grpc
        import checkout_pb2_grpc
        import ad_pb2_grpc
        import recommendation_pb2_grpc
        import shipping_pb2_grpc
        import currency_pb2_grpc
        import payment_pb2_grpc
        import email_pb2_grpc
        import shopping_assistant_pb2_grpc
    finally:
        sys.path.remove(str(genproto_python))
    # every backend service stub is available to the frontend stack
    assert hasattr(product_catalog_pb2_grpc, 'ProductCatalogServiceStub')
    assert hasattr(cart_pb2_grpc, 'CartServiceStub')
    assert hasattr(checkout_pb2_grpc, 'CheckoutServiceStub')
    assert hasattr(ad_pb2_grpc, 'AdServiceStub')
    assert hasattr(recommendation_pb2_grpc, 'RecommendationServiceStub')
    assert hasattr(shipping_pb2_grpc, 'ShippingServiceStub')
    assert hasattr(currency_pb2_grpc, 'CurrencyServiceStub')
    assert hasattr(payment_pb2_grpc, 'PaymentServiceStub')
    assert hasattr(email_pb2_grpc, 'EmailServiceStub')
    assert hasattr(shopping_assistant_pb2_grpc, 'ShoppingAssistantServiceStub')
    # shared domain messages are generated too (Product lives in demo.proto)
    assert hasattr(demo_pb2, 'Money')
    assert hasattr(demo_pb2, 'Product')
    assert hasattr(product_catalog_pb2, 'ListProductsResponse')


# --------------------------------------------------------------------------
# python entry point
# --------------------------------------------------------------------------

def test_python_entry_point_wired():
    import importlib.util
    spec = importlib.util.spec_from_file_location(
        'loadgenerator_main', SERVICE_DIR / '__main__.py')
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    assert callable(module.main)
