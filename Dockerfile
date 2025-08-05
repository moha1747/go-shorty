FROM golang:tip-alpine3.22
RUN apk add --no-cache make 
WORKDIR /app
COPY . .
RUN go build
CMD ["./go-shorty"]

