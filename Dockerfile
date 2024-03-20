FROM golang:1.22.1-alpine as builder

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/grobuzin

CMD ["grobuzin"]