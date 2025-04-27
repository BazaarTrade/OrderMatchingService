FROM golang:1.23.1 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy

WORKDIR /app/cmd
RUN go build -o matchingEngine .

FROM gcr.io/distroless/base

COPY --from=builder /app/cmd/matchingEngine /matchingEngine

EXPOSE 50051

CMD ["/matchingEngine"]