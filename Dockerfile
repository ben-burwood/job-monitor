FROM node:latest AS build-vue

WORKDIR /app/frontend

COPY frontend/package*.json ./

RUN npm ci

COPY frontend/. .

RUN npm run build

FROM --platform=$BUILDPLATFORM golang:alpine AS build-go

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /server

FROM scratch

WORKDIR /

COPY --from=build-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-go /server /server
COPY --from=build-vue /app/frontend/dist /frontend/dist

ENV HOME=/root

EXPOSE 8080

COPY --from=ghcr.io/tarampampam/microcheck:1 /bin/httpcheck /bin/httpcheck
HEALTHCHECK --interval=30s --timeout=3s --retries=3 --start-period=10s CMD ["httpcheck", "http://localhost:8080"]

ENTRYPOINT ["/server"]
