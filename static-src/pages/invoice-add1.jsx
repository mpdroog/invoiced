'use strict';
var React = require('react');
var Request = require('superagent');
require('./invoice.css');

module.exports = React.createClass({
  getInitialState: function() {
      return {
      };
  },

  componentDidMount: function() {
      this.ajax();
  },
  componentWillUnmount: function() {
  },

  ajax: function(range) {
  },

	render: function() {
		return <div className="normalheader">
		    <div className="hpanel hblue">
          <div className="panel-heading hbuilt">
            New Invoice
          </div>
          <div className="panel-body">

<div className="invoice group">
  <div className="company">
    <input type="text" value="RootDev"/>
  </div>
  <div className="entity">
    From
    <input type="text" value="M.P. Droog"/><br/>
    <input type="text" value="Dorpsstraat 236a"/><br/>
    <input type="text" value="Obdam, 1713HP, NL"/>
  </div>
  <div className="clearfix"></div>

  <div className="customer">
    Invoice For

    <input type="text" value="XS News B.V."/><br/>
    <input type="text" value="New Yorkstraat 9-13"/><br/>
    <input type="text" value="1175 RD Lijnden"/><br/>
  </div>
  <div className="meta">
    <input type="text" value="2016Q3-0001"/><br/>
    <input type="text" value="2016-05-23"/><br/>
    <input type="text" value="-"/><br/>
    <input type="text" value="2016-05-31"/><br/>
  </div>
  <div className="clearfix"></div>

  <div className="lines">
    <input type="text" value="PPF"/>
    <input type="text" value="50,00"/>
    <input type="text" value="42,50"/>
    2.1250,00
  </div>

  <div className="total">
    2.21250,00<br/>
    446,25<br/>
    2.571,25
  </div>
  <div className="clearfix"></div>

  <hr/>
  <div className="notes">
    <textarea value="Hello world..."/>
  </div>
  <div className="banking">
     <input type="text" value="TAXNR"/><br/>
     <input type="text" value="COCNR"/><br/>
     <input type="text" value="IBAN"/>
  </div>
</div>


	        </div>
		    </div>
    </div>;
	}
});
