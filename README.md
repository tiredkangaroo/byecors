# byecors

A minimal CORS proxy written in Go. Pass it a URL and it forwards the request, adding the CORS headers needed to call it from a browser. Single-file, no dependencies beyond the standard library.

## Usage

Send any method to the proxy with the target in the `url` query parameter:

```
GET http://localhost:3133/?url=https://example.com/api/data
```

The proxy forwards the method and body (POST, PUT, DELETE, etc.), streams the response back, and sets permissive CORS headers so browser clients can read it.

### Query parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `url`     | yes      | —       | Target URL. Only `http` and `https` schemes are allowed. |
| `timeout` | no       | `30`    | Upstream request timeout in seconds. |

### Special headers

- `X-Cookie` — value is forwarded upstream as the `Cookie` header. Use this to send cookies to the target.

The following request headers are **not** forwarded: `Host`, `Origin`, `Referer`, `Access-Control-Request-Method`, `Access-Control-Request-Headers`, `X-Cookie`.

`OPTIONS` requests are answered directly with `204 No Content` as part of the CORS preflight.

## Running

### Local

Requires Go 1.27+.

```sh
go run .
```

The server listens on port `3133` by default. Override with `PORT`:

```sh
PORT=8080 go run .
```

### Docker

Build and run:

```sh
docker build -t byecors .
docker run -p 3133:3133 byecors
```

Or pull the published image:

```sh
docker run -p 3133:3133 ghcr.io/tiredkangaroo/byecors:latest
```

## Deployment

Pushing to `main` triggers the GitHub Actions workflow (`.github/workflows/docker.yml`), which builds the image and pushes it to GitHub Container Registry (`ghcr.io`) tagged `latest` and with a short commit SHA.
