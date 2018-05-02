FROM golang:1.10

ENV GROUPIE_PORT=8080
ENV GROUPIE_MONGO_URL=localhost:27017
ENV GROUPIE_DATABASE=groupie
ENV GROUPIE_JWT_SECRET=secret

WORKDIR /go/src/wolszon.me/groupie

COPY . .

# Install golang/dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN dep ensure

CMD ["go", "run", "main.go"]