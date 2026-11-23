# Build stage
FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o sublinkX

# Final stage
FROM alpine:latest
WORKDIR /app

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 sublink \
    && adduser -S -D -H -u 10001 -G sublink sublink \
    && mkdir -p /app/db /app/logs /app/template \
    && chown -R sublink:sublink /app/db /app/logs /app/template \
    && chmod 750 /app/db /app/logs /app/template

COPY --from=builder --chown=sublink:sublink /app/sublinkX /app/sublinkX
USER sublink:sublink
EXPOSE 8000
CMD ["/app/sublinkX"]

