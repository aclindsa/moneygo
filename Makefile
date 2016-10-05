JS_SOURCES = $(wildcard js/*.js) $(wildcard js/*/*.js)

all: static/bundle.js static/react-widgets

node_modules:
	npm install

static/bundle.js: $(JS_SOURCES) node_modules
	browserify -t [ babelify --presets [ react ] ] js/main.js -o static/bundle.js

static/react-widgets: node_modules/react-widgets/dist node_modules
	rsync -a node_modules/react-widgets/dist/ static/react-widgets/

.PHONY = all
