'use strict';
var React = require('react');
var Request = require('superagent');

module.exports = React.createClass({
    getInitialState: function() {
        return {
          "pagination": {
            "from": 0,
            "count": 50
          },
          "invoices": []
        };
    },

    componentDidMount: function() {
        this.ajax();
    },
    componentWillUnmount: function() {
    },

    ajax: function(range) {
        var that = this;
        Request.get('/api/invoices')
        .set('Accept', 'application/json')
        .end(function(err, res) {
            if (err) {
              //Fn.error(err.message);
              return;
            }
            if (that.isMounted()) {
              var body = res.body;
              that.setState({invoices: body});
            }
        });
    },

	render: function() {
    var res = [];
    console.log("invoices=",this.state.invoices);
    if (this.state.invoices.length === 0) {
      res.push(<tr key="empty"><td colSpan="4">No invoices yet :)</td></tr>);
    } else {
      this.state.invoices.forEach(function(elem) {
        res.push(<tr key={elem}><td>#</td><td><a href={"#invoice-add/"+elem}>{elem}</a></td><td></td><td></td></tr>);
      });
    }

		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <a href="#invoice-add" className="showhide"><i className="fa fa-plus"></i> New</a>
            </div>
            Invoices
          </div>
          <div className="panel-body">
            <table className="table table-striped">
            	<thead><tr><th>#</th><th>Invoice</th><th>Customer</th><th>Amount</th></tr></thead>
            	<tbody>{res}</tbody>
            </table>
	        </div>
		    </div>
    </div>;
	}
});
