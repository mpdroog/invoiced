import * as React from "react";
import Axios from "axios";

export default class TaxesPage extends React.Component<{}, {}> {
  constructor(props) {
    super(props);
    this.state = {quarter: "Q1", data: null};
  }

  componentDidMount() {
    //this.ajax();
  }

  private ajax() {
    var that = this;
    Axios.post('/api/v1/taxes/'+this.props.entity+'/'+this.props.year+'/'+this.state.quarter)
    .then(res => {
      that.setState({data: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }

  private setQuarter(e) {
    console.log(e.target.value);
    this.setState({quarter: e.target.value});
  }

	render() {
    let sum = null;
    if (this.state.data) {
      sum = <table className="table">
        <thead>
          <tr><th>Rubriek</th><th>Omzet</th><th>Omzetbelasting</th></tr>
        </thead>
        <tr>
          <td>1a. Leveringen/diensten belast met hoog tarief</td>
          <td>&euro; {this.state.data.Ex}</td>
          <td>&euro; {this.state.data.Tax}</td>
        </tr>
        <tr>
          <td>3b. Leveringen naar of diensten in landen binnen de EU</td>
          <td>&euro; {this.state.data.EUEx}</td>
        </tr>
      </table>;
    }

		return <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          TAX Summary Generator
        </div>
        <div className="panel-body">
          <div className="row">
            <input type="text" value={this.state.quarter} onChange={this.setQuarter.bind(this)} />
            <button type="submit" onClick={this.ajax.bind(this)}>Get numbers</button>

            {sum}
          </div>
        </div>
      </div>
    </div>;
	}
}
