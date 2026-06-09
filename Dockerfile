FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ai-commit-gen .

FROM alpine:3.20
RUN apk add --no-cache git
COPY --from=build /out/ai-commit-gen /usr/local/bin/ai-commit-gen
WORKDIR /repo
ENTRYPOINT ["ai-commit-gen"]
