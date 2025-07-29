FROM golang:alpine-3.22
RUN apk add --no-cache make 
WORKDIR /app
COPY . .
RUN make
CMD ["./bin/go-shorty"]

