import * as React from "react";
import Axios from "axios";
import {IInvoiceState} from "./invoice-add";
import {DOM} from "../../lib/dom";

interface IInvoicePagination {
  from?: string
  count?: number
}
interface IInvoiceListState {
  pagination?: IInvoicePagination
  invoices?: IInvoiceState[]
  isBalance: bool
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
          "from": 0,
          "count": 50
        },
        "invoices": null,
        "isBalance": false,
      };
  }

  private delete(e: BrowserEvent) {
    e.preventDefault();
    let node = DOM.eventFilter(e, "A");
    let id = node.dataset["target"];
    let bucket = node.dataset.bucket;
    if (node.dataset["status"] === 'FINAL') {
      console.log("Cannot delete finalized invoices.");
      return;
    }

    Axios.delete(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${id}`)
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private setPaid(e) {
    e.preventDefault();
    let node = DOM.eventFilter(e, "A");
    let id = node.dataset.id;
    let bucket = node.dataset.bucket;

    Axios.post(`/api/v1/invoice/${this.props.entity}/${this.props.year}/${bucket}/${id}/paid`, {})
    .then(res => {
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private conceptLine(key: string, inv: IInvoiceState, bucket: string): React.JSX.Element {
    return <tr key={key}>
      <td>{key}</td>
      <td>{inv.Meta.Invoiceid}</td>
      <td>{inv.Customer.Name}</td>
      <td>&euro; {inv.Total.Total}</td>
      <td>
        <a className="btn btn-default btn-hover-primary" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/edit/"+bucket+"/"+key}><i className="fa fa-pencil"></i></a>
        <a disabled={inv.Meta.Status === 'FINAL'} className="btn btn-default btn-hover-danger faa-parent animated-hover" data-target={key} data-status={inv.Meta.Status} onClick={this.delete.bind(this)}><i className="fa fa-trash faa-flash"></i></a>
        <a disabled={bucket === 'concepts'} className="btn btn-default btn-hover-danger faa-parent animated-hover" data-id={key} data-bucket={bucket} onClick={this.setPaid.bind(this)}><i className="fa fa-check faa-flash"></i></a>
      </td>
    </tr>;
  }

  private finishedLine(key: string, inv: IInvoiceState, bucket: string): React.JSX.Element {
    console.log(bucket);
    return <tr key={key}>
      <td>{key}</td>
      <td>{inv.Meta.Invoiceid}</td>
      <td>{inv.Customer.Name}</td>
      <td>&euro; {inv.Total.Total}</td>
      <td>
        <a className="btn btn-default btn-hover-primary" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/edit/"+bucket+"/"+key}><i className="fa fa-pencil"></i></a>
      </td>
    </tr>;
  }

  private openUpload(e) {
    e.preventDefault();
    document.getElementById('js-balance-field').click();
  }
  private uploadBalance(e) {
    if (e.target.files.length === 0) {
      return;
    }

    console.log("Upload");
    let file = e.target.files[0];
    let form = new FormData();
    form.append('file', file, file.name);

    // api/v1/invoice/:entity/:year/:bucket/:id/balance
    Axios.post('/api/v1/invoice-balance/'+this.props.entity+'/'+this.props.year,
      form, {headers: {'Content-Type': 'multipart/form-data' }})
    .then(res => {
      // TODO: Report something to UI?
      location.reload();
    })
    .catch(err => {
      handleErr(err);
    });
  }

  // TODO: Ugly hack...
  private bucket(invoiceId) {
    return 'Q' + invoiceId.split("-")[0].split("Q")[1];
  }

	render() {
    let res:React.JSX.Element[] = [];
    if (this.props.items) {
      for (let dir in this.props.items) {
        if (! this.props.items.hasOwnProperty(dir)) {
          continue;
        }
        let bucket = dir.split("/")[3];

        this.props.items[dir].forEach((inv) => {
          let key: string = inv.Meta.Conceptid;
          if (this.props.bucket === "concepts" || this.props.bucket === "sales-invoices-unpaid") {
            if (this.props.bucket === "sales-invoices-unpaid") {
              bucket = this.bucket(inv.Meta.Invoiceid);
            }
            res.push(this.conceptLine(key, inv, bucket));
          } else {
            res.push(this.finishedLine(key, inv, this.bucket(inv.Meta.Invoiceid)));
          }
        });
      }
    }

    if (res.length === 0) {
      res.push(<tr key="empty"><td colSpan={5}>No invoices yet :)</td></tr>);
    }

    var headerButtons = <div/>;
    if (this.props.bucket === "concepts") {
      headerButtons = <div>
        <a id="js-new" href={"#"+this.props.entity+"/"+this.props.year+"/"+"invoices/add"} className="btn btn-default btn-hover-primary showhide">
          <i className="fa fa-plus"></i> New
        </a>
      </div>;
    }
    if (this.props.bucket === "sales-invoices-unpaid") {
      headerButtons = <div><a id="js-balance" onClick={this.openUpload.bind(this)} className="btn btn-default btn-hover-primary showhide">
          <i className="fa fa-upload"></i> Bankbalance
        </a>
      </div>;
    }

    let url = `/api/v1/invoice/${this.props.entity}/${this.props.year}/${this.props.bucket}/balance`;
    let balanceUpload = <div>
      <form className="form-inline hidden" method="post" encType="multipart/form-data" action={url}>
        <input id="js-balance-field" accept=".xml" className="form-control" name="file" type="file" onChange={this.uploadBalance.bind(this)} />
        <button className="btn btn-default btn-hover-primary" type="submit"><i className="fa fa-arrow-up"></i> Upload</button>
      </form>
    </div>;

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
            {balanceUpload}
            <table className="table table-striped">
            	<thead><tr><th>#</th><th>Invoice</th><th>Customer</th><th>Amount</th><th>I/O</th></tr></thead>
            	<tbody>{res}</tbody>
            </table>
	        </div>
		    </div>
    </div>;
	}
}
