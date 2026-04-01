import * as React from "react";
import Axios from "axios";
import { formatDate, daysFromNow } from "../../utils/date";
import { Autocomplete, LockedInput } from "../../shared/components";
import { ActionButton } from "../../shared/ActionButton";
import { InvoiceLineEdit } from "./edit-line";
import { InvoiceMail } from "./edit-mail";
import Big from "big.js";
import type { Invoice, InvoiceLine, InvoiceTotal } from "../../types/model";

interface InvoiceEditProps {
  entity: string;
  year: string;
  id?: string;
  bucket?: string;
}

interface InvoiceEditState extends Invoice {
  State: {
    email: boolean;
    currentBucket: string;
  };
}

export default class InvoiceEdit extends React.Component<InvoiceEditProps, InvoiceEditState> {
  private undoStack: InvoiceLine[][];
  private redoStack: InvoiceLine[][];
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
        Street2: "",
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
        Issuedate: "",
        Ponumber: "",
        Duedate: daysFromNow(14),
        Paydate: "",
        Freefield: "",
        HourFile: "",
      },
      Lines: [
        {
          Description: "",
          Quantity: "0.00",
          Price: "0.00",
          Total: "0.00",
        },
      ],
      Notes: "",
      Total: {
        Ex: "0.00",
        Tax: "0.00",
        Total: "0.00",
      },
      Bank: {
        Vat: "",
        Coc: "",
        Iban: "",
        Bic: "",
      },
      State: {
        email: false,
        currentBucket: props.bucket ?? "concepts",
      },
      Mail: {
        From: "",
        Subject: "",
        To: "",
        Body: "",
      },
    };
  }

  componentDidMount(): void {
    const params = this.props;
    if (params.id != null && params.id !== "" && params.bucket != null && params.bucket !== "") {
      this.ajax(params.bucket, params.id);
    } else {
      this.ajaxDefaults(params.entity);
    }
  }

  private ajaxDefaults(entity: string): void {
    interface EntityDetails {
      Entity: { Name: string; VAT: string; COC: string; IBAN: string; BIC: string };
      User: { Name: string; Address1: string; Address2: string };
    }
    Axios.get<EntityDetails>(`/api/v1/entities/${entity}/details`)
      .then((res) => {
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
            Bic: res.data.Entity.BIC,
          },
        });
      })
      .catch((err) => {
        handleErr(err);
      });
  }

  private ajax(bucket: string, name: string): void {
    Axios.get<Invoice>(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${name}`)
      .then((res) => {
        this.parseInput(res.data);
      })
      .catch((err) => {
        handleErr(err);
      });
  }

  private parseInput(data: Invoice, newbucket?: string): void {
    const meta = data.Meta;

    const bucket = newbucket ?? this.props.bucket ?? "concepts";
    const url = `#${this.props.entity}/${this.props.year}/invoices/edit/${bucket}/${meta.Conceptid}`;
    if (window.location.hash !== url) {
      // Update URL so refresh will keep the invoice open
      history.replaceState({}, "", url);
    }
    meta.Issuedate = formatDate(meta.Issuedate);
    meta.Duedate = formatDate(meta.Duedate);
    meta.Paydate = formatDate(meta.Paydate);

    // Update current bucket in state so it reflects the actual location after finalize/reset
    // Add local State to server response data
    this.setState({
      ...data,
      State: { ...this.state.State, currentBucket: bucket },
    });
  }

  private triggerChange(indices: string[], val: string): void {
    const meta = this.state.Meta;
    if (meta.Status === "FINAL") {
      console.log("Finalized, not allowing changes!");
      return;
    }
    // Dynamic nested object traversal requires any - this navigates paths like "Customer.Name"
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let node: any = this.state;
    for (let i = 0; i < indices.length - 1; i++) {
      const key = indices[i];
      if (key === undefined) return;
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
      node = node[key];
    }
    const lastKey = indices[indices.length - 1];
    if (lastKey === undefined) return;
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    node[lastKey] = val;

    // Any post-processing
    const lines = this.state.Lines;
    const lineIdx = indices[1];
    if (indices[0] === "Lines" && lineIdx !== undefined) {
      const idx = parseInt(lineIdx, 10);
      const line = lines[idx];
      if (line) {
        lines[idx] = this.lineUpdate(line);
      }
      const newState = { ...this.state, Lines: lines, Total: this.totalUpdate(lines) };
      this.setState(newState);
    } else {
      this.setState({ ...this.state });
    }
  }

  private defaultDecimal(val: string, isNeg: boolean): string {
    if (val === "") {
      return "0.00";
    }
    val = val.replace(/,/g, ".");
    if (isNeg) {
      val = val.replace(/[^\d.-]/g, "");
    } else {
      val = val.replace(/[^\d.]/g, "");
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

  private lineUpdate(line: InvoiceLine): InvoiceLine {
    line.Quantity = this.defaultDecimal(line.Quantity, false);
    line.Price = this.defaultDecimal(line.Price, true);

    line.Total = new Big(line.Price).times(line.Quantity).round(2).toFixed(2).toString();
    return line;
  }

  private totalUpdate(lines: InvoiceLine[]): InvoiceTotal {
    let ex = new Big(0);
    lines.forEach(function (val: InvoiceLine) {
      console.log("Add", val.Total);
      ex = ex.plus(val.Total);
    });

    let tax = new Big(0);
    const customer = this.state.Customer;
    if (customer.Tax === "NL21") {
      tax = ex.div("100").times("21");
    }
    const total = ex.plus(tax);
    console.log("totals (ex,tax,total)", ex.toString(), tax.toString(), total.toString());

    return {
      Ex: ex.round(2).toFixed(2).toString(),
      Tax: tax.round(2).toFixed(2).toString(),
      Total: total.round(2).toFixed(2).toString(),
    };
  }

  handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>): void {
    const key = e.target.dataset["key"];
    if (key == null) return;
    console.log("handleChange", key);
    const indices = key.split(".");
    this.triggerChange(indices, e.target.value);
  }

  private async save(): Promise<void> {
    const req = structuredClone(this.state);
    console.log(req);

    const res = await Axios.post<Invoice>("/api/v1/invoice/" + this.props.entity + "/" + this.props.year, req);
    this.parseInput.call(this, res.data);
  }

  private async reset(): Promise<void> {
    const meta = this.state.Meta;
    const res = await Axios.post<Invoice>(
      `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${meta.Conceptid}/reset`,
      {}
    );
    // Reset always moves to concepts bucket
    this.parseInput(res.data, "concepts");
  }

  private async finalize(): Promise<void> {
    const meta = this.state.Meta;
    const res = await Axios.post<Invoice>(
      `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${meta.Conceptid}/finalize`,
      {}
    );
    // Axios headers type returns any for custom headers - validated via typeof
    const bucketHeader: unknown = res.headers["x-bucket-change"];
    const bucket = typeof bucketHeader === "string" ? bucketHeader : undefined;
    this.parseInput(res.data, bucket);
  }

  private pdf(): void {
    const meta = this.state.Meta;
    if (meta.Status !== "FINAL") {
      console.log("PDF only available in finalized invoices");
      return;
    }
    const url = `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/${this.props.id}/pdf`;
    console.log(`Open PDF ${url}`);
    location.assign(url);
  }

  private selectCustomer(data: {
    Name: string;
    Street1?: string;
    Street2?: string;
    VAT?: string;
    COC?: string;
    NoteAdd?: string;
    BillingEmail?: string[];
  }): void {
    console.log("Select customer", data);
    this.setState({
      Customer: {
        Name: data.Name,
        Street1: data.Street1 ?? "",
        Street2: data.Street2 ?? "",
        Vat: data.VAT ?? "",
        Coc: data.COC ?? "",
        Tax: "NL21",
      },
      Notes: data.NoteAdd ?? "",
      Mail: {
        ...this.state.Mail,
        From: this.state.Mail.From,
        Subject: this.state.Mail.Subject,
        Body: this.state.Mail.Body,
        To: (data.BillingEmail ?? []).join(", "),
      },
    });
  }

  private email(): void {
    this.setState({ State: { ...this.state.State, email: !this.state.State.email } });
  }

  public pushUndo(): void {
    // Save current lines state for undo
    const linesCopy = structuredClone(this.state.Lines);
    this.undoStack.push(linesCopy);
    this.redoStack = []; // Clear redo stack on new change
  }

  private undo(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.undoStack.length === 0) return;
    const previousLines = this.undoStack.pop();
    if (previousLines == null) return;

    // Save current state to redo stack
    const currentLines = structuredClone(this.state.Lines);
    this.redoStack.push(currentLines);

    // Restore previous state
    this.setState({ Lines: previousLines, Total: this.totalUpdate(previousLines) });
  }

  private redo(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    if (this.redoStack.length === 0) return;
    const nextLines = this.redoStack.pop();
    if (nextLines == null) return;

    // Save current state to undo stack
    const currentLines = structuredClone(this.state.Lines);
    this.undoStack.push(currentLines);

    // Restore next state
    this.setState({ Lines: nextLines, Total: this.totalUpdate(nextLines) });
  }

  render(): React.JSX.Element {
    const inv = this.state;
    const that = this;
    const meta = inv.Meta;
    const entity = inv.Entity;
    const customer = inv.Customer;

    return (
      <form>
        <div>
          <div className="card">
            <div className="card-header">
              <div className="d-flex flex-wrap gap-1 justify-content-between align-items-center">
                <span>New Invoice</span>
                <div className="btn-toolbar flex-wrap gap-1">
                  <div className="btn-group btn-group-sm">
                    <button
                      type="button"
                      className="btn btn-warning"
                      disabled={this.undoStack.length === 0 || meta.Status === "FINAL"}
                      onClick={this.undo.bind(this)}
                    >
                      <i className="fas fa-rotate-left"></i>
                      <span className="d-none d-md-inline"> Undo</span>
                    </button>
                    <button
                      type="button"
                      className="btn btn-warning"
                      disabled={this.redoStack.length === 0 || meta.Status === "FINAL"}
                      onClick={this.redo.bind(this)}
                    >
                      <i className="fas fa-rotate-right"></i>
                      <span className="d-none d-md-inline"> Redo</span>
                    </button>
                  </div>
                  <div className="btn-group btn-group-sm">
                    <ActionButton
                      type="button"
                      className="btn btn-success"
                      disabled={meta.Status === "FINAL"}
                      onClick={this.save.bind(this)}
                    >
                      <i className="fas fa-floppy-disk"></i>
                      <span className="d-none d-md-inline"> Save</span>
                    </ActionButton>
                    <ActionButton
                      type="button"
                      className="btn btn-danger"
                      disabled={meta.Status !== "CONCEPT"}
                      onClick={this.finalize.bind(this)}
                    >
                      <i className="fas fa-lock"></i>
                      <span className="d-none d-md-inline"> Finalize</span>
                    </ActionButton>
                  </div>
                  <div className="btn-group btn-group-sm">
                    <button
                      type="button"
                      className={"btn btn-success" + (meta.Status !== "FINAL" ? " disabled" : "")}
                      onClick={this.pdf.bind(this)}
                    >
                      <i className="far fa-file-pdf"></i>
                      <span className="d-none d-md-inline"> PDF</span>
                    </button>
                    <button
                      type="button"
                      className={"btn btn-success" + (meta.Status !== "FINAL" ? " disabled" : "")}
                      onClick={this.email.bind(this)}
                    >
                      <i className="fas fa-paper-plane"></i>
                      <span className="d-none d-md-inline"> E-mail</span>
                    </button>
                    <ActionButton
                      type="button"
                      className="btn btn-danger"
                      disabled={meta.Status !== "FINAL"}
                      onClick={this.reset.bind(this)}
                    >
                      <i className="fas fa-unlock"></i>
                      <span className="d-none d-md-inline"> Reset</span>
                    </ActionButton>
                  </div>
                </div>
              </div>
            </div>
            <div className="card-body">
              <div className={"invoice group " + (meta.Status === "FINAL" ? "o50" : "")}>
                <div className="row g-3 mb-3">
                  <div className="col-12 col-md-4">
                    <label className="form-label text-muted small">Company</label>
                    <input
                      className="form-control"
                      type="text"
                      data-key="Company"
                      onChange={that.handleChange.bind(this)}
                      value={inv.Company}
                    />
                  </div>
                  <div className="col-12 col-md-4 offset-md-4">
                    <label className="form-label text-muted small">From</label>
                    <input
                      className="form-control"
                      type="text"
                      data-key="Entity.Name"
                      onChange={that.handleChange.bind(this)}
                      value={entity.Name}
                    />
                    <input
                      className="form-control mt-1"
                      type="text"
                      data-key="Entity.Street1"
                      onChange={that.handleChange.bind(this)}
                      value={entity.Street1}
                    />
                    <input
                      className="form-control mt-1"
                      type="text"
                      data-key="Entity.Street2"
                      onChange={that.handleChange.bind(this)}
                      value={entity.Street2}
                    />
                  </div>
                </div>

                <div className="row g-3 mb-3">
                  <div className="col-12 col-md-4">
                    <label className="form-label text-muted small">Invoice For</label>
                    <Autocomplete
                      data-key="Customer.Name"
                      onSelect={that.selectCustomer.bind(that)}
                      onChange={that.handleChange.bind(that)}
                      required={true}
                      placeholder="Company Name"
                      url={"/api/v1/debtors/" + that.props.entity + "/search"}
                      value={customer.Name}
                    />
                    <div className="pr mt-1">
                      <input
                        className="form-control"
                        type="text"
                        data-key="Customer.Street1"
                        onChange={that.handleChange.bind(this)}
                        value={customer.Street1}
                        placeholder="Street1"
                      />
                      <i className="fas fa-asterisk text-danger fa-input"></i>
                    </div>
                    <div className="pr mt-1">
                      <input
                        className="form-control"
                        type="text"
                        data-key="Customer.Street2"
                        onChange={that.handleChange.bind(this)}
                        value={customer.Street2}
                        placeholder="Street2"
                      />
                      <i className="fas fa-asterisk text-danger fa-input"></i>
                    </div>

                    <input
                      className="form-control mt-1"
                      type="text"
                      data-key="Customer.Vat"
                      onChange={that.handleChange.bind(this)}
                      value={customer.Vat}
                      placeholder="VAT-number"
                    />
                    <input
                      className="form-control mt-1"
                      type="text"
                      data-key="Customer.Coc"
                      onChange={that.handleChange.bind(this)}
                      value={customer.Coc}
                      placeholder="Chamber Of Commerce (CoC)"
                    />
                    <select
                      className="form-control mt-1"
                      data-key="Customer.Tax"
                      onChange={that.handleChange.bind(this)}
                      value={customer.Tax}
                    >
                      <option value="NL21">NL21 (Domestic)</option>
                      <option value="EU0">EU (ICP)</option>
                      <option value="WORLD0">Outside EU (export)</option>
                    </select>
                  </div>
                  <div className="col-12 col-md-5 offset-md-3">
                    <label className="form-label text-muted small">Invoice Details</label>
                    <table className="table table-sm">
                      <tbody>
                        <tr>
                          <td className="text">Invoice ID</td>
                          <td>
                            <LockedInput
                              type="text"
                              value={meta.Invoiceid}
                              placeholder="AUTOGENERATED"
                              onChange={that.handleChange.bind(that)}
                              locked={true}
                              data-key="Meta.Invoiceid"
                            />
                          </td>
                        </tr>
                        <tr>
                          <td className="text">Issue Date</td>
                          <td>
                            <LockedInput
                              type="date"
                              value={meta.Issuedate}
                              placeholder="AUTOGENERATED"
                              onChange={that.handleChange.bind(that)}
                              locked={true}
                              data-key="Meta.Issuedate"
                            />
                          </td>
                        </tr>
                        <tr>
                          <td className="text">PO Number</td>
                          <td>
                            <input
                              className="form-control"
                              type="text"
                              data-key="Meta.Ponumber"
                              onChange={that.handleChange.bind(that)}
                              value={meta.Ponumber}
                            />
                          </td>
                        </tr>
                        <tr>
                          <td className="text">Due Date</td>
                          <td>
                            <input
                              type="date"
                              value={meta.Duedate}
                              onChange={that.handleChange.bind(that)}
                              className="form-control"
                              data-key="Meta.Duedate"
                            />
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>

                <InvoiceLineEdit parent={this} />

                <div className="row g-3 mb-3">
                  <div className="col-12">
                    <label className="form-label text-muted small">Notes</label>
                    <small className="d-block text-muted mb-1">
                      <i className="fas fa-info"></i> This text is added to the invoice
                    </small>
                    <textarea
                      className="form-control"
                      data-key="Notes"
                      onChange={this.handleChange.bind(this)}
                      value={inv.Notes}
                    />
                  </div>
                </div>
                <div className="row g-3">
                  <div className="col-12 col-md-6 col-lg-4">
                    <label className="form-label text-muted small">Banking details</label>
                    <table className="table table-sm mb0">
                      <tbody>
                        <tr>
                          <td className="text">VAT</td>
                          <td className="pr">
                            <LockedInput
                              type="text"
                              value={inv.Bank.Vat}
                              onChange={this.handleChange.bind(this)}
                              locked={true}
                              data-key="Bank.Vat"
                              required={true}
                            />
                          </td>
                        </tr>
                        <tr>
                          <td className="text">CoC</td>
                          <td className="pr">
                            <LockedInput
                              type="text"
                              value={inv.Bank.Coc}
                              onChange={this.handleChange.bind(this)}
                              locked={true}
                              data-key="Bank.Coc"
                              required={true}
                            />
                          </td>
                        </tr>
                        <tr>
                          <td className="text">IBAN</td>
                          <td className="pr">
                            <LockedInput
                              type="text"
                              value={inv.Bank.Iban}
                              onChange={this.handleChange.bind(this)}
                              locked={true}
                              data-key="Bank.Iban"
                              required={true}
                            />
                          </td>
                        </tr>
                        <tr>
                          <td className="text">BIC</td>
                          <td className="pr">
                            <LockedInput
                              type="text"
                              value={inv.Bank.Bic}
                              onChange={this.handleChange.bind(this)}
                              locked={true}
                              data-key="Bank.Bic"
                              required={true}
                            />
                          </td>
                        </tr>
                      </tbody>
                    </table>
                    <small>
                      <i className="fas fa-info"></i> Edit these from your settings file.
                    </small>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <InvoiceMail parent={this} onHide={this.email.bind(this)} hide={this.state.State.email} />
      </form>
    );
  }
}
