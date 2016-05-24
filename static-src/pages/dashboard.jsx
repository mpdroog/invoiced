'use strict';
var React = require('react');

module.exports = React.createClass({
    getInitialState: function() {
        return {
        };
    },
	render: function() {
		return <div className="normalheader">
		    <div className="hpanel">
		        <div className="panel-body">
		            <h2 className="font-light m-b-xs">
		                InvoiceD
		            </h2>
		            <small>General billing stuff here?</small>
		        </div>
		    </div>
		</div>;
	}
});
