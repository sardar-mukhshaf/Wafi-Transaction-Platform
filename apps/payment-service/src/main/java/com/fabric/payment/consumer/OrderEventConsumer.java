package com.fabric.payment.consumer;

import com.fabric.payment.service.PaymentService;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.support.Acknowledgment;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;

@Component
@RequiredArgsConstructor
@Slf4j
public class OrderEventConsumer {

    private final PaymentService paymentService;
    private final ObjectMapper objectMapper;

    @KafkaListener(topics = "order-events", groupId = "payment-service")
    public void onOrderEvent(ConsumerRecord<String, String> record, Acknowledgment acknowledgment) {
        try {
            String correlationId = record.key();
            JsonNode payload = objectMapper.readTree(record.value());
            String eventType = extractEventType(payload);

            log.info("[CONSUMER] Received event type={} correlation={}", eventType, correlationId);

            switch (eventType) {
                case "OrderPaymentRequested" -> handlePaymentRequested(payload);
                default -> log.debug("Ignoring event type: {}", eventType);
            }

            acknowledgment.acknowledge();
        } catch (Exception e) {
            log.error("Failed to process order event", e);
            // In production: send to dead-letter queue after retries
            acknowledgment.acknowledge();
        }
    }

    private void handlePaymentRequested(JsonNode payload) {
        String orderId = payload.get("order_id").asText();
        String correlationId = payload.get("correlation_id").asText();
        BigDecimal amount = payload.get("amount").decimalValue();
        String currency = payload.get("currency").asText("SAR");
        String paymentMethod = payload.get("payment_method").asText("MADA");

        // Simulate fraud check / risk scoring
        if (amount.compareTo(new BigDecimal("50000")) > 0) {
            paymentService.failPayment(orderId, correlationId,
                    "Amount exceeds risk threshold", "RISK_001");
            return;
        }

        paymentService.authorizePayment(orderId, correlationId, amount, currency, paymentMethod);
    }

    private String extractEventType(JsonNode payload) {
        if (payload.has("payment_id") && payload.has("amount")) {
            return "OrderPaymentRequested";
        }
        return "UNKNOWN";
    }
}
