FROM golang:1.18 as builder

WORKDIR /go/src/github.com/SphericalPotatoInVacuum/soa-message-queues

COPY go.mod go.sum ./
RUN go mod download

COPY proto_gen ./proto_gen/
COPY internal ./internal/
COPY cmd/grabber ./cmd/grabber/
RUN CGO_ENABLED=0 go build -o grabber ./cmd/grabber/main.go

FROM scratch
WORKDIR /root/
COPY --from=builder /go/src/github.com/SphericalPotatoInVacuum/soa-message-queues/grabber ./

CMD ["./grabber"]
