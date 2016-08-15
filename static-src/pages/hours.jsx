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
    delete: function(e) {
      e.preventDefault()
      var id = e.target.dataset.target;

      var that = this;
      Request.delete('/api/hour/'+id)
      .set('Accept', 'application/json')
      .end(function(err, res) {
          if (err) {
            //Fn.error(err.message);
            return;
          }
      });
    },

  render: function() {
    var res = [];
    var that = this;
    console.log("hours=",this.state.hours);
    if (this.state.hours.length === 0) {
      res.push(<tr key="empty"><td colSpan="4">No hours yet :)</td></tr>);
    } else {
      this.state.hours.forEach(function(elem) {
       // var key = elem.replace(/[^A-Za-z0-9_-]*/g, "");
        res.push(<tr key={elem}><td><a href={"#hour-add/"+elem}>{elem}</a></td><td><a data-target={elem} onClick={that.delete}>Delete</a></td></tr>);
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
              <thead><tr><th>Name</th><th>I/O</th></tr></thead>
              <tbody>{res}</tbody>
            </table>
          </div>
        </div>
    </div>;
  }
});
