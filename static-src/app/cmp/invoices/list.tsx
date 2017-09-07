import * as React from "react";
import Invoices from "./list-bucket";
import Axios from "axios";
import * as Msgpack from 'msgpack-lite';
import * as Moment from "moment";

export default class InvoicesPage extends React.Component<{}, {}> {
  constructor(props) {
    super(props);
    this.state = {concepts: [], pending: [], paid: [], commits: []};
  }

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/v1/invoices/'+this.props.entity+'/'+this.props.year, {params: {
      from: 0,
      count: 0
    }, headers: {'Accept': 'application/x-msgpack'}, responseType: 'arraybuffer'})
    .then(res => {
      res.data = Msgpack.decode(new Uint8Array(res.data));
      let s = {concepts: [], pending: [], paid: [], commits: []};
      console.log(res.data);

      // invoices
      for (let key in res.data.Invoices) {
        if (! res.data.Invoices.hasOwnProperty(key)) {
          continue;
        }
        let item = res.data.Invoices[key];
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
      s.commits = res.data.Commits;
      this.setState(s);
    })
    .catch(err => {
      handleErr(err);
    });
  }

	render() {
    let commits = [];
    let that = this;
    this.state.commits.forEach(function(item) {
      item = item.Commit;
      let subject = `RE: ${item.Message}`;
      let body = `Open invoices ${location.href}`;
      let now = Moment.unix(item.Committer.When[0]);
      console.log(now);
      commits.push(
        <div className="vertical-timeline-block">
          <div className="vertical-timeline-icon navy-bg">
              <i className="fa fa-calendar"></i>
          </div>
          <div className="vertical-timeline-content">
              <div className="p-sm">
                  <span className="vertical-date pull-right"> {now.format('YYYY-MM-DD')} <br/> <small>{now.format('HH:mm:ss')}</small> </span>

                  <h2>{item.Message}</h2>
                  <p>{now.fromNow()}</p>
              </div>
              <div className="panel-footer">
                  {item.Committer.Name} (<a href={"mailto:"+item.Committer.Email+"?subject="+subject+"&body="+body}>{item.Committer.Email}</a>)
              </div>
          </div>
        </div>
      );
    });

		return <div className="row"><div className="col-sm-8">
      <Invoices title="Concepts" bucket="concepts" items={this.state.concepts} {...this.props} />
      <Invoices title="Pending" bucket="sales-invoices-unpaid" items={this.state.pending} {...this.props} />
      <Invoices title="Paid" bucket="sales-invoices-paid" items={this.state.paid} {...this.props} />
    </div><div className="col-sm-4">

<div className="v-timeline vertical-container animate-panel" data-child="vertical-timeline-block" data-delay="1">
  {commits}
</div>

    </div></div>;
	}
}
