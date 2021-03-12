FROM golang:1.16 AS build-env
WORKDIR /app
COPY . .
RUN go get .
RUN go build .

FROM alpine:3.12
RUN apk update && apk add curl && apk add ffmpeg
WORKDIR /app
COPY --from=build-env /app/TUM-Live-Worker .
ENTRYPOINT ./TUM-Live-Worker
