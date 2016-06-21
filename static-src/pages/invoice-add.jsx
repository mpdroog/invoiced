'use strict';
var React = require('react');
var Request = require('superagent');
require('./invoice.css');

module.exports = React.createClass({
  getInitialState: function() {
      return {
        company: "RootDev",
        entity: {
          name: "M.P. Droog",
          street1: "Dorpsstraat 236a",
          street2: "Obdam, 1713HP, NL"
        },
        customer: {
          name: "XSNews B.V.",
          street1: "New Yorkstraat 9-13",
          street2: "1175 RD Lijnden"
        },
        meta: {
          invoiceid: "",
          issuedate: "",
          ponumber: "P/O",
          duedate: "due"
        },
        lines: [{
          description: "description",
          quantity: "1",
          price: "12.00",
          total: "12.00"
        }],
        notes: "",
        total: {
          ex: "200",
          tax: "1000",
          total: "1200"
        },
        bank: {
          vat: "VAT",
          coc: "COC",
          iban: "IBEN"
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
    this.state.lines.push({
      description: "",
      quantity: "1",
      price: "0.00",
      total: "0.00"
    });
    this.setState({lines: this.state.lines});
  },
  lineRemove: function(key) {
    if (confirm("Are you sure you want to remove the invoiceline with description '" + this.state.lines[key].description + "'?")) {
      console.log("Remove invoice line with key=" + key);
      console.log("Deleted idx ", this.state.lines.splice(key, 1)[0]);
      this.setState({lines: this.state.lines});
    }
  },
  handleChange: function(e) {
    console.log("handleChange", e.target.dataset.key);
    var indices = e.target.dataset.key.split('.');

    var node = this.state;
    for (var i = 0; i < indices.length-1; i++) {
      node = node[ indices[i] ];
    }
    node[indices[indices.length-1]] = e.target.value;
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

    inv.lines.forEach(function(line, idx) {
      lines.push(
        <tr key={"line"+idx}>
          <td><a onClick={that.lineRemove.bind(null, idx)}><i className="fa fa-trash"></i></a></td>
          <td><input className="form-control" type="text" data-key={"lines."+idx+".description"} onChange={that.handleChange} value={line.description}/></td>
          <td><input className="form-control" type="text" data-key={"lines."+idx+".quantity"} onChange={that.handleChange} value={line.quantity}/></td>
          <td><input className="form-control" type="text" data-key={"lines."+idx+".price"} onChange={that.handleChange} value={line.price}/></td>
          <td><input className="form-control" readOnly="readOnly" disabled="disabled" type="text" data-key={"lines."+idx+".total"} readOnly="readOnly" value={line.total}/></td>
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
      <input className="form-control" type="text" data-key="company" onChange={that.handleChange} value={inv.company}/>
    </div>

    <div className="col-sm-offset-3 col-sm-1">
      From
    </div>
    <div className="entity col-sm-4">
      <input className="form-control" type="text" data-key="entity.name" onChange={that.handleChange} value={inv.entity.name}/>
      <input className="form-control" type="text" data-key="entity.street1" onChange={that.handleChange} value={inv.entity.street1}/>
      <input className="form-control" type="text" data-key="entity.street2" onChange={that.handleChange} value={inv.entity.street2}/>
    </div>
  </div>

  <div className="row">
    <div className="col-sm-1">
      Invoice For
    </div>
    <div className="col-sm-3">
      <input className="form-control" type="text" data-key="customer.name" onChange={that.handleChange} value={inv.customer.name}/>
      <input className="form-control" type="text" data-key="customer.street1" onChange={that.handleChange} value={inv.customer.street1}/>
      <input className="form-control" type="text" data-key="customer.street2" onChange={that.handleChange} value={inv.customer.street2}/>
    </div>
    <div className="meta col-sm-offset-3 col-sm-5">
      <table className="table">
        <tr>
          <td className="text">Invoice ID</td>
          <td><input className="form-control" disabled="disabled" type="text" readOnly="readOnly" value={inv.meta.invoiceid} placeholder="AUTOGENERATED"/></td>
        </tr>
        <tr>
          <td className="text">Issue Date</td>
          <td><input className="form-control" disabled="disabled" type="text" data-key="meta.issuedate" onChange={that.handleChange} value={inv.meta.issuedate} placeholder="AUTOGENERATED"/></td>
        </tr>
        <tr>
          <td className="text">PO Number</td>
          <td><input className="form-control" type="text" data-key="meta.ponumber" onChange={that.handleChange} value={inv.meta.ponumber}/></td>
        </tr>
        <tr>
          <td className="text">Due Date</td>
          <td><input className="form-control" type="text" data-key="meta.duedate" onChange={that.handleChange} value={inv.meta.duedate}/></td>
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
        <td><input className="form-control" disabled="disabled" type="text" data-key="total.ex" readOnly="readOnly" value={inv.total.ex}/></td>
      </tr>
      <tr>
        <td colSpan="3"></td>
        <td className="text">Tax (21%)</td>
        <td><input className="form-control" disabled="disabled" type="text" data-key="total.tax" readOnly="readOnly" value={inv.total.tax}/></td>
      </tr>
      <tr>
        <td colSpan="3">&nbsp;</td>
        <td className="text">Total</td>
        <td><input className="form-control" disabled="disabled" type="text" data-key="total.total" readOnly="readOnly" value={inv.total.total}/></td>
      </tr>
    </tfoot>
  </table>

  <div className="row notes col-sm-12">
    <p>Notes</p>
    <textarea className="form-control" data-key="notes" onChange={that.handleChange} value={inv.notes}/>
  </div>
  <div className="row banking">
    <div className="col-sm-4">
      <p>Banking details</p>
      <table className="table">
        <tr><td className="text">VAT</td><td><input className="form-control" type="text" data-key="bank.vat" onChange={that.handleChange} value={inv.bank.vat}/></td></tr>
        <tr><td className="text">CoC</td><td><input className="form-control" type="text" data-key="bank.coc" onChange={that.handleChange} value={inv.bank.coc}/></td></tr>
        <tr><td className="text">IBAN</td><td><input className="form-control" type="text" data-key="bank.iban" onChange={that.handleChange} value={inv.bank.iban}/></td></tr>
      </table>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div></form>;
	}
});
