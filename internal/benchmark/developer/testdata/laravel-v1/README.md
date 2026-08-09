# Order API fixture

This miniature Laravel project exposes `POST /api/orders`. The controller
validates the HTTP request and delegates order creation to `OrderService`.
The service charges a `PaymentGateway`, stores the order through
`OrderRepository`, and dispatches an `OrderCreated` event.

`CheckoutService` is a separate refactoring fixture that selects a preferred
payment gateway, then a fallback, and finally returns `null` when none are
available.
