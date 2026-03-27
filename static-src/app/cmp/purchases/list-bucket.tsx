import * as React from "react";
import Axios from "axios";

interface IPurchaseInvoice {
  ID: string
  Supplier: {
    Name: string
    VAT: string
  }
  Issuedate: string
  Duedate: string
  TotalEx: string
  TotalTax: string
  TotalInc: string
  Status: string
  Lines: IPurchaseLine[]
}

interface IPurchaseLine {
  Description: string
  Quantity: string
  Price: string
  Total: string
  TaxPercent: string
}

interface IState {
  sortField: string
  sortAsc: boolean
  showAddLine: boolean
  selectedInvoice: IPurchaseInvoice | null
}

interface IProps {
  bucket: string
  title: string
  items: any
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

  private toggleSort(field: string) {
    if (this.state.sortField === field) {
      this.setState({sortAsc: !this.state.sortAsc});
    } else {
      this.setState({sortField: field, sortAsc: true});
    }
  }

  private getSortValue(inv: IPurchaseInvoice, field: string): any {
    switch (field) {
      case "ID": return inv.ID || "";
      case "Supplier": return inv.Supplier?.Name || "";
      case "Amount": return parseFloat(inv.TotalInc) || 0;
      case "Duedate": return inv.Duedate || "";
      default: return "";
    }
  }

  private sortInvoices(invoices: {key: string, inv: IPurchaseInvoice, bucket: string}[]): {key: string, inv: IPurchaseInvoice, bucket: string}[] {
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

  private setPaid(e, key: string, bucket: string) {
    e.preventDefault();
    if (!confirm("Mark this purchase invoice as paid?")) {
      return;
    }

    Axios.post(`/api/v1/purchase/${this.props.entity}/${this.props.year}/${bucket}/${key}/paid`, {})
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private delete(e, key: string, bucket: string) {
    e.preventDefault();
    if (!confirm("Delete this purchase invoice?")) {
      return;
    }

    Axios.delete(`/api/v1/purchase/${this.props.entity}/${this.props.year}/${bucket}/${key}`)
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private showAddLineModal(inv: IPurchaseInvoice) {
    this.setState({showAddLine: true, selectedInvoice: inv});
  }

  private closeModal() {
    this.setState({showAddLine: false, selectedInvoice: null});
  }

  private sortHeader(field: string): React.JSX.Element {
    let icon = "";
    if (this.state.sortField === field) {
      icon = this.state.sortAsc ? " ▲" : " ▼";
    }
    return <th style={{cursor: "pointer"}} onClick={() => this.toggleSort(field)}>{field}{icon}</th>;
  }

  private openUpload(e) {
    e.preventDefault();
    document.getElementById('js-purchase-upload').click();
  }

  private uploadXML(e) {
    if (e.target.files.length === 0) {
      return;
    }

    let file = e.target.files[0];
    let form = new FormData();
    form.append('file', file, file.name);

    Axios.post('/api/v1/purchase/'+this.props.entity+'/'+this.props.year,
      form, {headers: {'Content-Type': 'multipart/form-data' }})
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  render() {
    let res: React.JSX.Element[] = [];
    let invoiceList: {key: string, inv: IPurchaseInvoice, bucket: string}[] = [];
    const isUnpaid = this.props.bucket === "purchase-invoices-unpaid";

    if (this.props.items) {
      for (let dir in this.props.items) {
        if (!this.props.items.hasOwnProperty(dir)) {
          continue;
        }
        let parts = dir.split("/").filter(p => p.length > 0);
        let bucket = parts[parts.length - 2]; // Get Q1, Q2, etc.

        this.props.items[dir].forEach((inv) => {
          // Use sanitized filename as key (from ID + Supplier.Name)
          let key = inv.ID ? (inv.Supplier?.Name?.toLowerCase().replace(/[^a-z0-9]/g, '-') + '-' + inv.ID.toLowerCase().replace(/[^a-z0-9]/g, '-')) : dir;
          invoiceList.push({key, inv, bucket});
        });
      }
    }

    invoiceList = this.sortInvoices(invoiceList);

    const today = new Date().toISOString().split('T')[0];

    invoiceList.forEach(({key, inv, bucket}) => {
      const expiryClass = isUnpaid && inv.Duedate && inv.Duedate <= today ? 'bg-danger' : '';

      res.push(<tr key={key}>
        <td>{inv.ID}</td>
        <td>{inv.Supplier?.Name}</td>
        <td>&euro; {inv.TotalInc}</td>
        <td className={expiryClass}>{inv.Duedate}</td>
        <td>
          <a className="btn btn-default btn-hover-primary" href={`/api/v1/purchase/${this.props.entity}/${this.props.year}/${bucket}/${key}/pdf`} target="_blank">
            <i className="fa fa-file-pdf-o"></i>
          </a>
          {isUnpaid && <a className="btn btn-default btn-hover-success faa-parent animated-hover" onClick={(e) => this.setPaid(e, key, bucket)}>
            <i className="fa fa-check faa-flash"></i>
          </a>}
          <a className="btn btn-default btn-hover-info" onClick={() => this.showAddLineModal(inv)}>
            <i className="fa fa-plus"></i> Add to Invoice
          </a>
          <a className="btn btn-default btn-hover-danger faa-parent animated-hover" onClick={(e) => this.delete(e, key, bucket)}>
            <i className="fa fa-trash faa-flash"></i>
          </a>
        </td>
      </tr>);
    });

    if (res.length === 0) {
      res.push(<tr key="empty"><td colSpan={5}>No purchase invoices yet</td></tr>);
    }

    let headerButtons = <div/>;
    if (isUnpaid) {
      headerButtons = <div>
        <a id="js-upload-btn" onClick={this.openUpload.bind(this)} className="btn btn-default btn-hover-primary showhide">
          <i className="fa fa-upload"></i> Upload XML
        </a>
      </div>;
    }

    let uploadForm = <form className="hidden">
      <input id="js-purchase-upload" accept=".xml" type="file" onChange={this.uploadXML.bind(this)} />
    </form>;

    // Add line modal
    let modal = null;
    if (this.state.showAddLine && this.state.selectedInvoice) {
      const inv = this.state.selectedInvoice;
      modal = <AddLineModal invoice={inv} onClose={() => this.closeModal()} {...this.props} />;
    }

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
  invoice: IPurchaseInvoice
  entity: string
  year: string
  onClose: () => void
}

interface IAddLineModalState {
  concepts: any[]
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

  componentDidMount() {
    // Fetch concept invoices to add lines to
    Axios.get('/api/v1/invoices/'+this.props.entity+'/'+this.props.year, {params: {from: 0, count: 0}})
    .then(res => {
      let concepts = [];
      for (let key in res.data.Invoices) {
        if (key.endsWith("/concepts/sales-invoices/")) {
          res.data.Invoices[key].forEach(inv => {
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

  private addLine() {
    if (!this.state.selectedConcept) {
      alert("Please select an invoice");
      return;
    }

    const line = this.props.invoice.Lines[this.state.selectedLine];
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
    concept.Lines.forEach(l => {
      totalEx += parseFloat(l.Total) || 0;
    });
    const taxRate = 0.21; // Default 21% VAT
    const totalTax = totalEx * taxRate;
    concept.Total.Ex = totalEx.toFixed(2);
    concept.Total.Tax = totalTax.toFixed(2);
    concept.Total.Total = (totalEx + totalTax).toFixed(2);

    // Save updated concept
    Axios.post('/api/v1/invoice/'+this.props.entity+'/'+this.props.year, concept)
    .then(res => {
      alert("Line added successfully!");
      this.props.onClose();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  render() {
    const inv = this.props.invoice;
    const s = {display: "block"};

    return <div className="modal" style={s} tabIndex={-1} role="dialog">
      <div className="modal-dialog modal-lg">
        <div className="modal-content">
          <div className="modal-header">
            <button onClick={this.props.onClose} className="close" type="button">
              <span>&times;</span>
            </button>
            <h4 className="modal-title">
              <i className="fa fa-plus"></i> Add Line from Purchase Invoice {inv.ID}
            </h4>
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
                      <td>&euro; {line.Price}</td>
                      <td>&euro; {line.Total}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <div className="modal-footer">
            <button onClick={this.props.onClose} className="btn btn-default">Cancel</button>
            <button onClick={() => this.addLine()} className="btn btn-primary">Add Line</button>
          </div>
        </div>
      </div>
    </div>;
  }
}
