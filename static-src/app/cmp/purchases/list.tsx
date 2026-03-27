import * as React from "react";
import PurchaseInvoices from "./list-bucket";
import Axios from "axios";
import Msgpack from 'msgpack-lite';

export default class PurchasesPage extends React.Component<{}, {}> {
  constructor(props) {
    super(props);
    this.state = {unpaid: [], paid: []};
  }

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/v1/purchases/'+this.props.entity+'/'+this.props.year, {params: {
      from: 0,
      count: 0
    }, headers: {'Accept': 'application/x-msgpack'}, responseType: 'arraybuffer'})
    .then(res => {
      res.data = Msgpack.decode(new Uint8Array(res.data));
      let s = {unpaid: [], paid: []};

      for (let key in res.data.Invoices) {
        if (! res.data.Invoices.hasOwnProperty(key)) {
          continue;
        }
        let item = res.data.Invoices[key];
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

  render() {
    return <div className="row"><div className="col-sm-12">
      <PurchaseInvoices title="Unpaid" bucket="purchase-invoices-unpaid" items={this.state.unpaid} {...this.props} />
      <PurchaseInvoices title="Paid" bucket="purchase-invoices-paid" items={this.state.paid} {...this.props} />
    </div></div>;
  }
}
