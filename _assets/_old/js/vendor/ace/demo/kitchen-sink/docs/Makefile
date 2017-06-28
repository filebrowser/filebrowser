.PHONY:    apf ext worker mode theme package test

default: apf worker

update: worker

# packages apf

# This is the first line of a comment \
and this is still part of the comment \
as is this, since I keep ending each line \
with a backslash character

apf:
    cd node_modules/packager; node package.js projects/apf_cloud9.apr
	cd node_modules/packager; cat build/apf_release.js | sed 's/\(\/\*FILEHEAD(\).*//g' > ../../plugins-client/lib.apf/www/apf-packaged/apf_release.js

# package debug version of apf
apfdebug:
	cd node_modules/packager/projects; cat apf_cloud9.apr | sed 's/<p:define name=\"__DEBUG\" value=\"0\" \/>/<p:define name=\"__DEBUG\" value=\"1\" \/>/g' > apf_cloud9_debug2.apr
	cd node_modules/packager/projects; cat apf_cloud9_debug2.apr | sed 's/apf_release/apf_debug/g' > apf_cloud9_debug.apr; rm apf_cloud9_debug2.apr
	cd node_modules/packager; node package.js projects/apf_cloud9_debug.apr
	cd node_modules/packager; cat build/apf_debug.js | sed 's/\(\/\*FILEHEAD(\).*\/apf\/\(.*\)/\1\2/g' > ../../plugins-client/lib.apf/www/apf-packaged/apf_debug.js

# package_apf--temporary fix for non-workering infra
pack_apf:
	mkdir -p build/src
	mv plugins-client/lib.apf/www/apf-packaged/apf_release.js build/src/apf_release.js
	node build/r.js -o name=./build/src/apf_release.js out=./plugins-client/lib.apf/www/apf-packaged/apf_release.js baseUrl=.

# makes ace; at the moment, requires dryice@0.4.2
ace:
	cd node_modules/ace; make clean pre_build; ./Makefile.dryice.js minimal


# packages core
core: ace
	mkdir -p build/src
	node build/r.js -o build/core.build.js

# generates packed template
helper: 
	node build/packed_helper.js

helper_clean:
	mkdir -p build/src
	node build/packed_helper.js 1

# packages ext
ext: 
	node build/r.js -o build/app.build.js

# calls dryice on worker & packages it
worker: plugins-client/lib.ace/www/worker/worker-language.js

plugins-client/lib.ace/www/worker/worker-language.js plugins-client/lib.ace/www/worker/worker-javascript.js : \
        $(wildcard node_modules/ace/*) $(wildcard node_modules/ace/*/*) $(wildcard node_modules/ace/*/*/mode/*) \
        $(wildcard plugins-client/ext.language/*) \
        $(wildcard plugins-client/ext.language/*/*) \
        $(wildcard plugins-client/ext.linereport/*) \
        $(wildcard plugins-client/ext.codecomplete/*) \
        $(wildcard plugins-client/ext.codecomplete/*/*) \
        $(wildcard plugins-client/ext.jslanguage/*) \
        $(wildcard plugins-client/ext.jslanguage/*/*) \
        $(wildcard plugins-client/ext.csslanguage/*) \
        $(wildcard plugins-client/ext.csslanguage/*/*) \
        $(wildcard plugins-client/ext.htmllanguage/*) \
        $(wildcard plugins-client/ext.htmllanguage/*/*) \
        $(wildcard plugins-client/ext.jsinfer/*) \
        $(wildcard plugins-client/ext.jsinfer/*/*) \
        $(wildcard node_modules/treehugger/lib/*) \
        $(wildcard node_modules/treehugger/lib/*/*) \
        $(wildcard node_modules/ace/lib/*) \
        $(wildcard node_modules/ace/*/*) \
        Makefile.dryice.js
	mkdir -p plugins-client/lib.ace/www/worker
	rm -rf /tmp/c9_worker_build
	mkdir -p /tmp/c9_worker_build/ext
	ln -s `pwd`/plugins-client/ext.language /tmp/c9_worker_build/ext/language
	ln -s `pwd`/plugins-client/ext.codecomplete /tmp/c9_worker_build/ext/codecomplete
	ln -s `pwd`/plugins-client/ext.jslanguage /tmp/c9_worker_build/ext/jslanguage
	ln -s `pwd`/plugins-client/ext.csslanguage /tmp/c9_worker_build/ext/csslanguage
	ln -s `pwd`/plugins-client/ext.htmllanguage /tmp/c9_worker_build/ext/htmllanguage
	ln -s `pwd`/plugins-client/ext.linereport /tmp/c9_worker_build/ext/linereport
	ln -s `pwd`/plugins-client/ext.linereport_php /tmp/c9_worker_build/ext/linereport_php
	node Makefile.dryice.js worker
	cp node_modules/ace/build/src/worker* plugins-client/lib.ace/www/worker

define 

ifeq

override

# copies built ace modes
mode:
	mkdir -p plugins-client/lib.ace/www/mode
	cp `find node_modules/ace/build/src | grep -E "mode-[a-zA-Z_0-9]+.js"`  plugins-client/lib.ace/www/mode

# copies built ace themes
theme:
	mkdir -p plugins-client/lib.ace/www/theme
	cp `find node_modules/ace/build/src | grep -E "theme-[a-zA-Z_0-9]+.js"` plugins-client/lib.ace/www/theme

gzip_safe:
	for i in `ls ./plugins-client/lib.packed/www/*.js`; do \
		gzip -9 -v -c -q -f $$i > $$i.gz ; \
	done

gzip:
	for i in `ls ./plugins-client/lib.packed/www/*.js`; do \
		gzip -9 -v -q -f $$i ; \
	done

c9core: apf ace core worker mode theme
    
package_clean: helper_clean c9core ext

package: helper c9core ext

test check:
	test/run-tests.sh	