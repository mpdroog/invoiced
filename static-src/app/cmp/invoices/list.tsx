import * as React from "react";
import Invoices from "./list-bucket";
import Axios from "axios";
import { decode as msgpackDecode } from '@msgpack/msgpack';
import {IInvoiceState} from "./edit-struct";

interface InvoicesPageProps {
  entity: string;
  year: string;
}

interface InvoicesPageState {
  concepts: Record<string, IInvoiceState[]>;
  pending: Record<string, IInvoiceState[]>;
  paid: Record<string, IInvoiceState[]>;
}

export default class InvoicesPage extends React.Component<InvoicesPageProps, InvoicesPageState> {
  constructor(props: InvoicesPageProps) {
    super(props);
    this.state = {concepts: {}, pending: {}, paid: {}};
  }

  componentDidMount(): void {
    this.ajax();
  }

  private ajax(): void {
    Axios.get('/api/v1/invoices/'+this.props.entity+'/'+this.props.year, {params: {
      from: 0,
      count: 0
    }, headers: {'Accept': 'application/x-msgpack'}, responseType: 'arraybuffer'})
    .then(res => {
      const data = msgpackDecode(new Uint8Array(res.data)) as { Invoices: Record<string, IInvoiceState[]> };
      const s: InvoicesPageState = {concepts: {}, pending: {}, paid: {}};

      // invoices
      for (const key in data.Invoices) {
        if (!Object.prototype.hasOwnProperty.call(data.Invoices, key)) {
          continue;
        }
        const item = data.Invoices[key];
        if (key.endsWith("/sales-invoices-paid/")) {
          s.paid[key] = item;
        } else if (key.endsWith("/sales-invoices-unpaid/")) {
          s.pending[key] = item;
        } else if (key.endsWith("/concepts/sales-invoices/")) {
          s.concepts[key] = item;
        } else {
          console.log("SKIP " + key);
        }
      }
      this.setState(s);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  render(): React.JSX.Element {
    return <div className="row"><div className="col-sm-12">
      <Invoices title="Concepts" bucket="concepts" items={this.state.concepts} {...this.props} />
      <Invoices title="Pending" bucket="sales-invoices-unpaid" items={this.state.pending} {...this.props} />
      <Invoices title="Paid" bucket="sales-invoices-paid" items={this.state.paid} {...this.props} />
    </div></div>;
  }
}
