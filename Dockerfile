
FROM golang:alpine


WORKDIR /app


COPY . /app


ENV GOOS=linux
ENV GOARCH=amd64


RUN go mod download


RUN go build -o main main.go


ENV PORT 8080


EXPOSE $PORT


CMD ["./main"]