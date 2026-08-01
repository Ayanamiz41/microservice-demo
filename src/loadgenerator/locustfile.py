#!/usr/bin/python
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

"""Locust load test for the Online Boutique replica.

Replicates the upstream GoogleCloudPlatform/microservices-demo loadgenerator:
an HTTP-only load script that drives the *frontend* REST endpoints
(``/``, ``/product/{id}``, ``/cart``, ``/cart/checkout``, ...) to simulate
real user shopping flows (browse, currency switch, add-to-cart, checkout).
The frontend itself fans out to the gRPC backend services, so the load
generator needs no server contract of its own (see src/README.md).

Run it after installing dependencies (``pip install -r requirements.txt``)::

    locust --host http://localhost:8080 --headless -u 10 -r 1

or equivalently through the Python entry point::

    python -m loadgenerator --host http://localhost:8080 --headless -u 10 -r 1

The web UI (default port 8089) is available when ``--headless`` is omitted.
"""

import datetime
import random

from faker import Faker
from locust import FastHttpUser, TaskSet, between

fake = Faker()

# Product ids served by productcatalogservice (upstream catalog).
products = [
    '0PUK6V6EV0',
    '1YMWWN1N4O',
    '2ZYFJ3GM2N',
    '66VCHSJNUP',
    '6E92ZMYYFZ',
    '9SIQT8TOJO',
    'L9ECAV7KIM',
    'LS4PSXUNUM',
    'OLJCESPC7Z']

# Currencies supported by currencyservice.
currencies = ['EUR', 'USD', 'JPY', 'CAD', 'GBP', 'TRY']


# ---------------------------------------------------------------------------
# Pure payload builders (kept separate so unit tests can exercise them
# without a running Locust instance).
# ---------------------------------------------------------------------------

def add_to_cart_form():
    """Form payload POSTed to ``/cart`` (product + quantity)."""
    return {
        'product_id': random.choice(products),
        'quantity': random.randint(1, 10),
    }


def checkout_form_data():
    """Form payload POSTed to ``/cart/checkout`` (customer + credit card)."""
    current_year = datetime.datetime.now().year + 1
    return {
        'email': fake.email(),
        'street_address': fake.street_address(),
        'zip_code': fake.zipcode(),
        'city': fake.city(),
        'state': fake.state_abbr(),
        'country': fake.country(),
        'credit_card_number': fake.credit_card_number(card_type="visa"),
        'credit_card_expiration_month': random.randint(1, 12),
        'credit_card_expiration_year': random.randint(current_year, current_year + 70),
        'credit_card_cvv': f"{random.randint(100, 999)}",
    }


# ---------------------------------------------------------------------------
# Task helpers. Each takes the TaskSet (``l``) as first argument, matching
# the upstream locustfile signature.
# ---------------------------------------------------------------------------

def index(l):
    l.client.get("/")


def setCurrency(l):
    l.client.post("/setCurrency",
                  {'currency_code': random.choice(currencies)})


def browseProduct(l):
    l.client.get("/product/" + random.choice(products))


def viewCart(l):
    l.client.get("/cart")


def addToCart(l):
    product = random.choice(products)
    l.client.get("/product/" + product)
    l.client.post("/cart", add_to_cart_form())


def empty_cart(l):
    l.client.post('/cart/empty')


def checkout(l):
    addToCart(l)
    l.client.post("/cart/checkout", checkout_form_data())


def logout(l):
    l.client.get('/logout')


class UserBehavior(TaskSet):
    """Simulates a shopper: lands on the home page, then browses the
    catalog, fiddles with the currency, fills a cart and occasionally
    checks out."""

    def on_start(self):
        index(self)

    tasks = {index: 1,
             setCurrency: 2,
             browseProduct: 10,
             addToCart: 2,
             viewCart: 3,
             checkout: 1}


class WebsiteUser(FastHttpUser):
    tasks = [UserBehavior]
    wait_time = between(1, 10)
