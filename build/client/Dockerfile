FROM golang:1.18 as builder

WORKDIR /go/src/github.com/SphericalPotatoInVacuum/soa-message-queues

COPY go.mod go.sum ./
RUN go mod download

COPY proto_gen ./proto_gen/
COPY internal ./internal/
COPY cmd/client ./cmd/client/
RUN CGO_ENABLED=0 go build -o client ./cmd/client/main.go

FROM scratch
WORKDIR /root/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs/
COPY --from=builder /go/src/github.com/SphericalPotatoInVacuum/soa-message-queues/client ./

CMD ["./client"]
