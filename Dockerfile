FROM golang

ENV CACHE_INTERVAL 600
RUN ls
COPY bonde-cache .

CMD ["bonde-cache"]
EXPOSE 443 80
