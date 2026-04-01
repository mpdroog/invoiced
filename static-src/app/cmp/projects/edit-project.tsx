import * as React from "react";
import Axios from "axios";
import { ActionButton } from "../../shared/ActionButton";
import type { ProjectItem, Project, DebtorItem } from "../../types/entities";

interface ProjectEditProps {
  entity: string;
  year: string;
  id?: string;
}

interface ProjectEditState {
  key: string;
  project: Project;
  isNew: boolean;
  debtors: DebtorItem[];
}

const emptyProject: Project = {
  Name: "",
  Debtor: "",
  BillingEmail: [],
  NoteAdd: "",
  HourRate: 75.0,
  DueDays: 14,
  PO: "",
  Street1: "",
};

export default class ProjectEdit extends React.Component<ProjectEditProps, ProjectEditState> {
  constructor(props: ProjectEditProps) {
    super(props);
    this.state = {
      key: "",
      project: { ...emptyProject },
      isNew: props.id === undefined || props.id === "",
      debtors: [],
    };
  }

  componentDidMount(): void {
    this.loadDebtors();
    if (this.props.id !== undefined && this.props.id !== "") {
      this.loadProject(this.props.id);
    }
  }

  private loadDebtors(): void {
    Axios.get<DebtorItem[]>(`/api/v1/debtors/${this.props.entity}`)
      .then((res) => {
        this.setState({ debtors: res.data });
      })
      .catch((err) => {
        handleErr(err);
      });
  }

  private loadProject(key: string): void {
    Axios.get<ProjectItem>(`/api/v1/project/${this.props.entity}/${key}`)
      .then((res) => {
        this.setState({
          key: res.data.key,
          project: res.data.project,
          isNew: false,
        });
      })
      .catch((err) => {
        handleErr(err);
      });
  }

  private updateField(field: keyof Project, value: string | number): void {
    this.setState((prev) => ({
      project: { ...prev.project, [field]: value },
    }));
  }

  private updateBillingEmail(value: string): void {
    const emails = value
      .split(",")
      .map((e) => e.trim())
      .filter((e) => e.length > 0);
    this.setState((prev) => ({
      project: { ...prev.project, BillingEmail: emails },
    }));
  }

  private async save(): Promise<void> {
    const { key, project } = this.state;
    if (key === "") {
      alert("Key is required");
      return;
    }
    if (project.Debtor === "") {
      alert("Debtor is required");
      return;
    }

    const payload: ProjectItem = { key, project };
    await Axios.post(`/api/v1/project/${this.props.entity}/${key}`, payload);

    location.href = `#${this.props.entity}/${this.props.year}/projects`;
  }

  render(): React.JSX.Element {
    const { key, project, isNew, debtors } = this.state;

    return (
      <div className="card">
        <div className="card-header">
          <div className="float-end">
            <ActionButton className="btn btn-success" onClick={this.save.bind(this)}>
              <i className="fas fa-floppy-disk"></i>&nbsp;Save
            </ActionButton>
          </div>
          {isNew ? "New Project" : `Edit Project: ${key}`}
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
                placeholder="e.g., acme-website"
              />
              <small className="text-muted">Lowercase letters, numbers, and hyphens only</small>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Debtor</label>
            <div className="col-sm-10">
              <select
                className="form-select"
                value={project.Debtor}
                onChange={(e) => this.updateField("Debtor", e.target.value)}
              >
                <option value="">-- Select Debtor --</option>
                {debtors.map((d) => (
                  <option key={d.key} value={d.key}>
                    {d.key} - {d.debtor.Name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Project Name</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={project.Name}
                onChange={(e) => this.updateField("Name", e.target.value)}
                placeholder="Website Development"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Hour Rate</label>
            <div className="col-sm-10">
              <div className="input-group">
                <span className="input-group-text">EUR</span>
                <input
                  type="number"
                  className="form-control"
                  value={project.HourRate}
                  onChange={(e) => {
                    const val = parseFloat(e.target.value);
                    this.updateField("HourRate", Number.isNaN(val) ? 0 : val);
                  }}
                  step="0.01"
                  min="0"
                />
              </div>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Due Days</label>
            <div className="col-sm-10">
              <input
                type="number"
                className="form-control"
                value={project.DueDays}
                onChange={(e) => {
                  const val = parseInt(e.target.value);
                  this.updateField("DueDays", Number.isNaN(val) ? 14 : val);
                }}
                min="1"
              />
              <small className="text-muted">Payment due within this many days</small>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">PO Number</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={project.PO}
                onChange={(e) => this.updateField("PO", e.target.value)}
                placeholder="Purchase Order number (optional)"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Contact</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={project.Street1}
                onChange={(e) => this.updateField("Street1", e.target.value)}
                placeholder="Contact person or location"
              />
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Billing Emails</label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                value={project.BillingEmail.join(", ")}
                onChange={(e) => this.updateBillingEmail(e.target.value)}
                placeholder="Override debtor emails (optional)"
              />
              <small className="text-muted">Leave empty to use debtor emails</small>
            </div>
          </div>
          <div className="row mb-3">
            <label className="col-sm-2 col-form-label">Invoice Note</label>
            <div className="col-sm-10">
              <textarea
                className="form-control"
                value={project.NoteAdd}
                onChange={(e) => this.updateField("NoteAdd", e.target.value)}
                placeholder="Additional notes to appear on invoices"
                rows={3}
              />
            </div>
          </div>
        </div>
      </div>
    );
  }
}
