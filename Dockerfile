FROM golang:1.16 AS build-env
WORKDIR /app
COPY . .
RUN go get .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker TUM-Live-Worker

FROM alpine:3.12
RUN apk update && apk add curl && apk add ffmpeg
WORKDIR /app
COPY --from=build-env /app/worker /app/worker

CMD "./worker"