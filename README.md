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

### Notes/Takeaways

#### `hx-boost`

Requests use AJAX rather than the browser built-in, and HTMX knows to only swap the `<body>` tag. This avoids a
"Flash of Unstyled Content" side-effect common to native HTML while the `<head>` is being processed, before
styles take effect on the page. Using `hx-boost` means the `<head>` stays, and only the `<body>` is swapped, so
all styles are already loaded. It can be inherited, so it can probably just be placed directly on
`<body hx-boost="true">`, and specific elements disabled with `hx-boost="false"` (Images, PDFs, etc.)
