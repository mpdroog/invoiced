import * as React from "react";
import Axios from "axios";
import {DOM} from "../lib/dom";

interface IDictionary {
[index: string]: IMetricDay;
}

interface IMetricDay {
  RevenueTotal: string
  RevenueEx: string
}
interface IMetrics {
  metrics?: IDictionary
}

export default class Dashboard extends React.Component<{}, IMetrics> {
  constructor() {
    super();
    this.state = {
      "metrics": {}
    };
  }

	render() {
	    let items:JSX.Element[] = [];
		let sorted:string[] = Object.keys(this.state.metrics).sort();

		var pref = "0.00";
		for (var i = 0; i < sorted.length; i++) {
			var key = sorted[i];
			var revenue = this.state.metrics[key].RevenueEx;
			var delta = (((revenue*100) / (pref*100) * 100) - 100).toFixed(0);
			items.push(<tr key={key}><td>{key}</td><td>&euro; {revenue}</td><td>{delta}%</td></tr>);

			pref = revenue;
		}

		return <div className="normalheader">
		    <div className="hpanel">
		        <div className="panel-body">
		            <h2 className="font-light m-b-xs">
		                Revenue
		            </h2>
		            <table className="table">
		            	<thead>
		            		<tr><th>Date</th><th>Revenue</th><th>Δ</th></tr>
		            	</thead>
		            	<tbody>{items}</tbody>
		            </table>
		        </div>
		    </div>
		</div>;
	}

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/v1/metrics', {})
    .then(res => {
      this.setState({metrics: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }
}
