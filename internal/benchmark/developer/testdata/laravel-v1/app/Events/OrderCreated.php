<?php

namespace App\Events;

final class OrderCreated
{
    public function __construct(public int $orderId) {}
}
