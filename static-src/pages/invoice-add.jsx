'use strict';
var React = require('react');
var Request = require('superagent');
var Big = require('big.js');
var Moment = require('moment');

require('./invoice.css');
var DatePicker = require('react-datepicker');
require('react-datepicker/dist/react-datepicker.css');

module.exports = React.createClass({
  getInitialState: function() {
      this.revisions = [];
      return {
        Company: "RootDev",
        Entity: {
          Name: "M.P. Droog",
          Street1: "Dorpsstraat 236a",
          Street2: "Obdam, 1713HP, NL"
        },
        Customer: {
          Name: "XSNews B.V.",
          Street1: "New Yorkstraat 9-13",
          Street2: "1175 RD Lijnden"
        },
        Meta: {
          Conceptid: "",
          Status: "NEW",
          Invoiceid: "",
          InvoiceidL: true,
          Issuedate: null,
          IssuedateL: true,
          Ponumber: "",
          Duedate: Moment().add(14, 'days')
        },
        Lines: [{
          Description: "",
          Quantity: "0",
          Price: "0.00",
          Total: "0.00"
        }],
        Notes: "",
        Total: {
          Ex: "0.00",
          Tax: "0.00",
          Total: "0.00"
        },
        Bank: {
          Vat: "",
          Coc: "",
          Iban: ""
        }
      };
  },

  componentDidMount: function() {
    if (this.props.args.length > 0) {
      console.log("Load invoice name=" + this.props.args[0]);
      this.ajax(this.props.args[0]);
    }
  },

  ajax: function(name) {
    var that = this;
    Request.get('/api/invoice/'+name)
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          handleErr(err);
          return;
        }
        if (that.isMounted()) {
          that.parseInput.call(that, res.body);
        }
    });
  },

  parseInput: function(data) {
    console.log(data);
    if (window.location.href != "?#invoice-add/" + data.Meta.Conceptid) {
      // Update URL so refresh will keep the invoice open
      history.replaceState({}, "", "?#invoice-add/" + data.Meta.Conceptid);
      this.props.args.push(data.Meta.Conceptid);
    }
    data.Meta.Issuedate = data.Meta.Issuedate ? Moment(data.Meta.Issuedate) : null;
    data.Meta.Duedate = data.Meta.Duedate ? Moment(data.Meta.Duedate) : null;
    data.Meta.InvoiceidL = true;
    data.Meta.IssuedateL = true;
    this.setState(data);
  },

  lineAdd: function() {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    console.log("Add invoice line");
    this.state.Lines.push({
      Description: "",
      Quantity: "0",
      Price: "0.00",
      Total: "0.00"
    });
    this.setState({Lines: this.state.Lines});
  },
  lineRemove: function(key) {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    var line = this.state.Lines[key];
    var isEmpty = line.Description === ""
      && line.Quantity === "0"
      && line.Price === "0.00"
      && line.Total === "0.00";
    var isOk = !isEmpty && confirm("Are you sure you want to remove the invoiceline with description '" + line.Description + "'?");

    if (isEmpty || isOk) {
      console.log("Remove invoice line with key=" + key);
      console.log("Deleted idx ", this.state.Lines.splice(key, 1)[0]);
      this.setState({Lines: this.state.Lines});
    }
  },
  lineUpdate: function(line) {
    line.Total = new Big(line.Price).times(line.Quantity).round(2).toFixed(2).toString();
    return line;
  },
  totalUpdate: function(lines) {
    var ex = new Big(0);
    lines.forEach(function(val) {
      console.log("Add", val.Total);
      ex = ex.plus(val.Total);
    });
    // TODO: Hardcoded to 21%
    var tax = ex.div("100").times("21");
    var total = ex.plus(tax);
    console.log("totals (ex,tax,total)", ex.toString(), tax.toString(), total.toString());

    return {
      Ex: ex.round(2).toFixed(2).toString(),
      Tax: tax.round(2).toFixed(2).toString(),
      Total: total.round(2).toFixed(2).toString()
    };
  },

  triggerChange: function(indices, val) {
    if (this.state.Meta.Status === 'FINAL') {
      console.log("Finalized, not allowing changes!");
      return;
    }
    var node = this.state;
    for (var i = 0; i < indices.length-1; i++) {
      node = node[ indices[i] ];
    }
    node[indices[indices.length-1]] = val;

    // Any post-processing
    if (indices[0] === "Lines") {
      this.state.Lines[indices[1]] = this.lineUpdate(this.state.Lines[indices[1]]);
      this.state.Total = this.totalUpdate(this.state.Lines);
    }
    this.setState(this.state);
    this.revisions.push({}); // TODO :)
  },

  handleChange: function(e) {
    console.log("handleChange", e.target.dataset.key);
    var indices = e.target.dataset.key.split('.');
    this.triggerChange(indices, e.target.value);
  },

  handleChangeDate: function(id) {
    var indices = id.split('.');
    var that = this;
    return function(val) {
      console.log("handleChangeDate", id, val);
      that.triggerChange.call(that, indices, val);
    };
  },
  toggleChange: function(id, val) {
    var indices = id.split('.');
    var that = this;
    val = !val; // Invert value
    return function() {
      console.log("toggleChange", id, val);
      that.triggerChange.call(that, indices, val);
    };
  },

  save: function(e) {
    var that = this;
    var req = JSON.parse(JSON.stringify(this.state)); // deepCopy
    req.Meta.Issuedate = this.state.Meta.Issuedate ? this.state.Meta.Issuedate.format('YYYY-MM-DD') : "";
    req.Meta.Duedate = this.state.Meta.Duedate ? this.state.Meta.Duedate.format('YYYY-MM-DD') : "";
    console.log(req);

    Request.post('/api/invoice')
    .send(req)
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          console.log(err);
          handleErr(err);
          return;
        }
        if (that.isMounted()) {
          that.parseInput.call(that, res.body);
        }
    });
  },

  finalize: function(e) {
    var that = this;

    Request.get('/api/invoice/' + this.state.Meta.Conceptid + '/finalize')
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          console.log(err);
          handleErr(err);
          return;
        }
        if (that.isMounted()) {
          console.log("Finalized");
          that.parseInput.call(that, res.body);
        }
    });
  },

  pdf: function() {
    if (this.state.Meta.Status !== 'FINAL') {
      console.log("PDF only available in finalized invoices");
      return;
    }
    var url = '/api/invoice/'+this.props.args[0]+'/pdf';
    console.log("Open PDF " + url);
    location.assign(url);
  },

	render: function() {
    var inv = this.state;
    var that = this;
    var lines = [];

    inv.Lines.forEach(function(line, idx) {
      console.log(inv.Meta.Status);
      lines.push(
        <tr key={"line"+idx}>
          <td><button disabled={inv.Meta.Status === 'FINAL' ? 'disabled': ''} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-danger faa-parent animated-hover' : '')} onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash faa-flash"></i></button></td>
          <td><input className="form-control" type="text" data-key={"Lines."+idx+".Description"} onChange={that.handleChange} value={line.Description}/></td>
          <td><input className="form-control" type="text" data-key={"Lines."+idx+".Quantity"} onChange={that.handleChange} value={line.Quantity}/></td>
          <td><input className="form-control" type="text" data-key={"Lines."+idx+".Price"} onChange={that.handleChange} value={line.Price}/></td>
          <td><input className="form-control" readOnly="readOnly" disabled="disabled" type="text" data-key={"Lines."+idx+".Total"} readOnly="readOnly" value={line.Total}/></td>
        </tr>
      );
    });

		return <form><div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <div className="btn-group nm7">
                <button className="btn btn-default btn-hover-success" disabled={this.revisions.length > 0 ? '' : 'disabled'} onClick={this.save}><i className="fa fa-floppy-o"></i> Save</button>
                <button className="btn btn-default btn-hover-danger" disabled={inv.Meta.Status === "CONCEPT" ? '' : 'disabled'} onClick={this.finalize}><i className="fa fa-lock"></i> Finalize</button>
                <a className="btn btn-default btn-hover-success" disabled={inv.Meta.Status === "FINAL" ? '' : 'disabled'} onClick={this.pdf}><i className="fa fa-file-pdf-o"></i> PDF</a>
              </div>

            </div>
            New Invoice
          </div>
          <div className="panel-body">

<div className={"invoice group " + (inv.Meta.Status === 'FINAL' ? 'o50' : '')}>
  <div className="row">
    <div className="company col-sm-4">
      <input className="form-control" type="text" data-key="Company" onChange={that.handleChange} value={inv.Company}/>
    </div>

    <div className="col-sm-offset-3 col-sm-1">
      From
    </div>
    <div className="entity col-sm-4">
      <input className="form-control" type="text" data-key="Entity.Name" onChange={that.handleChange} value={inv.Entity.Name}/>
      <input className="form-control" type="text" data-key="Entity.Street1" onChange={that.handleChange} value={inv.Entity.Street1}/>
      <input className="form-control" type="text" data-key="Entity.Street2" onChange={that.handleChange} value={inv.Entity.Street2}/>
    </div>
  </div>

  <div className="row">
    <div className="col-sm-1">
      Invoice For
    </div>
    <div className="col-sm-3">
      <input className="form-control" type="text" data-key="Customer.Name" onChange={that.handleChange} value={inv.Customer.Name}/>
      <input className="form-control" type="text" data-key="Customer.Street1" onChange={that.handleChange} value={inv.Customer.Street1}/>
      <input className="form-control" type="text" data-key="Customer.Street2" onChange={that.handleChange} value={inv.Customer.Street2}/>
    </div>
    <div className="meta col-sm-offset-3 col-sm-5">
      <table className="table">
        <tbody>
          <tr>
            <td className="text">
              Invoice ID
            </td>
            <td>
              <div className="input-group">
                <input className="form-control" disabled={inv.Meta.InvoiceidL?"disabled":""} type="text" data-key="Meta.Invoiceid" onChange={that.handleChange} value={inv.Meta.Invoiceid} placeholder="AUTOGENERATED"/>
                <div className="input-group-addon"><a className="" onClick={this.toggleChange('Meta.InvoiceidL', inv.Meta.InvoiceidL)}><i className={"fa " + (inv.Meta.InvoiceidL?"fa-lock":"fa-unlock")}></i></a></div>
              </div>
            </td>
          </tr>
          <tr>
            <td className="text">Issue Date</td>
            <td>
              <div className="input-group">
                <DatePicker
                className="form-control"
                disabled={inv.Meta.IssuedateL}
                dateFormat="YYYY-MM-DD"
                selected={inv.Meta.Issuedate}
                placeholderText="AUTOGENERATED"
                onChange={this.handleChangeDate('Meta.Issuedate')} />
                <div className="input-group-addon"><a className="" onClick={this.toggleChange('Meta.IssuedateL', inv.Meta.IssuedateL)}><i className={"fa " + (inv.Meta.IssuedateL?"fa-lock":"fa-unlock")}></i></a></div>
              </div>
            </td>
          </tr>
          <tr>
            <td className="text">PO Number</td>
            <td><input className="form-control" type="text" data-key="Meta.Ponumber" onChange={that.handleChange} value={inv.Meta.Ponumber}/></td>
          </tr>
          <tr>
            <td className="text">Due Date</td>
            <td>
                <DatePicker
                className="form-control"
                dateFormat="YYYY-MM-DD"
                selected={inv.Meta.Duedate}
                onChange={this.handleChangeDate('Meta.Duedate')} />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <table className="table table-striped">
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
        <td colSpan="3" className="text">
          <button disabled={inv.Meta.Status === 'FINAL' ? 'disabled': ''} className={"btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-success faa-parent animated-hover' : '')} onClick={this.lineAdd}><i className="fa fa-plus faa-bounce"></i> Add row</button>
        </td>
        <td className="text">Total (ex tax)</td>
        <td><input className="form-control" disabled="disabled" type="text" data-key="Total.Ex" readOnly="readOnly" value={inv.Total.Ex}/></td>
      </tr>
      <tr>
        <td colSpan="3"></td>
        <td className="text">Tax (21%)</td>
        <td><input className="form-control" disabled="disabled" type="text" data-key="Total.Tax" readOnly="readOnly" value={inv.Total.Tax}/></td>
      </tr>
      <tr>
        <td colSpan="3">&nbsp;</td>
        <td className="text">Total</td>
        <td><input className="form-control" disabled="disabled" type="text" data-key="Total.Total" readOnly="readOnly" value={inv.Total.Total}/></td>
      </tr>
    </tfoot>
  </table>

  <div className="row notes col-sm-12">
    <p>Notes</p>
    <textarea className="form-control" data-key="Notes" onChange={that.handleChange} value={inv.Notes}/>
  </div>
  <div className="row banking">
    <div className="col-sm-4">
      <p>Banking details</p>
      <table className="table"><tbody>
        <tr><td className="text">VAT</td><td><input className="form-control" type="text" data-key="Bank.Vat" onChange={that.handleChange} value={inv.Bank.Vat}/></td></tr>
        <tr><td className="text">CoC</td><td><input className="form-control" type="text" data-key="Bank.Coc" onChange={that.handleChange} value={inv.Bank.Coc}/></td></tr>
        <tr><td className="text">IBAN</td><td><input className="form-control" type="text" data-key="Bank.Iban" onChange={that.handleChange} value={inv.Bank.Iban}/></td></tr>
      </tbody></table>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div></form>;
	}
});
