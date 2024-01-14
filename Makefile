dev:
	@wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run main.go serve