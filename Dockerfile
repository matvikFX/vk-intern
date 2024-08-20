FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN GOOS=linux CGO_ENABLED=0 go build -o /vk-intern cmd/main.go

ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 8080

CMD ["/vk-intern"]
