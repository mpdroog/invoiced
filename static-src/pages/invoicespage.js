"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var React = require("react");
var invoices_1 = require("./invoices");
var InvoicesPage = (function (_super) {
    __extends(InvoicesPage, _super);
    function InvoicesPage() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    InvoicesPage.prototype.render = function () {
        return React.createElement("div", null,
            React.createElement(invoices_1.default, { title: "Pending Invoices", bucket: "invoices" }),
            React.createElement(invoices_1.default, { title: "Paid Invoices", bucket: "invoices-paid" }));
    };
    return InvoicesPage;
}(React.Component));
exports.default = InvoicesPage;
