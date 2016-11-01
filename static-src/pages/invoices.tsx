import * as React from "react";
import Axios from "axios";
import {IInvoiceState} from "./invoice-add";
import {IInjectedProps} from "react-router";

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
    Axios.get('/api/invoices', {params: {
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
    e.preventDefault()
    let id = e.target.dataset['target'];

    Axios.delete(`/api/invoice/${id}?bucket=${this.props.bucket}`)
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

	render() {
    let res:JSX.Element[] = [];
    console.log("invoices=", this.state.invoices);
    if (this.state.invoices && this.state.invoices.length > 0) {
      this.state.invoices.forEach((inv) => {
        let key: string = inv.Meta.Conceptid;
        res.push(<tr key={key}>
          <td>{key}</td>
          <td>{inv.Meta.Invoiceid}</td>
          <td>{inv.Customer.Name}</td>
          <td>{inv.Total.Total}</td>
          <td>
            <a className="btn btn-default btn-hover-primary" href={"#invoice-add/"+this.props.bucket+"/"+key}><i className="fa fa-pencil"></i></a>
            <a disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? "btn-hover-danger faa-parent animated-hover" : "")} data-target={key} onClick={this.delete.bind(this)}><i className="fa fa-trash faa-flash"></i></a>
          </td></tr>);
      });
    } else {
      res.push(<tr key="empty"><td colSpan={5}>No invoices yet :)</td></tr>);
    }

		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <a href={"#invoice-add/"+this.props.bucket} className="btn btn-default btn-hover-primary showhide"><i className="fa fa-plus"></i> New</a>
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
