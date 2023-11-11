FROM golang:1.21.4 AS builder

RUN mkdir /app
WORKDIR /app
COPY . /app

RUN go get -v . 
RUN CGO_ENABLED=0 go build -v -o app .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/app /

ENTRYPOINT ["/app"]