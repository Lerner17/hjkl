FROM golang:latest

WORKDIR $GOPATH/src/github.com/lerner17/hjkl

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["bin/hjkl"]