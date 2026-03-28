import * as React from "react";
import Axios from "axios";
import {ActionButton} from "../../shared/ActionButton";

interface TaxData {
  Ex: string;
  Tax: string;
  EUEx: string;
  EUCompany: Record<string, string>;
}

interface TaxesPageProps {
  entity: string;
  year: string;
}

interface TaxesPageState {
  quarter: string;
  data: TaxData | null;
}

export default class TaxesPage extends React.Component<TaxesPageProps, TaxesPageState> {
  constructor(props: TaxesPageProps) {
    super(props);
    this.state = {quarter: "Q1", data: null};
  }

  componentDidMount(): void {
    //this.ajax();
  }

  private async ajax(): Promise<void> {
    const res = await Axios.post('/api/v1/taxes/'+this.props.entity+'/'+this.props.year+'/'+this.state.quarter);
    this.setState({data: res.data});
  }

  private setQuarter(e: React.ChangeEvent<HTMLSelectElement>): void {
    this.setState({quarter: e.target.value});
  }

  render(): React.JSX.Element {
    let sum = null;
    if (this.state.data) {
      const icp = [];
      for (const vat in this.state.data.EUCompany) {
        if (!Object.prototype.hasOwnProperty.call(this.state.data.EUCompany, vat)) {
          continue;
        }
        icp.push(<tr key={"icp"+vat}>
          <td>3a. Intracommunautaire leveringen en diensten</td>
          <td><code>{vat}</code></td>
          <td className="text-right">&euro; {this.state.data.EUCompany[vat]}</td>
        </tr>);
      }

      sum = <div className="m-t-lg">
        <div className="hpanel hgreen">
          <div className="panel-heading hbuilt">
            <i className="fas fa-file-invoice"></i> Aangifte omzetbelasting
          </div>
          <div className="panel-body">
            <table className="table table-striped">
              <thead>
                <tr><th>Rubriek</th><th className="text-right">Omzet</th><th className="text-right">Omzetbelasting</th></tr>
              </thead>
              <tbody>
                <tr>
                  <td>1a. Leveringen/diensten belast met hoog tarief</td>
                  <td className="text-right">&euro; {this.state.data.Ex}</td>
                  <td className="text-right">&euro; {this.state.data.Tax}</td>
                </tr>
                <tr>
                  <td>3b. Leveringen naar of diensten in landen binnen de EU</td>
                  <td className="text-right">&euro; {this.state.data.EUEx}</td>
                  <td></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        {icp.length > 0 && <div className="hpanel horange m-t-md">
          <div className="panel-heading hbuilt">
            <i className="fas fa-globe-europe"></i> Aangifte Intracommunautaire prestaties (ICP)
          </div>
          <div className="panel-body">
            <table className="table table-striped">
              <thead>
                <tr><th>Rubriek</th><th>BTW-nummer</th><th className="text-right">Levering</th></tr>
              </thead>
              <tbody>
                {icp}
              </tbody>
            </table>
          </div>
        </div>}
      </div>;
    }

    return <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <i className="fas fa-calculator"></i> Dutch TAX Summary Generator
          <div className="panel-tools">
            <div className="btn-group nm7">
              <a className="btn btn-default btn-hover-success" href={'/api/v1/summary/'+this.props.entity+'/'+this.props.year+'?excel=1'}>
                <i className="fas fa-file-excel"></i> XLSX Accountant
              </a>
            </div>
          </div>
        </div>
        <div className="panel-body">
          <div className="row">
            <div className="col-md-8">
              <div className="form-inline">
                <div className="form-group m-r-sm">
                  <label className="m-r-sm">Quarter</label>
                  <select className="form-control" value={this.state.quarter} onChange={this.setQuarter.bind(this)}>
                    <option value="Q1">Q1 (Jan - Mar)</option>
                    <option value="Q2">Q2 (Apr - Jun)</option>
                    <option value="Q3">Q3 (Jul - Sep)</option>
                    <option value="Q4">Q4 (Oct - Dec)</option>
                  </select>
                </div>
                <ActionButton className="btn btn-primary" onClick={this.ajax.bind(this)}>
                  <i className="fas fa-calculator"></i> Get numbers
                </ActionButton>
              </div>
            </div>
            <div className="col-md-4">
              <div className="hpanel">
                <div className="panel-heading hbuilt">
                  <i className="fas fa-calendar"></i> Quarter Reference
                </div>
                <div className="panel-body">
                  <table className="table table-bordered table-striped">
                    <thead>
                      <tr><th>Quarter</th><th>Months</th><th>Deadline</th></tr>
                    </thead>
                    <tbody>
                      <tr className={this.state.quarter === "Q1" ? "active" : ""}>
                        <td><strong>Q1</strong></td>
                        <td>Jan - Mar</td>
                        <td>30 April</td>
                      </tr>
                      <tr className={this.state.quarter === "Q2" ? "active" : ""}>
                        <td><strong>Q2</strong></td>
                        <td>Apr - Jun</td>
                        <td>31 July</td>
                      </tr>
                      <tr className={this.state.quarter === "Q3" ? "active" : ""}>
                        <td><strong>Q3</strong></td>
                        <td>Jul - Sep</td>
                        <td>31 October</td>
                      </tr>
                      <tr className={this.state.quarter === "Q4" ? "active" : ""}>
                        <td><strong>Q4</strong></td>
                        <td>Oct - Dec</td>
                        <td>31 January</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      {sum}
    </div>;
  }
}
