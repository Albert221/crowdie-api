FROM golang:1.10 as builder

WORKDIR /go/src/wolszon.me/groupie

COPY . .

# Install golang/dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN dep ensure

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crowdie-server .

FROM alpine:latest

ENV GROUPIE_PORT=8080
ENV GROUPIE_MONGO_URL=localhost:27017
ENV GROUPIE_DATABASE=groupie
ENV GROUPIE_JWT_SECRET=secret

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/wolszon.me/groupie/crowdie-server .

CMD ["./crowdie-server"]