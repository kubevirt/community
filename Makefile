generate:
	go run ./validators/cmd/sigs --dry-run=false
	go run ./generators/cmd/sigs
	go run ./generators/cmd/alumni

validate-sigs:
	go run ./validators/cmd/sigs
