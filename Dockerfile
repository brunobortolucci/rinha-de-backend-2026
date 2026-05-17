FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git
WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -ldflags="-s -w" --trimpath -o /out/api ./cmd/api

FROM scratch

COPY --from=builder /out/api /api

EXPOSE 9999

ENTRYPOINT ["/api"]