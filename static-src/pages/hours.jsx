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
          "hours": []
        };
    },

    componentDidMount: function() {
        this.ajax();
    },
    componentWillUnmount: function() {
    },

    ajax: function(range) {
        var that = this;
        Request.get('/api/hours')
        .set('Accept', 'application/json')
        .end(function(err, res) {
            if (err) {
              //Fn.error(err.message);
              return;
            }
            if (that.isMounted()) {
              var body = res.body;
              that.setState({hours: body});
            }
        });
    },

  render: function() {
    var res = [];
    console.log("hours=",this.state.hours);
    if (this.state.hours.length === 0) {
      res.push(<tr><td colSpan="4">No hours yet :)</td></tr>);
    } else {
      this.state.hours.forEach(function(elem) {
        res.push(<tr key={elem}><td>#</td><td><a href={"#hour-add/"+elem}>{elem}</a></td><td></td><td></td></tr>);
      });
    }

    return <div className="normalheader">
        <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <a href="#hour-add" className="showhide"><i className="fa fa-plus"></i> New</a>
            </div>
            Hour registration
          </div>
          <div className="panel-body">
            <table className="table table-striped">
              <thead><tr><th>#</th><th>Name</th><th>Customer</th><th>Amount</th></tr></thead>
              <tbody>{res}</tbody>
            </table>
          </div>
        </div>
    </div>;
  }
});
