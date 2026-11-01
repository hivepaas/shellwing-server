# shellwing server (web)

The web app of the shellwing server, where people sign in and an admin manages users and settings. Its server is the
Go module at the root of this repository, which serves the app's build beside the API.

It was made from [hivepaas-dashboard](https://github.com/hivepaas/hivepaas-dashboard) and keeps its stack and
conventions: React, Vite, TypeScript, TanStack Query, Tailwind and shadcn/ui. The dashboard's guides in [docs/](docs/)
apply here too. It keeps:

- **Authentication**: sign-in with a password, a passcode or SSO; sign-up and invitations; password reset; two-factor
  sign-in.
- **User Management** and **Your Account**.
- **Integrations**: OAuth, cloud storages, email accounts, IM platforms and notification targets.
- **Operations**: tasks and audit logs.

## Development

Run the server first. By default it listens on `http://localhost:10200`.
Then:

```bash
yarn install
yarn dev        # http://localhost:4200
```

The dev server listens on 4200, so it can run beside the portal's on 4100.
The app calls the API on its own origin, `http://localhost:4200/api`, and the
dev server forwards those calls to the server server, as one server does in
production. That is what lets the browser send the session's refresh cookie: an
app calling `:10200` directly loses its session as soon as the access token
expires.

| Script         | Does                                  |
| -------------- | ------------------------------------- |
| `yarn dev`     | the dev server                        |
| `yarn build`   | type-checks, then builds into `dist/` |
| `yarn lint:ci` | type-checks and lints, as CI does     |
| `yarn format`  | formats with Prettier                 |

The server serves the build from its `dist-dashboard/`: `make local-build-fe`,
at the root, builds it there.

### Environment

`.env.development` holds the shared settings. Yours go in
`.env.development.local`, which git ignores (`*.local`) and which overrides
them:

| Variable                      | What it is                                                                                       |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| `PORT`                        | the dev server's port: 4200                                                                      |
| `VITE_SHW_DASHBOARD_BASE_URL` | empty: the API is on the app's own origin. An absolute URL calls it directly, without the cookie |
| `VITE_SHW_API_BASE_PATH`      | the API's path: `/api`                                                                           |
| `VITE_SHW_API_PROXY_TARGET`   | where the dev server forwards `/api`: `http://localhost:10200` unless set                        |

Paddle.js (`@paddle/paddle-js`) opens the checkout and shows prices in the
buyer's currency. The app takes the Paddle environment and the client-side
token from the server server, with the checkout it opens, so it has no Paddle
variable of its own: they are set once, on the server (`SHW_PADDLE_ENVIRONMENT`,
`SHW_PADDLE_CLIENT_TOKEN`). Every `VITE_` variable ends up in the bundle anyone
can read, so a Paddle API key or webhook secret never goes in one.

Paddle's default payment link (Paddle > Checkout > Checkout configuration) is
the app's `/pay`, `https://server.hivepaas.com/pay` in production. Paddle's
emails to update a payment method, and a transaction's `checkout.url`, open it
with `?_ptxn=txn_…`, and Paddle.js opens that transaction's checkout by itself.
Without `_ptxn` the page sends the visitor to `/pricing`.
