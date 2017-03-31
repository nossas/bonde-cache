FROM golang

ENV CACHE_INTERVAL 600
RUN ls bin
COPY bin/bonde-cache .

CMD ["bonde-cache"]
EXPOSE 443 80
