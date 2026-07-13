FROM golang:1.26

LABEL maintaner="b2flow.tech"

WORKDIR /app

COPY . .

RUN go mod download

COPY go.mod go.sum ./

RUN go build -o labs cmd/labs/main.go

CMD ["./labs"]
