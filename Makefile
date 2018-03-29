all:
	@echo 'no default'

chart:
	go run cryptocharts/cryptocharts.go -coin ethereum

table:
	go run cryptocharts/cryptocharts.go -table
