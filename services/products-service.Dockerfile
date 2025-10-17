FROM golang:1.25 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN cd products-service && CGO_ENABLED=0 GOOS=linux go build -o /app/products-service .

FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=builder /app/products-service .

EXPOSE 8080

CMD ["/app/products-service"]