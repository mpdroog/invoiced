import * as React from "react";
import {DOM} from "../../lib/dom";
import type {IInvoiceLine, IInvoiceState} from "./edit-struct";

interface InvoiceLineEditProps {
  parent: {
    state: IInvoiceState;
    setState: (state: Partial<IInvoiceState>) => void;
    handleChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
    pushUndo: () => void;
  };
}

export class InvoiceLineEdit extends React.Component<InvoiceLineEditProps, Record<string, never>> {
  constructor(props: InvoiceLineEditProps) {
    super(props);
  }

  private lineAdd(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    const parent = this.props.parent;
    if (parent.state.Meta?.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    parent.pushUndo();
    const lines = [...(parent.state.Lines ?? [])];
    console.log("Add invoice line");
    lines.push({
      Description: "",
      Quantity: "0.00",
      Price: "0.00",
      Total: "0.00"
    });
    parent.setState({Lines: lines});
  }

  private lineRemove(e: React.MouseEvent<HTMLButtonElement>): void {
    e.preventDefault();
    const node = DOM.eventFilter(e, "BUTTON");
    if (!node) return;
    const key = node.dataset["idx"];
    if (key === undefined) return;
    const parent = this.props.parent;

    if (parent.state.Meta?.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    const keyNum = parseInt(key, 10);
    const lines = parent.state.Lines;
    if (!lines) return;

    parent.pushUndo();
    const newLines = [...lines];
    console.log(`Deleted idx (${key})`, newLines.splice(keyNum, 1)[0]);
    parent.setState({Lines: newLines});
  }

  render(): React.JSX.Element {
  	const that = this;
  	const parent = this.props.parent;
  	const inv = parent.state;
  	const invLines = inv.Lines ?? [];
  	const invTotal = inv.Total ?? { Ex: "0.00", Tax: "0.00", Total: "0.00" };
  	const invStatus = inv.Meta?.Status;
  	const lines: React.JSX.Element[] = [];
	invLines.forEach(function(line: IInvoiceLine, idx: number) {
      lines.push(
        <tr key={"line"+idx}>
          <td><button type="button" disabled={invStatus === 'FINAL'} className={"btn " + (invStatus !== 'FINAL' ? 'btn-danger' : 'btn-secondary')} onClick={that.lineRemove.bind(that)} data-idx={idx}><i className="fas fa-trash"></i></button></td>
          <td className="pr"><input className="form-control" type="text" data-key={"Lines."+idx+".Description"} onChange={parent.handleChange.bind(parent)} value={line.Description}/><i className="fas fa-asterisk text-danger fa-input"></i></td>
          <td className="pr"><input className="form-control" type="text" data-key={"Lines."+idx+".Quantity"} onChange={parent.handleChange.bind(parent)} value={line.Quantity}/><i className="fas fa-asterisk text-danger fa-input"></i></td>
          <td className="pr"><input className="form-control" type="text" data-key={"Lines."+idx+".Price"} onChange={parent.handleChange.bind(parent)} value={line.Price}/><i className="fas fa-asterisk text-danger fa-input"></i></td>
          <td><input className="form-control" readOnly={true} disabled={true} type="text" data-key={"Lines."+idx+".Total"} value={line.Total}/></td>
        </tr>
      );
    });

    return <div className="table-responsive mb-3">
    <table className="table table-striped table-sm">
    <thead>
      <tr>
        <th style={{width: '50px'}}></th>
        <th>Description</th>
        <th style={{width: '100px'}}>Qty</th>
        <th style={{width: '100px'}}>Price</th>
        <th style={{width: '120px'}}>Total</th>
      </tr>
    </thead>
    <tbody>{lines}</tbody>
    <tfoot>
      <tr>
        <td colSpan={3} className="text">
          <button type="button" disabled={invStatus === 'FINAL'} className={"btn " + (invStatus !== 'FINAL' ? 'btn-success' : 'btn-secondary')} onClick={that.lineAdd.bind(that)}><i className="fas fa-plus"></i> Add row</button>
        </td>
        <td className="text">Total (ex tax)</td>
        <td><input className="form-control" disabled={true} type="text" data-key="Total.Ex" readOnly={true} value={invTotal.Ex}/></td>
      </tr>
      <tr>
        <td colSpan={3}></td>
        <td className="text">Tax (21%)</td>
        <td><input className="form-control" onChange={parent.handleChange.bind(parent)} disabled={true} type="text" data-key="Total.Tax" readOnly={true} value={invTotal.Tax}/></td>
      </tr>
      <tr>
        <td colSpan={3}>&nbsp;</td>
        <td className="text">Total</td>
        <td><input className="form-control" onChange={parent.handleChange.bind(parent)} disabled={true} type="text" data-key="Total.Total" readOnly={true} value={invTotal.Total}/></td>
      </tr>
    </tfoot>
  </table>
  </div>
  }
}