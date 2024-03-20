FROM golang:1.22.1 as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/grobuzin

FROM golang:1.22.1

COPY --from=builder /usr/local/bin/grobuzin /usr/local/bin/grobuzin

CMD ["grobuzin"]