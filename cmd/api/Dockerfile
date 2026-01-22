FROM golang:1.25.6-alpine

WORKDIR /edtronautinterview
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/api ./cmd/api \
    && go build -o ./bin/migrate ./cmd/migrate

CMD ["/edtronautinterview/bin/api"]
EXPOSE 8080
