import * as React from "react";
import Axios from "axios";
import type {IInvoiceState} from "./edit-struct";
import {DOM} from "../../lib/dom";
import {ActionLink} from "../../shared/ActionButton";

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
  selectedKeys: Set<string>;
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
        selectedKeys: new Set<string>(),
      };
  }

  private toggleSort(field: string): void {
    if (this.state.sortField === field) {
      this.setState({sortAsc: !this.state.sortAsc});
    } else {
      this.setState({sortField: field, sortAsc: true});
    }
  }

  private toggleSelect(key: string): void {
    const selected = new Set(this.state.selectedKeys);
    if (selected.has(key)) {
      selected.delete(key);
    } else {
      selected.add(key);
    }
    this.setState({selectedKeys: selected});
  }

  private toggleSelectAll(keys: string[]): void {
    const allSelected = keys.every(k => this.state.selectedKeys.has(k));
    if (allSelected) {
      this.setState({selectedKeys: new Set<string>()});
    } else {
      this.setState({selectedKeys: new Set(keys)});
    }
  }

  private clearSelection(): void {
    this.setState({selectedKeys: new Set<string>()});
  }

  private getSelectedSum(invoiceList: {key: string, inv: IInvoiceState}[]): number {
    let sum = 0;
    for (const {key, inv} of invoiceList) {
      if (this.state.selectedKeys.has(key)) {
        const val = parseFloat(inv.Total?.Total ?? "0");
        if (!Number.isNaN(val)) {
          sum += val;
        }
      }
    }
    return sum;
  }

  private getSortValue(inv: IInvoiceState, field: string): string | number {
    switch (field) {
      case "Invoice": return inv.Meta?.Invoiceid ?? inv.Meta?.Conceptid ?? "";
      case "Customer": return inv.Customer?.Name ?? "";
      case "Amount": {
        const val = parseFloat(inv.Total?.Total ?? "0");
        return Number.isNaN(val) ? 0 : val;
      }
      case "Duedate": return inv.Meta?.Duedate ?? "";
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

  private async delete(e: React.MouseEvent<HTMLAnchorElement>): Promise<void> {
    const node = DOM.eventFilter(e, "A");
    if (!node) return;
    const id = node.dataset["target"];
    const bucket = node.dataset["bucket"];

    await Axios.delete(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${id}`);
    location.reload();
  }

  private async setPaid(e: React.MouseEvent<HTMLAnchorElement>): Promise<void> {
    const node = DOM.eventFilter(e, "A");
    if (!node) return;
    const id = node.dataset["id"];
    const bucket = node.dataset["bucket"];

    await Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${id}/paid`, {});
    location.reload();
  }

  private conceptLine(key: string, inv: IInvoiceState, bucket: string, isPending: boolean): React.JSX.Element {
    const today = new Date().toISOString().split('T')[0] ?? "";
    let expiryClass = "";
    if (isPending && inv.Meta?.Duedate != null && inv.Meta.Duedate !== '' && inv.Meta.Duedate <= today) {
      expiryClass = 'bg-danger';
    }
    const meta = inv.Meta ?? { Status: "NEW", Invoiceid: "", Conceptid: "", Duedate: "", Ponumber: "", HourFile: "" };
    const customer = inv.Customer ?? { Name: "", Street1: "", Street2: "", Vat: "", Coc: "", Tax: "NL21" };
    const total = inv.Total ?? { Ex: "0.00", Tax: "0.00", Total: "0.00" };
    const isDeleteDisabled = meta.Status === 'FINAL' || meta.Invoiceid.length > 0;
    const isPaidDisabled = bucket === 'concepts';
    const isSelected = this.state.selectedKeys.has(key);
    return <tr key={key} className={isSelected ? "table-active" : ""}>
      <td><input type="checkbox" className="form-check-input" checked={isSelected} onChange={() => this.toggleSelect(key)} /></td>
      <td className="d-none d-md-table-cell">{key}</td>
      <td>{meta.Invoiceid}</td>
      <td><span className="d-inline-block text-truncate-sm" title={customer.Name}>{customer.Name}</span></td>
      <td className="text-end text-nowrap">&euro; {total.Total}</td>
      <td className={expiryClass}>{meta.Duedate}</td>
      <td className="text-end">
        <div className="btn-group">
          <a className="btn btn-primary" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/edit/"+bucket+"/"+key}><i className="fas fa-pencil"></i></a>
          <ActionLink className="btn btn-danger" disabled={isDeleteDisabled} data-target={key} data-bucket={bucket} onClick={this.delete.bind(this)}><i className="fas fa-trash"></i></ActionLink>
          <ActionLink className="btn btn-success" disabled={isPaidDisabled} data-id={key} data-bucket={bucket} onClick={this.setPaid.bind(this)}><i className="fas fa-check"></i></ActionLink>
        </div>
      </td>
    </tr>;
  }

  private finishedLine(key: string, inv: IInvoiceState, bucket: string): React.JSX.Element {
    const isSelected = this.state.selectedKeys.has(key);
    return <tr key={key} className={isSelected ? "table-active" : ""}>
      <td><input type="checkbox" className="form-check-input" checked={isSelected} onChange={() => this.toggleSelect(key)} /></td>
      <td className="d-none d-md-table-cell">{key}</td>
      <td>{inv.Meta?.Invoiceid}</td>
      <td><span className="d-inline-block text-truncate-sm" title={inv.Customer?.Name}>{inv.Customer?.Name}</span></td>
      <td className="text-end text-nowrap">&euro; {inv.Total?.Total}</td>
      <td>{inv.Meta?.Duedate}</td>
      <td className="text-end">
        <div className="btn-group">
          <a className="btn btn-primary" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/edit/"+bucket+"/"+key}><i className="fas fa-pencil"></i></a>
          <a className="btn btn-primary" href={"/api/v1/invoice/" + this.props.entity + "/" + this.props.year + "/" + bucket + "/" + key + "/pdf"}><i className="far fa-file-pdf"></i></a>
        </div>
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

  private sortHeader(field: string, extraClass?: string): React.JSX.Element {
    let icon = "";
    if (this.state.sortField === field) {
      icon = this.state.sortAsc ? " ▲" : " ▼";
    }
    const className = extraClass != null && extraClass !== "" ? `sortable ${extraClass}` : "sortable";
    return <th className={className} onClick={() => this.toggleSort(field)}>{field}{icon}</th>;
  }

	render(): React.JSX.Element {
    const res:React.JSX.Element[] = [];
    let invoiceList: {key: string, inv: IInvoiceState, bucket: string}[] = [];
    const isPending = this.props.bucket === "sales-invoices-unpaid";

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
      res.push(<tr key="empty"><td colSpan={7}>No invoices yet :)</td></tr>);
    }

    // Calculate selection info
    const allKeys = invoiceList.map(i => i.key);
    const selectedCount = this.state.selectedKeys.size;
    const selectedSum = this.getSelectedSum(invoiceList);
    const allSelected = allKeys.length > 0 && allKeys.every(k => this.state.selectedKeys.has(k));

    let headerButtons = <div/>;
    if (this.props.bucket === "concepts") {
      headerButtons = <div>
        <a id="js-new" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/add"} className="btn btn-primary showhide">
          <i className="fas fa-plus"></i> New
        </a>
      </div>;
    }
    if (this.props.bucket === "sales-invoices-unpaid") {
      headerButtons = <div><a id="js-balance" onClick={this.openUpload.bind(this)} className="btn btn-primary showhide">
          <i className="fas fa-upload"></i> Bankbalance
        </a>
      </div>;
    }

    const selectionInfo = selectedCount > 0 ? (
      <span className="ms-3 badge bg-success">
        {selectedCount} selected: &euro;{selectedSum.toFixed(2)}
        <button type="button" className="btn-close btn-close-white ms-2" style={{fontSize: '0.6em'}} onClick={() => this.clearSelection()} aria-label="Clear"></button>
      </span>
    ) : null;

    const url = `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/balance`;
    const balanceUpload = <form className="d-none" method="post" encType="multipart/form-data" action={url}>
      <input id="js-balance-field" accept=".xml" name="file" type="file" onChange={this.uploadBalance.bind(this)} />
    </form>;

		return <div className="mb-4">
		    <div className="card">
          <div className="card-header d-flex align-items-center position-sticky" style={{top: '56px', zIndex: 1020}}>
            <span>{this.props.title}</span>
            {selectionInfo}
            <div className="ms-auto">
              <div className="btn-group nm7">
                {headerButtons}
              </div>
            </div>
          </div>
          <div className="card-body">
            {balanceUpload}
            <div className="table-responsive">
            <table className="table table-striped">
            	<thead><tr>
                <th><input type="checkbox" className="form-check-input" checked={allSelected} onChange={() => this.toggleSelectAll(allKeys)} title="Select all" /></th>
                <th className="d-none d-md-table-cell">#</th>{this.sortHeader("Invoice")}{this.sortHeader("Customer")}{this.sortHeader("Amount", "text-end")}{this.sortHeader("Duedate")}<th className="text-end">Actions</th>
              </tr></thead>
            	<tbody>{res}</tbody>
            </table>
            </div>
	        </div>
		    </div>
    </div>;
	}
}
