<?php

namespace App\Http\Controllers;

use App\Services\OrderService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

final class OrderController
{
    public function __construct(private OrderService $orders) {}

    public function store(Request $request): JsonResponse
    {
        $payload = $request->validate([
            'customer_id' => ['required', 'integer'],
            'items' => ['required', 'array', 'min:1'],
        ]);

        $order = $this->orders->create($payload);

        return response()->json(['data' => $order], 201);
    }
}
