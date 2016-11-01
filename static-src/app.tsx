import * as React from "react";
import * as ReactDOM from "react-dom";
import {Router, Route, hashHistory} from "react-router";

import Dashboard from "./pages/dashboard";
import HourEdit from "./pages/hour-add";
import Hours from "./pages/hours";
import InvoiceEdit from "./pages/invoice-add";
import Invoices from "./pages/invoicespage";

try {
	ReactDOM.render(
      <Router history={hashHistory}>
      	<Route path="/" component={Dashboard} />
      	<Route path="/hour-add/:id" component={HourEdit} />
      	<Route path="/hours" component={Hours} />

      	<Route path="/invoice-add/:bucket" component={InvoiceEdit} />
            <Route path="/invoice-add/:bucket/:id" component={InvoiceEdit} />
            <Route component={Invoices} path="/invoices"/>
      </Router>,
	  document.getElementById('root')
	);
} catch (e) {
	console.log(e);
	handleErr(e);
	throw e;
}
