FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git
WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY resources/embed.go ./resources/
COPY resources/mcc_risk.json ./resources/
COPY resources/normalization.json ./resources/
COPY resources/references.json.gz ./resources/

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go run ./cmd/preprocess
RUN go build -ldflags="-s -w" --trimpath -o /out/api ./cmd/api

FROM scratch

COPY --from=builder /out/api /api
COPY --from=builder /build/resources/references.bin /references.bin

EXPOSE 9999

ENTRYPOINT ["/api"]