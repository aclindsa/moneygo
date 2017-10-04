JS_SOURCES = $(wildcard js/*.js) $(wildcard js/*/*.js)

all: static/bundle.js static/react-widgets static/codemirror/codemirror.css

node_modules:
	npm install

static/bundle.js: $(JS_SOURCES) node_modules
	browserify -t [ babelify --presets [ react es2015 ] ] js/main.js -o static/bundle.js

static/react-widgets: node_modules/react-widgets/dist node_modules
	rsync -a node_modules/react-widgets/dist/ static/react-widgets/

static/codemirror/codemirror.css: node_modules/codemirror/lib/codemirror.js node_modules
	mkdir -p static/codemirror
	cp node_modules/codemirror/lib/codemirror.css static/codemirror/codemirror.css

.PHONY = all
