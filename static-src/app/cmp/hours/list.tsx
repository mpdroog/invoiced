import * as React from "react";
import Axios from "axios";
import {DOM} from "../../lib/dom";
import * as Msgpack from 'msgpack-lite';

interface IHourPagination {
  from?: string
  count?: number
}
interface IHourState {
  pagination?: IHourPagination
  hours?: string[]
}

export default class Hours extends React.Component<{}, IHourState> {
  constructor(p, s) {
    super(p, s);
    this.state = {
      "pagination": {
        "from": 0,
        "count": 50
      },
      "hours": null
    };
  }

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    let entity = this.props.entity;
    let year = this.props.year;
    Axios.get('/api/v1/hours/'+entity+'/'+year, {params: this.state.pagination, headers: {'Accept': 'application/x-msgpack'}, responseType: 'arraybuffer'})
    .then(res => {
      res.data = Msgpack.decode(new Uint8Array(res.data));
      this.setState({hours: res.data});
      window.rootdev = {
        invoiced: res.data
      };
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private delete(e: BrowserEvent) {
    e.preventDefault();
    let id = DOM.eventFilter(e, "A").dataset["target"];

    Axios.delete(`/api/v1/hour/${this.props.entity}/${this.props.year}/${this.props.bucket}/${id}`)
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  render() {
    let res:React.JSX.Element[] = [];
    let that = this;
    let items = 0;
    console.log("hours=",this.state.hours);

    if (this.state.hours) {
      for (let bucket in this.state.hours) {
        if (! this.state.hours.hasOwnProperty(bucket)) {
          continue;
        }
        items++;
        this.state.hours[bucket].forEach(function(elem) {
          res.push(<tr key={bucket+elem}>
            <td>{bucket}</td>
            <td>{elem}</td>
            <td>
              <a className="btn btn-default btn-hover-primary" href={"#"+that.props.entity+"/"+that.props.year+"/hours/edit/"+bucket+"/"+elem}><i className="fa fa-pencil"></i></a>
              <a className="btn btn-default btn-hover-danger faa-parent animated-hover" data-target={elem} onClick={that.delete.bind(that)}><i className="fa fa-trash faa-flash"></i></a>
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
                <a href={"#"+that.props.entity+"/"+that.props.year+"/hours/add"} id="js-new" className="btn btn-default btn-hover-primary showhide"><i className="fa fa-plus"></i> New</a>
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
