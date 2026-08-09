<?php

namespace App\Services;

use App\Contracts\PaymentGateway;

final class CheckoutService
{
    /** @param list<PaymentGateway> $gateways */
    public function selectGateway(array $gateways, ?string $preferred): ?PaymentGateway
    {
        $fallback = null;

        foreach ($gateways as $gateway) {
            if (!$gateway->available()) {
                continue;
            }

            if ($preferred !== null && $gateway->name() === $preferred) {
                return $gateway;
            }

            if ($fallback === null) {
                $fallback = $gateway;
            }
        }

        if ($fallback !== null) {
            return $fallback;
        }

        return null;
    }
}
