<?php

namespace App\Contracts;

interface PaymentGateway
{
    public function name(): string;

    public function available(): bool;

    public function charge(int $customerId, array $items): void;
}
