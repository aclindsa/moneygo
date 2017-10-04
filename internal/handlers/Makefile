all: security_templates.go

security_templates.go: cusip_list.csv scripts/gen_security_list.py
	./scripts/gen_security_list.py > security_templates.go

cusip_list.csv:
	./scripts/gen_cusip_csv.sh > cusip_list.csv

.PHONY = all
