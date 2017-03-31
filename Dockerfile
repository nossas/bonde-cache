FROM golang

ENV CACHE_INTERVAL 600
RUN ls bin
COPY . .
RUN ls
RUN ls bin
CMD ["bin/bonde-cache"]
EXPOSE 443 80
