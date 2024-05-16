FROM golang:1.21

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download && go mod verify

RUN apt-get update && apt-get install -y wget

RUN wget -O /usr/local/bin/migrate https://github.com/golang-migrate/migrate/releases/download/v4.16.0/migrate.linux-amd64.tar.gz && \
    tar -xvzf /usr/local/bin/migrate -C /usr/local/bin/ && \
    chmod +x /usr/local/bin/migrate

COPY . .

RUN go build -o main ./cmd/api/

CMD ["sh", "-c", "migrate -path ./migrations -database postgres://postgres:postgres@db:5432/task?sslmode=disable up && ./main"]


#FROM golang:1.21
#
#RUN apt-get update && apt-get install -y
#
#WORKDIR /app
#COPY ./go.mod ./go.sum ./
#RUN go mod download && go mod verify
#RUN go get -v -u github.com/golang-migrate/migrate/v4
#
#COPY . .
#
#RUN go build -o main ./cmd/api/
#
#
##RUN migrate -path=./migrations -database=postgres://watch_admin:watch@db:5432/watch_database?sslmode=disable up
#
##CMD ["migrate -path ./migrations -database postgres://watch_admin:watch@db:5432/watch_database?sslmode=disable up"]
#CMD ["./main"]