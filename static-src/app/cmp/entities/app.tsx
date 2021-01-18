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

	    // TODO: More dynamic years?
	    let years:string[] = [
		    new Date().getFullYear()-1,
		    new Date().getFullYear()
	    ];
	    for (var key in this.state.entities) {
	    	if (! this.state.entities.hasOwnProperty(key)) {
	    		// ignore
	    		continue;
	    	}
			let ukey = "entity_" + key;
			let entity = this.state.entities[key];

			items.push(<tr key={ukey+"company"}><td colSpan={2}>{entity.Name} - {entity.VAT}</td></tr>);
			for (var k in years) {
				if (! years.hasOwnProperty(k)) {
		    		// ignore
		    		continue;
		    	}
		    	let year:string = years[k];
				items.push(<tr key={ukey+"company"+year}><td><a href={"#"+key+"/"+year}>{year}</a></td><td>0,00EUR</td></tr>);
			}
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
