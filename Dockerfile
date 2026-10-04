# Build stage
FROM golang:1.22 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /hello-politecnico-scheduler .

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /hello-politecnico-scheduler /usr/local/bin/kube-scheduler
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/kube-scheduler"]
