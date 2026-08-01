"""Rule-based shopping assistant engine.

The upstream ``shoppingassistantservice`` (GoogleCloudPlatform/
microservices-demo) is a cloud-only Flask app backed by Gemini + AlloyDB.
This local replica keeps the same *intent scope* — greeting, product
recommendation, product details, order tracking, small talk — but answers
with deterministic rules over the local ProductCatalogService, so it runs
fully offline with no cloud dependency.

``get_completion`` returns ``(reply_text, context)``; ``context`` is an
opaque dict the caller (the gRPC servicer) stores per conversation so a
follow-up like "tell me more" can expand on what was recommended before.
"""

import re

DEFAULT_RECOMMENDATION_COUNT = 3

GREETING_RE = re.compile(
    r"^\s*(hi|hello|hey|howdy|good\s+(morning|afternoon|evening)|"
    r"你好|哈喽|嗨|早上好|下午好|晚上好)[!?.,。！？]*\s*$",
    re.IGNORECASE,
)
THANKS_RE = re.compile(r"(thank\s*you|thanks|thank|thx|谢谢|感谢)", re.IGNORECASE)
RECOMMEND_RE = re.compile(
    r"(recommend|recommendation|suggest|suggestion|gift|gifts|ideas?|"
    r"what\s+should\s+(i|we|my|our)\s+(buy|get|pick)|推荐|建议|礼物)",
    re.IGNORECASE,
)
DETAILS_RE = re.compile(
    r"(tell\s+me\s+(more\s+)?about|details?|price|how\s+much|costs?|"
    r"info(rmation)?\s+(about|on)|what\s+is|介绍|详情|价格|多少钱)",
    re.IGNORECASE,
)
TRACKING_RE = re.compile(
    r"(order|track(ing)?|shipping|delivery|shipment|where\s+is\s+my|"
    r"订单|物流|发货|快递)",
    re.IGNORECASE,
)
FOLLOWUP_RE = re.compile(
    r"(tell\s+me\s+more|more\s+(about|details)|learn\s+more|go\s+on|"
    r"继续|更多|还有吗|展开)",
    re.IGNORECASE,
)

GENERIC_FALLBACK = (
    "I can help with a few things:\n"
    "• \"recommend a gift\" — get product suggestions\n"
    "• \"tell me about the sunglasses\" — product details & prices\n"
    "• \"where is my order?\" — order support\n"
    "What would you like to do?"
)


def format_price(price):
    """Formats a demo.proto ``Money`` (units + nanos) as a USD string."""
    if price is None:
        return "n/a"
    units = int(getattr(price, "units", 0) or 0)
    nanos = int(getattr(price, "nanos", 0) or 0)
    sign = "-" if (units < 0 or nanos < 0) else ""
    dollars = abs(units)
    cents = abs(nanos) // 10_000_000  # 1e9 nanos = 1 unit
    return f"{sign}${dollars}.{cents:02d}"


class Assistant:
    """Deterministic chat assistant for the Online Boutique replica.

    ``catalog`` is an optional ``ProductCatalogClient``; when it is absent
    (or unreachable) the assistant falls back to generic replies so the
    service still answers and passes health checks standalone.
    """

    def __init__(self, catalog=None):
        self._catalog = catalog

    # ------------------------------------------------------------------ API

    def get_completion(self, message, context_product_ids=(), last_context=None):
        """Returns ``(reply_text, new_context)`` for a user message.

        ``context_product_ids`` are products currently on screen (from the
        request); ``last_context`` is the context returned for the previous
        message of the same conversation (for follow-up expansion).
        """
        message = (message or "").strip()
        if not message:
            return GENERIC_FALLBACK, {}

        if GREETING_RE.match(message):
            return self._greeting_reply(), {}

        if THANKS_RE.search(message):
            return (
                "You're welcome! Anything else I can help you find — "
                "a recommendation, details on a product, or an order update?",
                {},
            )

        if RECOMMEND_RE.search(message):
            reply, ids = self._recommendation_reply(message, context_product_ids)
            return reply, {"recommended_ids": ids}

        if FOLLOWUP_RE.search(message) and last_context and last_context.get("recommended_ids"):
            ids = last_context["recommended_ids"]
            reply = self._details_for_ids(ids)
            return reply, {"detailed_ids": ids}

        if DETAILS_RE.search(message):
            reply, ids = self._details_reply(message)
            return reply, {"detailed_ids": ids}

        if TRACKING_RE.search(message):
            return (
                "I can't look up order status from here, but our support team "
                "can — just send them your order number and they'll help right away.",
                {},
            )

        return GENERIC_FALLBACK, {}

    # ------------------------------------------------------------- replies

    def _greeting_reply(self):
        return (
            "Hello! Welcome to Online Boutique — I'm your shopping assistant. "
            "I can recommend products, tell you about an item, or help with an "
            "order. What are you looking for today?"
        )

    def _recommendation_reply(self, message, context_product_ids):
        products = self._catalog.list_products() if self._catalog else []
        if not products:
            return (
                "I'd love to suggest something, but the product catalog is "
                "temporarily unreachable. Please try again in a moment!",
                [],
            )

        picks = self._pick_recommendations(products, context_product_ids)
        if not picks:
            return (
                "I couldn't find anything to recommend right now — please try again!",
                [],
            )

        lines = [f"• {p.name} — {format_price(p.price_usd)}" for p in picks]
        if context_product_ids:
            intro = (
                "Since you're looking at the catalog right now, here are a few "
                "similar picks you might like:\n"
            )
        else:
            intro = "Here are a few picks I think you'll like:\n"
        return intro + "\n".join(lines), [p.id for p in picks]

    def _details_reply(self, message):
        query = self._details_query(message)
        results = []
        if self._catalog:
            results = self._catalog.search_products(query) if query else []
            if not results:
                results = self._catalog.list_products()[:DEFAULT_RECOMMENDATION_COUNT]
        if not results:
            return (
                "I couldn't find a matching product in the catalog right now — "
                "it may be temporarily unreachable. Please try again in a moment!",
                [],
            )
        top = results[:3]
        lines = [f"• {p.name} — {format_price(p.price_usd)}: {p.description}" for p in top]
        return "Here's what I found:\n" + "\n".join(lines), [p.id for p in top]

    def _details_for_ids(self, product_ids):
        if not self._catalog:
            return (
                "I'd love to tell you more, but the product catalog is "
                "temporarily unreachable. Please try again in a moment!"
            )
        products = [self._catalog.get_product(pid) for pid in product_ids]
        products = [p for p in products if p is not None]
        if not products:
            return "Sorry, I lost track of those products — could you ask again?"
        lines = [f"• {p.name} — {format_price(p.price_usd)}: {p.description}" for p in products]
        return "Sure! Here's more on those:\n" + "\n".join(lines)

    # ------------------------------------------------------------ helpers

    def _pick_recommendations(self, products, context_product_ids):
        """Picks up to N products, preferring ones related to on-screen items."""
        if context_product_ids:
            context_products = [
                p for p in products if p.id in set(context_product_ids)
            ]
            related = [
                p
                for p in products
                if p.id not in set(context_product_ids)
                and any(c in p.categories for c in self._categories_of(context_products))
            ]
            if related:
                return related[:DEFAULT_RECOMMENDATION_COUNT]
        return products[:DEFAULT_RECOMMENDATION_COUNT]

    @staticmethod
    def _categories_of(products):
        categories = set()
        for product in products:
            categories.update(product.categories)
        return categories

    def _details_query(self, message):
        """Strips detail keywords/punctuation to leave a searchable query."""
        query = DETAILS_RE.sub(" ", message)
        query = re.sub(r"[^0-9a-zA-Z\u4e00-\u9fff ]+", " ", query)
        return query.strip()
