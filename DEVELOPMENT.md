# DEVELOPMENT

## How to build diffpunk


Here's the basic idea: Go cli code lives in cli/. React frontend in
frontend/.

* We build the frontend first: `npm run build`
* The cli uses Go [embed](https://pkg.go.dev/embed) to embed the frontend.
(cli/assets/assets.go)
* Build the backend:
