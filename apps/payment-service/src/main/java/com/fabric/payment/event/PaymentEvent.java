package com.fabric.payment.event;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.Instant;

public class PaymentEvent {

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class PaymentAuthorized {
        private String paymentId;
        private String orderId;
        private String correlationId;
        private BigDecimal amount;
        private String currency;
        private String provider;
        private String providerReference;
        private Instant authorizedAt;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class PaymentCaptured {
        private String paymentId;
        private String orderId;
        private String correlationId;
        private BigDecimal amount;
        private Instant capturedAt;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class PaymentFailed {
        private String paymentId;
        private String orderId;
        private String correlationId;
        private String reason;
        private String errorCode;
        private Instant failedAt;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class PaymentRefunded {
        private String paymentId;
        private String orderId;
        private String correlationId;
        private BigDecimal amount;
        private Instant refundedAt;
    }
}
