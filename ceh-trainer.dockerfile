FROM golang:latest AS build

RUN mkdir /src
WORKDIR /src
COPY . /src

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cehTrainer ./cmd/trainer/main.go
RUN ls -la

FROM scratch

COPY --from=build /src/cehTrainer /app/cehTrainer
COPY --from=build /src/question.data /app/question.data
WORKDIR /app

ENTRYPOINT ["/app/cehTrainer"]