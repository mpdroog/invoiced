import * as React from "react";
import Axios from "axios";
import * as Big from "big.js";
import * as Moment from "moment";
import Import from "./edit-import";
import {Autocomplete, LockedInput} from "../../shared/components";

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
  day?: Moment.Moment
  Lines?: IHourLineState[]
  Name?: string
  Project?: string
  Status?: string
  Total?: string
}
export default class HourEdit extends React.Component<{}, IHourState> {
  constructor(props) {
    super(props);
    this.state = {
      start: "",
      stop: "",
      description: "",
      day: Moment(),
      import: false,
      hourRate: 0,

      Lines: [],
      Name: "",
      Project: "",
      Status: "NEW",
      Total: "0"
    };
  }

  componentDidMount() {
    if (this.props.id) {
      this.ajax(this.props.id);
    }
  }

  private ajax(name: string) {
    Axios.get(`/api/v1/hour/${this.props.entity}/${this.props.year}/${this.props.bucket}/${name}`)
    .then(res => {
      this.setState(res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private importLine(lines) {
    let total = new Big("0.00");
    let out = this.state.Lines;
    for (let i = 0; i < lines.length; i++) {
      let day = lines[i];
      for (let i2 = 0; i2 < day.fromTo.length; i2++) {
        let fromTo = day.fromTo[i2];
        let start = Moment.duration(fromTo[0]);
        let stop = Moment.duration(fromTo[1]);
        let sum = stop.subtract(start);

        out.push({
          Start: fromTo[0],
          Stop: fromTo[1],
          Hours: sum.asHours(),
          Description: day.text,
          Day: day.day
        });
        total = total.plus(sum.asHours());
      }
    }

    this.setState({
      Lines: out,
      Total: total.toFixed(2).toString()
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

    let total = new Big(this.state.Total);
    this.state.Lines.push({
      Start: this.state.start,
      Stop: this.state.stop,
      Hours: sum.asHours(),
      Description: this.state.description,
      Day: this.state.day.format("YYYY-MM-DD")
    });
    this.setState({
      Lines: this.state.Lines,
      Total: total.plus(sum.asHours()).toFixed(2).toString()
    });
  }

  private updateDate(date: Moment.Moment) {
    this.setState({day: date});
  }

  private updateTotal() {
    let total = new Big("0.00");
    this.state.Lines.forEach(function(val) {
      total = total.plus(val.Hours);
    });
    return total.toFixed(2).toString();
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
    if (elem.id === "hour-project") {
      let prevMonth = Moment().subtract(1, 'months');
      this.setState({
        Project: e.target.value,
        Name: e.target.value + "-" + prevMonth.format("YYYY-MM")
      });
    }
  }

  private selectProject(prj) {
    console.log("Change", prj);
    let prevMonth = Moment().subtract(1, 'months');
    this.setState({
       Project: prj.Name,
       Name: prj.Name + "-" + prevMonth.format("YYYY-MM"),
       hourRate: prj.HourRate
    });
  }

  private lineRemove(key: number) {
    console.log(`Deleted ${key} idx `, this.state.Lines.splice(key, 1)[0]);
    this.setState({Lines: this.state.Lines});
  }

  private save(e: BrowserEvent) {
    e.preventDefault();
    let req = this.state;
    req.Total = this.updateTotal();

    Axios.post(`/api/v1/hour/${this.props.entity}/${this.props.year}/concepts`, req)
    .then(res => {
      console.log(res.data);
      this.props.id = res.data.Name;
      this.setState(res.data);
      history.replaceState({}, "", `#${this.props.entity}/${this.props.year}/hours/edit/concepts/${res.data.Name}`);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private bill(e: BrowserEvent) {
    e.preventDefault();
    let args = this.state;
    Axios.post(`/api/v1/hour/${this.props.entity}/${this.props.year}/concepts/${args.Name}/bill`, args)
    .then(res => {
      location.href = `#${this.props.entity}/${this.props.year}/invoices/edit/concepts/${res.headers["x-redirect-invoice"]}`
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private shortHand(d: number): string {
    var date = new Date(1000*60*60*d);
    var str = '';
    if (date.getUTCDate()-1 > 0) {
      str += date.getUTCDate()-1 + "d";
    }
    str += date.getUTCHours() + "h";
    str += date.getUTCMinutes() + "m";
    return str;
  }

  private toggleImport(e) {
    e.preventDefault();
    this.setState({import: !this.state.import});
  }

	render() {
    let lines: React.JSX.Element[] = [];
    let that = this;
    let isEditable = this.state.Status === "NEW" || this.state.Status === "CONCEPT";

    this.state.Lines.forEach(function(item:IHourLineState, idx:number) {
      lines.push(<tr key={idx}>
        <td><button className="btn btn-default btn-hover-danger faa-parent animated-hover" onClick={that.lineRemove.bind(that, idx)}><i className="fa fa-trash faa-flash"></i></button></td>
        <td>{item.Day}</td>
        <td>{item.Start}</td><td>{item.Stop}</td><td>{that.shortHand(item.Hours)}</td><td>{item.Description}</td>
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
              <input type="date" id="hour-day" className="form-control" value={this.state.day.format("YYYY-MM-DD")} onChange={this.update.bind(this)}/>
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
    <Import hide={this.state.import} onHide={this.toggleImport.bind(this)} importFn={this.importLine.bind(this)} />

    <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <div className="panel-tools">
            <div className="btn-group nm7">
              <button className="btn btn-default btn-hover-success" disabled={!isEditable} onClick={this.toggleImport.bind(this)}><i className="fa fa-arrow-up"></i>&nbsp;Import</button>
              <button className="btn btn-default btn-hover-success" disabled={!isEditable} onClick={this.save.bind(this)}><i className="fa fa-floppy-o"></i>&nbsp;Save</button>
              <button className="btn btn-default btn-hover-danger" disabled={this.state.Status !== "CONCEPT"} onClick={this.bill.bind(this)}><i className="fa fa-share"></i>&nbsp;Bill</button>
            </div>
          </div>
          Sum ({this.state.Total} hours/{this.state.Total * this.state.hourRate } EUR)
        </div>
        <div className="panel-body">
          <div className="row">
            <div className="col-sm-6">
              <Autocomplete id="hour-project" onSelect={this.selectProject.bind(this)} onChange={this.update.bind(this)} placeholder="Project name" url={"/api/v1/projects/"+that.props.entity+"/search"} value={this.state.Project} />
            </div>
            <div className="col-sm-6">
              <LockedInput type="text" id="hour-name" value={this.state.Name} placeholder="AUTOGENERATED" onChange={this.update.bind(this)} locked={true} />
            </div>
          </div>
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
