import * as React from "react";
import Axios from "axios";
import * as Moment from "moment";
import {Autocomplete, LockedInput} from "../../shared/components";
import {InvoiceLineEdit} from "./edit-line";
import {InvoiceMail} from "./edit-mail";
import * as Big from "big.js";
import * as Struct from "./edit-struct";

export default class InvoiceEdit extends React.Component<{}, Struct.IInvoiceState> {
  private revisions: IInvoiceState[];
  constructor(props) {
    super(props);
    this.revisions = [];
    this.state = {
      Company: props.entity,
      Entity: {
        Name: "",
        Street1: "",
        Street2: ""
      },
      Customer: {
        Name: "",
        Street1: "",
        Street2: "",
        Vat: "",
        Coc: ""
      },
      Meta: {
        Conceptid: "",
        Status: "NEW",
        Invoiceid: "",
        Issuedate: null,
        Ponumber: "",
        Duedate: Moment().add(14, 'days').format('YYYY-MM-DD'),
        Paydate: null,
        HourFile: ""
      },
      Lines: [{
        Description: "",
        Quantity: "0.00",
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
        Iban: "",
        Bic: ""
      },
      State: {
        email: false
      },
      Mail: {
        From: "",
        Subject: "",
        To: "",
        Body: ""
      }
    };
  }

  componentDidMount() {
    let params = this.props;
    if (params.id) {
      this.ajax(params["bucket"], params.id);
    } else {
      this.ajaxDefaults(params.entity);
    }
  }

  private ajaxDefaults(entity: string) {
    let that = this;
    Axios.get(`/api/v1/entities/${entity}/details`)
    .then(res => {
      that.setState({
        Company: res.data.Entity.Name,
        Entity: {
          Name: res.data.User.Name,
          Street1: res.data.User.Address1,
          Street2: res.data.User.Address2,
        },
        Bank: {
          Vat: res.data.Entity.VAT,
          Coc: res.data.Entity.COC,
          Iban: res.data.Entity.IBAN,
          Bic: res.data.Entity.BIC
        }
      });
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private ajax(bucket: string, name: string) {
    Axios.get(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${name}`)
    .then(res => {
      this.parseInput.call(this, res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private parseInput(data: IInvoiceState, newbucket) {
    if (newbucket) {
      this.props.bucket = newbucket;
    }
    let url = `#${this.props.entity}/${this.props.year}/invoices/edit/${this.props.bucket}/${data.Meta.Conceptid}`;
    if (window.location.href != url) {
      // Update URL so refresh will keep the invoice open
      history.replaceState({}, "", url);
      this.props.id = data.Meta.Conceptid;
    }
    data.Meta.Issuedate = data.Meta.Issuedate ? Moment(data.Meta.Issuedate).format('YYYY-MM-DD') : null;
    data.Meta.Duedate = data.Meta.Duedate ? Moment(data.Meta.Duedate).format('YYYY-MM-DD') : null;
    data.Meta.Paydate = data.Meta.Paydate ? Moment(data.Meta.Paydate).format('YYYY-MM-DD') : null;

    this.setState(data);
  }

  private triggerChange(indices: string[], val: string) {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    let node: any = this.state;
    for (let i = 0; i < indices.length-1; i++) {
      node = node[ indices[i] ];
    }
    node[indices[indices.length-1]] = val;

    // Any post-processing
    if (indices[0] === "Lines") {
      let idx = indices[1] as any;
      this.state.Lines[idx] = this.lineUpdate(this.state.Lines[idx]);
      this.state.Total = this.totalUpdate(this.state.Lines);
    }
    this.setState(this.state);
    this.revisions.push({}); // TODO :)
  }

  private defaultDecimal(val) {
    if (val === "") {
      return "0.00";
    }
    val = val.replace(/,/g, ".");
    val = val.replace(/[^\d.]/g, '')

    let idx = val.indexOf(".");
    if (idx === -1) {
      return val + ".00";
    }
    if (idx === 0) {
      val = "0" + val;
    }

    if (val.length - idx === 1) {
      return val + "0";
    }
    if (val.length - idx === 2) {
      return val + "0";
    }
    return val;
  }

  private lineUpdate(line: IInvoiceLine) {
    line.Quantity = this.defaultDecimal(line.Quantity);
    line.Price = this.defaultDecimal(line.Price);

    line.Total = new Big(line.Price).times(line.Quantity).round(2).toFixed(2).toString();
    return line;
  }

  private totalUpdate(lines: IInvoiceLine[]) {
    let ex = new Big(0);
    lines.forEach(function(val: IInvoiceLine) {
      console.log("Add", val.Total);
      ex = ex.plus(val.Total);
    });

    let tax = ex.div("100").times("21");
    if (this.state.Customer.Vat.length > 0) {
      var country = this.state.Customer.Vat.substr(0, 2).toUpperCase();
      console.log("Country " + country);
      if (country !== "NL") {
        tax = new Big(0);
      }
    }
    let total = ex.plus(tax);
    console.log("totals (ex,tax,total)", ex.toString(), tax.toString(), total.toString());

    return {
      Ex: ex.round(2).toFixed(2).toString(),
      Tax: tax.round(2).toFixed(2).toString(),
      Total: total.round(2).toFixed(2).toString()
    };
  }

  handleChange(e: InputEvent) {
    console.log("handleChange", e.target.dataset["key"]);
    let indices = e.target.dataset["key"].split('.');
    this.triggerChange(indices, e.target.value);
  }

  private save(e: BrowserEvent) {
    e.preventDefault();
    let req = JSON.parse(JSON.stringify(this.state));
    console.log(req);

    Axios.post('/api/v1/invoice/'+this.props.entity+'/'+this.props.year, req)
    .then(res => {
      this.parseInput.call(this, res.data);
    })
    .catch(err => {
      if (err.response && err.response.status === 417) {
        prettyErr(err.response.data);
        return;
      }
      handleErr(err);
    });
  }

  private reset(e: BrowserEvent) {
    e.preventDefault();
    Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${this.state.Meta.Conceptid}/reset`, {})
    .then(res => {
      this.parseInput.call(this, res.data, res.headers["x-bucket-change"]);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private finalize(e: BrowserEvent) {
    e.preventDefault();
    Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${this.state.Meta.Conceptid}/finalize`, {})
    .then(res => {
      this.parseInput.call(this, res.data, res.headers["x-bucket-change"]);
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
    let url = `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${this.props.id}/pdf`;
    console.log(`Open PDF ${url}`);
    location.assign(url);
  }

  private selectCustomer(data) {
    console.log("Select customer", data);
    this.setState({
      Customer: {
        Name: data.Name,
        Street1: data.Street1,
        Street2: data.Street2,
        Vat: data.VAT,
        Coc: data.COC
      },
      Notes: data.NoteAdd,
      Mail: {
        To: data.BillingEmail.join(", ")
      }
    });
  }

  private email() {
    this.setState({State: {email: !this.state.State.email}});
  }

	render() {
    let inv = this.state;
    let that = this;

		return <form><div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <button className="btn btn-default btn-hover-success" disabled={this.revisions.length === 0 || inv.Meta.Status === "FINAL"} onClick={this.save.bind(this)}><i className="fa fa-floppy-o"></i> Save</button>
                <button className="btn btn-default btn-hover-danger" disabled={inv.Meta.Status !== "CONCEPT"} onClick={this.finalize.bind(this)}><i className="fa fa-lock"></i> Finalize</button>
                <a className="btn btn-default btn-hover-success" disabled={inv.Meta.Status !== "FINAL"} onClick={this.pdf.bind(this)}><i className="fa fa-file-pdf-o"></i> PDF</a>
                <a className="btn btn-default btn-hover-success" disabled={inv.Meta.Status !== "FINAL"} onClick={this.email.bind(this)}><i className="fa fa-send"></i> E-mail</a>

                <button className="btn btn-default btn-hover-danger" disabled={inv.Meta.Status !== "FINAL"} onClick={this.reset.bind(this)}><i className="fa fa-unlock"></i> Reset</button>

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
      <Autocomplete data-key="Customer.Name" onSelect={that.selectCustomer.bind(that)} onChange={that.handleChange.bind(that)} required={true} placeholder="Company Name" url={"/api/v1/debtors/"+that.props.entity+"/search"} value={inv.Customer.Name} />
      <div className="pr"><input className="form-control" type="text" data-key="Customer.Street1" onChange={that.handleChange.bind(this)} value={inv.Customer.Street1} placeholder="Street1" /><i className="fa fa-asterisk text-danger fa-input"></i></div>
      <div className="pr"><input className="form-control" type="text" data-key="Customer.Street2" onChange={that.handleChange.bind(this)} value={inv.Customer.Street2} placeholder="Street2" /><i className="fa fa-asterisk text-danger fa-input"></i></div>

      <input className="form-control" type="text" data-key="Customer.Vat" onChange={that.handleChange.bind(this)} value={inv.Customer.Vat} placeholder="VAT-number"/>
      <input className="form-control" type="text" data-key="Customer.Coc" onChange={that.handleChange.bind(this)} value={inv.Customer.Coc} placeholder="Chamber Of Commerce (CoC)"/>

    </div>
    <div className="meta col-sm-offset-3 col-sm-5">
      <table className="table">
        <tbody>
          <tr>
            <td className="text">
              Invoice ID
            </td>
            <td>
              <LockedInput type="text" value={inv.Meta.Invoiceid} placeholder="AUTOGENERATED" onChange={that.handleChange.bind(that)} locked={true} data-key="Meta.Invoiceid"/>
            </td>
          </tr>
          <tr>
            <td className="text">Issue Date</td>
            <td>
              <LockedInput type="date" value={inv.Meta.Issuedate} placeholder="AUTOGENERATED" onChange={that.handleChange.bind(that)} locked={true} data-key="Meta.Issuedate"/>
            </td>
          </tr>
          <tr>
            <td className="text">PO Number</td>
            <td><input className="form-control" type="text" data-key="Meta.Ponumber" onChange={that.handleChange.bind(that)} value={inv.Meta.Ponumber}/></td>
          </tr>
          <tr>
            <td className="text">Due Date</td>
            <td>
                <input type="date" value={inv.Meta.Duedate} onChange={that.handleChange.bind(that)} className="form-control" data-key="Meta.Duedate" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <InvoiceLineEdit parent={this} />

  <div className="row notes col-sm-12">
    <p>Notes</p>
    <textarea className="form-control" data-key="Notes" onChange={this.handleChange.bind(this)} value={inv.Notes}/>
  </div>
  <div className="row banking">
    <div className="col-sm-4">
      <p>Banking details</p>
      <table className="table mb0"><tbody>
        <tr><td className="text">VAT</td><td className="pr">
          <LockedInput type="text" value={inv.Bank.Vat} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Vat" required={true} /></td></tr>
        <tr><td className="text">CoC</td><td className="pr"><LockedInput type="text" value={inv.Bank.Coc} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Coc" required={true} /></td></tr>
        <tr><td className="text">IBAN</td><td className="pr"><LockedInput type="text" value={inv.Bank.Iban} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Iban" required={true} /></td></tr>
      </tbody></table>
      <small><i className="fa fa-info"></i> Edit these from your settings file.</small>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div><InvoiceMail parent={this} onHide={this.email.bind(this)} hide={this.state.State.email} /></form>;
	}
}
