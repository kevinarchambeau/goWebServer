# goWebServer

Demo of RESTful API using the Go builtin HTTP server.
Includes Auth, with generation of JWT and refresh tokens.

You will need to create a .env file that contains:
* `JWT_SECRET=$SECRET`
* `POLKA_KEY=$API_KEY`

Can be executed with `go build && ./goWebServer` or with the ` --debug` flag. The debug flag will delete the database file.

The routes are:
* GET /app/
* GET /admin/metrics
* GET /api/reset
* GET /api/healthz
* POST /api/chirps
  * Body: `{"body":"wee"}`
  * Requires JWT auth token
* GET /api/chirps
  * takes optional query params: `author_id=$id` and `sort=asc|desc`
* GET /api/chirps/{chirpID}
* DELETE /api/chirps/{chirpID}
  * Requires JWT auth token
* POST /api/users
  * Body: `{"email":"$email", "password":"$password"}`
* PUT /api/users
  * Body: `{"email":"$email", "password":"$password"}`
  * Requires JWT auth token
* POST /api/login
  * Body: `{"email":"$email", "password":"$password", "expires": $seconds}`
  * Returns a JWT auth and a refresh token. `expires` is optional, valid range is 1 - 86400
* POST /api/revoke
  * revokes the refresh token in the auth header of the request
* POST /api/refresh
  * takes a refresh bearer token
* POST /api/polka/webhooks
  * JSON body in the format of:
    * `{"event": "user.upgraded", "data": {"user_id": $id} }`