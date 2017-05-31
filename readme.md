Bonde Cache

1. http web server with maximum performance
2. read json from mobilizations
3. parse html to save in bolt at startup script
4. configure bolt as in-memory db
5. serve domains based on custom_domain in mobilizations
6. add worker to listen queue and update bolt cache from mobilization
7. download static files and serve dynamic (as found in html downloaded)
8. add support to auto tls custom_domain
9. production test with multiple domains and certificates
10. verificar se a requisição voltar vazia e não sobrescrever cache
```
docker build -t nossas/bonde-cache .
docker run -it --rm -p 3000:3000 -v "$PWD":/go/src/app -w /go/src/app -e IS_DEV=true --name bonde-cache-app nossas/bonde-cache gin

docker build -f Dockerfile.dev -t nossas/bonde-cache .
docker run -it --rm -p 3000:3000 -v "$PWD":/go/src/app -w /go/src/app -e IS_DEV=true --name bonde-cache-app nossas/bonde-cache gin
```
