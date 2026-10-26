FROM golang:1.27.3-trixie AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY internal ./internal
RUN go build -o ./bin/hndelegate

FROM gcr.io/distroless/base-debian13
COPY --from=build /app/bin/hndelegate /usr/local/bin/hndelegate
ENV DB_PATH=/root/.config/hndelegate
CMD ["hndelegate"]
