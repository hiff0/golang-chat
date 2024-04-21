## Простой чат на Go, html и js. Также по CDN подключен Bootstrap

## Start 
### С Docker
```
docker-compose up -d
```

### Без Docker
Без докера необходимо локально поднять Redis и в .env прописать адрес `REDIS_URL`, если он поменялся
```
go mod download  
go run main.go
```