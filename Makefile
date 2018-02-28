all:
	@echo 'no default'

run:
	go run cryptodash/main.go

table:
	go run cryptodash/main.go -table
