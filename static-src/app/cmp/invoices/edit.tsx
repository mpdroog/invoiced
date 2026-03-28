import * as React from "react";
import Axios from "axios";
import Moment from "moment";
import {Autocomplete, LockedInput} from "../../shared/components";
import {ActionButton} from "../../shared/ActionButton";
import {InvoiceLineEdit} from "./edit-line";
import {InvoiceMail} from "./edit-mail";
import Big from "big.js";
import * as Struct from "./edit-struct";

interface InvoiceEditProps {
  entity: string;
  year: string;
  id?: string;
  bucket?: string;
}

interface InvoiceEditState extends Struct.IInvoiceState {
  State: {
    email: boolean;
  };
}

export default class InvoiceEdit extends React.Component<InvoiceEditProps, InvoiceEditState> {
  private undoStack: Struct.IInvoiceLine[][];
  private redoStack: Struct.IInvoiceLine[][];
  public errors: Record<string, string>;

  constructor(props: InvoiceEditProps) {
    super(props);
    this.undoStack = [];
    this.redoStack = [];
    this.errors = {};
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
        Coc: "",
        Tax: "NL21",
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

  componentDidMount(): void {
    const params = this.props;
    if (params.id && params.bucket) {
      this.ajax(params.bucket, params.id);
    } else {
      this.ajaxDefaults(params.entity);
    }
  }

  private ajaxDefaults(entity: string): void {
    Axios.get(`/api/v1/entities/${entity}/details`)
    .then(res => {
      this.setState({
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

  private ajax(bucket: string, name: string): void {
    Axios.get(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${name}`)
    .then(res => {
      this.parseInput(res.data);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private parseInput(data: Struct.IInvoiceState, newbucket?: string): void {
    const meta = data.Meta;
    if (!meta) return;

    const bucket = newbucket || this.props.bucket || 'concepts';
    const url = `#${this.props.entity}/${this.props.year}/invoices/edit/${bucket}/${meta.Conceptid}`;
    if (window.location.hash !== url) {
      // Update URL so refresh will keep the invoice open
      history.replaceState({}, "", url);
    }
    meta.Issuedate = meta.Issuedate ? Moment(meta.Issuedate).format('YYYY-MM-DD') : null;
    meta.Duedate = meta.Duedate ? Moment(meta.Duedate).format('YYYY-MM-DD') : null;
    meta.Paydate = meta.Paydate ? Moment(meta.Paydate).format('YYYY-MM-DD') : null;

    this.setState(data as InvoiceEditState);
  }

  private triggerChange(indices: string[], val: string): void {
    const meta = this.state.Meta;
    if (meta?.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let node: any = this.state;
    for (let i = 0; i < indices.length-1; i++) {
      const key = indices[i];
      if (key === undefined) return;
      node = node[key];
    }
    const lastKey = indices[indices.length-1];
    if (lastKey === undefined) return;
    node[lastKey] = val;

    // Any post-processing
    const lines = this.state.Lines;
    const lineIdx = indices[1];
    if (indices[0] === "Lines" && lines && lineIdx !== undefined) {
      const idx = parseInt(lineIdx, 10);
      const line = lines[idx];
      if (line) {
        lines[idx] = this.lineUpdate(line);
      }
      const newState = {...this.state, Lines: lines, Total: this.totalUpdate(lines)};
      this.setState(newState);
    } else {
      this.setState({...this.state});
    }
  }

  private defaultDecimal(val: string, isNeg: boolean): string {
    if (val === "") {
      return "0.00";
    }
    val = val.replace(/,/g, ".");
    if (isNeg) {
      val = val.replace(/[^\d.-]/g, '');
    } else {
      val = val.replace(/[^\d.]/g, '');
    }

    const idx = val.indexOf(".");
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

  private lineUpdate(line: Struct.IInvoiceLine): Struct.IInvoiceLine {
    line.Quantity = this.defaultDecimal(line.Quantity, false);
    line.Price = this.defaultDecimal(line.Price, true);

    line.Total = new Big(line.Price).times(line.Quantity).round(2).toFixed(2).toString();
    return line;
  }

  private totalUpdate(lines: Struct.IInvoiceLine[]): Struct.IInvoiceTotal {
    let ex = new Big(0);
    lines.forEach(function(val: Struct.IInvoiceLine) {
      console.log("Add", val.Total);
      ex = ex.plus(val.Total);
    });

    let tax = new Big(0);
    const customer = this.state.Customer;
    if (customer?.Tax === "NL21") {
      tax = ex.div("100").times("21");
    }
    const total = ex.plus(tax);
    console.log("totals (ex,tax,total)", ex.toString(), tax.toString(), total.toString());

    return {
      Ex: ex.round(2).toFixed(2).toString(),
      Tax: tax.round(2).toFixed(2).toString(),
      Total: total.round(2).toFixed(2).toString()
    };
  }

  handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>): void {
    const key = e.target.dataset["key"];
    if (!key) return;
    console.log("handleChange", key);
    const indices = key.split('.');
    this.triggerChange(indices, e.target.value);
  }

  private async save(): Promise<void> {
    const req = JSON.parse(JSON.stringify(this.state));
    console.log(req);

    const res = await Axios.post('/api/v1/invoice/'+this.props.entity+'/'+this.props.year, req);
    this.parseInput.call(this, res.data);
  }

  private async reset(): Promise<void> {
    const meta = this.state.Meta;
    if (!meta) return;
    const res = await Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${meta.Conceptid}/reset`, {});
    // Reset always moves to concepts bucket
    this.parseInput(res.data, "concepts");
  }

  private async finalize(): Promise<void> {
    const meta = this.state.Meta;
    if (!meta) return;
    const res = await Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${meta.Conceptid}/finalize`, {});
    this.parseInput(res.data, res.headers["x-bucket-change"]);
  }

  private pdf(): void {
    const meta = this.state.Meta;
    if (meta?.Status !== 'FINAL') {
      console.log("PDF only available in finalized invoices");
      return;
    }
    const url = `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${this.props.id}/pdf`;
    console.log(`Open PDF ${url}`);
    location.assign(url);
  }

  private selectCustomer(data: {Name: string; Street1?: string; Street2?: string; VAT?: string; COC?: string; NoteAdd?: string; BillingEmail?: string[]}): void {
    console.log("Select customer", data);
    this.setState({
      Customer: {
        Name: data.Name,
        Street1: data.Street1 || "",
        Street2: data.Street2 || "",
        Vat: data.VAT || "",
        Coc: data.COC || "",
        Tax: "NL21"
      },
      Notes: data.NoteAdd,
      Mail: {
        ...this.state.Mail,
        From: this.state.Mail?.From || "",
        Subject: this.state.Mail?.Subject || "",
        Body: this.state.Mail?.Body || "",
        To: (data.BillingEmail || []).join(", ")
      }
    });
  }

  private email(): void {
    this.setState({State: {email: !this.state.State.email}});
  }

  public pushUndo(): void {
    // Save current lines state for undo
    const linesCopy = JSON.parse(JSON.stringify(this.state.Lines || []));
    this.undoStack.push(linesCopy);
    this.redoStack = []; // Clear redo stack on new change
  }

  private undo(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.undoStack.length === 0) return;
    const previousLines = this.undoStack.pop();
    if (!previousLines) return;

    // Save current state to redo stack
    const currentLines = JSON.parse(JSON.stringify(this.state.Lines || []));
    this.redoStack.push(currentLines);

    // Restore previous state
    this.setState({Lines: previousLines, Total: this.totalUpdate(previousLines)});
  }

  private redo(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.redoStack.length === 0) return;
    const nextLines = this.redoStack.pop();
    if (!nextLines) return;

    // Save current state to undo stack
    const currentLines = JSON.parse(JSON.stringify(this.state.Lines || []));
    this.undoStack.push(currentLines);

    // Restore next state
    this.setState({Lines: nextLines, Total: this.totalUpdate(nextLines)});
  }

	render(): React.JSX.Element {
    const inv = this.state;
    const that = this;
    const meta = inv.Meta || { Status: "NEW", Conceptid: "", Invoiceid: "", Ponumber: "", HourFile: "" };
    const entity = inv.Entity || { Name: "", Street1: "", Street2: "" };
    const customer = inv.Customer || { Name: "", Street1: "", Street2: "", Vat: "", Coc: "", Tax: "NL21" };

		return <form><div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <button type="button" className="btn btn-default btn-hover-warning" disabled={this.undoStack.length === 0 || meta.Status === "FINAL"} onClick={this.undo.bind(this)}><i className="fas fa-rotate-left"></i> Undo</button>
                <button type="button" className="btn btn-default btn-hover-warning" disabled={this.redoStack.length === 0 || meta.Status === "FINAL"} onClick={this.redo.bind(this)}><i className="fas fa-rotate-right"></i> Redo</button>
                <ActionButton type="button" className="btn btn-default btn-hover-success" disabled={meta.Status === "FINAL"} onClick={this.save.bind(this)}><i className="fas fa-floppy-disk"></i> Save</ActionButton>
                <ActionButton type="button" className="btn btn-default btn-hover-danger" disabled={meta.Status !== "CONCEPT"} onClick={this.finalize.bind(this)}><i className="fas fa-lock"></i> Finalize</ActionButton>
                <button type="button" className={"btn btn-default btn-hover-success" + (meta.Status !== "FINAL" ? " disabled" : "")} onClick={this.pdf.bind(this)}><i className="far fa-file-pdf"></i> PDF</button>
                <button type="button" className={"btn btn-default btn-hover-success" + (meta.Status !== "FINAL" ? " disabled" : "")} onClick={this.email.bind(this)}><i className="fas fa-paper-plane"></i> E-mail</button>

                <ActionButton type="button" className="btn btn-default btn-hover-danger" disabled={meta.Status !== "FINAL"} onClick={this.reset.bind(this)}><i className="fas fa-unlock"></i> Reset</ActionButton>

              </div>

            </div>
            New Invoice
          </div>
          <div className="panel-body">

<div className={"invoice group " + (meta.Status === 'FINAL' ? 'o50' : '')}>
  <div className="row">
    <div className="company col-sm-4">
      <input className="form-control" type="text" data-key="Company" onChange={that.handleChange.bind(this)} value={inv.Company}/>
    </div>

    <div className="col-sm-offset-3 col-sm-1">
      From
    </div>
    <div className="entity col-sm-4">
      <input className="form-control" type="text" data-key="Entity.Name" onChange={that.handleChange.bind(this)} value={entity.Name}/>
      <input className="form-control" type="text" data-key="Entity.Street1" onChange={that.handleChange.bind(this)} value={entity.Street1}/>
      <input className="form-control" type="text" data-key="Entity.Street2" onChange={that.handleChange.bind(this)} value={entity.Street2}/>
    </div>
  </div>

  <div className="row">
    <div className="col-sm-1">
      Invoice For
    </div>
    <div className="col-sm-3">
      <Autocomplete data-key="Customer.Name" onSelect={that.selectCustomer.bind(that)} onChange={that.handleChange.bind(that)} required={true} placeholder="Company Name" url={"/api/v1/debtors/"+that.props.entity+"/search"} value={customer.Name} />
      <div className="pr"><input className="form-control" type="text" data-key="Customer.Street1" onChange={that.handleChange.bind(this)} value={customer.Street1} placeholder="Street1" /><i className="fas fa-asterisk text-danger fa-input"></i></div>
      <div className="pr"><input className="form-control" type="text" data-key="Customer.Street2" onChange={that.handleChange.bind(this)} value={customer.Street2} placeholder="Street2" /><i className="fas fa-asterisk text-danger fa-input"></i></div>

      <input className="form-control" type="text" data-key="Customer.Vat" onChange={that.handleChange.bind(this)} value={customer.Vat} placeholder="VAT-number"/>
      <input className="form-control" type="text" data-key="Customer.Coc" onChange={that.handleChange.bind(this)} value={customer.Coc} placeholder="Chamber Of Commerce (CoC)"/>
      <select className="form-control" data-key="Customer.Tax" onChange={that.handleChange.bind(this)} value={customer.Tax}>
        <option value="NL21">NL21 (Domestic)</option>
        <option value="EU0">EU (ICP)</option>
        <option value="WORLD0">Outside EU (export)</option>
      </select>

    </div>
    <div className="meta col-sm-offset-3 col-sm-5">
      <table className="table">
        <tbody>
          <tr>
            <td className="text">
              Invoice ID
            </td>
            <td>
              <LockedInput type="text" value={meta.Invoiceid} placeholder="AUTOGENERATED" onChange={that.handleChange.bind(that)} locked={true} data-key="Meta.Invoiceid"/>
            </td>
          </tr>
          <tr>
            <td className="text">Issue Date</td>
            <td>
              <LockedInput type="date" value={meta.Issuedate || ""} placeholder="AUTOGENERATED" onChange={that.handleChange.bind(that)} locked={true} data-key="Meta.Issuedate"/>
            </td>
          </tr>
          <tr>
            <td className="text">PO Number</td>
            <td><input className="form-control" type="text" data-key="Meta.Ponumber" onChange={that.handleChange.bind(that)} value={meta.Ponumber}/></td>
          </tr>
          <tr>
            <td className="text">Due Date</td>
            <td>
                <input type="date" value={meta.Duedate || ""} onChange={that.handleChange.bind(that)} className="form-control" data-key="Meta.Duedate" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <InvoiceLineEdit parent={this} />

  <div className="row notes col-sm-12">
    <p>Notes</p>
    <small><i className="fas fa-info"></i> This text is added to the invoice</small>
    <textarea className="form-control" data-key="Notes" onChange={this.handleChange.bind(this)} value={inv.Notes}/>
  </div>
  <div className="row banking">
    <div className="col-sm-4">
      <p>Banking details</p>
      <table className="table mb0"><tbody>
        <tr><td className="text">VAT</td><td className="pr">
          <LockedInput type="text" value={(inv.Bank || {Vat: "", Coc: "", Iban: "", Bic: ""}).Vat} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Vat" required={true} /></td></tr>
        <tr><td className="text">CoC</td><td className="pr"><LockedInput type="text" value={(inv.Bank || {Vat: "", Coc: "", Iban: "", Bic: ""}).Coc} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Coc" required={true} /></td></tr>
        <tr><td className="text">IBAN</td><td className="pr"><LockedInput type="text" value={inv.Bank?.Iban ?? ""} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Iban" required={true} /></td></tr>
        <tr><td className="text">BIC</td><td className="pr"><LockedInput type="text" value={inv.Bank?.Bic ?? ""} onChange={this.handleChange.bind(this)} locked={true} data-key="Bank.Bic" required={true} /></td></tr>
      </tbody></table>
      <small><i className="fas fa-info"></i> Edit these from your settings file.</small>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div><InvoiceMail parent={this} onHide={this.email.bind(this)} hide={this.state.State.email} /></form>;
	}
}
