<?php

namespace App\Repositories;

final class OrderRepository
{
    public function create(array $payload): array
    {
        return ['id' => 42, ...$payload];
    }
}
