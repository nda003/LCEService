FROM golang:1.25.6-trixie

WORKDIR /edtronautinterview
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apt-get update \
    && apt-get install -y python3 \
    && go build -o ./bin/consumer ./cmd/consumer

CMD ["/edtronautinterview/bin/consumer"]
