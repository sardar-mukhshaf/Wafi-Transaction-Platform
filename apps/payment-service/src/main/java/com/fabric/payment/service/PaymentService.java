package com.fabric.payment.service;

import com.fabric.payment.event.PaymentEvent;
import com.fabric.payment.model.Payment;
import com.fabric.payment.repository.PaymentRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class PaymentService {

    private final PaymentRepository paymentRepository;
    private final KafkaTemplate<String, String> kafkaTemplate;
    private final ObjectMapper objectMapper;

    @Transactional
    public Payment authorizePayment(String orderId, String correlationId,
                                     BigDecimal amount, String currency,
                                     String paymentMethod) {
        log.info("[PAYMENT] Authorizing payment for order={}, correlation={}", orderId, correlationId);

        // Simulate provider call (MADA/STC Pay/Apple Pay)
        String providerRef = "PROV-" + UUID.randomUUID().toString().substring(0, 8);

        Payment payment = Payment.builder()
                .orderId(orderId)
                .correlationId(correlationId)
                .status("AUTHORIZED")
                .amount(amount)
                .currency(currency)
                .paymentMethod(paymentMethod)
                .providerReference(providerRef)
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build();

        Payment saved = paymentRepository.save(payment);

        // Publish event
        PaymentEvent.PaymentAuthorized event = PaymentEvent.PaymentAuthorized.builder()
                .paymentId(saved.getId())
                .orderId(orderId)
                .correlationId(correlationId)
                .amount(amount)
                .currency(currency)
                .provider(paymentMethod)
                .providerReference(providerRef)
                .authorizedAt(Instant.now())
                .build();

        publishEvent("payment-events", correlationId, event);
        return saved;
    }

    @Transactional
    public void failPayment(String orderId, String correlationId, String reason, String errorCode) {
        log.warn("[PAYMENT] Payment failed for order={}, reason={}", orderId, reason);

        PaymentEvent.PaymentFailed event = PaymentEvent.PaymentFailed.builder()
                .paymentId("N/A")
                .orderId(orderId)
                .correlationId(correlationId)
                .reason(reason)
                .errorCode(errorCode)
                .failedAt(Instant.now())
                .build();

        publishEvent("payment-events", correlationId, event);
    }

    private void publishEvent(String topic, String correlationId, Object event) {
        try {
            String payload = objectMapper.writeValueAsString(event);
            kafkaTemplate.send(topic, correlationId, payload)
                    .whenComplete((result, ex) -> {
                        if (ex != null) {
                            log.error("Failed to publish event to {}", topic, ex);
                        } else {
                            log.debug("Published event to {}: {}", topic, correlationId);
                        }
                    });
        } catch (Exception e) {
            log.error("Failed to serialize event", e);
            throw new RuntimeException("Event serialization failed", e);
        }
    }
}
