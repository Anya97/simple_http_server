FROM golang:1.21

WORKDIR /app

ENV PORT 3333

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /simple_http_server

CMD ["/simple_http_server"]
