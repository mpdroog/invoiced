import * as React from "react";
import Axios from "axios";
import {IInvoiceState} from "./invoice-add";

interface IInvoicePagination {
  from?: number
  count?: number
}
interface IInvoiceListState {
  pagination?: IInvoicePagination
  invoices?: IInvoiceState[]
}

export default class Hours extends React.Component<{}, IInvoiceListState> {
  constructor() {
      super();
      this.state = { pagination: {from:0, count:50}, invoices: null };
  }

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/invoices')
    .then(res => {
      this.setState({invoices: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private delete(e: BrowserEvent) {
    e.preventDefault()
    var id = e.target.dataset['target'];

    Axios.delete('/api/invoice/'+id)
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

	render() {
    var res:JSX.Element[] = [];
    var that = this;
    console.log("invoices=", this.state.invoices);
    if (this.state.invoices && this.state.invoices.length > 0) {
      this.state.invoices.forEach(function(inv) {
        var key: string = inv.Meta.Conceptid;
        res.push(<tr key={key}>
          <td>{key}</td>
          <td>{inv.Meta.Invoiceid}</td>
          <td>{inv.Customer.Name}</td>
          <td>{inv.Total.Total}</td>
          <td>
            <a className="btn btn-default btn-hover-primary" href={"#invoice-add/"+key}><i className="fa fa-pencil"></i></a>
            <a disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? "btn-hover-danger faa-parent animated-hover" : "")} data-target={key} onClick={that.delete.bind(this)}><i className="fa fa-trash faa-flash"></i></a>
          </td></tr>);
      });
    } else {
      res.push(<tr key="empty"><td colSpan={4}>No invoices yet :)</td></tr>);
    }

		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <a href="#invoice-add" className="btn btn-default btn-hover-primary showhide"><i className="fa fa-plus"></i> New</a>
              </div>
            </div>
            Invoices
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
