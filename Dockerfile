# Build Stage
FROM golang:bookworm AS builder
RUN mkdir /build
COPY . /build/
WORKDIR /build
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o tunnel cmd/main.go

# Deploy Stage
FROM alpine:3.20
RUN apk add --no-cache iproute2
COPY --from=builder /build/tunnel /usr/local/bin/tunnel
ENTRYPOINT [ "/usr/local/bin/tunnel" ]