FROM golang:alpine AS builder

WORKDIR /var/app

COPY . .

RUN go build cmd/servicea/main.go

FROM scratch

ARG HTTP_PORT
ARG API_URL

WORKDIR /var/app

COPY --from=builder /var/app/main .

EXPOSE 8080

ENTRYPOINT [ "./main" ]