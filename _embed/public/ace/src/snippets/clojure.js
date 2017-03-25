define("ace/snippets/clojure",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet comm\n\
	(comment\n\
	  ${1}\n\
	  )\n\
snippet condp\n\
	(condp ${1:pred} ${2:expr}\n\
	  ${3})\n\
snippet def\n\
	(def ${1})\n\
snippet defm\n\
	(defmethod ${1:multifn} \"${2:doc-string}\" ${3:dispatch-val} [${4:args}]\n\
	  ${5})\n\
snippet defmm\n\
	(defmulti ${1:name} \"${2:doc-string}\" ${3:dispatch-fn})\n\
snippet defma\n\
	(defmacro ${1:name} \"${2:doc-string}\" ${3:dispatch-fn})\n\
snippet defn\n\
	(defn ${1:name} \"${2:doc-string}\" [${3:arg-list}]\n\
	  ${4})\n\
snippet defp\n\
	(defprotocol ${1:name}\n\
	  ${2})\n\
snippet defr\n\
	(defrecord ${1:name} [${2:fields}]\n\
	  ${3:protocol}\n\
	  ${4})\n\
snippet deft\n\
	(deftest ${1:name}\n\
	    (is (= ${2:assertion})))\n\
	  ${3})\n\
snippet is\n\
	(is (= ${1} ${2}))\n\
snippet defty\n\
	(deftype ${1:Name} [${2:fields}]\n\
	  ${3:Protocol}\n\
	  ${4})\n\
snippet doseq\n\
	(doseq [${1:elem} ${2:coll}]\n\
	  ${3})\n\
snippet fn\n\
	(fn [${1:arg-list}] ${2})\n\
snippet if\n\
	(if ${1:test-expr}\n\
	  ${2:then-expr}\n\
	  ${3:else-expr})\n\
snippet if-let \n\
	(if-let [${1:result} ${2:test-expr}]\n\
		(${3:then-expr} $1)\n\
		(${4:else-expr}))\n\
snippet imp\n\
	(:import [${1:package}])\n\
	& {:keys [${1:keys}] :or {${2:defaults}}}\n\
snippet let\n\
	(let [${1:name} ${2:expr}]\n\
		${3})\n\
snippet letfn\n\
	(letfn [(${1:name) [${2:args}]\n\
	          ${3})])\n\
snippet map\n\
	(map ${1:func} ${2:coll})\n\
snippet mapl\n\
	(map #(${1:lambda}) ${2:coll})\n\
snippet met\n\
	(${1:name} [${2:this} ${3:args}]\n\
	  ${4})\n\
snippet ns\n\
	(ns ${1:name}\n\
	  ${2})\n\
snippet dotimes\n\
	(dotimes [_ 10]\n\
	  (time\n\
	    (dotimes [_ ${1:times}]\n\
	      ${2})))\n\
snippet pmethod\n\
	(${1:name} [${2:this} ${3:args}])\n\
snippet refer\n\
	(:refer-clojure :exclude [${1}])\n\
snippet require\n\
	(:require [${1:namespace} :as [${2}]])\n\
snippet use\n\
	(:use [${1:namespace} :only [${2}]])\n\
snippet print\n\
	(println ${1})\n\
snippet reduce\n\
	(reduce ${1:(fn [p n] ${3})} ${2})\n\
snippet when\n\
	(when ${1:test} ${2:body})\n\
snippet when-let\n\
	(when-let [${1:result} ${2:test}]\n\
		${3:body})\n\
";
exports.scope = "clojure";

});
