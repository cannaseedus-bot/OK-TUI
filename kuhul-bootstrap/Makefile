.PHONY: all bootstrap stage1 stage2 verify

BOOTSTRAP=bootstrap/target/release/kuhulc_stage0
STAGE1=build/kuhulc_stage1
STAGE2=build/kuhulc_stage2

all: bootstrap stage1 stage2 verify

bootstrap:
	cd bootstrap && cargo build --release

stage1:
	mkdir -p build
	$(BOOTSTRAP) kuhulc/main.khl -o $(STAGE1)

stage2:
	$(STAGE1) kuhulc/main.khl -o $(STAGE2)

verify:
	cmp $(STAGE1) $(STAGE2)
