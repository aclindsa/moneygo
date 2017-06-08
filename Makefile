JS_SOURCES = $(wildcard js/*.js) $(wildcard js/*/*.js)

all: static/bundle.js static/react-widgets security_templates.go

node_modules:
	npm install

static/bundle.js: $(JS_SOURCES) node_modules
	browserify -t [ babelify --presets [ react es2015 ] ] js/main.js -o static/bundle.js

static/react-widgets: node_modules/react-widgets/dist node_modules
	rsync -a node_modules/react-widgets/dist/ static/react-widgets/

security_templates.go: cusip_list.csv
	./scripts/gen_security_list.py > security_templates.go

cusip_list.csv:
	./scripts/gen_cusip_csv.sh > cusip_list.csv

.PHONY = all
