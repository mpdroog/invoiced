import * as React from "react";
import Axios from "axios";
import { ActionButton } from "../../shared/ActionButton";
import type { DebtorItem, Debtor } from "../../types/entities";

interface DebtorEditProps {
  entity: string;
  year: string;
  id?: string;
}

interface DebtorEditState {
  key: string;
  debtor: Debtor;
  isNew: boolean;
}

const emptyDebtor: Debtor = {
  Name: "",
  Street1: "",
  Street2: "",
  VAT: "",
  COC: "",
  TAX: "NL21",
  NoteAdd: "",
  BillingEmail: [],
  AccountingCode: "",
};

export default class DebtorEdit extends React.Component<DebtorEditProps, DebtorEditState> {
  constructor(props: DebtorEditProps) {
    super(props);
    this.state = {
      key: "",
      debtor: { ...emptyDebtor },
      isNew: props.id === undefined || props.id === "",
    };
  }

  componentDidMount(): void {
    if (this.props.id !== undefined && this.props.id !== "") {
      this.loadDebtor(this.props.id);
    }
  }

  private loadDebtor(key: string): void {
    Axios.get<DebtorItem>(`/api/v1/debtor/${this.props.entity}/${key}`)
      .then((res) => {
        this.setState({
          key: res.data.key,
          debtor: res.data.debtor,
          isNew: false,
        });
      })
      .catch((err) => {
        handleErr(err);
      });
  }

  private updateField(field: keyof Debtor, value: string): void {
    this.setState((prev) => ({
      debtor: { ...prev.debtor, [field]: value },
    }));
  }

  private updateBillingEmail(value: string): void {
    const emails = value
      .split(",")
      .map((e) => e.trim())
      .filter((e) => e.length > 0);
    this.setState((prev) => ({
      debtor: { ...prev.debtor, BillingEmail: emails },
    }));
  }

  private async save(): Promise<void> {
    const { key, debtor } = this.state;
    if (key === "") {
      alert("Key is required");
      return;
    }

    const payload: DebtorItem = { key, debtor };
    await Axios.post(`/api/v1/debtor/${this.props.entity}/${key}`, payload);

    location.href = `#${this.props.entity}/${this.props.year}/projects`;
  }

  render(): React.JSX.Element {
    const { key, debtor, isNew } = this.state;

    return (
      <div className="card">
        <div className="card-header">
          <div className="float-end">
            <ActionButton className="btn btn-success" onClick={this.save.bind(this)}>
              <i className="fas fa-floppy-disk"></i>&nbsp;Save
            </ActionButton>
          </div>
          {isNew ? "New Debtor" : `Edit Debtor: ${key}`}
        </div>
        <div className="card-body">
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Key (slug)</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={key}
                onChange={(e) => this.setState({ key: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, "") })}
                disabled={!isNew}
                placeholder="e.g., acme-corp"
              />
              <small className="text-muted">Lowercase letters, numbers, and hyphens only</small>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Company Name</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.Name}
                onChange={(e) => this.updateField("Name", e.target.value)}
                placeholder="ACME Corporation B.V."
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Street 1</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.Street1}
                onChange={(e) => this.updateField("Street1", e.target.value)}
                placeholder="123 Main Street"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Street 2</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.Street2}
                onChange={(e) => this.updateField("Street2", e.target.value)}
                placeholder="1234 AB Amsterdam"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">VAT Number</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.VAT}
                onChange={(e) => this.updateField("VAT", e.target.value)}
                placeholder="NL123456789B01"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">COC (KvK)</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.COC}
                onChange={(e) => this.updateField("COC", e.target.value)}
                placeholder="12345678"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Tax Category</label>
            <div className="col-sm-10">
              <select
                className="form-select"
                value={debtor.TAX}
                onChange={(e) => this.updateField("TAX", e.target.value)}
              >
                <option value="NL21">NL21 - Dutch 21% VAT</option>
                <option value="EU0">EU0 - EU Reverse Charge</option>
                <option value="WORLD0">WORLD0 - Export (0% VAT)</option>
              </select>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Billing Emails</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.BillingEmail.join(", ")}
                onChange={(e) => this.updateBillingEmail(e.target.value)}
                placeholder="billing@example.com, finance@example.com"
              />
              <small className="text-muted">Comma-separated email addresses</small>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Invoice Note</label>
            <div className="col-sm-10">
              <textarea
                className="form-control"
                value={debtor.NoteAdd}
                onChange={(e) => this.updateField("NoteAdd", e.target.value)}
                placeholder="Additional notes to appear on invoices"
                rows={3}
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Accounting Code</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={debtor.AccountingCode}
                onChange={(e) => this.updateField("AccountingCode", e.target.value)}
                placeholder="e.g., 3, 4, 5..."
              />
              <small className="text-muted">Relation code used in accounting software export (XLSX)</small>
            </div>
          </div>
        </div>
      </div>
    );
  }
}
