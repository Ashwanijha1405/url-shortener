# ---------- Build stage ----------
FROM golang:1.26.3-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /url-shortener \
    ./cmd/server


# ---------- Runtime stage ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /url-shortener /url-shortener

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/url-shortener"]
