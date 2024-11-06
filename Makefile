generate:
	go run ./validators/cmd/sigs --dry-run=false
	go run ./generators/cmd/sigs

validate-sigs:
	go run ./validators/cmd/sigs
