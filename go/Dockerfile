from golang:alpine

RUN apk update && apk upgrade && apk add --no-cache git

COPY ./src/main.go /go/main.go

COPY ./src/.remote-sync.json /go/.remote-sync.json

WORKDIR /go

RUN go get golang.org/x/crypto/ssh

RUN go get github.com/pkg/sftp

RUN go get github.com/mholt/archiver

CMD go run main.go
