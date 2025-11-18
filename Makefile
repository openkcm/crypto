.PHONY: reuse-lint
reuse-lint:
	docker run --rm --volume $(PWD):/data fsfe/reuse lint
