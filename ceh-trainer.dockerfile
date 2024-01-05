FROM golang:latest AS build

RUN mkdir /src
WORKDIR /src
COPY . /src

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cehTrainer ./cmd/trainer/main.go
RUN ls -la

RUN mkdir -p .empty/dir

FROM scratch

COPY --from=build /src/cehTrainer /app/cehTrainer
COPY --from=build /src/ceh-12-cehtest.org /app/ceh-12-cehtest.org
COPY --from=build /src/.empty/dir /app/data
WORKDIR /app

ENTRYPOINT ["/app/cehTrainer"]