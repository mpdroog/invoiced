import * as React from "react";
import Axios from "axios";
import {ActionButton, ActionLink} from "../../shared/ActionButton";
import { openModal, closeModal } from "../../shared/Modal";
import type { PurchaseInvoice } from "../../types/purchase";
import type { Invoice } from "../../types/model";

interface IState {
  sortField: string
  sortAsc: boolean
  showAddLine: boolean
  selectedInvoice: PurchaseInvoice | null
}

interface IProps {
  bucket: string
  title: string
  items: Record<string, PurchaseInvoice[]>
  entity: string
  year: string
}

export default class PurchaseInvoices extends React.Component<IProps, IState> {
  constructor(props: IProps) {
    super(props);
    this.state = {
      sortField: "ID",
      sortAsc: true,
      showAddLine: false,
      selectedInvoice: null
    };
  }

  private toggleSort(field: string): void {
    if (this.state.sortField === field) {
      this.setState({sortAsc: !this.state.sortAsc});
    } else {
      this.setState({sortField: field, sortAsc: true});
    }
  }

  private getSortValue(inv: PurchaseInvoice, field: string): string | number {
    switch (field) {
      case "ID": return inv.ID !== '' ? inv.ID : "";
      case "Supplier": return inv.Supplier.Name !== '' ? inv.Supplier.Name : "";
      case "Amount": {
        const val = parseFloat(inv.TotalInc);
        return Number.isNaN(val) ? 0 : val;
      }
      case "Duedate": return inv.Duedate !== '' ? inv.Duedate : "";
      default: return "";
    }
  }

  private sortInvoices(invoices: {key: string, inv: PurchaseInvoice, bucket: string}[]): {key: string, inv: PurchaseInvoice, bucket: string}[] {
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

  private async setPaid(key: string, bucket: string): Promise<void> {
    if (!confirm("Mark this purchase invoice as paid?")) {
      return;
    }

    await Axios.post(`/api/v1/purchase/${this.props.entity}/${this.props.year}/${bucket}/${key}/paid`, {});
    location.reload();
  }

  private async delete(key: string, bucket: string): Promise<void> {
    if (!confirm("Delete this purchase invoice?")) {
      return;
    }

    await Axios.delete(`/api/v1/purchase/${this.props.entity}/${this.props.year}/${bucket}/${key}`);
    location.reload();
  }

  private showAddLineModal(inv: PurchaseInvoice): void {
    this.setState({showAddLine: true, selectedInvoice: inv});
  }

  private closeModal(): void {
    this.setState({showAddLine: false, selectedInvoice: null});
  }

  private sortHeader(field: string): React.JSX.Element {
    let icon = "";
    if (this.state.sortField === field) {
      icon = this.state.sortAsc ? " ▲" : " ▼";
    }
    return <th className="sortable" onClick={() => this.toggleSort(field)}>{field}{icon}</th>;
  }

  private openUpload(e: React.MouseEvent<HTMLAnchorElement>): void {
    e.preventDefault();
    document.getElementById('js-purchase-upload')?.click();
  }

  private uploadXML(e: React.ChangeEvent<HTMLInputElement>): void {
    if (!e.target.files || e.target.files.length === 0) {
      return;
    }

    const file = e.target.files[0];
    if (!file) return;
    const form = new FormData();
    form.append('file', file, file.name);

    Axios.post('/api/v1/purchase/'+this.props.entity+'/'+this.props.year,
      form, {headers: {'Content-Type': 'multipart/form-data' }})
    .then(() => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  render(): React.JSX.Element {
    const res: React.JSX.Element[] = [];
    let invoiceList: {key: string, inv: PurchaseInvoice, bucket: string}[] = [];
    const isUnpaid = this.props.bucket === "purchase-invoices-unpaid";

    for (const dir in this.props.items) {
      if (!Object.prototype.hasOwnProperty.call(this.props.items, dir)) {
        continue;
      }
      const parts = dir.split("/").filter(p => p.length > 0);
      const bucket = parts[parts.length - 2] ?? ""; // Get Q1, Q2, etc.

      const items = this.props.items[dir];
      if (items == null) continue;
      items.forEach((inv: PurchaseInvoice) => {
        // Use sanitized filename as key (from ID + Supplier.Name)
        const key = inv.ID !== '' ? (inv.Supplier.Name.toLowerCase().replace(/[^a-z0-9]/g, '-') + '-' + inv.ID.toLowerCase().replace(/[^a-z0-9]/g, '-')) : dir;
        invoiceList.push({key, inv, bucket});
      });
    }

    invoiceList = this.sortInvoices(invoiceList);

    const today = new Date().toISOString().split('T')[0] ?? "";

    invoiceList.forEach(({key, inv, bucket}) => {
      const expiryClass = isUnpaid && inv.Duedate !== '' && inv.Duedate <= today ? 'bg-danger' : '';

      res.push(<tr key={key}>
        <td>{inv.ID}</td>
        <td>{inv.Supplier.Name}</td>
        <td className="text-end text-nowrap">&euro; {inv.TotalInc}</td>
        <td className={expiryClass}>{inv.Duedate}</td>
        <td>
          <a className="btn btn-primary" href={`/api/v1/purchase/${this.props.entity}/${this.props.year}/${bucket}/${key}/pdf`} target="_blank" rel="noreferrer">
            <i className="far fa-file-pdf"></i>
          </a>
          {isUnpaid && <ActionLink className="btn btn-success" onClick={() => this.setPaid(key, bucket)}>
            <i className="fas fa-check"></i>
          </ActionLink>}
          <a className="btn btn-info" onClick={() => this.showAddLineModal(inv)}>
            <i className="fas fa-plus"></i> Add to Invoice
          </a>
          <ActionLink className="btn btn-danger" onClick={() => this.delete(key, bucket)}>
            <i className="fas fa-trash"></i>
          </ActionLink>
        </td>
      </tr>);
    });

    if (res.length === 0) {
      res.push(<tr key="empty"><td colSpan={5}>No purchase invoices yet</td></tr>);
    }

    let headerButtons = <div/>;
    if (isUnpaid) {
      headerButtons = <div>
        <a id="js-upload-btn" onClick={this.openUpload.bind(this)} className="btn btn-primary showhide">
          <i className="fas fa-upload"></i> Upload XML
        </a>
      </div>;
    }

    const uploadForm = <form className="d-none">
      <input id="js-purchase-upload" accept=".xml" type="file" onChange={this.uploadXML.bind(this)} />
    </form>;

    // Add line modal
    let modal = null;
    if (this.state.showAddLine && this.state.selectedInvoice) {
      const inv = this.state.selectedInvoice;
      modal = <AddLineModal invoice={inv} onClose={() => this.closeModal()} {...this.props} />;
    }

    return <div className="mb-4">
      <div className="card">
        <div className="card-header">
          <div className="float-end">
            <div className="btn-group nm7">
              {headerButtons}
            </div>
          </div>
          {this.props.title}
        </div>
        <div className="card-body">
          {uploadForm}
          <table className="table table-striped">
            <thead><tr>{this.sortHeader("ID")}{this.sortHeader("Supplier")}{this.sortHeader("Amount")}{this.sortHeader("Duedate")}<th>Actions</th></tr></thead>
            <tbody>{res}</tbody>
          </table>
        </div>
      </div>
      {modal}
    </div>;
  }
}

// Modal for adding line to existing invoice
interface IAddLineModalProps {
  invoice: PurchaseInvoice
  entity: string
  year: string
  onClose: () => void
}

interface IAddLineModalState {
  concepts: Invoice[]
  selectedConcept: string
  selectedLine: number
}

class AddLineModal extends React.Component<IAddLineModalProps, IAddLineModalState> {
  constructor(props: IAddLineModalProps) {
    super(props);
    this.state = {
      concepts: [],
      selectedConcept: "",
      selectedLine: 0
    };
  }

  componentDidMount(): void {
    openModal();
    // Fetch concept invoices to add lines to
    interface InvoicesResponse {
      Invoices: Record<string, Invoice[]>;
    }
    Axios.get<InvoicesResponse>('/api/v1/invoices/'+this.props.entity+'/'+this.props.year, {params: {from: 0, count: 0}})
    .then(res => {
      const concepts: Invoice[] = [];
      for (const key in res.data.Invoices) {
        if (key.endsWith("/concepts/sales-invoices/")) {
          res.data.Invoices[key]?.forEach((inv: Invoice) => {
            concepts.push(inv);
          });
        }
      }
      this.setState({concepts});
    })
    .catch(err => {
      console.error("Failed to load concepts", err);
    });
  }

  componentWillUnmount(): void {
    closeModal();
  }

  private async addLine(): Promise<void> {
    if (this.state.selectedConcept === '') {
      alert("Please select an invoice");
      return;
    }

    const line = this.props.invoice.Lines[this.state.selectedLine];
    if (!line) {
      alert("Line not found");
      return;
    }
    const concept = this.state.concepts.find(c => c.Meta.Conceptid === this.state.selectedConcept);

    if (!concept) {
      alert("Concept not found");
      return;
    }

    // Add new line to the concept invoice
    concept.Lines.push({
      Description: line.Description,
      Quantity: line.Quantity,
      Price: line.Price,
      Total: line.Total
    });

    // Recalculate totals
    let totalEx = 0;
    concept.Lines.forEach((l: { Total: string }) => {
      const val = parseFloat(l.Total);
      totalEx += Number.isNaN(val) ? 0 : val;
    });
    const taxRate = 0.21; // Default 21% VAT
    const totalTax = totalEx * taxRate;
    concept.Total.Ex = totalEx.toFixed(2);
    concept.Total.Tax = totalTax.toFixed(2);
    concept.Total.Total = (totalEx + totalTax).toFixed(2);

    // Save updated concept
    await Axios.post('/api/v1/invoice/'+this.props.entity+'/'+this.props.year, concept);
    alert("Line added successfully!");
    this.props.onClose();
  }

  render(): React.JSX.Element {
    const inv = this.props.invoice;

    return <div className="modal modal-show" tabIndex={-1} role="dialog">
      <div className="modal-dialog modal-lg">
        <div className="modal-content">
          <div className="modal-header">
            <h4 className="modal-title">
              <i className="fas fa-plus"></i> Add Line from Purchase Invoice {inv.ID}
            </h4>
            <button onClick={this.props.onClose} className="btn-close" type="button" aria-label="Close"></button>
          </div>
          <div className="modal-body">
            <div className="form-group">
              <label>Select Target Invoice (Concept)</label>
              <select className="form-control" value={this.state.selectedConcept} onChange={(e) => this.setState({selectedConcept: e.target.value})}>
                <option value="">-- Select Invoice --</option>
                {this.state.concepts.map(c => (
                  <option key={c.Meta.Conceptid} value={c.Meta.Conceptid}>
                    {c.Meta.Conceptid} - {c.Customer.Name}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label>Select Line to Add</label>
              <table className="table table-bordered">
                <thead>
                  <tr><th></th><th>Description</th><th>Qty</th><th>Price</th><th>Total</th></tr>
                </thead>
                <tbody>
                  {inv.Lines.map((line, idx) => (
                    <tr key={idx}>
                      <td>
                        <input type="radio" name="selectedLine" checked={this.state.selectedLine === idx} onChange={() => this.setState({selectedLine: idx})} />
                      </td>
                      <td>{line.Description}</td>
                      <td>{line.Quantity}</td>
                      <td className="text-nowrap">&euro; {line.Price}</td>
                      <td className="text-nowrap">&euro; {line.Total}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <div className="modal-footer">
            <button type="button" onClick={this.props.onClose} className="btn btn-secondary">Cancel</button>
            <ActionButton onClick={() => this.addLine()} className="btn btn-primary">Add Line</ActionButton>
          </div>
        </div>
      </div>
    </div>;
  }
}
