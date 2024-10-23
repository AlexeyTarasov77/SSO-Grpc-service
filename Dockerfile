FROM golang:1.22.8-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download 

COPY . .

ENTRYPOINT [ "go", "run", "./cmd/sso", "-config=./config/local.yaml" ]

