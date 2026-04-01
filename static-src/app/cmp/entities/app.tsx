import * as React from "react";
import Axios from "axios";

// Format number with space as thousands separator: 51868.65 -> 51 868.65
function formatCurrency(value: number): string {
  const parts = value.toFixed(2).split(".");
  const intPart = parts[0] ?? "0";
  parts[0] = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  return parts.join(",");
}

interface IDictionary {
  [index: string]: IMetricDay;
}

interface IMetricDay {
  RevenueTotal: string;
  RevenueEx: string;
}
interface IEntityYear {
  Years?: string[];
  YearRevenue?: Record<string, string>;
  Name: string;
  VAT: string;
}

interface IEntitiesState {
  entities: Record<string, IEntityYear>;
  metrics?: IDictionary;
}

export default class Entities extends React.Component<Record<string, never>, IEntitiesState> {
  constructor(props: Record<string, never>) {
    super(props);
    this.state = {
      entities: {},
    };
  }

  render(): React.JSX.Element {
    let items: React.JSX.Element[] = [];
    const curYear: number = new Date().getFullYear();
    for (const key in this.state.entities) {
      if (!Object.prototype.hasOwnProperty.call(this.state.entities, key)) {
        // ignore
        continue;
      }
      const ukey = "entity_" + key;
      const entity = this.state.entities[key];
      if (!entity) continue;

      let hasCurYear: boolean = false;
      const accountingYears: React.JSX.Element[] = [];
      let prevRevenue: number = 0;

      // Sort years ascending for proper comparison
      const sortedYears = (entity.Years ?? []).slice().sort((a, b) => parseInt(a) - parseInt(b));

      for (const yearStr of sortedYears) {
        const year: number = parseInt(yearStr);
        if (year === curYear) {
          hasCurYear = true;
        }

        const revenueStr = entity.YearRevenue?.[yearStr] ?? "0.00";
        const revenue = parseFloat(revenueStr);
        const formattedRevenue = formatCurrency(revenue);

        // Calculate delta vs previous year
        let deltaEl: React.JSX.Element | null = null;
        if (prevRevenue > 0) {
          const delta = revenue - prevRevenue;
          const pct = ((delta / prevRevenue) * 100).toFixed(0);
          const sign = delta >= 0 ? "+" : "";
          const badgeClass = delta >= 0 ? "m-l-sm badge bg-success" : "m-l-sm badge bg-danger";
          deltaEl = (
            <span className={badgeClass}>
              {sign}&euro; {formatCurrency(Math.abs(delta))} ({sign}
              {pct}%)
            </span>
          );
        }

        accountingYears.push(
          <tr key={ukey + "company" + year}>
            <td>
              <a href={"#" + key + "/" + year}>{year}</a>
            </td>
            <td>
              &euro; {formattedRevenue}
              {deltaEl}
            </td>
          </tr>
        );

        prevRevenue = revenue;
      }

      let link: React.JSX.Element | null = null;
      if (!hasCurYear) {
        link = <a href={"/api/v1/entities/" + key + "/open/" + curYear}>Open {curYear}</a>;
      }
      items.push(
        <tr key={ukey + "company"}>
          <td>
            {entity.Name} - {entity.VAT}
          </td>
          <td>{link}</td>
        </tr>
      );
      items = items.concat(accountingYears);
    }

    return (
      <div className="container">
        <div className="row">
          <div className="col-md-6 offset-md-3">
            <div className="card">
              <div className="card-body">
                <h2 className="font-light m-b-xs">
                  <i className="fas fa-building"></i>
                  &nbsp;Your Companies
                </h2>
                <table className="table">
                  <thead>
                    <tr>
                      <th>Company</th>
                      <th>Revenue</th>
                    </tr>
                  </thead>
                  <tbody>{items}</tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  componentDidMount(): void {
    this.ajax();
  }

  private ajax(): void {
    Axios.get<Record<string, IEntityYear>>("/api/v1/entities", {})
      .then((res) => {
        this.setState({ entities: res.data });
      })
      .catch((err) => {
        handleErr(err);
      });
  }
}
