# build
FROM golang:1.12-alpine AS builder

WORKDIR /app
COPY . .
RUN apk update && apk add --no-cache git ca-certificates
RUN CGO_ENABLED=0 GOOS=linux go build -o /remediator cmd/remediator/app.go


# pack
FROM gcr.io/docker-images-180022/base/alpine:3.10

WORKDIR .

ADD config config

COPY --from=builder /remediator .

CMD ["./remediator"]