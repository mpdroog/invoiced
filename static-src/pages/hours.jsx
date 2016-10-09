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

    ajax: function(range) {
        var that = this;
        Request.get('/api/hours')
        .set('Accept', 'application/json')
        .end(function(err, res) {
            if (err) {
              handleErr(err);
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
            handleErr(err);
            return;
          }
          location.reload();
      });
    },

  render: function() {
    var res = [];
    var that = this;
    console.log("hours=",this.state.hours);
    if (this.state.hours && this.state.hours.length > 0) {
      this.state.hours.forEach(function(elem) {
        res.push(<tr key={elem}>
          <td>{elem}</td>
          <td>
            <a className="btn btn-default btn-hover-primary" href={"#hour-add/"+elem}><i className="fa fa-pencil"></i></a>
            <a className="btn btn-default btn-hover-danger faa-parent animated-hover" data-target={elem} onClick={that.delete}><i className="fa fa-trash faa-flash"></i></a>
          </td></tr>);
      });
    } else {
      res.push(<tr key="empty"><td colSpan="4">No hours yet :)</td></tr>);
    }

    return <div className="normalheader">
        <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <a href="#hour-add" className="btn btn-default btn-hover-primary showhide"><i className="fa fa-plus"></i> New</a>
              </div>
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
