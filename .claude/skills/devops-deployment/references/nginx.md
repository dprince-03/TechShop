# Nginx Reverse Proxy

Template: `assets/nginx.conf`.

- TLS via Certbot (`certbot --nginx -d example.com -d www.example.com`) with auto-renew timer, or terminate TLS at Cloudflare/ALB.
- Redirect HTTP→HTTPS; HSTS.
- `server_tokens off;`, `client_max_body_size` sized for uploads.
- Proxy headers: `Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`; set app `trust proxy` accordingly.
- WebSockets: `proxy_http_version 1.1; proxy_set_header Upgrade $http_upgrade; proxy_set_header Connection "upgrade";`
- Rate limit auth routes with `limit_req_zone`.
- Gzip/Brotli for text assets; long cache headers for hashed static files.
- Test config: `nginx -t` then `systemctl reload nginx`.
