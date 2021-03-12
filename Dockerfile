FROM golang:1.16

WORKDIR /app
COPY . .
RUN go get .
RUN go build .
RUN ls

ENTRYPOINT ./TUM-Live-Worker