'use strict';
var React = require('react');

window.onhashchange = function() {
	// TODO: Something more awesome?
	location.reload();
};

var page = document.location.hash || "#dashboard";
try {
	var Ctx = require('./pages/' + page.substr(1) + '.jsx');

	React.render(
	  <Ctx/>,
	  document.getElementById('root')
	);
} catch (e) {
	console.log(e);
}
