'use strict';
var React = require('react');

window.onhashchange = function() {
	// TODO: Something more awesome?
	location.reload();
};

var page = document.location.hash || "#dashboard";
try {
	var cmp = page.substr(1).split("/");
	page = cmp[0];
	var args = cmp.slice(1);

	console.log("Open " + cmp[0] + " with args=", args);
	var Ctx = require('./pages/' + cmp[0] + '.jsx');

	React.render(
	  <Ctx args={args} />,
	  document.getElementById('root')
	);
} catch (e) {
	console.log(e);
}
