FROM golang:1.24

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY .env ./
RUN go mod download

COPY . .

RUN go build -o /app/main 

EXPOSE 8080

CMD [ "/app/main" ]