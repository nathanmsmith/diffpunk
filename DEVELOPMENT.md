# DEVELOPMENT

## How to build diffpunk


Here's the basic idea: Go cli code lives in cli/. React frontend in
frontend/.

* We build the frontend first: `npm run build`
* The cli uses Go [embed](https://pkg.go.dev/embed) to embed the frontend.
(cli/assets/assets.go)
* Build the backend:

## CLI coverage

`make test` writes Go test coverage for all packages under `cli/` to
`coverage.out`. Inspect the report with `go tool cover -func=coverage.out` or
`go tool cover -html=coverage.out`.

The Checks workflow uploads the Linux report to Codecov with the `cli` flag.
Enable the repository in Codecov and add its upload token as the GitHub Actions
repository secret `CODECOV_TOKEN`.
