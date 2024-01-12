package views

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
)

func Render(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response())
}
