import * as React from "react";
import Axios from "axios";
import { formatDate, today } from "../../utils/date";
import type { QuarterStatus } from "../../types/model";

export interface TaxState {
  label: string;
  cls: string; // bootstrap badge-class
  open: boolean; // VAT-return still needs filing (quarter ended)
}

export async function loadTaxStatus(entity: string, year: number | string): Promise<QuarterStatus[]> {
  const res = await Axios.get<QuarterStatus[]>("/api/v1/taxes/" + entity + "/" + String(year) + "/status");
  return res.data;
}

export function formatDeadline(d: string): string {
  return formatDate(d, "D MMM YYYY");
}

export function daysText(days: number): string {
  if (days < 0) {
    return "overdue " + String(-days) + (days === -1 ? " day" : " days");
  }
  if (days === 0) {
    return "today";
  }
  return String(days) + (days === 1 ? " day" : " days") + " remaining";
}

export function taxState(s: QuarterStatus): TaxState {
  if (s.Filed !== "" && s.Paid !== "") {
    return { label: "Filed + paid", cls: "bg-success", open: false };
  }
  if (s.Filed !== "") {
    // Nothing to pay, or the receipt isn't saved in the tax-dir
    return { label: "Filed, no payment receipt", cls: "bg-info", open: false };
  }
  if (!s.Active) {
    return { label: "No data", cls: "bg-secondary", open: false };
  }
  if (today() <= s.PeriodEnd) {
    return { label: "Quarter running", cls: "bg-secondary", open: false };
  }
  if (s.DaysLeft < 0) {
    return { label: "Overdue", cls: "bg-danger", open: true };
  }
  return { label: "To file", cls: s.DaysLeft <= 7 ? "bg-danger" : "bg-warning", open: true };
}

interface TaxStatusTableProps {
  data: QuarterStatus[];
}

export class TaxStatusTable extends React.Component<TaxStatusTableProps> {
  render(): React.JSX.Element {
    const rows = this.props.data.map((s) => {
      const st = taxState(s);
      const days = s.Filed !== "" || !s.Active ? "–" : daysText(s.DaysLeft);
      return (
        <tr key={s.Quarter}>
          <td>
            <strong>{s.Quarter}</strong>
          </td>
          <td>
            <span className={"badge " + st.cls}>{st.label}</span>
          </td>
          <td title={s.Filed}>
            {s.Filed !== "" ? <i className="fas fa-check"></i> : null}
            {s.ICP !== "" ? <span title={s.ICP}> (+ICP)</span> : null}
          </td>
          <td title={s.Paid}>{s.Paid !== "" ? <i className="fas fa-check"></i> : null}</td>
          <td>{formatDeadline(s.Deadline)}</td>
          <td>{days}</td>
        </tr>
      );
    });

    return (
      <table className="table table-bordered table-striped">
        <thead>
          <tr>
            <th>Quarter</th>
            <th>Status</th>
            <th>Filed</th>
            <th>Paid</th>
            <th>Deadline</th>
            <th>Days left</th>
          </tr>
        </thead>
        <tbody>{rows}</tbody>
      </table>
    );
  }
}
