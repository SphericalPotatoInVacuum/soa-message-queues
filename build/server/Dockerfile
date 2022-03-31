FROM golang:1.18 as builder

WORKDIR /go/src/github.com/SphericalPotatoInVacuum/soa-message-queues

COPY go.mod go.sum ./
RUN go mod download

COPY internal ./internal/
COPY proto_gen ./proto_gen/
COPY cmd/server ./cmd/server/
RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go

FROM alpine
WORKDIR /root/
COPY --from=builder /go/src/github.com/SphericalPotatoInVacuum/soa-message-queues/server ./

CMD ["/bin/sh", "-c", "while ! nc -z rabbitmq 15672; do echo waiting for rabbitmq; sleep 1; done; ./server"]
