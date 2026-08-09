<?php

namespace App\Services;

use App\Contracts\PaymentGateway;
use App\Events\OrderCreated;
use App\Repositories\OrderRepository;
use Illuminate\Contracts\Events\Dispatcher;

final class OrderService
{
    public function __construct(
        private OrderRepository $orders,
        private PaymentGateway $payments,
        private Dispatcher $events,
    ) {}

    public function create(array $payload): array
    {
        $this->payments->charge($payload['customer_id'], $payload['items']);
        $order = $this->orders->create($payload);
        $this->events->dispatch(new OrderCreated($order['id']));

        return $order;
    }
}
