import * as React from "react";
import Axios from "axios";
import ChartistGraph from 'react-chartist';

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

var formatMoney = function(n, c, d, t){
var c = isNaN(c = Math.abs(c)) ? 2 : c, 
    d = d == undefined ? "." : d, 
    t = t == undefined ? "," : t, 
    s = n < 0 ? "-" : "", 
    i = String(parseInt(n = Math.abs(Number(n) || 0).toFixed(c))), 
    j = (j = i.length) > 3 ? j % 3 : 0;
   return s + (j ? i.substr(0, j) + t : "") + i.substr(j).replace(/(\d{3})(?=\d)/g, "$1" + t) + (c ? d + Math.abs(n - i).toFixed(c).slice(2) : "");
};

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

		let pref:string = "0.00";
		let sum:number = 0;
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

			revstats.labels.push(key);
			revstats.series[0].push(parseInt(revenue));
			hourstats.labels.push(key);
			hourstats.series[0].push(parseInt(hours));

			pref = revenue;
			sum += revenue*100;
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
			                <span style={smallHead}>{fullyear}: &euro; {formatMoney(sum/100, '.', ',', ' ')}</span>
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
