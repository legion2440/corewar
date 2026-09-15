.PHONY: all build asm-bin vm-bin test audit contracts clean

all: build

build: asm-bin vm-bin

asm-bin:
	cd assembler && cargo build --release
	cp assembler/target/release/asm ./asm

vm-bin:
	cd vm && go build -o ../corewar ./cmd/corewar

test:
	cd assembler && cargo test
	cd vm && go test ./...

contracts:
	python3 scripts/validate_agent_contracts.py

audit: build test contracts
	bash scripts/audit.sh

clean:
	rm -f asm corewar testdata/*.cor testdata/*.dis.s champions/*.cor champions/*.dis.s
	cd assembler && cargo clean
	cd vm && go clean ./...
