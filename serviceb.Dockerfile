FROM golang:alpine AS builder

WORKDIR /var/app

COPY . .

RUN go build cmd/serviceb/main.go

FROM scratch

ARG HTTP_PORT
ARG API_KEY

WORKDIR /var/app

COPY --from=builder /var/app/main .

EXPOSE 8080

ENTRYPOINT [ "./main" ]