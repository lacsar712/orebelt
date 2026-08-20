# Orebelt

矿石皮带振动尖峰

## Build

```bash
export GOTOOLCHAIN=local
go build ./...
```

## Test

```bash
export GOTOOLCHAIN=local
go test ./... -count=1
```

## Docker (benzhi)

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh orebelt linux/amd64
./build_benzhi_docker.sh orebelt linux/arm64
```