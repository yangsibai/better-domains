run: assets build
	./domains

build:
	go build -o domains

assets:
	gulp build

.PHONY: run build assets
