import * as React from "react";
import {IInjectedProps} from "react-router";
import Axios from "axios";
import * as Big from "big.js";
import * as Moment from "moment";
import * as DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
//require('react-datepicker/dist/react-datepicker.css');

interface IHourLineState {
  Hours: number
  Day: string
  Start: string
  Stop: string
  Description: string
}
interface IHourState {
  start?: string
  stop?: string
  description?: string
  day?: moment.Moment
  Lines?: IHourLineState[]
  Name?: string
}
export default class HourEdit extends React.Component<IInjectedProps, IHourState> {
  constructor(props: IInjectedProps) {
    super(props);
    this.state = {
      start: "",
      stop: "",
      description: "",
      day: Moment(),

      Lines: [],
      Name: ""
    };
  }

  componentDidMount() {
    console.log("componentDidMount", this.props);
    console.log("Load hour name=" + this.props.params["id"]);
    this.ajax(this.props.params["id"]);
  }

  private ajax(name: string) {
    Axios.get('/api/hour/'+name)
    .then(res => {
      this.setState(res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private recalc(e: Event) {
    e.preventDefault();
    if (this.state.start === "" || this.state.stop === "") {
      console.log("Empty state");
      return;
    }

    let start = Moment.duration(this.state.start);
    let stop = Moment.duration(this.state.stop);
    let sum = stop.subtract(start);
    console.log("Start=" + start + " Stop=" + stop + " to hours=" + sum.humanize());

    this.state.Lines.push({
      Start: this.state.start,
      Stop: this.state.stop,
      Hours: sum.asHours(),
      Description: this.state.description,
      Day: this.state.day.format("YYYY-MM-DD")
    });
    this.setState({
      Lines: this.state.Lines
    });
  }

  private updateDate(date: moment.Moment) {
    this.setState({day: date});
  }

  private update(e: InputEvent) {
    console.log(e.target.value);
    let elem = e.target as any;

    if (elem.id === "hour-start") {
      this.setState({start: e.target.value});
    }
    if (elem.id === "hour-stop") {
      this.setState({stop: e.target.value});
    }
    if (elem.id === "hour-description") {
      this.setState({description: e.target.value});
    }
    if (elem.id === "hour-name") {
      this.setState({Name: e.target.value});
    }
    if (elem.id === "hour-day") {
      this.setState({day: Moment(e.target.value)});
    }
  }

  private lineRemove(key: number) {
    console.log("Remove hour line with key=" + key);
    console.log("Deleted idx ", this.state.Lines.splice(key, 1)[0]);
    this.setState({Lines: this.state.Lines});
  }

  private save() {
    Axios.post('/api/hour', this.state)
    .then(res => {
      console.log(res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

	render() {
    let lines: JSX.Element[] = [];
    let that = this;
    let sum = new Big("0.00");
    this.state.Lines.forEach(function(item:IHourLineState, idx:number) {
      sum = sum.plus(item.Hours);
      lines.push(<tr key={idx}>
        <td><button className="btn btn-default btn-hover-danger faa-parent animated-hover" onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash faa-flash"></i></button></td>
        <td>{item.Day}</td>
        <td>{item.Start}</td><td>{item.Stop}</td><td>{item.Hours}</td><td>{item.Description}</td>
      </tr>);
    });

		return <form>
      <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            Project Hour Calc
          </div>
          <div className="panel-body">
            <div className="col-sm-2">
              <DatePicker
                id="hour-day"
                className="form-control"
                dateFormat="YYYY-MM-DD"
                selected={this.state.day}
                onChange={this.updateDate.bind(this)} />
            </div>
            <div className="col-sm-2">
              <input type="text" id="hour-start" className="form-control" placeholder="HH:mm" value={this.state.start} onChange={this.update.bind(this)}/>
            </div>
            <div className="col-sm-2">
              <input type="text" id="hour-stop" className="form-control" placeholder="HH:mm" value={this.state.stop} onChange={this.update.bind(this)}/>
            </div>
            <div className="col-sm-5">
              <input type="text" id="hour-description" className="form-control" placeholder="Description" value={this.state.description} onChange={this.update.bind(this)}/>
            </div>
            <div className="col-sm-1">            
              <button onClick={this.recalc.bind(this)} className="btn btn-default btn-hover-success">
                <i className="fa fa-plus"></i>
                &nbsp;Add
              </button>
	          </div>
          </div>
		    </div>
    </div>

    <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <div className="panel-tools">
            <div className="btn-group nm7">
              <button className="btn btn-default btn-hover-success" onClick={this.save.bind(this)}><i className="fa fa-floppy-o"></i>&nbsp;Save</button>
            </div>
          </div>
          Sum ({sum.toString()} hours)
        </div>
        <div className="panel-body">
          <input type="text" id="hour-name" className="form-control" placeholder="Hour name" value={this.state.Name} onChange={this.update.bind(this)}/>
          <table className="table table-striped">
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
}
