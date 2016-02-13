JS_SOURCES = $(wildcard js/*.js)

all: static/bundle.js static/react-widgets

static/bundle.js: $(JS_SOURCES)
	browserify -t [ babelify --presets [ react ] ] js/main.js -o static/bundle.js

static/react-widgets: node_modules/react-widgets/dist
	rsync -a node_modules/react-widgets/dist/ static/react-widgets/

.PHONY = all
