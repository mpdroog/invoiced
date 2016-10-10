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

    ajax: function(range) {
        var that = this;
        Request.get('/api/invoices')
        .set('Accept', 'application/json')
        .end(function(err, res) {
            if (err) {
              handleErr(err);
              return;
            }
            if (that.isMounted()) {
              var body = res.body;
              that.setState({invoices: body});
            }
        });
    },

    delete: function(e) {
      e.preventDefault()
      var id = e.target.dataset.target;

      var that = this;
      Request.delete('/api/invoice/'+id)
      .set('Accept', 'application/json')
      .end(function(err, res) {
          if (err) {
            handleErr(err);
            return;
          }
          location.reload();
      });
    },

	render: function() {
    var res = [];
    var that = this;
    console.log("invoices=",this.state.invoices);
    if (this.state.invoices && this.state.invoices.length > 0) {
      this.state.invoices.forEach(function(elem) {
        res.push(<tr key={elem}>
          <td>
            <a className="btn btn-default btn-hover-primary" href={"#invoice-add/"+elem}><i className="fa fa-pencil"></i></a>
            <a className="btn btn-default btn-hover-danger faa-parent animated-hover" data-target={elem} onClick={that.delete}><i className="fa fa-trash faa-flash"></i></a>
          </td><td>{elem}</td><td></td><td></td></tr>);
      });
    } else {
      res.push(<tr key="empty"><td colSpan="4">No invoices yet :)</td></tr>);
    }

		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <a href="#invoice-add" className="btn btn-default btn-hover-primary showhide"><i className="fa fa-plus"></i> New</a>
              </div>
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
