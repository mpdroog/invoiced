import * as React from "react";
import Axios from "axios";
import {DOM} from "../../lib/dom";
import { decode as msgpackDecode } from '@msgpack/msgpack';
import {ActionLink} from "../../shared/ActionButton";

interface IHourPagination {
  from?: number;
  count?: number;
}

interface HoursListProps {
  entity: string;
  year: string;
  bucket?: string;
}

interface IHourState {
  pagination: IHourPagination;
  hours: Record<string, string[]> | null;
}

export default class Hours extends React.Component<HoursListProps, IHourState> {
  constructor(props: HoursListProps) {
    super(props);
    this.state = {
      pagination: {
        from: 0,
        count: 50
      },
      hours: null
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
      const data = msgpackDecode(new Uint8Array(res.data)) as Record<string, string[]>;
      this.setState({hours: data});
      (window as Window & { rootdev?: { invoiced?: unknown } }).rootdev = {
        invoiced: data
      };
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
    if (!id || !bucket) return;

    await Axios.delete(`/api/v1/hour/${this.props.entity}/${this.props.year}/${bucket}/${id}`);
    location.reload();
  }

  render(): React.JSX.Element {
    const res:React.JSX.Element[] = [];
    const that = this;
    let items = 0;
    console.log("hours=",this.state.hours);

    if (this.state.hours) {
      for (const bucket in this.state.hours) {
        if (!Object.prototype.hasOwnProperty.call(this.state.hours, bucket)) {
          continue;
        }
        const bucketItems = this.state.hours[bucket];
        if (!bucketItems) continue;
        items++;
        bucketItems.forEach(function(elem) {
          res.push(<tr key={bucket+elem}>
            <td>{bucket}</td>
            <td>{elem}</td>
            <td>
              <a className="btn btn-default btn-hover-primary" href={"#"+that.props.entity+"/"+that.props.year+"/hours/edit/"+bucket+"/"+elem}><i className="fas fa-pencil"></i></a>
              <ActionLink className="btn btn-default btn-hover-danger" data-target={elem} data-bucket={bucket} onClick={that.delete.bind(that)}><i className="fas fa-trash"></i></ActionLink>
            </td></tr>);
        });
      }
    }
    if (items === 0) {
      res.push(<tr key="empty"><td colSpan={5}>No hours yet :)</td></tr>);
    }

    return <div className="normalheader">
        <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <a href={"#"+that.props.entity+"/"+that.props.year+"/hours/add"} id="js-new" className="btn btn-default btn-hover-primary showhide"><i className="fas fa-plus"></i> New</a>
              </div>
            </div>
            Hour registration
          </div>
          <div className="panel-body">
            <table className="table table-striped">
              <thead><tr><th>Bucket</th><th>Name</th><th>I/O</th></tr></thead>
              <tbody>{res}</tbody>
            </table>
          </div>
        </div>
    </div>;
  }
}
