FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN mkdir /chat
WORKDIR /chat

COPY web/go.mod .
COPY web/go.sum .
RUN go mod download

COPY . .

RUN touch .env

RUN echo "DB_SOURCE=postgres:///hsp_staging?host=stage-db.hs-prod.svc.cluster.local&port=5432&user=h_s_prod&password=%26Rx3%3Fdh%40cs" >> .env &&\
    echo "SOCKET=0.0.0.0:8090" >> .env &&\
    echo "API_BASE_URL=http://stage.hsprod.tech:30073" >> .env

WORKDIR web

RUN go build -o web .

#FROM alpine:latest
FROM scratch

COPY --from=builder /chat/.env /
COPY --from=builder /chat/web /

EXPOSE 8090

ENTRYPOINT ["/web"]
