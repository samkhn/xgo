# gin for server development

## setup

run locally...

``
go run .
``

## CLI

### Post an album

``
curl http://localhost:8080/albums --include --header "Content-Type: application/json" --request "POST" --data '{"id": "4","title": "The Modern Sound of Betty Carter","artist": "Betty Carter","price": 49.99}'
``

### Get albums

``
curl http://localhost:8080/albums --header "Content-Type: application/json" --request "GET"
``

### Get a specific album

``
curl http://localhost:8080/albums/2
``