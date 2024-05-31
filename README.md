# distributedservice-gin

# steps
go run main.go

curl -X GET http://localhost:5000/anila

go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger version

swagger generate spec -o ./swagger.json

swagger serve ./swagger.json

swagger serve -F swagger ./swagger.json

docker run -d --name mongodb -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=password -p 27017:27017 mongo:4.4.3

"mongodb://admin:password@localhost:27017"

export MONGO_URI=mongodb://admin:password@localhost:27017 && export MONGO_DATABASE=demo && go run main.go

mongoimport --username admin --password password --authenticationDatabase admin --db demo --collection recipes --file recipes.json --jsonArray

docker run -d --name redis -p 6379:6379 redis:6.0

docker run -d -v $PWD/conf:/usr/local/etc/redis --name redis -p 6379:6379 redis:6.0     ##eviction file

docker exec -it 3dbc28f50955 bash
redis-cli
EXISTS recipes

docker run -d --name redisinsight --link redis -p 5540:5540 redislabs/redisinsight  ##redis GUI

benchmarking

ab -n 2000 -c 100 -g without-cache.data http://localhost:8080/recipes  #without cache

ab -n 2000 -c 100 -g with-cache.data http://localhost:8080/recipes

start gnuplot $gnuplot

gnuplot apache-benchmark.p   #image creation