FROM golang:1.24.2

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o main main.go
RUN go build -o worker ./tasks/worker/main.go
EXPOSE 8080
CMD [ "./main" ]