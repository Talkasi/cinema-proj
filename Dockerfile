FROM golang:1.24-alpine

WORKDIR /app

COPY . .

RUN apk add --no-cache make git

RUN make build

CMD ["make", "run"]