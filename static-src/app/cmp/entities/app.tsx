import * as React from "react";
import Axios from "axios";

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
			for (var k in entity.Years) {
				if (! entity.Years.hasOwnProperty(k)) {
		    	// ignore
		    	continue;
		    }
		    let year:int = parseInt(entity.Years[k]);
		    if (year === curYear) {
		    	hasCurYear = true;
		    }
				accountingYears.push(<tr key={ukey+"company"+year}><td><a href={"#"+key+"/"+year}>{year}</a></td><td>0,00EUR</td></tr>);
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
