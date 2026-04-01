import * as React from "react";
import type { Root } from "react-dom/client";
import { createRoot } from "react-dom/client";
import type { AxiosError } from "axios";
import Axios from "axios";
import {Design} from "./shared/design";

let isLoginModalOpen = false;

// Show login modal - dispatches event that Design component listens for
function showLoginModal(): void {
      if (isLoginModalOpen) return;
      isLoginModalOpen = true;
      window.dispatchEvent(new Event('show-login-modal'));
}

// Called by LoginModal when closed
window.addEventListener('login-modal-closed', () => {
      isLoginModalOpen = false;
});

// Refresh git status badge after successful write operations
// Show login modal on 401 Unauthorized
Axios.interceptors.response.use(function (response) {
      if (response.config.method === 'post' || response.config.method === 'delete') {
            window.dispatchEvent(new Event('git-refresh'));
      }
      return response;
}, function (error: AxiosError) {
      if (error.response?.status === 401) {
            showLoginModal();
      }
      return Promise.reject(error);
});

// Static imports for all page components
import EntitiesApp from "./cmp/entities/app";
import DashboardApp from "./cmp/dashboard/app";
import HoursList from "./cmp/hours/list";
import HoursEdit from "./cmp/hours/edit";
import InvoicesList from "./cmp/invoices/list";
import InvoicesEdit from "./cmp/invoices/edit";
import PurchasesList from "./cmp/purchases/list";
import TaxesList from "./cmp/taxes/list";
import GitPage from "./cmp/git/app";

declare function handleErr(e: unknown): void;

interface RouteProps {
      entity: string;
      year: string;
      id?: string;
      bucket?: string;
      key?: string;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type PageComponent = React.ComponentType<any>;

function hashChange(): void {
      const url = location.hash.substr(1).split("/");
      let inject: PageComponent | null = null;
      let props: RouteProps | null = null;
      let pageName = "Entities";

      if (url[0] === '') {
            url.shift();
      }
      if (url.length === 0) {
            inject = EntitiesApp;
      }

      if (inject === null) {
            props = {
                  entity: url.shift() ?? "",
                  year: url.shift() ?? ""
            };
            switch (url.shift()) {
                  case "hours":
                        switch (url.shift()) {
                              case "add":
                              case "edit":
                                    props.bucket = url.shift() ?? "";
                                    inject = HoursEdit;
                                    pageName = "Hours Edit";
                                    break;

                              case undefined:
                              default:
                                    inject = HoursList;
                                    pageName = "Hours";
                                    break;
                        }
                        break;

                  case "invoices":
                        switch (url.shift()) {
                              case "add":
                              case "edit":
                                    props.bucket = url.shift() ?? "";
                                    inject = InvoicesEdit;
                                    pageName = "Invoice Edit";
                                    break;

                              case undefined:
                              default:
                                    inject = InvoicesList;
                                    pageName = "Invoices";
                                    break;
                        }
                        break;

                  case "purchases":
                        inject = PurchasesList;
                        pageName = "Purchases";
                        break;

                  case "taxes":
                        inject = TaxesList;
                        pageName = "Taxes";
                        break;

                  case "git":
                        inject = GitPage;
                        pageName = "Git Sync";
                        break;

                  case "":
                  case undefined:
                        inject = DashboardApp;
                        pageName = "Dashboard";
                        break;

                  default:
                        throw new Error("Invalid path: " + location.hash);
            }
      }

      // Update document title
      if (props?.entity !== undefined && props.entity !== "" && props.year !== "") {
            document.title = `${pageName} - ${props.year} - ${props.entity} | Boekhoud.cloud`;
      } else {
            document.title = `${pageName} | Boekhoud.cloud`;
      }

      let page = null;
      // Use location.hash as key to force unmount/remount on navigation
      // This matches React 15 behavior where we called unmountComponentAtNode
      const routeKey = location.hash !== '' ? location.hash : '/';
      if (props !== null) {
            props.id = url.shift() ?? "";
            props.key = routeKey;
            page = React.createElement(inject, props);
            page = React.createElement(Design, {...props, key: routeKey}, page);
      } else {
            page = React.createElement(inject, {key: routeKey});
      }
      console.log("root.render()", props);
      reactRoot.render(page);
}

let reactRoot: Root;

try {
      const splash = document.getElementById("js-splash");
      const rootEl = document.getElementById('root');
      if (rootEl === null) {
            throw new Error("document.getElementById(root) returned null");
      }

      reactRoot = createRoot(rootEl);
      hashChange();
      splash?.remove();
      window.onhashchange = function(): void {
            hashChange();
      };

} catch (e) {
      handleErr(e);
      throw e;
}
