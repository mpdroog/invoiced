import * as React from "react";
import Axios from "axios";
import {DOM} from "../lib/dom";
import { LineChart, Line, XAxis, YAxis, Tooltip } from 'recharts';

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
		var years = {"2016": 0, "2017": 0};
		for (var i = 0; i < sorted.length; i++) {
			var key = sorted[i];
			var revenue = this.state.metrics[key].RevenueEx;
			var delta = (((revenue*100) - (pref*100)) / 100).toFixed(0);
			var change = {};
			if (delta > 0) {
				change = {backgroundColor: "green", color: "white"};
			}
			items.push(<tr key={key}><td>{key}</td><td>&euro; {revenue}</td><td style={change}>&euro; {delta}</td></tr>);

			pref = revenue;
			console.log(revenue*100);
			years[ key.substr(0, key.indexOf("-")) ] += revenue*100;
		}
		console.log(years);

		var stats = [];
		for (var i = 0; i < sorted.length; i++) {
			var key = sorted[i];
			var vals = this.state.metrics[key];
			vals.RevenueEx = parseInt(vals.RevenueEx);
			vals.name = key;
			stats.push(vals);
		}

		var smallHead = {
			fontSize: "12px",
			float: "right",
			border: "1px solid gray",
			padding: "10px",
			marginLeft: "5px"
		};
		return <div>
			<div className="normalheader col-md-6">
			    <div className="hpanel">
			        <div className="panel-body">
			            <h2 className="font-light m-b-xs">
			                <i className="fa fa-bank"></i>
			                Revenue
			                <span style={smallHead}>2017: &euro; {years[2017]/100}</span>
			                <span style={smallHead}>2016: &euro; {years[2016]/100}</span>
			            </h2>
			            <table className="table">
			            	<thead>
			            		<tr><th>Date</th><th>Revenue</th><th>Δ</th></tr>
			            	</thead>
			            	<tbody>{items}</tbody>
			            </table>
			        </div>
			    </div>
			</div>

			<div className="normalheader col-md-6">
			    <div className="hpanel">
			        <div className="panel-body">
			            <h2 className="font-light m-b-xs pa">
			            	<i className="fa fa-area-chart"></i>
			                Graph
			            </h2>
			            <LineChart width={600} height={200} data={stats}>
						<XAxis dataKey="name"/>
						<Line type="monotone" dataKey="RevenueEx" stroke='#82ca9d' fill='#82ca9d' />
						<Tooltip/>
						</LineChart>
			        </div>
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
