FROM golang:1.24.2

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./application

FROM alpine:latest

WORKDIR /

COPY --from=0 /app/application .

CMD ["./application"]