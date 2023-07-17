The server is divided into four base endpoints:

- `/api` for the CRUD operations
- `/docs` for API documentation
- `/dex` for OIDC authentication
- `/public` for static files

### Authentication

Authentication requires an OIDC bearer token.
The issuer is `{domain}/dex`.

<details>

<summary>Details on user authentication</summary>

#### Discovery

`{domain}/dex/.well-known/openid-configuration`

#### Resource Owner Password Flow (username + password)

- [Resource Owner Password Flow with OIDC](https://auth0.com/docs/authenticate/login/oidc-conformant-authentication/oidc-adoption-rop-flow#oidc-conformant)

```
POST {domain}/dex/token
Content-Type: application/x-www-form-urlencoded
Body: {
    grant_type: "password",
    username: "{email}",
    password: "{password}",
    client_id: "{client_id}",
    client_secret: "{client_secret}",
    scope: "openid profile email offline_access",
}
```

#### Authorization Flow (3rd party)

- [Requesting an ID token from dex](https://dexidp.io/docs/using-dex/#requesting-an-id-token-from-dex)
- [Authorization Code Flow with OIDC](https://auth0.com/docs/authenticate/login/oidc-conformant-authentication/oidc-adoption-auth-code-flow#oidc-conformant)

```
GET {domain}/dex/auth(/{connector_id})?
    response_type=code
    &scope=openid profile email offline_access
    &client_id={client_id}
    &state={state}
    &redirect_uri={redirect_uri}
```

The optional `connector_id` will redirect directly to the 3rd party login,
instead of the intermediate page provided by Dex.

After the user completes the login with the third party, they will be
redirected to the `redirect_uri`, with the `state` provided above and a code:

```
HTTP/1.1 302 Found
Location: {redirect_uri}?
    code=SplxlOBeZQQYbYS6WxSbIA
    &state={state}
```

The `state` should be checked against the one provided in the auth request.

- [Prevent Attacks and Redirect Users with OAuth 2.0 State Parameters](https://auth0.com/docs/secure/attack-protection/state-parameters)

This `code` is then used to redeem the token:

- [Code exchange request](https://auth0.com/docs/authenticate/login/oidc-conformant-authentication/oidc-adoption-auth-code-flow#code-exchange-request-authorization-code-flow)

```
POST {domain}/dex/token
Content-Type: application/x-www-form-urlencoded
Body: {
  grant_type: "authorization_code",
  client_id: "{client_id}",
  client_secret: "{client_secret}",
  code: "{code}",
  redirect_uri: "{redirect_uri}",
}
```

#### Refresh token

- [Refresh Tokens with OIDC](https://auth0.com/docs/authenticate/login/oidc-conformant-authentication/oidc-adoption-refresh-tokens#oidc-conformant-token-endpoint-)

```
POST {domain}/dex/token
Content-Type: application/x-www-form-urlencoded
Body: {
  grant_type: "refresh_token",
  client_id: "{client_id}",
  client_secret: "{client_secret}",
  refresh_token: "{refresh_token}",
}
```

### User registration

#### Username + password

1. Create a new user with `[POST] {domain}/api/users`, providing the password.
2. Authenticate with Dex using the Resource Owner Password Flow, to obtain a
token.

#### With a third party account

1. Authenticate with Dex to the chosen third party, using the Authorization
Flow.
2. Create a new user with `[POST] {domain}/api/users`, providing the `sub`
field from the identification token received from step 1.

</details>

### FCM Notifications

Register FCM Tokens with `POST /fcm/register`, and the server will handle the
rest. Old or invalid tokens are deleted periodically. The apps should refresh
their tokens according to the FCM documentation.

Below are documented the FCM notifications that the server sends.

<details>

<summary>Details on FCM Notifications</summary>

#### New user achievement

``` go
func NewAchievementMessage(token string, achievement models.Achievement) messaging.Message {
	return messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title:    achievement.Name,
			Body:     achievement.Desc,
			ImageURL: achievement.Image,
		},
		Data: map[string]string{
			"type": "achievement",
			"code": achievement.Code,
		},
	}
}
```

</details>
