import * as React from "react";
import PurchaseInvoices from "./list-bucket";
import Axios from "axios";
import { decode as msgpackDecode } from '@msgpack/msgpack';

interface IPurchaseInvoice {
  ID: string;
  Supplier: { Name: string; VAT: string };
  Issuedate: string;
  Duedate: string;
  TotalEx: string;
  TotalTax: string;
  TotalInc: string;
  Status: string;
  Lines: { Description: string; Quantity: string; Price: string; Total: string; TaxPercent: string }[];
}

interface PurchasesPageProps {
  entity: string;
  year: string;
}

interface PurchasesPageState {
  unpaid: Record<string, IPurchaseInvoice[]>;
  paid: Record<string, IPurchaseInvoice[]>;
}

export default class PurchasesPage extends React.Component<PurchasesPageProps, PurchasesPageState> {
  constructor(props: PurchasesPageProps) {
    super(props);
    this.state = {unpaid: {}, paid: {}};
  }

  componentDidMount(): void {
    this.ajax();
  }

  private ajax(): void {
    Axios.get('/api/v1/purchases/'+this.props.entity+'/'+this.props.year, {params: {
      from: 0,
      count: 0
    }, headers: {'Accept': 'application/x-msgpack'}, responseType: 'arraybuffer'})
    .then(res => {
      // Server returns a known shape - runtime validation would be overkill for internal API
      // eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
      const data = msgpackDecode(new Uint8Array(res.data)) as { Invoices: Record<string, IPurchaseInvoice[]> };
      const s: PurchasesPageState = {unpaid: {}, paid: {}};

      for (const key in data.Invoices) {
        if (!Object.prototype.hasOwnProperty.call(data.Invoices, key)) {
          continue;
        }
        const item = data.Invoices[key];
        if (!item) continue;
        if (key.endsWith("/purchase-invoices-paid/")) {
          s.paid[key] = item;
        } else if (key.endsWith("/purchase-invoices-unpaid/")) {
          s.unpaid[key] = item;
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
      <PurchaseInvoices title="Unpaid" bucket="purchase-invoices-unpaid" items={this.state.unpaid} entity={this.props.entity} year={this.props.year} />
      <PurchaseInvoices title="Paid" bucket="purchase-invoices-paid" items={this.state.paid} entity={this.props.entity} year={this.props.year} />
    </div></div>;
  }
}
