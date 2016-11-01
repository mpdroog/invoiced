import * as React from "react";
import Axios from "axios";
import {IInvoiceState} from "./invoice-add";
import {IInjectedProps} from "react-router";
import {DOM} from "../lib/dom";

interface IInvoicePagination {
  from?: string
  count?: number
}
interface IInvoiceListState {
  pagination?: IInvoicePagination
  invoices?: IInvoiceState[]
}

interface IInvoiceListProps {
  bucket: string
  title: string
}

export default class Invoices extends React.Component<IInvoiceListProps, IInvoiceListState> {
  constructor(props: IInvoiceListProps) {
      super(props);
      this.state = {
        "pagination": {
          "from": "",
          "count": 50
        },
        "invoices": null
      };
  }

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/v1/invoices', {params: {
      from: this.state.pagination.from,
      count: this.state.pagination.count,
      bucket: this.props.bucket
    }})
    .then(res => {
      this.setState({invoices: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private delete(e: BrowserEvent) {
    e.preventDefault();
    let node = DOM.eventFilter(e, "A");
    let id = node.dataset["target"];
    if (node.dataset["status"] === 'FINAL') {
      console.log("Cannot delete finalized invoices.");
      return;
    }

    Axios.delete(`/api/v1/invoice/${id}?bucket=${this.props.bucket}`)
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private setPaid(id: string) {
    Axios.post('/api/v1/invoice/'+id+'/paid', {params: {
      bucket: this.props.bucket
    }})
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private conceptLine(key: string, inv: IInvoiceState): JSX.Element {
    return <tr key={key}>
      <td>{key}</td>
      <td>{inv.Meta.Invoiceid}</td>
      <td>{inv.Customer.Name}</td>
      <td>&euro; {inv.Total.Total}</td>
      <td>
        <a className="btn btn-default btn-hover-primary" href={"#invoice-add/"+this.props.bucket+"/"+key}><i className="fa fa-pencil"></i></a>
        <a disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? "btn-hover-danger faa-parent animated-hover" : "")} data-target={key} data-status={inv.Meta.Status} onClick={this.delete.bind(this, key)}><i className="fa fa-trash faa-flash"></i></a>
        <a className="btn btn-default btn-hover-primary" onClick={this.setPaid.bind(this, key)}><i className="fa fa-check"></i></a>
      </td>
    </tr>;
  }

  private finishedLine(key: string, inv: IInvoiceState): JSX.Element {
    return <tr key={key}>
      <td>{key}</td>
      <td>{inv.Meta.Invoiceid}</td>
      <td>{inv.Customer.Name}</td>
      <td>&euro; {inv.Total.Total}</td>
      <td>
        <a className="btn btn-default btn-hover-primary" href={"#invoice-add/"+this.props.bucket+"/"+key}><i className="fa fa-pencil"></i></a>
      </td>
    </tr>;
  }

	render() {
    let res:JSX.Element[] = [];
    console.log("invoices=", this.state.invoices);
    if (this.state.invoices && this.state.invoices.length > 0) {
      this.state.invoices.forEach((inv) => {
        let key: string = inv.Meta.Conceptid;
        if (this.props.bucket === "invoices") {
          res.push(this.conceptLine(key, inv));
        } else {
          res.push(this.finishedLine(key, inv));
        }
      });
    } else {
      res.push(<tr key="empty"><td colSpan={5}>No invoices yet :)</td></tr>);
    }

    var headerButtons = <div/>;
    if (this.props.bucket === "invoices") {
      headerButtons = <a href={"#invoice-add/"+this.props.bucket} className="btn btn-default btn-hover-primary showhide">
        <i className="fa fa-plus"></i> New
      </a>;
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
            <table className="table table-striped">
            	<thead><tr><th>#</th><th>Invoice</th><th>Customer</th><th>Amount</th><th>I/O</th></tr></thead>
            	<tbody>{res}</tbody>
            </table>
	        </div>
		    </div>
    </div>;
	}
}
