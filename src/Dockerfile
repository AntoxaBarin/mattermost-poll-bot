FROM golang:1.24 as builder

WORKDIR /
COPY . .
RUN go mod tidy -e
RUN cd poll_bot && CGO_ENABLED=0 GOOS=linux go build .

FROM alpine:latest
COPY --from=builder /poll_bot/poll_bot .
CMD ["./poll_bot"]
