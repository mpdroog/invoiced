import * as React from "react";
import Axios from "axios";
import {IInvoiceState} from "./edit-struct";
import {DOM} from "../../lib/dom";

interface IInvoicePagination {
  from?: number;
  count?: number;
}
interface IInvoiceListState {
  pagination?: IInvoicePagination;
  invoices?: IInvoiceState[];
  isBalance: boolean;
  sortField: string;
  sortAsc: boolean;
}

interface IInvoiceListProps {
  bucket: string;
  title: string;
  entity: string;
  year: string;
  items: Record<string, IInvoiceState[]>;
}

export default class Invoices extends React.Component<IInvoiceListProps, IInvoiceListState> {
  constructor(props: IInvoiceListProps) {
      super(props);
      this.state = {
        pagination: {
          from: 0,
          count: 50
        },
        invoices: [],
        isBalance: false,
        sortField: "Invoice",
        sortAsc: true,
      };
  }

  private toggleSort(field: string): void {
    if (this.state.sortField === field) {
      this.setState({sortAsc: !this.state.sortAsc});
    } else {
      this.setState({sortField: field, sortAsc: true});
    }
  }

  private getSortValue(inv: IInvoiceState, field: string): string | number {
    switch (field) {
      case "Invoice": return inv.Meta?.Invoiceid || inv.Meta?.Conceptid || "";
      case "Customer": return inv.Customer?.Name || "";
      case "Amount": return parseFloat(inv.Total?.Total ?? "0") || 0;
      case "Duedate": return inv.Meta?.Duedate || "";
      default: return "";
    }
  }

  private sortInvoices(invoices: {key: string, inv: IInvoiceState, bucket: string}[]): {key: string, inv: IInvoiceState, bucket: string}[] {
    const field = this.state.sortField;
    const asc = this.state.sortAsc;

    return invoices.sort((a, b) => {
      let valA = this.getSortValue(a.inv, field);
      let valB = this.getSortValue(b.inv, field);

      if (typeof valA === "number" && typeof valB === "number") {
        return asc ? valA - valB : valB - valA;
      }

      valA = String(valA).toLowerCase();
      valB = String(valB).toLowerCase();
      if (valA < valB) return asc ? -1 : 1;
      if (valA > valB) return asc ? 1 : -1;
      return 0;
    });
  }

  private delete(e: React.MouseEvent<HTMLAnchorElement>): void {
    e.preventDefault();
    const node = DOM.eventFilter(e, "A");
    if (!node) return;
    const id = node.dataset["target"];
    const bucket = node.dataset["bucket"];
    const isDisabled = node.dataset["disabled"] === "true";

    if (isDisabled) {
      console.log("btn disabled");
      return;
    }

    Axios.delete(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${id}`)
    .then(() => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private setPaid(e: React.MouseEvent<HTMLAnchorElement>): void {
    e.preventDefault();
    const node = DOM.eventFilter(e, "A");
    if (!node) return;
    const id = node.dataset["id"];
    const bucket = node.dataset["bucket"];
    const isDisabled = node.dataset["disabled"] === "true";

    if (isDisabled) {
      console.log("btn disabled");
      return;
    }

    Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${id}/paid`, {})
    .then(() => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private conceptLine(key: string, inv: IInvoiceState, bucket: string, isPending: boolean): React.JSX.Element {
    const today = new Date().toISOString().split('T')[0] ?? "";
    let expiryClass = "";
    if (isPending && inv.Meta?.Duedate && inv.Meta.Duedate <= today) {
      expiryClass = 'bg-danger';
    }
    const meta = inv.Meta || { Status: "NEW", Invoiceid: "", Conceptid: "", Duedate: "", Ponumber: "", HourFile: "" };
    const customer = inv.Customer || { Name: "", Street1: "", Street2: "", Vat: "", Coc: "", Tax: "NL21" };
    const total = inv.Total || { Ex: "0.00", Tax: "0.00", Total: "0.00" };
    const isDeleteDisabled = meta.Status === 'FINAL' || meta.Invoiceid.length > 0;
    const isPaidDisabled = bucket === 'concepts';
    return <tr key={key}>
      <td>{key}</td>
      <td>{meta.Invoiceid}</td>
      <td>{customer.Name}</td>
      <td>&euro; {total.Total}</td>
      <td className={expiryClass}>{meta.Duedate}</td>
      <td>
        <a className="btn btn-default btn-hover-primary" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/edit/"+bucket+"/"+key}><i className="fas fa-pencil"></i></a>
        <a className={"btn btn-default btn-hover-danger " + (isDeleteDisabled ? " disabled" : "")} data-target={key} data-status={meta.Status} data-disabled={String(isDeleteDisabled)} onClick={this.delete.bind(this)}><i className="fas fa-trash"></i></a>
        <a className={"btn btn-default btn-hover-danger " + (isPaidDisabled ? " disabled" : "")} data-id={key} data-bucket={bucket} data-disabled={String(isPaidDisabled)} onClick={this.setPaid.bind(this)}><i className="fas fa-check"></i></a>
      </td>
    </tr>;
  }

  private finishedLine(key: string, inv: IInvoiceState, bucket: string): React.JSX.Element {
    console.log(bucket);
    return <tr key={key}>
      <td>{key}</td>
      <td>{inv.Meta?.Invoiceid}</td>
      <td>{inv.Customer?.Name}</td>
      <td>&euro; {inv.Total?.Total}</td>
      <td>{inv.Meta?.Duedate}</td>
      <td>
        <a className="btn btn-default btn-hover-primary" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/edit/"+bucket+"/"+key}><i className="fas fa-pencil"></i></a>
        <a className="btn btn-default btn-hover-primary" href={"/api/v1/invoice/" + this.props.entity + "/" + this.props.year + "/" + bucket + "/" + key + "/pdf"}><i className="far fa-file-pdf"></i></a>
      </td>
    </tr>;
  }

  private openUpload(e: React.MouseEvent<HTMLAnchorElement>): void {
    e.preventDefault();
    document.getElementById('js-balance-field')?.click();
  }
  private uploadBalance(e: React.ChangeEvent<HTMLInputElement>): void {
    if (!e.target.files || e.target.files.length === 0) {
      return;
    }

    console.log("Upload");
    const file = e.target.files[0];
    if (!file) return;
    const form = new FormData();
    form.append('file', file, file.name);

    // api/v1/invoice/:entity/:year/:bucket/:id/balance
    Axios.post('/api/v1/invoice-balance/'+this.props.entity+'/'+this.props.year,
      form, {headers: {'Content-Type': 'multipart/form-data' }})
    .then(() => {
      // TODO: Report something to UI?
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private sortHeader(field: string): React.JSX.Element {
    let icon = "";
    if (this.state.sortField === field) {
      icon = this.state.sortAsc ? " ▲" : " ▼";
    }
    return <th style={{cursor: "pointer"}} onClick={() => this.toggleSort(field)}>{field}{icon}</th>;
  }

	render(): React.JSX.Element {
    const res:React.JSX.Element[] = [];
    let invoiceList: {key: string, inv: IInvoiceState, bucket: string}[] = [];
    const isPending = this.props.bucket === "sales-invoices-unpaid";

    if (this.props.items) {
      for (const dir in this.props.items) {
        if (!Object.prototype.hasOwnProperty.call(this.props.items, dir)) {
          continue;
        }
        // Extract quarter from directory path (3rd element from end): .../Q1/sales-invoices-paid/ -> Q1
        const parts = dir.split("/").filter(p => p.length > 0);
        const bucket = this.props.bucket === "concepts" ? "concepts" : (parts[parts.length - 2] ?? "");

        const items = this.props.items[dir];
        if (!items) continue;
        items.forEach((inv) => {
          const key: string = inv.Meta?.Conceptid ?? "";
          invoiceList.push({key, inv, bucket});
        });
      }
    }

    // Sort the invoices
    invoiceList = this.sortInvoices(invoiceList);

    // Render sorted invoices
    invoiceList.forEach(({key, inv, bucket}) => {
      if (this.props.bucket === "concepts" || isPending) {
        res.push(this.conceptLine(key, inv, bucket, isPending));
      } else {
        res.push(this.finishedLine(key, inv, bucket));
      }
    });

    if (res.length === 0) {
      res.push(<tr key="empty"><td colSpan={6}>No invoices yet :)</td></tr>);
    }

    let headerButtons = <div/>;
    if (this.props.bucket === "concepts") {
      headerButtons = <div>
        <a id="js-new" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/add"} className="btn btn-default btn-hover-primary showhide">
          <i className="fas fa-plus"></i> New
        </a>
      </div>;
    }
    if (this.props.bucket === "sales-invoices-unpaid") {
      headerButtons = <div><a id="js-balance" onClick={this.openUpload.bind(this)} className="btn btn-default btn-hover-primary showhide">
          <i className="fas fa-upload"></i> Bankbalance
        </a>
      </div>;
    }

    const url = `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/balance`;
    const balanceUpload = <div>
      <form className="form-inline hidden" method="post" encType="multipart/form-data" action={url}>
        <input id="js-balance-field" accept=".xml" className="form-control" name="file" type="file" onChange={this.uploadBalance.bind(this)} />
        <button className="btn btn-default btn-hover-primary" type="submit"><i className="fas fa-arrow-up"></i> Upload</button>
      </form>
    </div>;

		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                {headerButtons}
              </div>
            </div>
            {this.props.title}
          </div>
          <div className="panel-body">
            {balanceUpload}
            <table className="table table-striped">
            	<thead><tr><th>#</th>{this.sortHeader("Invoice")}{this.sortHeader("Customer")}{this.sortHeader("Amount")}{this.sortHeader("Duedate")}<th>I/O</th></tr></thead>
            	<tbody>{res}</tbody>
            </table>
	        </div>
		    </div>
    </div>;
	}
}
