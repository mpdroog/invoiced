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
var axios_1 = require("axios");
var dom_1 = require("../lib/dom");
var Invoices = (function (_super) {
    __extends(Invoices, _super);
    function Invoices(props) {
        var _this = _super.call(this, props) || this;
        _this.state = {
            "pagination": {
                "from": "",
                "count": 50
            },
            "invoices": null
        };
        return _this;
    }
    Invoices.prototype.componentDidMount = function () {
        this.ajax();
    };
    Invoices.prototype.ajax = function () {
        var _this = this;
        axios_1.default.get('/api/v1/invoices', { params: {
                from: this.state.pagination.from,
                count: this.state.pagination.count,
                bucket: this.props.bucket
            } })
            .then(function (res) {
            _this.setState({ invoices: res.data });
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    Invoices.prototype.delete = function (e) {
        e.preventDefault();
        var node = dom_1.DOM.eventFilter(e, "A");
        var id = node.dataset["target"];
        if (node.dataset["status"] === 'FINAL') {
            console.log("Cannot delete finalized invoices.");
            return;
        }
        axios_1.default.delete("/api/v1/invoice/" + id + "?bucket=" + this.props.bucket)
            .then(function (res) {
            location.reload();
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    Invoices.prototype.setPaid = function (id) {
        axios_1.default.post('/api/v1/invoice/' + id + '/paid', { params: {
                bucket: this.props.bucket
            } })
            .then(function (res) {
            location.reload();
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    Invoices.prototype.conceptLine = function (key, inv) {
        return React.createElement("tr", { key: key },
            React.createElement("td", null, key),
            React.createElement("td", null, inv.Meta.Invoiceid),
            React.createElement("td", null, inv.Customer.Name),
            React.createElement("td", null,
                "\u20AC ",
                inv.Total.Total),
            React.createElement("td", null,
                React.createElement("a", { className: "btn btn-default btn-hover-primary", href: "#invoice-add/" + this.props.bucket + "/" + key },
                    React.createElement("i", { className: "fa fa-pencil" })),
                React.createElement("a", { disabled: inv.Meta.Status === 'FINAL', className: "btn btn-default " + (inv.Meta.Status !== 'FINAL' ? "btn-hover-danger faa-parent animated-hover" : ""), "data-target": key, "data-status": inv.Meta.Status, onClick: this.delete.bind(this, key) },
                    React.createElement("i", { className: "fa fa-trash faa-flash" })),
                React.createElement("a", { className: "btn btn-default btn-hover-primary", onClick: this.setPaid.bind(this, key) },
                    React.createElement("i", { className: "fa fa-check" }))));
    };
    Invoices.prototype.finishedLine = function (key, inv) {
        return React.createElement("tr", { key: key },
            React.createElement("td", null, key),
            React.createElement("td", null, inv.Meta.Invoiceid),
            React.createElement("td", null, inv.Customer.Name),
            React.createElement("td", null,
                "\u20AC ",
                inv.Total.Total),
            React.createElement("td", null,
                React.createElement("a", { className: "btn btn-default btn-hover-primary", href: "#invoice-add/" + this.props.bucket + "/" + key },
                    React.createElement("i", { className: "fa fa-pencil" }))));
    };
    Invoices.prototype.render = function () {
        var _this = this;
        var res = [];
        console.log("invoices=", this.state.invoices);
        if (this.state.invoices && this.state.invoices.length > 0) {
            this.state.invoices.forEach(function (inv) {
                var key = inv.Meta.Conceptid;
                if (_this.props.bucket === "invoices") {
                    res.push(_this.conceptLine(key, inv));
                }
                else {
                    res.push(_this.finishedLine(key, inv));
                }
            });
        }
        else {
            res.push(React.createElement("tr", { key: "empty" },
                React.createElement("td", { colSpan: 5 }, "No invoices yet :)")));
        }
        var headerButtons = React.createElement("div", null);
        if (this.props.bucket === "invoices") {
            headerButtons = React.createElement("a", { href: "#invoice-add/" + this.props.bucket, className: "btn btn-default btn-hover-primary showhide" },
                React.createElement("i", { className: "fa fa-plus" }),
                " New");
        }
        else {
            headerButtons = React.createElement("a", { href: "#invoice-add/" + this.props.bucket, className: "btn btn-default btn-hover-primary showhide" },
                React.createElement("i", { className: "fa fa-upload" }),
                " Bankbalance");
        }
        return React.createElement("div", { className: "normalheader" },
            React.createElement("div", { className: "hpanel hblue" },
                React.createElement("div", { className: "panel-heading hbuilt" },
                    React.createElement("div", { className: "panel-tools" },
                        React.createElement("div", { className: "btn-group nm7" }, headerButtons)),
                    this.props.title),
                React.createElement("div", { className: "panel-body" },
                    React.createElement("table", { className: "table table-striped" },
                        React.createElement("thead", null,
                            React.createElement("tr", null,
                                React.createElement("th", null, "#"),
                                React.createElement("th", null, "Invoice"),
                                React.createElement("th", null, "Customer"),
                                React.createElement("th", null, "Amount"),
                                React.createElement("th", null, "I/O"))),
                        React.createElement("tbody", null, res)))));
    };
    return Invoices;
}(React.Component));
exports.default = Invoices;
