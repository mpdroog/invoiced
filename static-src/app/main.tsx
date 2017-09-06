import * as React from "react";
import * as ReactDOM from "react-dom";
import {Design} from "./shared/design";

function hashChange() {
      let url = location.hash.substr(1).split("/");
      let inject = null;
      let props = null;

      if (url[0] === '') {
            url.shift();
      }
      if (url.length === 0) {
            inject = require("./cmp/entities/app");
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
                                    inject = require("./cmp/hours/edit");
                                    break;

                              default:
                                    inject = require("./cmp/hours/list");
                                    break;
                        }
                        break;

                  case "invoices":
                        switch (url.shift()) {
                              case "add":
                              case "edit":
                                    props.bucket = url.shift();
                                    inject = require("./cmp/invoices/edit");
                                    break;

                              default:
                                    inject = require("./cmp/invoices/list");
                                    break;                        
                        }
                        break;

                  case "":
                  case undefined:
                        inject = require("./cmp/dashboard/app");
                        break;

                  default:
                        throw "Invalid path: " + location.hash;
            }
      }

      let page = null;
      if (props !== null) {
            props.id = url.shift();
            page = React.createElement(inject.default, props);
            page = React.createElement(Design, props, page);
      } else {
            page = React.createElement(inject.default, props);
      }
      ReactDOM.render(page, root);
}

try {
      let splash = document.getElementById("js-splash");
      let root = document.getElementById('root');

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