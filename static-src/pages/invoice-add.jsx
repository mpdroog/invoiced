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
          ponumber: "",
          duedate: ""
        },
        lines: [{
          description: "",
          quantity: 1,
          price: "12.00",
          total: "12.00"
        }],
        notes: "",
        total: {
          ex: "",
          tax: "",
          total: ""
        },
        bank: {
          vat: "",
          coc: "",
          iban: ""
        }
      };
  },

  componentDidMount: function() {
      this.ajax();
  },
  componentWillUnmount: function() {
  },

  ajax: function(range) {
  },

  updateField: function() {
    // TODO
  },

	render: function() {
		return <form><div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            <div className="panel-tools">
              <a href="#"><i className="fa fa-plus"></i></a>
            </div>
            New Invoice
          </div>
          <div className="panel-body">

<div className="invoice group">
  <div className="row">
    <div className="company col-sm-4">
      <input className="form-control" type="text" value="RootDev"/>
    </div>

    <div className="col-sm-offset-3 col-sm-1">
      From
    </div>
    <div className="entity col-sm-4">
      <input className="form-control" type="text" value="M.P. Droog"/>
      <input className="form-control" type="text" value="Dorpsstraat 236a"/>
      <input className="form-control" type="text" value="Obdam, 1713HP, NL"/>
    </div>
  </div>

  <div className="row">
    <div className="col-sm-1">
      Invoice For
    </div>
    <div className="col-sm-3">
      <input className="form-control" type="text" value="XS News B.V."/>
      <input className="form-control" type="text" value="New Yorkstraat 9-13"/>
      <input className="form-control" type="text" value="1175 RD Lijnden"/>
    </div>
    <div className="meta col-sm-offset-3 col-sm-5">
      <table className="table">
        <tr>
          <td>Invoice ID</td>
          <td><input className="form-control" type="text" value="2016Q3-0001"/></td>
        </tr>
        <tr>
          <td>Issue Date</td>
          <td><input className="form-control" type="text" value="2016-05-23"/></td>
        </tr>
        <tr>
          <td>PO Number</td>
          <td><input className="form-control" type="text" value="-"/></td>
        </tr>
        <tr>
          <td>Due Date</td>
          <td><input className="form-control" type="text" value="2016-05-31"/></td>
        </tr>
      </table>
    </div>
  </div>

  <table className="table table-striped">
    <thead>
      <tr>
        <th>Description</th>
        <th>Quantity</th>
        <th>Price</th>
        <th>Line Total</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td><input className="form-control" type="text" value="PPF"/></td>
        <td><input className="form-control" type="text" value="50,00"/></td>
        <td><input className="form-control" type="text" value="42,50"/></td>
        <td><input className="form-control" type="text" value="2.1250,00"/></td>
      </tr>
    </tbody>
    <tfoot>
      <tr>
        <td colSpan="2">&nbsp;</td>
        <td>Total (ex tax)</td>
        <td><input className="form-control" type="text" value="2.21250,00"/></td>
      </tr>
      <tr>
        <td colSpan="2"></td>
        <td>Tax (21%)</td>
        <td><input className="form-control" type="text" value="446,25"/></td>
      </tr>
      <tr>
        <td colSpan="2"></td>
        <td>Total</td>
        <td><input className="form-control" type="text" value="2.571,25"/></td>
      </tr>
    </tfoot>
  </table>

  <div className="row notes col-sm-12">
    <p>Notes</p>
    <textarea className="form-control" value="Hello world..."/>
  </div>
  <div className="row banking">
    <div className="col-sm-4">
      <p>Banking details</p>
      <table className="table">
        <tr><td>VAT</td><td><input className="form-control" type="text" value="TAXNR"/></td></tr>
        <tr><td>CoC</td><td><input className="form-control" type="text" value="COCNR"/></td></tr>
        <tr><td>IBAN</td><td><input className="form-control" type="text" value="IBAN"/></td></tr>
      </table>
    </div>
  </div>
</div>


	        </div>
		    </div>
    </div></form>;
	}
});
