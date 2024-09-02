# goWebServer

Demo of RESTful API using the Go builtin HTTP server.
Includes Auth, with generation of JWT and refresh tokens.

You will need to create a .env file that contains:
* `JWT_SECRET=$SECRET`
* `POLKA_KEY=$API_KEY`

The routes are:
* GET /app/
* GET /admin/metrics
* GET /api/reset
* GET /api/healthz
* POST /api/chirps
* GET /api/chirps
* GET /api/chirps/{chirpID}
* DELETE /api/chirps/{chirpID}
* POST /api/users
* PUT /api/users
* POST /api/login
* POST /api/revoke
* POST /api/refresh
* POST /api/polka/webhooks