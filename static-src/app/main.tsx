import * as React from "react";
import * as ReactDOM from "react-dom";
import {Design} from "./shared/design";

// Static imports for all page components
import EntitiesApp from "./cmp/entities/app";
import DashboardApp from "./cmp/dashboard/app";
import HoursList from "./cmp/hours/list";
import HoursEdit from "./cmp/hours/edit";
import InvoicesList from "./cmp/invoices/list";
import InvoicesEdit from "./cmp/invoices/edit";
import TaxesList from "./cmp/taxes/list";

declare function handleErr(e: any): void;

function hashChange() {
      let url = location.hash.substr(1).split("/");
      let inject: any = null;
      let props: any = null;

      if (url[0] === '') {
            url.shift();
      }
      if (url.length === 0) {
            inject = EntitiesApp;
      }

      if (inject === null) {
            props = {
                  entity: url.shift(),
                  year: url.shift()
            };
            switch (url.shift()) {
                  case "hours":
                        switch (url.shift()) {
                              case "add":
                              case "edit":
                                    props.bucket = url.shift();
                                    inject = HoursEdit;
                                    break;

                              default:
                                    inject = HoursList;
                                    break;
                        }
                        break;

                  case "invoices":
                        switch (url.shift()) {
                              case "add":
                              case "edit":
                                    props.bucket = url.shift();
                                    inject = InvoicesEdit;
                                    break;

                              default:
                                    inject = InvoicesList;
                                    break;
                        }
                        break;

                  case "taxes":
                        inject = TaxesList;
                        break;

                  case "":
                  case undefined:
                        inject = DashboardApp;
                        break;

                  default:
                        throw "Invalid path: " + location.hash;
            }
      }

      let page = null;
      if (props !== null) {
            props.id = url.shift();
            page = React.createElement(inject, props);
            page = React.createElement(Design, props, page);
      } else {
            page = React.createElement(inject, props);
      }
      console.log("ReactDOM.render()", props);
      ReactDOM.render(page, root);
}

try {
      let splash = document.getElementById("js-splash");
      let root = document.getElementById('root');
      if (root === null) {
            throw "document.getElementById(root) returned null?";
      }

      hashChange();
      splash && splash.remove();
      window.onhashchange = function() {
            ReactDOM.unmountComponentAtNode(root);
            hashChange();
      };

} catch (e) {
      handleErr(e);
      throw e;
}
