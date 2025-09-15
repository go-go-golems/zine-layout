.PHONY: gifs

all: gifs

VERSION=v0.1.14

TAPES := $(wildcard doc/vhs/*.tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v

lint:
	golangci-lint run -v

lintmax:
	golangci-lint run -v --max-same-issues=100

gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude=G101,G304,G301,G306 -exclude-dir=.history ./...

govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

test:
	go test ./...

build:
	go generate ./...
	go build ./...

goreleaser:
	goreleaser release --skip=sign --snapshot --clean

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push origin --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/zine-layout@$(shell svu current)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/clay@latest
	go mod tidy

# Example rendering helpers
EX_OUT=./dist/examples
EX_SIZE=600px,800px

.PHONY: examples examples-clean
examples:
	@mkdir -p $(EX_OUT)
	# Two inputs on one page (from examples/layouts)
	go run ./cmd/zine-layout render \
		--spec ./examples/layouts/two_pages_two_inputs.yaml \
		--output-dir $(EX_OUT)/two-pages \
		--ppi 300 --test --test-dimensions $(EX_SIZE)
	# Selected specs from examples/tests
	go run ./cmd/zine-layout render \
		--spec ./examples/tests/01_single_input_single_output.yaml \
		--output-dir $(EX_OUT)/01 \
		--ppi 300 --test --test-dimensions $(EX_SIZE)
	go run ./cmd/zine-layout render \
		--spec ./examples/tests/04_two_input_single_output_rotation.yaml \
		--output-dir $(EX_OUT)/04 \
		--ppi 300 --test --test-dimensions $(EX_SIZE)
	go run ./cmd/zine-layout render \
		--spec ./examples/tests/06_eight_inputs_two_outputs.yaml \
		--output-dir $(EX_OUT)/06 \
		--ppi 300 --test --test-dimensions $(EX_SIZE)
	go run ./cmd/zine-layout render \
		--spec ./examples/tests/10_8_sheet_zine.yaml \
		--output-dir $(EX_OUT)/10 \
		--ppi 300 --test --test-dimensions $(EX_SIZE)

examples-clean:
	rm -rf $(EX_OUT)

ZINE_LAYOUT_BINARY=$(shell which zine-layout || echo /usr/local/bin/zine-layout)
install:
	go build -o ./dist/zine-layout ./cmd/zine-layout && \
		cp ./dist/zine-layout $(ZINE_LAYOUT_BINARY)

.PHONY: web-build serve
web-build:
	cd cmd/zine-layout && go generate

serve: web-build
	go run ./cmd/zine-layout serve --addr :8088 --root ./cmd/zine-layout/dist --data-root ./data --log-level debug --with-caller
