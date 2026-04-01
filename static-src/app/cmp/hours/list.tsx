import * as React from "react";
import Axios from "axios";
import {DOM} from "../../lib/dom";
import { decode as msgpackDecode } from '@msgpack/msgpack';
import {ActionLink} from "../../shared/ActionButton";

interface IHourPagination {
  from?: number;
  count?: number;
}

interface HourListItem {
  Name: string;
  Total: string;
}

interface HoursListProps {
  entity: string;
  year: string;
  bucket?: string;
}

type SortField = "name" | "bucket" | "hours";
type SortDirection = "asc" | "desc";

interface IHourState {
  pagination: IHourPagination;
  hours: Record<string, HourListItem[]> | null;
  sortField: SortField;
  sortDirection: SortDirection;
}

interface FlattenedHour {
  bucket: string;
  name: string;
  total: string;
}

export default class Hours extends React.Component<HoursListProps, IHourState> {
  constructor(props: HoursListProps) {
    super(props);
    this.state = {
      pagination: {
        from: 0,
        count: 50
      },
      hours: null,
      sortField: "name",
      sortDirection: "asc"
    };
  }

  componentDidMount(): void {
    this.ajax();
  }

  private ajax(): void {
    const entity = this.props.entity;
    const year = this.props.year;
    Axios.get('/api/v1/hours/'+entity+'/'+year, {params: this.state.pagination, headers: {'Accept': 'application/x-msgpack'}, responseType: 'arraybuffer'})
    .then(res => {
      // Server returns a known shape - runtime validation would be overkill for internal API
      // eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
      const data = msgpackDecode(new Uint8Array(res.data)) as Record<string, HourListItem[]>;
      this.setState({hours: data});
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private async delete(e: React.MouseEvent<HTMLAnchorElement>): Promise<void> {
    const node = DOM.eventFilter(e, "A");
    if (!node) return;
    const id = node.dataset["target"];
    const bucket = node.dataset["bucket"];
    if (id == null || bucket == null) return;

    await Axios.delete(`/api/v1/hour/${this.props.entity}/${this.props.year}/${bucket}/${id}`);
    location.reload();
  }

  private toggleSort(field: SortField): void {
    if (this.state.sortField === field) {
      this.setState({sortDirection: this.state.sortDirection === "asc" ? "desc" : "asc"});
    } else {
      this.setState({sortField: field, sortDirection: "asc"});
    }
  }

  private getSortIcon(field: SortField): string {
    if (this.state.sortField !== field) return "fa-sort";
    return this.state.sortDirection === "asc" ? "fa-sort-up" : "fa-sort-down";
  }

  private flattenAndSort(): FlattenedHour[] {
    const hours = this.state.hours;
    if (hours === null) return [];

    const flattened: FlattenedHour[] = [];
    for (const bucket in hours) {
      if (!Object.prototype.hasOwnProperty.call(hours, bucket)) continue;
      const items = hours[bucket];
      if (items === undefined) continue;
      for (const item of items) {
        flattened.push({bucket, name: item.Name, total: item.Total});
      }
    }

    const {sortField, sortDirection} = this.state;
    flattened.sort((a, b) => {
      let cmp: number;
      switch (sortField) {
        case "name":
          cmp = a.name.localeCompare(b.name);
          break;
        case "bucket":
          cmp = a.bucket.localeCompare(b.bucket);
          break;
        case "hours":
          cmp = parseFloat(a.total) - parseFloat(b.total);
          break;
      }
      return sortDirection === "asc" ? cmp : -cmp;
    });

    return flattened;
  }

  render(): React.JSX.Element {
    const that = this;
    const sorted = this.flattenAndSort();

    const rows = sorted.map((item) => (
      <tr key={item.bucket + item.name}>
        <td>{item.bucket}</td>
        <td>{item.name}</td>
        <td>{item.total}h</td>
        <td className="text-end">
          <div className="btn-group">
            <a className="btn btn-primary btn-sm" href={"#"+that.props.entity+"/"+that.props.year+"/hours/edit/"+item.bucket+"/"+item.name}><i className="fas fa-pencil"></i></a>
            <ActionLink className="btn btn-danger btn-sm" data-target={item.name} data-bucket={item.bucket} onClick={that.delete.bind(that)}><i className="fas fa-trash"></i></ActionLink>
          </div>
        </td>
      </tr>
    ));

    if (rows.length === 0 && this.state.hours !== null) {
      rows.push(<tr key="empty"><td colSpan={4}>No hours yet :)</td></tr>);
    }

    return <div className="mb-4">
        <div className="card">
          <div className="card-header">
            <div className="float-end">
              <div className="btn-group nm7">
                <a href={"#"+that.props.entity+"/"+that.props.year+"/hours/add"} id="js-new" className="btn btn-primary showhide"><i className="fas fa-plus"></i> New</a>
              </div>
            </div>
            Hour registration
          </div>
          <div className="card-body">
            <table className="table table-striped">
              <thead>
                <tr>
                  <th className="sortable" onClick={() => this.toggleSort("bucket")}>
                    Bucket <i className={"fas " + this.getSortIcon("bucket")}></i>
                  </th>
                  <th className="sortable" onClick={() => this.toggleSort("name")}>
                    Name <i className={"fas " + this.getSortIcon("name")}></i>
                  </th>
                  <th className="sortable" onClick={() => this.toggleSort("hours")}>
                    Hours <i className={"fas " + this.getSortIcon("hours")}></i>
                  </th>
                  <th className="text-end">I/O</th>
                </tr>
              </thead>
              <tbody>{rows}</tbody>
            </table>
          </div>
        </div>
    </div>;
  }
}
