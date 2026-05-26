# Build stage
FROM node:20-alpine AS builder

WORKDIR /build

COPY apps/catalog-service/package*.json ./
RUN npm ci

COPY apps/catalog-service/ .
RUN npm run build

# Runtime stage - distroless for security
FROM gcr.io/distroless/nodejs20-debian12:nonroot

WORKDIR /app

COPY --from=builder /build/dist ./dist
COPY --from=builder /build/node_modules ./node_modules
COPY --from=builder /build/package*.json ./

USER nonroot:nonroot

EXPOSE 8080

CMD ["dist/main.js"]
