# Go HTTP Demo Design

## Goal

Create a small Go HTTP demo application and deploy it to `10.235.106.18` over SSH.

## Architecture

The application is a single Go HTTP server built with the standard library. It listens on port `8080` and exposes two endpoints:

- `/` returns a short welcome message.
- `/healthz` returns `ok` for deployment health checks.

## Deployment

The deployment builds a Linux binary locally, uploads it to `/opt/go-demo/go-demo` on `10.235.106.18`, installs a `systemd` unit named `go-demo.service`, starts it, and verifies it with `curl`.

## Testing

Unit tests cover both HTTP handlers using `net/http/httptest`. Deployment is verified by checking the service status and calling `/healthz`.
