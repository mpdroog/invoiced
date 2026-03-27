import * as React from "react";
import Axios from "axios";
import ChartistGraph from 'react-chartist';
import './chartist.css';

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

// Format number with space as thousands separator: 51868.65 -> 51 868,65
function formatCurrency(value: number): string {
	let parts = value.toFixed(2).split(".");
	parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, " ");
	return parts.join(",");
}

export default class Dashboard extends React.Component<{}, IMetrics> {
  constructor(p, s) {
    super(p, s);
    this.state = {
      "metrics": {}
    };
  }

	render() {
	    let items:React.JSX.Element[] = [];
		let sorted:string[] = Object.keys(this.state.metrics).sort();
		let revstats = {labels: [], series: [[]]};
		let hourstats = {labels: [], series: [[]]};

		let prevRevenue:number = 0;
		let sum:number = 0;
		for (var i = 0; i < sorted.length; i++) {
			let key:string = sorted[i];
			let revenue:number = parseFloat(this.state.metrics[key].RevenueEx) || 0;
			let hours:number = parseFloat(this.state.metrics[key].Hours) || 0;

			// Calculate delta vs previous month
			let deltaEl: React.JSX.Element = null;
			if (prevRevenue > 0) {
				let delta = revenue - prevRevenue;
				let pct = ((delta / prevRevenue) * 100).toFixed(0);
				let sign = delta >= 0 ? "+" : "";
				let badgeClass = delta >= 0 ? "m-l-sm label label-success" : "m-l-sm label label-danger";
				deltaEl = <span className={badgeClass}>
					{sign}&euro; {formatCurrency(Math.abs(delta))} ({sign}{pct}%)
				</span>;
			}

			items.push(<tr key={key}>
				<td>{key}</td>
				<td>&euro; {formatCurrency(revenue)}{deltaEl}</td>
				<td>{formatCurrency(hours)}</td>
			</tr>);

			revstats.labels.push(key);
			revstats.series[0].push(revenue);
			hourstats.labels.push(key);
			hourstats.series[0].push(hours);

			prevRevenue = revenue;
			sum += revenue * 100;
		}

		let smallHead = {
			fontSize: "12px",
			float: "right",
			border: "1px solid gray",
			padding: "10px",
			marginLeft: "5px"
		};
		let options = {
	      axisX: {
	        labelInterpolationFnc: function(value, index) {
	          return index % 2 === 0 ? value : null;
	        }
	      }
	    };
	    let fullyear = new Date().getFullYear();
		return <div>
			<div className="normalheader col-md-6">
			    <div className="hpanel">
			        <div className="panel-body">
			            <h2 className="font-light m-b-xs">
			                <i className="fa fa-bank"></i>
			                Revenue
			                <span style={smallHead}>{fullyear}: &euro; {formatCurrency(sum/100)}</span>
			            </h2>
			            <table className="table">
			            	<thead>
			            		<tr><th>Date</th><th>Revenue</th><th>Hours</th></tr>
			            	</thead>
			            	<tbody>{items}</tbody>
			            </table>
			        </div>
			    </div>
			</div>

			<div className="normalheader col-md-6">
			    <div className="hpanel">
			        <div className="panel-body">
			            <h2 className="font-light m-b-xs">
			            	<i className="fa fa-area-chart"></i>
			                Revenue Graph
			            </h2>
			            <ChartistGraph data={revstats} options={options} type={"Line"} />
			        </div>
			    </div>
			</div>

			<div className="normalheader col-md-6">
			    <div className="hpanel">
			        <div className="panel-body">
			            <h2 className="font-light m-b-xs">
			            	<i className="fa fa-area-chart"></i>
			                Hour Graph
			            </h2>
						<ChartistGraph data={hourstats} options={options} type={"Line"} />
			        </div>
			    </div>
			</div>

		</div>;
	}

  componentDidMount() {
  	console.log('componentDidMount');
    this.ajax();
  }

  private ajax() {
  	let entity = this.props.entity;
  	let year = this.props.year;

    Axios.get('/api/v1/metrics/'+entity+'/'+year, {})
    .then(res => {
      this.setState({metrics: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }
}
