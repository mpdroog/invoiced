'use strict';
var React = require('react');
var Request = require('superagent');

module.exports = React.createClass({
    getInitialState: function() {
        return {
          "pagination": {
            "from": 0,
            "count": 50
          }
        };
    },

    componentDidMount: function() {
        this.ajax();
    },
    componentWillUnmount: function() {
    },

    ajax: function(range) {
        var that = this;
        var query = "select name FROM invoice ORDER BY id LIMIT " + this.state.pagination.from + "," + this.state.pagination.count;
        console.log(query);
        Request.get('/api/sql/all?sql=' + encodeURIComponent(query))
        .set('Accept', 'application/json')
        .end(function(err, res) {
            if (err) {
              //Fn.error(err.message);
              return;
            }
            if (that.isMounted()) {
              /*var body = res.body;
              body.loading = false;
              that.setState(body);*/
            }
        });
    },

	render: function() {
		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <a href="#invoice-add" className="showhide"><i className="fa fa-plus"></i></a>
            </div>
            Invoices
          </div>
          <div className="panel-body">
            <table className="table table-striped">
            	<thead><tr><th>#</th><th>Invoice</th><th>Customer</th><th>Amount</th></tr></thead>
            	<tbody><tr><td colSpan="4">No invoices yet :)</td></tr></tbody>
            </table>
	        </div>
		    </div>
    </div>;
	}
});
