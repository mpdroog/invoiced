import * as React from "react";
import Invoices from "./list-bucket";
import Axios from "axios";

export default class InvoicesPage extends React.Component<{}, {}> {
  constructor(props) {
    super(props);
    this.state = {concepts: [], pending: [], paid: []};
  }

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/v1/invoices/'+this.props.entity+'/'+this.props.year, {params: {
      from: 0,
      count: 0
    }})
    .then(res => {
      let s = {concepts: [], pending: [], paid: []};
      for (let key in res.data) {
        if (! res.data.hasOwnProperty(key)) {
          continue;
        }
        let item = res.data[key];
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

	render() {
		return <div>
      <Invoices title="Concepts" bucket="concepts" items={this.state.concepts} {...this.props} />
      <Invoices title="Pending" bucket="sales-invoices-unpaid" items={this.state.pending} {...this.props} />
      <Invoices title="Paid" bucket="sales-invoices-paid" items={this.state.paid} {...this.props} />
    </div>;
	}
}
