import * as React from "react";
import Axios from "axios";
import { ActionLink } from "../../shared/ActionButton";
import type { DebtorItem, ProjectItem } from "../../types/entities";

type SortDir = "asc" | "desc";

interface ProjectsListProps {
  entity: string;
  year: string;
}

interface ProjectsListState {
  debtors: DebtorItem[] | null;
  projects: ProjectItem[] | null;
  debtorSort: { col: string; dir: SortDir };
  projectSort: { col: string; dir: SortDir };
}

export default class ProjectsList extends React.Component<ProjectsListProps, ProjectsListState> {
  constructor(props: ProjectsListProps) {
    super(props);
    this.state = {
      debtors: null,
      projects: null,
      debtorSort: { col: "key", dir: "asc" },
      projectSort: { col: "key", dir: "asc" },
    };
  }

  componentDidMount(): void {
    this.loadDebtors();
    this.loadProjects();
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

  private loadProjects(): void {
    Axios.get<ProjectItem[]>(`/api/v1/projects/${this.props.entity}`)
      .then((res) => {
        this.setState({ projects: res.data });
      })
      .catch((err) => {
        handleErr(err);
      });
  }

  private async deleteDebtor(e: React.MouseEvent<HTMLAnchorElement>): Promise<void> {
    const target = e.currentTarget;
    const key = target.dataset["target"];
    if (key === undefined || key === "") return;

    if (!confirm(`Delete debtor "${key}"?`)) return;

    await Axios.delete(`/api/v1/debtor/${this.props.entity}/${key}`);
    this.loadDebtors();
  }

  private async deleteProject(e: React.MouseEvent<HTMLAnchorElement>): Promise<void> {
    const target = e.currentTarget;
    const key = target.dataset["target"];
    if (key === undefined || key === "") return;

    if (!confirm(`Delete project "${key}"?`)) return;

    await Axios.delete(`/api/v1/project/${this.props.entity}/${key}`);
    this.loadProjects();
  }

  private sortDebtors(col: string): void {
    this.setState((prev) => ({
      debtorSort: {
        col,
        dir: prev.debtorSort.col === col && prev.debtorSort.dir === "asc" ? "desc" : "asc",
      },
    }));
  }

  private sortProjects(col: string): void {
    this.setState((prev) => ({
      projectSort: {
        col,
        dir: prev.projectSort.col === col && prev.projectSort.dir === "asc" ? "desc" : "asc",
      },
    }));
  }

  private sortIcon(currentCol: string, sortState: { col: string; dir: SortDir }): React.JSX.Element | null {
    if (sortState.col !== currentCol) return null;
    return <i className={`fas fa-sort-${sortState.dir === "asc" ? "up" : "down"} ms-1`}></i>;
  }

  private sortedDebtors(): DebtorItem[] {
    const { debtors, debtorSort } = this.state;
    if (debtors === null) return [];

    const sorted = [...debtors];
    const dir = debtorSort.dir === "asc" ? 1 : -1;

    sorted.sort((a, b) => {
      let cmp = 0;
      switch (debtorSort.col) {
        case "key":
          cmp = a.key.localeCompare(b.key);
          break;
        case "name":
          cmp = a.debtor.Name.localeCompare(b.debtor.Name);
          break;
        case "address":
          cmp = a.debtor.Street1.localeCompare(b.debtor.Street1);
          break;
        case "vat":
          cmp = a.debtor.VAT.localeCompare(b.debtor.VAT);
          break;
        case "tax":
          cmp = a.debtor.TAX.localeCompare(b.debtor.TAX);
          break;
        case "lastInvoice":
          cmp = (a.lastInvoice ?? "").localeCompare(b.lastInvoice ?? "");
          break;
        default:
          cmp = a.key.localeCompare(b.key);
      }
      return cmp * dir;
    });

    return sorted;
  }

  private sortedProjects(): ProjectItem[] {
    const { projects, projectSort } = this.state;
    if (projects === null) return [];

    const sorted = [...projects];
    const dir = projectSort.dir === "asc" ? 1 : -1;

    sorted.sort((a, b) => {
      let cmp = 0;
      switch (projectSort.col) {
        case "key":
          cmp = a.key.localeCompare(b.key);
          break;
        case "debtor":
          cmp = a.project.Debtor.localeCompare(b.project.Debtor);
          break;
        case "hourRate":
          cmp = a.project.HourRate - b.project.HourRate;
          break;
        case "dueDays":
          cmp = a.project.DueDays - b.project.DueDays;
          break;
        case "note":
          cmp = a.project.NoteAdd.localeCompare(b.project.NoteAdd);
          break;
        default:
          cmp = a.key.localeCompare(b.key);
      }
      return cmp * dir;
    });

    return sorted;
  }

  render(): React.JSX.Element {
    const { entity, year } = this.props;
    const { debtors, projects, debtorSort, projectSort } = this.state;

    const sortedDebtors = this.sortedDebtors();
    const sortedProjects = this.sortedProjects();

    const debtorRows = sortedDebtors.map((item) => (
      <tr key={item.key}>
        <td>
          <span className="d-inline-block text-truncate-sm" title={item.key}>
            {item.key}
          </span>
        </td>
        <td>
          <span className="d-inline-block text-truncate-sm" title={item.debtor.Name}>
            {item.debtor.Name}
          </span>
        </td>
        <td className="d-none d-md-table-cell">{item.debtor.Street1}</td>
        <td className="d-none d-lg-table-cell">{item.debtor.VAT}</td>
        <td>{item.debtor.TAX}</td>
        <td className="d-none d-md-table-cell">{item.lastInvoice ?? "-"}</td>
        <td className="text-end">
          <div className="btn-group">
            <a className="btn btn-primary btn-sm" href={`#${entity}/${year}/projects/debtor/edit/${item.key}`}>
              <i className="fas fa-pencil"></i>
            </a>
            <ActionLink className="btn btn-danger btn-sm" data-target={item.key} onClick={this.deleteDebtor.bind(this)}>
              <i className="fas fa-trash"></i>
            </ActionLink>
          </div>
        </td>
      </tr>
    ));

    const projectRows = sortedProjects.map((item) => (
      <tr key={item.key}>
        <td>
          <span className="d-inline-block text-truncate-sm" title={item.key}>
            {item.key}
          </span>
        </td>
        <td>
          <span className="d-inline-block text-truncate-sm" title={item.project.Debtor}>
            {item.project.Debtor}
          </span>
        </td>
        <td>{item.project.HourRate.toFixed(2)}</td>
        <td className="d-none d-md-table-cell">{item.project.DueDays}</td>
        <td className="d-none d-lg-table-cell">
          <span className="d-inline-block text-truncate-sm" title={item.project.NoteAdd}>
            {item.project.NoteAdd}
          </span>
        </td>
        <td className="text-end">
          <div className="btn-group">
            <a className="btn btn-primary btn-sm" href={`#${entity}/${year}/projects/project/edit/${item.key}`}>
              <i className="fas fa-pencil"></i>
            </a>
            <ActionLink
              className="btn btn-danger btn-sm"
              data-target={item.key}
              onClick={this.deleteProject.bind(this)}
            >
              <i className="fas fa-trash"></i>
            </ActionLink>
          </div>
        </td>
      </tr>
    ));

    return (
      <div>
        <div className="card mb-4">
          <div className="card-header">
            <div className="float-end">
              <a href={`#${entity}/${year}/projects/debtor/add`} className="btn btn-primary btn-sm">
                <i className="fas fa-plus"></i> New Debtor
              </a>
            </div>
            Debtors (Customers)
          </div>
          <div className="card-body">
            <table className="table table-striped">
              <thead>
                <tr>
                  <th className="sortable" onClick={() => this.sortDebtors("key")}>
                    Key{this.sortIcon("key", debtorSort)}
                  </th>
                  <th className="sortable" onClick={() => this.sortDebtors("name")}>
                    Name{this.sortIcon("name", debtorSort)}
                  </th>
                  <th className="sortable d-none d-md-table-cell" onClick={() => this.sortDebtors("address")}>
                    Address{this.sortIcon("address", debtorSort)}
                  </th>
                  <th className="sortable d-none d-lg-table-cell" onClick={() => this.sortDebtors("vat")}>
                    VAT{this.sortIcon("vat", debtorSort)}
                  </th>
                  <th className="sortable" onClick={() => this.sortDebtors("tax")}>
                    Tax{this.sortIcon("tax", debtorSort)}
                  </th>
                  <th className="sortable d-none d-md-table-cell" onClick={() => this.sortDebtors("lastInvoice")}>
                    Last Invoice{this.sortIcon("lastInvoice", debtorSort)}
                  </th>
                  <th className="text-end">Actions</th>
                </tr>
              </thead>
              <tbody>
                {debtorRows.length > 0 ? (
                  debtorRows
                ) : (
                  <tr>
                    <td colSpan={7}>{debtors === null ? "Loading..." : "No debtors yet"}</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        <div className="card mb-4">
          <div className="card-header">
            <div className="float-end">
              <a href={`#${entity}/${year}/projects/project/add`} className="btn btn-primary btn-sm">
                <i className="fas fa-plus"></i> New Project
              </a>
            </div>
            Projects
          </div>
          <div className="card-body">
            <table className="table table-striped">
              <thead>
                <tr>
                  <th className="sortable" onClick={() => this.sortProjects("key")}>
                    Key{this.sortIcon("key", projectSort)}
                  </th>
                  <th className="sortable" onClick={() => this.sortProjects("debtor")}>
                    Debtor{this.sortIcon("debtor", projectSort)}
                  </th>
                  <th className="sortable" onClick={() => this.sortProjects("hourRate")}>
                    Hour Rate{this.sortIcon("hourRate", projectSort)}
                  </th>
                  <th className="sortable d-none d-md-table-cell" onClick={() => this.sortProjects("dueDays")}>
                    Due Days{this.sortIcon("dueDays", projectSort)}
                  </th>
                  <th className="sortable d-none d-lg-table-cell" onClick={() => this.sortProjects("note")}>
                    Note{this.sortIcon("note", projectSort)}
                  </th>
                  <th className="text-end">Actions</th>
                </tr>
              </thead>
              <tbody>
                {projectRows.length > 0 ? (
                  projectRows
                ) : (
                  <tr>
                    <td colSpan={6}>{projects === null ? "Loading..." : "No projects yet"}</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}
