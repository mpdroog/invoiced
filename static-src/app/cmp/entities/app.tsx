import * as React from "react";
import Axios from "axios";

// Format number with space as thousands separator: 51868.65 -> 51 868.65
function formatCurrency(value: number): string {
	let parts = value.toFixed(2).split(".");
	parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, " ");
	return parts.join(",");
}

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

export default class Entities extends React.Component<{}, IMetrics> {
  constructor() {
    super();
    this.state = {
      "entities": []
    };
  }

	render() {
    let items:React.JSX.Element[] = [];
    let curYear:int = new Date().getFullYear();
    for (var key in this.state.entities) {
    	if (! this.state.entities.hasOwnProperty(key)) {
    		// ignore
    		continue;
    	}
			let ukey = "entity_" + key;
			let entity = this.state.entities[key];

			let hasCurYear: bool = false;
			let accountingYears: React.JSX.Element[] = [];
			let prevRevenue: number = 0;

			// Sort years ascending for proper comparison
			let sortedYears = (entity.Years || []).slice().sort((a, b) => parseInt(a) - parseInt(b));

			for (var k in sortedYears) {
				if (! sortedYears.hasOwnProperty(k)) {
		    	// ignore
		    	continue;
		    }
		    let yearStr = sortedYears[k];
		    let year:int = parseInt(yearStr);
		    if (year === curYear) {
		    	hasCurYear = true;
		    }

				let revenueStr = entity.YearRevenue && entity.YearRevenue[yearStr] ? entity.YearRevenue[yearStr] : "0.00";
				let revenue = parseFloat(revenueStr);
				let formattedRevenue = formatCurrency(revenue);

				// Calculate delta vs previous year
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

				accountingYears.push(<tr key={ukey+"company"+year}>
					<td><a href={"#"+key+"/"+year}>{year}</a></td>
					<td>&euro; {formattedRevenue}{deltaEl}</td>
				</tr>);

				prevRevenue = revenue;
			}

			let link: React.JSX.Element = null;
			if (! hasCurYear) {
				link = <a href={"/api/v1/entities/" + key + "/open/" + curYear}>Open {curYear}</a>;
			}
			items.push(<tr key={ukey+"company"}><td>{entity.Name} - {entity.VAT}</td><td>{link}</td></tr>);
			items = items.concat(accountingYears);
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
			                <i className="fa fa-building"></i>
			                &nbsp;Your Companies
			            </h2>
			            <table className="table">
			            	<thead>
			            		<tr><th>Company</th><th>Revenue</th></tr>
			            	</thead>
			            	<tbody>{items}</tbody>
			            </table>
			        </div>
			    </div>
			</div>
		</div>;
	}

  componentDidMount() {
    this.ajax();
  }

  private ajax() {
    Axios.get('/api/v1/entities', {})
    .then(res => {
      	this.setState({entities: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }
}
