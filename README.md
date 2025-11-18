# Omnibasement

## Development

To upgrade internal dependencies:

```bash
go get github.com/omnsight/omniscent-library@<branch>
```

## Run Locally

Manual buf action to manage protobuf

```bash
buf registry login buf.build

buf dep update

buf format -w
buf lint

buf generate
buf push

go mod tidy
```

Run unit tests. You can view arangodb dashboard at http://localhost:8529.

```bash
docker-compose up -d arangodb
go test -v ./... -run <test name>
docker-compose down
```
