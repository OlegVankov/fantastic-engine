APP=./bin/gophermart
ACCRUAL=./cmd/accrual/accrual_linux_amd64
PORT=8080
ADDRESS=localhost
DSN='postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable'

.PHONY: build
build:
	go test -v -cover ./...
	go build -o ${APP} ./cmd/gophermart/...

.PHONY: test
test: build
	gophermarttest -test.v -test.run="^TestGophermart$$" \
		-gophermart-binary-path=${APP} \
        -gophermart-host=${ADDRESS} \
        -gophermart-port=${PORT} \
        -gophermart-database-uri=${DSN} \
        -accrual-binary-path=${ACCRUAL} \
        -accrual-host=${ADDRESS} \
        -accrual-port=34567 \
        -accrual-database-uri=${DSN}
