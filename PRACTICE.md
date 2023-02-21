```shell
curl localhost:9003/orders

curl -X POST -H "Content-Type: application/json" localhost:9003/orders -d "@./order.json

curl localhost:9003/orders/5
```


```shell
curl localhost:9003/products

curl -X POST -H "Content-Type: application/json" localhost:9003/products -d "@./product.json

curl localhost:9003/products/5
```
