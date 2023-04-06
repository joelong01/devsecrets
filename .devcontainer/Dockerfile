FROM golang:1.19

WORKDIR /usr/src/devsecrets

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go get -v golang.org/x/tools/cmd/tidy
RUN go install -v github.com/cweill/gotests/gotests@v1.6.0
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/devsecrets ./...

CMD ["app"]
