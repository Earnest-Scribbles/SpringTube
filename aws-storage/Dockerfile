FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY ./videos ./videos
RUN CGO_ENABLED=0 GOOS=linux go build -o /springtube

CMD ["/springtube"]