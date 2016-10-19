import * as React from "react";
import {IInjectedProps} from "react-router";
import Axios from "axios";
import * as Big from "big.js";
import * as Moment from "moment";
import * as DatePicker from "react-datepicker";

import "./invoice.css";
import "react-datepicker/dist/react-datepicker.css";
//require('react-datepicker/dist/react-datepicker.css');

type IInvoiceStatus = "NEW" | "CONCEPT" | "FINAL";
interface IInvoiceProps extends React.Props<InvoiceEdit> {
  id : string
}
interface IInvoiceEntity {
  Name: string
  Street1: string
  Street2: string
}
interface IInvoiceCustomer {
  Name: string
  Street1: string
  Street2: string
}
interface IInvoiceMeta {
  Conceptid: string
  Status: IInvoiceStatus
  Invoiceid: string
  InvoiceidL: boolean
  Issuedate?: moment.Moment
  IssuedateL: boolean
  Ponumber: string
  Duedate?: moment.Moment
}
interface IInvoiceLine {
  Description: string
  Quantity: string //number
  Price: string //number
  Total: string //number
}
interface IInvoiceTotal {
  Ex: string //number
  Tax: string //number
  Total: string //number
}
interface IInvoiceBank {
  Vat: string
  Coc: string
  Iban: string
}
export interface IInvoiceState {
  Company?: string
  Entity?: IInvoiceEntity
  Customer?: IInvoiceCustomer
  Meta?: IInvoiceMeta
  Lines?: IInvoiceLine[]
  Notes?: string
  Total?: IInvoiceTotal
  Bank?: IInvoiceBank
}

export default class InvoiceEdit extends React.Component<IInjectedProps, IInvoiceState> {
  private revisions: IInvoiceState[];
  constructor(props: IInjectedProps) {
    super(props);
    this.revisions = [];
    this.state = {
      Company: "RootDev",
      Entity: {
        Name: "M.P. Droog",
        Street1: "Dorpsstraat 236a",
        Street2: "Obdam, 1713HP, NL"
      },
      Customer: {
        Name: "XSNews B.V.",
        Street1: "New Yorkstraat 9-13",
        Street2: "1175 RD Lijnden"
      },
      Meta: {
        Conceptid: "",
        Status: "NEW",
        Invoiceid: "",
        InvoiceidL: true,
        Issuedate: null,
        IssuedateL: true,
        Ponumber: "",
        Duedate: Moment().add(14, 'days')
      },
      Lines: [{
        Description: "",
        Quantity: "0",
        Price: "0.00",
        Total: "0.00"
      }],
      Notes: "",
      Total: {
        Ex: "0.00",
        Tax: "0.00",
        Total: "0.00"
      },
      Bank: {
        Vat: "",
        Coc: "",
        Iban: ""
      }
    };
  }

  componentDidMount() {
    console.log("Load hour name=" + this.props.params["id"]);
    this.ajax(this.props.params["id"]);
  }

  private ajax(name: string) {
    Axios.get('/api/invoice/'+name)
    .then(res => {
      this.parseInput.call(this, res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private parseInput(data: IInvoiceState) {
    console.log(data);
    if (window.location.href != "?#invoice-add/" + data.Meta.Conceptid) {
      // Update URL so refresh will keep the invoice open
      history.replaceState({}, "", "?#invoice-add/" + data.Meta.Conceptid);
      this.props.params["id"] = data.Meta.Conceptid;
    }
    data.Meta.Issuedate = data.Meta.Issuedate ? Moment(data.Meta.Issuedate) : null;
    data.Meta.Duedate = data.Meta.Duedate ? Moment(data.Meta.Duedate) : null;
    data.Meta.InvoiceidL = true;
    data.Meta.IssuedateL = true;
    this.setState(data);
  }

  private lineAdd() {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    console.log("Add invoice line");
    this.state.Lines.push({
      Description: "",
      Quantity: "0",
      Price: "0.00",
      Total: "0.00"
    });
    this.setState({Lines: this.state.Lines});
  }

  private lineRemove(key: number) {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    var line: IInvoiceLine = this.state.Lines[key];
    var isEmpty = line.Description === ""
      && line.Quantity === "0"
      && line.Price === "0.00"
      && line.Total === "0.00";
    var isOk = !isEmpty && confirm("Are you sure you want to remove the invoiceline with description '" + line.Description + "'?");

    if (isEmpty || isOk) {
      console.log("Remove invoice line with key=" + key);
      console.log("Deleted idx ", this.state.Lines.splice(key, 1)[0]);
      this.setState({Lines: this.state.Lines});
    }
  }

  lineUpdate(line: IInvoiceLine) {
    line.Total = new Big(line.Price).times(line.Quantity).round(2).toFixed(2).toString();
    return line;
  }

  private totalUpdate(lines: IInvoiceLine[]) {
    var ex = new Big(0);
    lines.forEach(function(val: IInvoiceLine) {
      console.log("Add", val.Total);
      ex = ex.plus(val.Total);
    });
    // TODO: Hardcoded to 21%
    var tax = ex.div("100").times("21");
    var total = ex.plus(tax);
    console.log("totals (ex,tax,total)", ex.toString(), tax.toString(), total.toString());

    return {
      Ex: ex.round(2).toFixed(2).toString(),
      Tax: tax.round(2).toFixed(2).toString(),
      Total: total.round(2).toFixed(2).toString()
    };
  }

  private triggerChange(indices: string[], val: string) {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    var node: any = this.state;
    for (var i = 0; i < indices.length-1; i++) {
      node = node[ indices[i] ];
    }
    node[indices[indices.length-1]] = val;

    // Any post-processing
    if (indices[0] === "Lines") {
      var idx = indices[1] as any;
      this.state.Lines[idx] = this.lineUpdate(this.state.Lines[idx]);
      this.state.Total = this.totalUpdate(this.state.Lines);
    }
    this.setState(this.state);
    this.revisions.push({}); // TODO :)
  }

  private handleChange(e: InputEvent) {
    console.log("handleChange", e.target.dataset["key"]);
    var indices = e.target.dataset["key"].split('.');
    this.triggerChange(indices, e.target.value);
  }

  private handleChangeDate(id: string) {
    var indices = id.split('.');
    var that = this;
    return function(val: string) {
      console.log("handleChangeDate", id, val);
      that.triggerChange.call(that, indices, val);
    };
  }

  private toggleChange(id: string, val: boolean) {
    var indices = id.split('.');
    var that = this;
    val = !val; // Invert value
    return function() {
      console.log("toggleChange", id, val);
      that.triggerChange.call(that, indices, val);
    };
  }

  private save() {
    var req = JSON.parse(JSON.stringify(this.state)); // deepCopy
    req.Meta.Issuedate = this.state.Meta.Issuedate ? this.state.Meta.Issuedate.format('YYYY-MM-DD') : "";
    req.Meta.Duedate = this.state.Meta.Duedate ? this.state.Meta.Duedate.format('YYYY-MM-DD') : "";
    console.log(req);

    Axios.post('/api/invoice', req)
    .then(res => {
      this.parseInput.call(this, res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private finalize() {
    Axios.post('/api/invoice/' + this.state.Meta.Conceptid + '/finalize', {})
    .then(res => {
      this.parseInput.call(this, res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private pdf() {
    if (this.state.Meta.Status !== 'FINAL') {
      console.log("PDF only available in finalized invoices");
      return;
    }
    var url = '/api/invoice/'+this.props.params["id"]+'/pdf';
    console.log("Open PDF " + url);
    location.assign(url);
  }

	render() {
    var inv = this.state;
    var that = this;
    var lines: JSX.Element[] = [];

    inv.Lines.forEach(function(line: IInvoiceLine, idx: number) {
      console.log(inv.Meta.Status);
      lines.push(
        <tr key={"line"+idx}>
          <td><button disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-danger faa-parent animated-hover' : '')} onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash faa-flash"></i></button></td>
          <td><input className="form-control" type="text" data-key={"Lines."+idx+".Description"} onChange={that.handleChange.bind(this)} value={line.Description}/></td>
          <td><input className="form-control" type="text" data-key={"Lines."+idx+".Quantity"} onChange={that.handleChange.bind(this)} value={line.Quantity}/></td>
          <td><input className="form-control" type="text" data-key={"Lines."+idx+".Price"} onChange={that.handleChange.bind(this)} value={line.Price}/></td>
          <td><input className="form-control" readOnly={true} disabled={true} type="text" data-key={"Lines."+idx+".Total"} value={line.Total}/></td>
        </tr>
      );
    });

		return <form><div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <button className="btn btn-default btn-hover-success" disabled={this.revisions.length === 0 || inv.Meta.Status !== "FINAL"} onClick={this.save.bind(this)}><i className="fa fa-floppy-o"></i> Save</button>
                <button className="btn btn-default btn-hover-danger" disabled={inv.Meta.Status !== "CONCEPT"} onClick={this.finalize.bind(this)}><i className="fa fa-lock"></i> Finalize</button>
                <a className="btn btn-default btn-hover-success" disabled={inv.Meta.Status !== "FINAL"} onClick={this.pdf.bind(this)}><i className="fa fa-file-pdf-o"></i> PDF</a>
              </div>

            </div>
            New Invoice
          </div>
          <div className="panel-body">

<div className={"invoice group " + (inv.Meta.Status === 'FINAL' ? 'o50' : '')}>
  <div className="row">
    <div className="company col-sm-4">
      <input className="form-control" type="text" data-key="Company" onChange={that.handleChange.bind(this)} value={inv.Company}/>
    </div>

    <div className="col-sm-offset-3 col-sm-1">
      From
    </div>
    <div className="entity col-sm-4">
      <input className="form-control" type="text" data-key="Entity.Name" onChange={that.handleChange.bind(this)} value={inv.Entity.Name}/>
      <input className="form-control" type="text" data-key="Entity.Street1" onChange={that.handleChange.bind(this)} value={inv.Entity.Street1}/>
      <input className="form-control" type="text" data-key="Entity.Street2" onChange={that.handleChange.bind(this)} value={inv.Entity.Street2}/>
    </div>
  </div>

  <div className="row">
    <div className="col-sm-1">
      Invoice For
    </div>
    <div className="col-sm-3">
      <input className="form-control" type="text" data-key="Customer.Name" onChange={that.handleChange.bind(this)} value={inv.Customer.Name}/>
      <input className="form-control" type="text" data-key="Customer.Street1" onChange={that.handleChange.bind(this)} value={inv.Customer.Street1}/>
      <input className="form-control" type="text" data-key="Customer.Street2" onChange={that.handleChange.bind(this)} value={inv.Customer.Street2}/>
    </div>
    <div className="meta col-sm-offset-3 col-sm-5">
      <table className="table">
        <tbody>
          <tr>
            <td className="text">
              Invoice ID
            </td>
            <td>
              <div className="input-group">
                <input className="form-control" disabled={inv.Meta.InvoiceidL} type="text" data-key="Meta.Invoiceid" onChange={that.handleChange.bind(this)} value={inv.Meta.Invoiceid} placeholder="AUTOGENERATED"/>
                <div className="input-group-addon"><a className="" onClick={this.toggleChange('Meta.InvoiceidL', inv.Meta.InvoiceidL)}><i className={"fa " + (inv.Meta.InvoiceidL?"fa-lock":"fa-unlock")}></i></a></div>
              </div>
            </td>
          </tr>
          <tr>
            <td className="text">Issue Date</td>
            <td>
              <div className="input-group">
                <DatePicker
                className="form-control"
                disabled={inv.Meta.IssuedateL}
                dateFormat="YYYY-MM-DD"
                selected={inv.Meta.Issuedate}
                placeholderText="AUTOGENERATED"
                onChange={this.handleChangeDate('Meta.Issuedate').bind(this)} />
                <div className="input-group-addon"><a className="" onClick={this.toggleChange('Meta.IssuedateL', inv.Meta.IssuedateL)}><i className={"fa " + (inv.Meta.IssuedateL?"fa-lock":"fa-unlock")}></i></a></div>
              </div>
            </td>
          </tr>
          <tr>
            <td className="text">PO Number</td>
            <td><input className="form-control" type="text" data-key="Meta.Ponumber" onChange={that.handleChange.bind(this)} value={inv.Meta.Ponumber}/></td>
          </tr>
          <tr>
            <td className="text">Due Date</td>
            <td>
                <DatePicker
                className="form-control"
                dateFormat="YYYY-MM-DD"
                selected={inv.Meta.Duedate}
                onChange={this.handleChangeDate('Meta.Duedate').bind(this)} />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <table className="table table-striped">
    <thead>
      <tr>
        <th>&nbsp;</th>
        <th>Description</th>
        <th>Quantity</th>
        <th>Price</th>
        <th>Line Total</th>
      </tr>
    </thead>
    <tbody>{lines}</tbody>
    <tfoot>
      <tr>
        <td colSpan={3} className="text">
          <button disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-success faa-parent animated-hover' : '')} onClick={this.lineAdd.bind(this)}><i className="fa fa-plus faa-bounce"></i> Add row</button>
        </td>
        <td className="text">Total (ex tax)</td>
        <td><input className="form-control" disabled={true} type="text" data-key="Total.Ex" readOnly={true} value={inv.Total.Ex}/></td>
      </tr>
      <tr>
        <td colSpan={3}></td>
        <td className="text">Tax (21%)</td>
        <td><input className="form-control" disabled={true} type="text" data-key="Total.Tax" readOnly={true} value={inv.Total.Tax}/></td>
      </tr>
      <tr>
        <td colSpan={3}>&nbsp;</td>
        <td className="text">Total</td>
        <td><input className="form-control" disabled={true} type="text" data-key="Total.Total" readOnly={true} value={inv.Total.Total}/></td>
      </tr>
    </tfoot>
  </table>

  <div className="row notes col-sm-12">
    <p>Notes</p>
    <textarea className="form-control" data-key="Notes" onChange={this.handleChange.bind(this)} value={inv.Notes}/>
  </div>
  <div className="row banking">
    <div className="col-sm-4">
      <p>Banking details</p>
      <table className="table"><tbody>
        <tr><td className="text">VAT</td><td><input className="form-control" type="text" data-key="Bank.Vat" onChange={this.handleChange.bind(this)} value={inv.Bank.Vat}/></td></tr>
        <tr><td className="text">CoC</td><td><input className="form-control" type="text" data-key="Bank.Coc" onChange={this.handleChange.bind(this)} value={inv.Bank.Coc}/></td></tr>
        <tr><td className="text">IBAN</td><td><input className="form-control" type="text" data-key="Bank.Iban" onChange={this.handleChange.bind(this)} value={inv.Bank.Iban}/></td></tr>
      </tbody></table>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div></form>;
	}
}
