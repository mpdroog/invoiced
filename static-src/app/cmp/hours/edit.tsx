import * as React from "react";
import Axios from "axios";
import Big from "big.js";
import Moment from "moment";
import Import from "./edit-import";
import {Autocomplete, LockedInput} from "../../shared/components";
import {ActionButton} from "../../shared/ActionButton";

interface IHourLineState {
  Hours: number;
  Day: string;
  Start: string;
  Stop: string;
  Description: string;
  HourRate?: number;
}

interface HourEditProps {
  entity: string;
  year: string;
  id?: string;
  bucket?: string;
}

interface IHourState {
  start: string;
  stop: string;
  description: string;
  day: Moment.Moment;
  import: boolean;
  HourRate: number;
  Lines: IHourLineState[];
  Name: string;
  Project: string;
  Status: string;
  Total: string;
}

export default class HourEdit extends React.Component<HourEditProps, IHourState> {
  private undoStack: IHourLineState[][];
  private redoStack: IHourLineState[][];

  constructor(props: HourEditProps) {
    super(props);
    this.undoStack = [];
    this.redoStack = [];
    this.state = {
      start: "",
      stop: "",
      description: "",
      day: Moment(),
      import: false,
      HourRate: 0,

      Lines: [],
      Name: "",
      Project: "",
      Status: "NEW",
      Total: "0"
    };
  }

  componentDidMount(): void {
    if (this.props.id != null && this.props.id !== '') {
      this.ajax(this.props.id);
    }
  }

  private ajax(name: string): void {
    Axios.get<IHourState>(`/api/v1/hour/${this.props.entity}/${this.props.year}/${this.props.bucket}/${name}`)
    .then(res => {
      this.setState(res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private importLine(lines: Array<{day: string | null; text: string; fromTo: string[][]}>): void {
    let total = new Big("0.00");
    const out = [...this.state.Lines];
    for (const lineItem of lines) {
      if (lineItem.day == null) continue;
      for (const fromTo of lineItem.fromTo) {
        if (fromTo.length < 2) continue;
        const startTime = fromTo[0];
        const stopTime = fromTo[1];
        if (startTime == null || stopTime == null) continue;
        const start = Moment(startTime, 'HH:mm')
        const stop = Moment(stopTime, 'HH:mm');
        if (! start.isValid()) {
          throw new Error("Failed parsing start=" + startTime);
        }
        if (! stop.isValid()) {
          throw new Error("Failed parsing stop=" + stopTime);
        }
        // Momentjs fails us, do the math ourselves..
        const diff = stop.diff(start)/1000/60/60;
        console.log(diff);

        out.push({
          Start: startTime,
          Stop: stopTime,
          Hours: diff,
          Description: lineItem.text,
          Day: lineItem.day
        });
        total = total.plus(diff);
      }
    }

    this.setState({
      Lines: out,
      Total: total.toFixed(2).toString()
    });
  }

  private recalc(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.state.start === "" || this.state.stop === "") {
      console.log("Empty state");
      return;
    }

    const start = Moment(this.state.start, 'HH:mm')
    const stop = Moment(this.state.stop, 'HH:mm');
    if (! start.isValid()) {
      throw new Error("Failed parsing start=" + this.state.start);
    }
    if (! stop.isValid()) {
      throw new Error("Failed parsing stop=" + this.state.stop);
    }
    // Momentjs fails us, do the math ourselves..
    const sum = stop.diff(start)/1000/60/60;

    console.log("Start=" + start + " Stop=" + stop + " to hours=" + sum);

    const total = new Big(this.state.Total);
    this.state.Lines.push({
      Start: this.state.start,
      Stop: this.state.stop,
      Hours: sum,
      Description: this.state.description,
      Day: this.state.day.format("YYYY-MM-DD")
    });
    this.setState({
      Lines: this.state.Lines,
      Total: total.plus(sum).toFixed(2).toString()
    });
  }

  private updateTotal(): string {
    let total = new Big("0.00");
    this.state.Lines.forEach(function(val) {
      total = total.plus(val.Hours);
    });
    return total.toFixed(2).toString();
  }

  private update(e: React.ChangeEvent<HTMLInputElement>): void {
    const elem = e.target;

    if (elem.id === "hour-start") {
      this.setState({start: elem.value});
    }
    if (elem.id === "hour-stop") {
      this.setState({stop: elem.value});
    }
    if (elem.id === "hour-description") {
      this.setState({description: elem.value});
    }
    if (elem.id === "hour-name") {
      this.setState({Name: elem.value});
    }
    if (elem.id === "hour-day") {
      this.setState({day: Moment(elem.value)});
    }
    if (elem.id === "hour-project") {
      const prevMonth = Moment().subtract(1, 'months');

      const diff: Partial<IHourState> = {};
      if (elem.value !== this.state.Project) {
        diff.Project = elem.value;
      }
      if (this.state.Name === "") {
        diff.Name = elem.value + "-" + prevMonth.format("YYYY-MM");
      }

      console.log("diff", diff);
      if (Object.keys(diff).length > 0) this.setState(diff as IHourState);
    }
  }

  private selectProject(prj: {Name: string; HourRate?: number}): void {
    console.log("Change", prj);
    const prevMonth = Moment().subtract(1, 'months');
    const s: Partial<IHourState> = {
      Project: prj.Name,
      HourRate: prj.HourRate ?? 0,
    };
    if (this.state.Name === "") {
      s.Name = prj.Name + "-" + prevMonth.format("YYYY-MM");
    }
    this.setState(s as IHourState);
  }

  private pushUndo(): void {
    const linesCopy = JSON.parse(JSON.stringify(this.state.Lines)) as IHourLineState[];
    this.undoStack.push(linesCopy);
    this.redoStack = [];
  }

  private undo(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.undoStack.length === 0) return;
    const previousLines = this.undoStack.pop();
    if (!previousLines) return;

    const currentLines = JSON.parse(JSON.stringify(this.state.Lines)) as IHourLineState[];
    this.redoStack.push(currentLines);

    this.setState({Lines: previousLines, Total: this.updateTotalFromLines(previousLines)});
  }

  private redo(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.redoStack.length === 0) return;
    const nextLines = this.redoStack.pop();
    if (!nextLines) return;

    const currentLines = JSON.parse(JSON.stringify(this.state.Lines)) as IHourLineState[];
    this.undoStack.push(currentLines);

    this.setState({Lines: nextLines, Total: this.updateTotalFromLines(nextLines)});
  }

  private updateTotalFromLines(lines: IHourLineState[]): string {
    let total = new Big("0.00");
    lines.forEach(function(val) {
      total = total.plus(val.Hours);
    });
    return total.toFixed(2).toString();
  }

  private lineRemove(key: number): void {
    this.pushUndo();
    const newLines = [...this.state.Lines];
    console.log(`Deleted ${key} idx `, newLines.splice(key, 1)[0]);
    this.setState({Lines: newLines, Total: this.updateTotalFromLines(newLines)});
  }

  private async save(): Promise<void> {
    const req = {...this.state};
    req.Total = this.updateTotal();

    const res = await Axios.post<IHourState>(`/api/v1/hour/${this.props.entity}/${this.props.year}/concepts`, req);
    console.log(res.data);
    this.setState(res.data);
    history.replaceState({}, "", `#${this.props.entity}/${this.props.year}/hours/edit/concepts/${res.data.Name}`);
  }

  private async bill(): Promise<void> {
    const args = this.state;
    const res = await Axios.post(`/api/v1/hour/${this.props.entity}/${this.props.year}/concepts/${args.Name}/bill`, args);
    location.href = `#${this.props.entity}/${this.props.year}/invoices/edit/concepts/${res.headers["x-redirect-invoice"]}`;
  }

  private shortHand(d: number): string {
    const date = new Date(1000*60*60*d);
    let str = '';
    if (date.getUTCDate()-1 > 0) {
      str += date.getUTCDate()-1 + "d";
    }
    str += date.getUTCHours() + "h";
    str += date.getUTCMinutes() + "m";
    return str;
  }

  private toggleImport(e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>): void {
    e.preventDefault();
    this.setState({import: !this.state.import});
  }

	render(): React.JSX.Element {
    const lines: React.JSX.Element[] = [];
    const that = this;
    const isEditable = this.state.Status === "NEW" || this.state.Status === "CONCEPT";

    this.state.Lines.forEach(function(item:IHourLineState, idx:number) {
      lines.push(<tr key={idx}>
        <td><button type="button" className="btn btn-default btn-hover-danger" onClick={that.lineRemove.bind(that, idx)}><i className="fas fa-trash"></i></button></td>
        <td>{item.Day}</td>
        <td>{item.Start}</td><td>{item.Stop}</td><td>{that.shortHand(item.Hours)}</td><td>{item.Description}</td>
      </tr>);
    });

		return <form>
      <div>
		    <div className="panel panel-primary">
          <div className="panel-heading">
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
              <button type="button" onClick={this.recalc.bind(this)} className="btn btn-default btn-hover-success">
                <i className="fas fa-plus"></i>
                &nbsp;Add
              </button>
	          </div>
          </div>
		    </div>
    </div>
    <Import hide={this.state.import} onHide={this.toggleImport.bind(this)} importFn={this.importLine.bind(this)} />

    <div>
      <div className="panel panel-primary">
        <div className="panel-heading">
          <div className="pull-right">
            <div className="btn-group nm7">
              <button type="button" className="btn btn-default btn-hover-warning" disabled={this.undoStack.length === 0 || !isEditable} onClick={this.undo.bind(this)}><i className="fas fa-rotate-left"></i>&nbsp;Undo</button>
              <button type="button" className="btn btn-default btn-hover-warning" disabled={this.redoStack.length === 0 || !isEditable} onClick={this.redo.bind(this)}><i className="fas fa-rotate-right"></i>&nbsp;Redo</button>
              <button type="button" className="btn btn-default btn-hover-success" disabled={!isEditable} onClick={this.toggleImport.bind(this)}><i className="fas fa-arrow-up"></i>&nbsp;Import</button>
              <ActionButton className="btn btn-default btn-hover-success" disabled={!isEditable} onClick={this.save.bind(this)}><i className="fas fa-floppy-disk"></i>&nbsp;Save</ActionButton>
              <ActionButton className="btn btn-default btn-hover-danger" disabled={this.state.Status !== "CONCEPT"} onClick={this.bill.bind(this)}><i className="fas fa-share"></i>&nbsp;Bill</ActionButton>
            </div>
          </div>
          Sum ({this.state.Total} hours/{parseFloat(this.state.Total) * this.state.HourRate} EUR)
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
