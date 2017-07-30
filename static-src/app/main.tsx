import * as React from "react";
import * as ReactDOM from "react-dom";
import {Router, Route, Switch} from "react-router";
import { createHashHistory } from 'history';

import Dashboard from "./cmp/dashboard/app";
import HoursEdit from "./cmp/hours/edit";
import HoursList from "./cmp/hours/list";

import InvoiceEdit from "./cmp/invoices/edit";
import InvoicesPage from "./cmp/invoices/list";

try {
      const history = createHashHistory();
      ReactDOM.render(
            <Router history={history}><Switch>
                  <Route exact path="/" component={Dashboard} />

                  <Route path="/hour-add" component={HoursEdit} />
                  <Route path="/hour-add/:id" component={HoursEdit} />
                  <Route path="/hours" component={HoursList} />

                  <Route path="/invoice-add" component={InvoiceEdit} />
                  <Route path="/invoice-add/:id" component={InvoiceEdit} />
                  <Route path="/invoices" component={InvoicesPage} />
            </Switch></Router>,
            document.getElementById('root')
      );
} catch (e) {
      //console.log(e);
      handleErr(e);
      //throw e;
}