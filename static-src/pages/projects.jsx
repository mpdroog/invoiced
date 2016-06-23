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
      lines: []
    };
  },

  componentDidMount: function() {
    console.log(Moment.duration("13:33").subtract("12:15").asHours());
  },
  componentWillUnmount: function() {
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

    this.state.lines.push({
      start: this.state.start,
      stop: this.state.stop,
      hours: sum.asHours()
    });
    this.setState({
      start: null,
      stop: null,
      lines: this.state.lines
    });
  },

  update: function(e) {
    var valid = /^[0-9]{2}:[0-9]{2}$/;

    if (e.target.value.match(valid)) {
      console.log(e.target.value);

      if (e.target.id === "hour-start") {
        this.setState({start: e.target.value});
      }
      if (e.target.id === "hour-stop") {
        this.setState({stop: e.target.value});
      }
    }
  },
  lineRemove: function(key) {
    console.log("Remove invoice line with key=" + key);
    console.log("Deleted idx ", this.state.lines.splice(key, 1)[0]);
    this.setState({lines: this.state.lines});
  },

	render: function() {
    var lines = [];
    var that = this;
    var sum = new Big("0.00");
    this.state.lines.forEach(function(item, idx) {
      sum = sum.plus(item.hours);
      lines.push(<tr key={idx}>
        <td><a onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash"></i></a></td>
        <td>{item.start}</td><td>{item.stop}</td><td>{item.hours}</td></tr>);
    });

		return <form>
      <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            Project Hour Calc
          </div>
          <div className="panel-body">
            <input type="text" id="hour-start" placeholder="HH:mm" value={this.state.start} onChange={this.update}/>
            <input type="text" id="hour-stop" placeholder="HH:mm" value={this.state.stop} onChange={this.update}/>
            <button onClick={this.recalc} className="form-field">Add</button>
	        </div>
		    </div>
    </div>

    <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          Sum ({sum.toString()} hours)
        </div>
        <div className="panel-body">
          <table className="table">
            <thead>
              <tr>
                <th>#</th>
                <th>From</th>
                <th>To</th>
                <th>Hours</th>
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
