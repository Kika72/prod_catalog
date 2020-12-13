FROM alpine:latest

COPY .build/prod-catalog /
COPY etc/config.yml etc/config.yml

EXPOSE 5000

CMD ["/prod-catalog", "--config=etc/config.yml"]