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
      let icp = [];
      for (let vat in this.state.data.EUCompany) {
        if (! this.state.data.EUCompany.hasOwnProperty(vat)) {
          continue;
        }
        icp.push(<tr key={"icp"+vat}>
          <td>3a. Intracommunautaire leveringen en diensten</td>
          <td>{vat}</td>
          <td>&euro; {this.state.data.EUCompany[vat]}</td>
        </tr>);
      };

      sum = <div><div><h2>Aangifte omzetbelasting</h2><table className="table">
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
      </table></div><div>
        <h2>Aangifte Intracommunautaire prestaties</h2><table className="table">
        <thead>
          <tr><th>Rubriek</th><th>BTW-nummer</th><th>Levering</th></tr>
        </thead>
        {icp}
      </table>
      </div></div>;
    }

		return <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          Dutch TAX Summary Generator
        </div>
        <div className="panel-body">
          <div className="row">
          <div className="col-md-8">
            <input type="text" value={this.state.quarter} onChange={this.setQuarter.bind(this)} />
            <button type="submit" onClick={this.ajax.bind(this)}>Get numbers</button>

            {sum}
          </div>
          <div className="col-md-4">
            <table className="table table-bordered">
              <thead>
                <tr><th>Quarter</th><th>F - T</th><th>From - To (including)</th><th>Deadline</th></tr>
              </thead>
              <tr>
                <td>Q1</td>
                <td>01-03</td>
                <td>Januari - March</td>
                <td>31 April</td>
              </tr>
              <tr>
                <td>Q2</td>
                <td>04-06</td>
                <td>April - June</td>
                <td>31 July</td>
              </tr>
              <tr>
                <td>Q3</td>
                <td>07-09</td>
                <td>July - September</td>
                <td>31 October</td>
              </tr>
              <tr>
                <td>Q4</td>
                <td>10-12</td>
                <td>October - December</td>
                <td>31 Januari</td>
              </tr>
            </table>
          </div>
          </div>
        </div>
      </div>
    </div>;
	}
}
