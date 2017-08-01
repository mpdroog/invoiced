import * as React from "react";
import Axios from "axios";
//import {DOM} from "../lib/dom";
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
	    let items:React.JSX.Element[] = [];
		let sorted:string[] = Object.keys(this.state.metrics).sort();

		let pref:string = "0.00";
		let years = {"2016": 0, "2017": 0};
		for (var i = 0; i < sorted.length; i++) {
			let key:string = sorted[i];
			let revenue:number = this.state.metrics[key].RevenueEx;
			let delta:number = (((+revenue*100) - (+pref*100)) / 100).toFixed(0);
			let hours:number = this.state.metrics[key].Hours;
			var change = {};
			if (delta > 0) {
				change = {backgroundColor: "green", color: "white"};
			}
			items.push(<tr key={key}><td>{key}</td><td>&euro; {revenue}</td><td style={change}>&euro; {delta}</td><td>{hours}</td></tr>);

			pref = revenue;
			years[ key.substr(0, key.indexOf("-")) ] += revenue*100;
		}

		let stats = [];
		for (var i = 0; i < sorted.length; i++) {
			let key = sorted[i];
			let vals = this.state.metrics[key];
			vals.RevenueEx = parseInt(vals.RevenueEx);
			vals.Hours = parseInt(vals.Hours);
			vals.name = key;
			stats.push(vals);
		}

		let smallHead = {
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
			            		<tr><th>Date</th><th>Revenue</th><th>Δ</th><th>Hours</th></tr>
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
			                Revenue Graph
			            </h2>
			            <LineChart width={600} height={200} data={stats}>
							<XAxis dataKey="name"/>
							<Line type="monotone" dataKey="RevenueEx" stroke='#82ca9d' fill='#82ca9d' />
							<Tooltip/>
						</LineChart>
			        </div>
			    </div>
			</div>

			<div className="normalheader col-md-6">
			    <div className="hpanel">
			        <div className="panel-body">
			            <h2 className="font-light m-b-xs pa">
			            	<i className="fa fa-area-chart"></i>
			                Hour Graph
			            </h2>
			            <LineChart width={600} height={200} data={stats}>
							<XAxis dataKey="name"/>
							<Line type="monotone" dataKey="Hours" stroke='#82ca9d' fill='#82ca9d' />
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
