FROM golang:1.10

ENV GROUPIE_PORT=8080
# You need to provide it by yourself
# ENV GROUPIE_MONGO_URL=localhost:27017
ENV GROUPIE_DATABASE=groupie

WORKDIR /go/src/wolszon.me/groupie

COPY . .

# Install golang/dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN dep ensure
RUN go run main.go