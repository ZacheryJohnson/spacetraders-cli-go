# Spacetrader CLI
A side project for myself to learn Go while still playing a game.

## Generating Latest OpenAPI spec
Prerequisites:
- `npm`

1. Install [openapi-generator-cli](https://github.com/OpenAPITools/openapi-generator-cli) via `npm install -g @openapitools/openapi-generator-cli`.
2. Run `git submodule update --init --recursive` to initialize our submodule, which holds the OpenAPI spec.
3. Run `npx @openapitools/openapi-generator-cli generate -i spacetraders-spec/reference/SpaceTraders.json -g go -o generated/ --git-user-id ZacheryJohnson --git-repo-id spacetraders-cli-go` to generate our code from the spec