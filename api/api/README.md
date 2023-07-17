## Command-line arguments

| Flag        | Default          | Description                           |
| ----------- | ---------------- | ------------------------------------- |
| `-seed`     | `false`          | Insert seed data into the database    |
| `-env`      | `.env`           | Path to the env file to load          |
| `-dex`      | `.dexrc.json`    | Path to the Dex config file           |
| `-firebase` | `.firebase.json` | Path to the Firebase credentials file |

## Setup

Run `make install` to install the necessary dependencies.

Two databases need to be created, one for the application and another for
[Dex](https://dexidp.io/).

## Development

### `make`

Generates the docs and compiles the project. Use `make compile` to skip doc
generation and speed up compilation.

### `make run`

Calls `make` and runs the binary.

### `make test`

Runs all tests.

### `make format`

Formats the code and the `swag` annotations.

Remember to format before committing. Consider a pre-commit hook like
[this one](https://github.com/edsrzf/gofmt-git-hook/blob/master/fmt-check).

## API docs

The docs are generated with [swag](https://github.com/swaggo/swag).
Update with `make docs`.

They can be accessed at `{api url}/docs/index.html`.

## Authorization tokens

[Dex](https://dexidp.io/) is used for authentication.
The [example app](https://dexidp.io/docs/getting-started/#running-a-client) can be used
to generate tokens. Run the example app with:

`./bin/example-app --issuer "http://127.0.0.1:8080/dex"`

Or `./bin/example-app --issuer "api.cycleforlisbon.com/dex"` for a remote token.

More documentation on user registration and authentication can be found [here](api.md).
