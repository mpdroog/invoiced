import * as React from "react";
import Big from "big.js";
import {DOM} from "../../lib/dom";

export class InvoiceLineEdit extends React.Component<{}, {}> {
  private revisions: IInvoiceState[];
  constructor(props) {
    super(props);
  }

  private lineAdd(e) {
    e.preventDefault();
    let parent = this.props.parent;
    if (parent.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    console.log("Add invoice line");
    parent.state.Lines.push({
      Description: "",
      Quantity: "0.00",
      Price: "0.00",
      Total: "0.00"
    });
    parent.setState({Lines: parent.state.Lines});
  }

  private lineRemove(e) {
    e.preventDefault();
    let node = DOM.eventFilter(e, "BUTTON");
    let key = node.dataset["idx"];
    let parent = this.props.parent;

    if (parent.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    let line: IInvoiceLine = parent.state.Lines[key];
    let isEmpty = line.Description === ""
      && line.Quantity === "0.00"
      && line.Price === "0.00"
      && line.Total === "0.00";
    let isOk = !isEmpty && confirm(`Are you sure you want to remove the invoiceline with description '${line.Description}'?`);

    if (isEmpty || isOk) {
      console.log(`Deleted idx (${key})`, parent.state.Lines.splice(key, 1)[0]);
      parent.setState({Lines: parent.state.Lines});
    }
  }

  render() {
  	let that = this;
  	let parent = this.props.parent;
  	let inv = parent.state;
  	let lines = [];
	inv.Lines.forEach(function(line: IInvoiceLine, idx: number) {
      lines.push(
        <tr key={"line"+idx}>
          <td><button disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-danger faa-parent animated-hover' : '')} onClick={that.lineRemove.bind(that)} data-idx={idx}><i className="fa fa-trash faa-flash"></i></button></td>
          <td className="pr"><input className="form-control" type="text" data-key={"Lines."+idx+".Description"} onChange={parent.handleChange.bind(parent)} value={line.Description}/><i className="fa fa-asterisk text-danger fa-input"></i></td>
          <td className="pr"><input className="form-control" type="text" data-key={"Lines."+idx+".Quantity"} onChange={parent.handleChange.bind(parent)} value={line.Quantity}/><i className="fa fa-asterisk text-danger fa-input"></i></td>
          <td className="pr"><input className="form-control" type="text" data-key={"Lines."+idx+".Price"} onChange={parent.handleChange.bind(parent)} value={line.Price}/><i className="fa fa-asterisk text-danger fa-input"></i></td>
          <td><input className="form-control" readOnly={true} disabled={true} type="text" data-key={"Lines."+idx+".Total"} value={line.Total}/></td>
        </tr>
      );
    });

    return <table className="table table-striped">
    <thead>
      <tr>
        <th>&nbsp;</th>
        <th>Description</th>
        <th>Quantity</th>
        <th>Price</th>
        <th>Line Total</th>
      </tr>
    </thead>
    <tbody>{lines}</tbody>
    <tfoot>
      <tr>
        <td colSpan={3} className="text">
          <button disabled={inv.Meta.Status === 'FINAL'} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-success faa-parent animated-hover' : '')} onClick={that.lineAdd.bind(that)}><i className="fa fa-plus faa-bounce"></i> Add row</button>
        </td>
        <td className="text">Total (ex tax)</td>
        <td><input className="form-control" disabled={true} type="text" data-key="Total.Ex" readOnly={true} value={inv.Total.Ex}/></td>
      </tr>
      <tr>
        <td colSpan={3}></td>
        <td className="text">Tax (21%)</td>
        <td><input className="form-control" onChange={parent.handleChange.bind(parent)} disabled={true} type="text" data-key="Total.Tax" readOnly={true} value={inv.Total.Tax}/></td>
      </tr>
      <tr>
        <td colSpan={3}>&nbsp;</td>
        <td className="text">Total</td>
        <td><input className="form-control" onChange={parent.handleChange.bind(parent)} disabled={true} type="text" data-key="Total.Total" readOnly={true} value={inv.Total.Total}/></td>
      </tr>
    </tfoot>
  </table>
  }
}