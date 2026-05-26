# Build stage
FROM eclipse-temurin:17-jdk-alpine AS builder

WORKDIR /build

COPY apps/payment-service/pom.xml .
COPY apps/payment-service/src ./src

RUN apk add --no-cache maven
RUN mvn clean package -DskipTests

# Runtime stage - distroless for security
FROM gcr.io/distroless/java17-debian12:nonroot

WORKDIR /app

COPY --from=builder /build/target/payment-service-1.0.0.jar app.jar

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "app.jar"]
