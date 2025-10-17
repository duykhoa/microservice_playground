FROM golang:1.25 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN cd warehouse && CGO_ENABLED=0 GOOS=linux go build -o /app/warehouse .

FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=builder /app/warehouse .

EXPOSE 8080

CMD ["/app/warehouse"]