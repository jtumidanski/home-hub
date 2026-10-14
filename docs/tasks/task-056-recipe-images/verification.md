# Task 056 — Verification Record

Date: 2026-10-14. Object store is RustFS (`rustfs/rustfs:1.0.1`), not MinIO; there is no `mc`, so bucket checks use sigv4 `curl` from the `rustfs-init` image.

## Benchmark

`cd services/recipe-service && go test ./internal/recipeimage/imaging/ -run=^$ -bench=BenchmarkProcess12MP -benchtime=5x`

```
BenchmarkProcess12MP-24    5    540229533 ns/op    0.36 MB/s
```

About 0.54 s/op against a 2 s budget. Machine: x86_64, AMD Ryzen 9 5900X 12-Core Processor (24 threads), WSL2 (Linux 5.15 microsoft-standard).

## Manual compose acceptance

Scope limit: the full stack was **not** started. Two blockers, both environmental:

- Authentication is Google OIDC only (`/api/v1/auth/login/google` redirects to a browser consent flow; see `services/auth-service/docs/rest.md`). There is no password or dev-login endpoint, so no session cookie could be obtained from the command line. The only alternative would be minting a JWT with the signing key from the developer's `.env`, which was judged out of scope for this task.
- Compose services connect to an external PostgreSQL (`DB_HOST=postgres.home`), i.e. the developer's real dev database, so starting recipe-service here would run migrations and write test data into it.

Only the object-store part was exercised, in an isolated compose project (`-p t056accept`, own volume) with `--no-deps`, then removed with `docker compose ... down` (no `-v` on the default project; no volumes of the default project were touched).

1. PASS — `rustfs` became healthy; `rustfs-init` ran `curl --aws-sigv4 ... -X PUT http://rustfs:9000/$S3_BUCKET` and exited 0. Bucket verified with:
   `docker compose -p t056accept -f deploy/compose/docker-compose.yml run --rm --no-deps --entrypoint sh rustfs-init -c 'curl -sS -o /dev/null -w "HEAD bucket %{http_code}\n" -I --aws-sigv4 "aws:amz:$S3_REGION:s3" --user "$S3_ACCESS_KEY:$S3_SECRET_KEY" "http://rustfs:9000/$S3_BUCKET"; curl -fsS --aws-sigv4 ... "http://rustfs:9000/$S3_BUCKET?list-type=2&prefix="'`
   Output: `HEAD bucket 200`; ListObjectsV2 returned `<Name>home-hub-recipes</Name> ... <KeyCount>0</KeyCount>`.
2. NOT RUN — upload just under 10 MiB (200) / 11 MiB (413) through nginx: no authenticated session (see scope limit). Server-side 413 behaviour is covered by `internal/recipeimage/resource_test.go`.
3. NOT RUN — create a recipe with a photo; hero and card thumbnail render (API and UI): no session.
4. NOT RUN — replace photo; old image id disappears from the bucket listing within about 5 minutes: no session. Outbox drain is covered by `internal/recipeimage/reaper_test.go`.
5. NOT RUN — remove photo; `image` becomes null: no session.
6. NOT RUN — EXIF orientation 6 upload returns `width < height`: no session. Orientation handling is covered by `internal/recipeimage/imaging` tests.
7. NOT RUN — `If-None-Match` returns 304: no session.
8. NOT RUN — stop `rustfs`; CRUD works, image endpoints 503, restart recovers without service restart: requires recipe-service running.
9. NOT RUN — soft-delete and restore a recipe with a photo: no session.

UI visual check: pending user (items 3-6).

Suggested follow-up for the user: log in through the browser against `scripts/local-up.sh`, then run items 2-9 with the `access_token` cookie (for example `curl -b access_token=<jwt> -F file=@photo.jpg -X PUT http://localhost:8080/api/v1/recipes/<id>/image`).
