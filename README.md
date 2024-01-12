# htmx-contact-app

Go/HTMX Contact App built while reading the Hypermedia Systems Book

- https://hypermedia.systems/a-web-1-0-application/
- https://github.com/bigskysoftware/contact-app (Python Implementation Reference)

### Dependencies

- PocketBase SQLite - Extended with Go - https://pocketbase.io/docs/go-overview/
  - Echo v5 Alpha - Routing - https://github.com/labstack/echo/tree/v5_alpha
- TEMPL - Strongly typed HTML Go templating language - https://github.com/a-h/templ
  ```sh
  go install github.com/a-h/templ/cmd/templ@latest
  ```
- `wgo` - Simple live reload - https://github.com/bokwoon95/wgo
  ```sh
  go install github.com/bokwoon95/wgo@latest
  ```

### Running

```sh
make dev

# or the long way

templ generate

go run main.go serve
```

#### Up next

Chapter 4 - https://hypermedia.systems/extending-html-as-hypermedia/
