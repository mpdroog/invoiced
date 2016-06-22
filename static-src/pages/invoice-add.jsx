'use strict';
var React = require('react');
var Request = require('superagent');
var Big = require('big.js');
require('./invoice.css');

module.exports = React.createClass({
  getInitialState: function() {
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
          Invoiceid: "",
          Issuedate: "",
          Ponumber: "P/O",
          Duedate: "due"
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
  componentWillUnmount: function() {
  },

  ajax: function(name) {
    var that = this;
    Request.get('/api/invoice/'+name)
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          //Fn.error(err.message);
          return;
        }
        if (that.isMounted()) {
          var body = res.body;
          that.setState(body);
        }
    });
  },

  updateField: function() {
    // TODO
  },

  lineAdd: function() {
    console.log("Add invoice line");
    this.state.Lines.push({
      Description: "",
      Quantity: "1",
      Price: "0.00",
      Total: "0.00"
    });
    this.setState({Lines: this.state.Lines});
  },
  lineRemove: function(key) {
    if (confirm("Are you sure you want to remove the invoiceline with description '" + this.state.Lines[key].Description + "'?")) {
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

  handleChange: function(e) {
    console.log("handleChange", e.target.dataset.key);
    var indices = e.target.dataset.key.split('.');

    var node = this.state;
    for (var i = 0; i < indices.length-1; i++) {
      node = node[ indices[i] ];
    }
    node[indices[indices.length-1]] = e.target.value;

    // Any post-processing
    if (indices[0] === "Lines") {
      this.state.Lines[indices[1]] = this.lineUpdate(this.state.Lines[indices[1]]);
      this.state.Total = this.totalUpdate(this.state.Lines);
    }
    this.setState(this.state);
  },
  save: function(e) {
    var that = this;
    Request.post('/api/invoice')
    .send(this.state)
    .set('Accept', 'application/json')
    .end(function(err, res) {
        if (err) {
          console.log(err);
          //Fn.error(err.message);
          return;
        }
        if (that.isMounted()) {
          console.log(res.body);
          /*var body = res.body;
          body.loading = false;
          that.setState(body);*/
        }
    });
  },
  pdf: function() {
    location.href = '/api/invoice/'+this.props.args[0]+'/pdf';
  },

	render: function() {
    var inv = this.state;
    var that = this;
    var lines = [];

    inv.Lines.forEach(function(line, idx) {
      lines.push(
        <tr key={"line"+idx}>
          <td><a onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash"></i></a></td>
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
              <a onClick={this.save}><i className="fa fa-floppy-o"></i> Save</a>
              <a onClick={this.pdf}><i className="fa fa-file-pdf-o"></i> PDF</a>
            </div>
            New Invoice
          </div>
          <div className="panel-body">

<div className="invoice group">
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
        <tr>
          <td className="text">Invoice ID</td>
          <td><input className="form-control" disabled="disabled" type="text" readOnly="readOnly" value={inv.Meta.Invoiceid} placeholder="AUTOGENERATED"/></td>
        </tr>
        <tr>
          <td className="text">Issue Date</td>
          <td><input className="form-control" disabled="disabled" type="text" data-key="Meta.Issuedate" onChange={that.handleChange} value={inv.Meta.Issuedate} placeholder="AUTOGENERATED"/></td>
        </tr>
        <tr>
          <td className="text">PO Number</td>
          <td><input className="form-control" type="text" data-key="Meta.Ponumber" onChange={that.handleChange} value={inv.Meta.Ponumber}/></td>
        </tr>
        <tr>
          <td className="text">Due Date</td>
          <td><input className="form-control" type="text" data-key="Meta.Duedate" onChange={that.handleChange} value={inv.Meta.Duedate}/></td>
        </tr>
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
          <a onClick={this.lineAdd}><i className="fa fa-plus"></i> Add row</a>
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
      <table className="table">
        <tr><td className="text">VAT</td><td><input className="form-control" type="text" data-key="Bank.Vat" onChange={that.handleChange} value={inv.Bank.Vat}/></td></tr>
        <tr><td className="text">CoC</td><td><input className="form-control" type="text" data-key="Bank.Coc" onChange={that.handleChange} value={inv.Bank.Coc}/></td></tr>
        <tr><td className="text">IBAN</td><td><input className="form-control" type="text" data-key="Bank.Iban" onChange={that.handleChange} value={inv.Bank.Iban}/></td></tr>
      </table>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div></form>;
	}
});
