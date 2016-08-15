'use strict';
var React = require('react');
var Request = require('superagent');
var Big = require('big.js');
var Moment = require('moment');
require('./invoice.css');

module.exports = React.createClass({
  getInitialState: function() {
    return {
      start: null,
      stop: null,
      description: null,
      day: null,

      Lines: [],
      Name: ""
    };
  },

  componentDidMount: function() {
    if (this.props.args.length > 0) {
      console.log("Load invoice name=" + this.props.args[0]);
      this.ajax(this.props.args[0]);
    }
  },
  componentWillUnmount: function() {
  },

  ajax: function(name) {
    var that = this;
    Request.get('/api/hour/'+name)
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          //Fn.error(err.message);
          return;
        }
        if (that.isMounted()) {
          var body = res.body;
          that.setState(body);
        }
    });
  },

  recalc: function(e) {
    if (this.state.start === null || this.state.stop === null) {
      console.log("Empty state");
      return;
    }

    var start = Moment.duration(this.state.start);
    var stop = Moment.duration(this.state.stop);
    var sum = stop.subtract(start);
    console.log("Start=" + start + " Stop=" + stop + " to hours=" + sum.humanize());

    this.state.Lines.push({
      Start: this.state.start,
      Stop: this.state.stop,
      Hours: sum.asHours(),
      Description: this.state.description,
      Day: this.state.day
    });
    this.setState({
      start: null,
      stop: null,
      description: null,
      Lines: this.state.Lines
    });
  },

  update: function(e) {
    var valid = /^[0-9]{2}:[0-9]{2}$/;

    //if (e.target.value.match(valid)) {
      console.log(e.target.value);

      if (e.target.id === "hour-start") {
        this.setState({start: e.target.value});
      }
      if (e.target.id === "hour-stop") {
        this.setState({stop: e.target.value});
      }
    //}
    if (e.target.id === "hour-description") {
      this.setState({description: e.target.value});
    }
    if (e.target.id === "hour-name") {
      this.setState({Name: e.target.value});
    }
    if (e.target.id === "hour-day") {
      this.setState({day: e.target.value});
    }
  },
  lineRemove: function(key) {
    console.log("Remove invoice line with key=" + key);
    console.log("Deleted idx ", this.state.Lines.splice(key, 1)[0]);
    this.setState({Lines: this.state.Lines});
  },

  save: function(e) {
    var that = this;
    Request.post('/api/hour')
    .send(this.state)
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          console.log(err);
          //Fn.error(err.message);
          return;
        }
        if (that.isMounted()) {
          console.log(res.body);
          /*var body = res.body;
          body.loading = false;
          that.setState(body);*/
        }
    });
  },

	render: function() {
    var lines = [];
    var that = this;
    var sum = new Big("0.00");
    this.state.Lines.forEach(function(item, idx) {
      sum = sum.plus(item.Hours);
      lines.push(<tr key={idx}>
        <td><a onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash"></i></a></td>
        <td>{item.Day}</td>
        <td>{item.Start}</td><td>{item.Stop}</td><td>{item.Hours}</td><td>{item.Description}</td></tr>);
    });

		return <form>
      <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            Project Hour Calc
          </div>
          <div className="panel-body">
            <div className="col-sm-2">
              <input type="text" id="hour-day" className="form-control" placeholder="YYYY-mm-dd" value={this.state.day} onChange={this.update}/>
            </div>
            <div className="col-sm-2">
              <input type="text" id="hour-start" className="form-control" placeholder="HH:mm" value={this.state.start} onChange={this.update}/>
            </div>
            <div className="col-sm-2">
              <input type="text" id="hour-stop" className="form-control" placeholder="HH:mm" value={this.state.stop} onChange={this.update}/>
            </div>
            <div className="col-sm-5">
              <input type="text" id="hour-description" className="form-control" placeholder="Description" value={this.state.description} onChange={this.update}/>
            </div>
            <div className="col-sm-1">            
              <button onClick={this.recalc} className="btn btn-default">Add</button>
	          </div>
          </div>
		    </div>
    </div>

    <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <div className="panel-tools">
            <a onClick={this.save}><i className="fa fa-floppy-o"></i> Save</a>
          </div>
          Sum ({sum.toString()} hours)
        </div>
        <div className="panel-body">
          <input type="text" id="hour-name" className="form-control" placeholder="Hour name" value={this.state.Name} onChange={this.update}/>
          <table className="table">
            <thead>
              <tr>
                <th>#</th>
                <th>Day</th>
                <th>From</th>
                <th>To</th>
                <th>Hours</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody>
              {lines}
            </tbody>
          </table>
        </div>
      </div>
    </div>
    </form>;
	}
});
