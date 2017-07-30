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
var Big = require("big.js");
var Moment = require("moment");
var DatePicker = require("react-datepicker");
require("./invoice.css");
require("react-datepicker/dist/react-datepicker.css");
var InvoiceEdit = (function (_super) {
    __extends(InvoiceEdit, _super);
    function InvoiceEdit(props) {
        var _this = _super.call(this, props) || this;
        _this.revisions = [];
        _this.state = {
            Company: "RootDev",
            Entity: {
                Name: "M.P. Droog",
                Street1: "Dorpsstraat 236a",
                Street2: "Obdam, 1713HP, NL"
            },
            Customer: {
                Name: "XSNews B.V.",
                Street1: "New Yorkstraat 9-13",
                Street2: "1175 RD Lijnden",
                Vat: "",
                Coc: ""
            },
            Meta: {
                Conceptid: "",
                Status: "NEW",
                Invoiceid: "",
                InvoiceidL: true,
                Issuedate: null,
                IssuedateL: true,
                Ponumber: "",
                Duedate: Moment().add(14, 'days'),
                Paydate: null
            },
            Lines: [{
                    Description: "",
                    Quantity: "0",
                    Price: "0.00",
                    Total: "0.00"
                }],
            Notes: "",
            Total: {
                Ex: "0.00",
                Tax: "0.00",
                Total: "0.00"
            },
            Bank: {
                Vat: "",
                Coc: "",
                Iban: ""
            }
        };
        return _this;
    }
    InvoiceEdit.prototype.componentDidMount = function () {
        var params = this.props.params;
        if (params["id"]) {
            console.log("Load invoice name=" + params["id"] + " from bucket=" + params["bucket"]);
            this.ajax(params["bucket"], params["id"]);
        }
    };
    InvoiceEdit.prototype.ajax = function (bucket, name) {
        var _this = this;
        axios_1.default.get("/api/v1/invoice/" + name, { params: { bucket: bucket } })
            .then(function (res) {
            _this.parseInput.call(_this, res.data);
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    InvoiceEdit.prototype.parseInput = function (data) {
        console.log(data);
        if (window.location.href != "#/invoice-add/" + this.props.params["bucket"] + "/" + data.Meta.Conceptid) {
            history.replaceState({}, "", "#/invoice-add/" + this.props.params["bucket"] + "/" + data.Meta.Conceptid);
            this.props.params["id"] = data.Meta.Conceptid;
        }
        data.Meta.Issuedate = data.Meta.Issuedate ? Moment(data.Meta.Issuedate) : null;
        data.Meta.Duedate = data.Meta.Duedate ? Moment(data.Meta.Duedate) : null;
        data.Meta.Paydate = data.Meta.Paydate ? Moment(data.Meta.Paydate) : null;
        data.Meta.InvoiceidL = true;
        data.Meta.IssuedateL = true;
        this.setState(data);
    };
    InvoiceEdit.prototype.lineAdd = function () {
        if (this.state.Meta.Status === 'FINAL') {
            console.log("Finalized, not allowing changes!");
            return;
        }
        console.log("Add invoice line");
        this.state.Lines.push({
            Description: "",
            Quantity: "0",
            Price: "0.00",
            Total: "0.00"
        });
        this.setState({ Lines: this.state.Lines });
    };
    InvoiceEdit.prototype.lineRemove = function (key) {
        if (this.state.Meta.Status === 'FINAL') {
            console.log("Finalized, not allowing changes!");
            return;
        }
        var line = this.state.Lines[key];
        var isEmpty = line.Description === ""
            && line.Quantity === "0"
            && line.Price === "0.00"
            && line.Total === "0.00";
        var isOk = !isEmpty && confirm("Are you sure you want to remove the invoiceline with description '" + line.Description + "'?");
        if (isEmpty || isOk) {
            console.log("Remove invoice line with key=" + key);
            console.log("Deleted idx ", this.state.Lines.splice(key, 1)[0]);
            this.setState({ Lines: this.state.Lines });
        }
    };
    InvoiceEdit.prototype.lineUpdate = function (line) {
        line.Total = new Big(line.Price).times(line.Quantity).round(2).toFixed(2).toString();
        return line;
    };
    InvoiceEdit.prototype.totalUpdate = function (lines) {
        var ex = new Big(0);
        lines.forEach(function (val) {
            console.log("Add", val.Total);
            ex = ex.plus(val.Total);
        });
        var tax = ex.div("100").times("21");
        if (this.state.Customer.Vat.length > 0) {
            var country = this.state.Customer.Vat.substr(0, 2).toUpperCase();
            console.log("Country " + country);
            if (country !== "NL") {
                tax = new Big(0);
            }
        }
        var total = ex.plus(tax);
        console.log("totals (ex,tax,total)", ex.toString(), tax.toString(), total.toString());
        return {
            Ex: ex.round(2).toFixed(2).toString(),
            Tax: tax.round(2).toFixed(2).toString(),
            Total: total.round(2).toFixed(2).toString()
        };
    };
    InvoiceEdit.prototype.triggerChange = function (indices, val) {
        if (this.state.Meta.Status === 'FINAL') {
            console.log("Finalized, not allowing changes!");
            return;
        }
        var node = this.state;
        for (var i = 0; i < indices.length - 1; i++) {
            node = node[indices[i]];
        }
        node[indices[indices.length - 1]] = val;
        if (indices[0] === "Lines") {
            var idx = indices[1];
            this.state.Lines[idx] = this.lineUpdate(this.state.Lines[idx]);
            this.state.Total = this.totalUpdate(this.state.Lines);
        }
        this.setState(this.state);
        this.revisions.push({});
    };
    InvoiceEdit.prototype.handleChange = function (e) {
        console.log("handleChange", e.target.dataset["key"]);
        var indices = e.target.dataset["key"].split('.');
        this.triggerChange(indices, e.target.value);
    };
    InvoiceEdit.prototype.handleChangeDate = function (id) {
        var indices = id.split('.');
        var that = this;
        return function (val) {
            console.log("handleChangeDate", id, val);
            that.triggerChange.call(that, indices, val);
        };
    };
    InvoiceEdit.prototype.toggleChange = function (id, val) {
        var indices = id.split('.');
        var that = this;
        val = !val;
        return function () {
            console.log("toggleChange", id, val);
            that.triggerChange.call(that, indices, val);
        };
    };
    InvoiceEdit.prototype.save = function (e) {
        var _this = this;
        e.preventDefault();
        var req = JSON.parse(JSON.stringify(this.state));
        req.Meta.Issuedate = this.state.Meta.Issuedate ? this.state.Meta.Issuedate.format('YYYY-MM-DD') : "";
        req.Meta.Duedate = this.state.Meta.Duedate ? this.state.Meta.Duedate.format('YYYY-MM-DD') : "";
        console.log(req);
        axios_1.default.post('/api/v1/invoice', req)
            .then(function (res) {
            _this.parseInput.call(_this, res.data);
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    InvoiceEdit.prototype.reset = function (e) {
        var _this = this;
        e.preventDefault();
        axios_1.default.post("/api/v1/invoice/" + this.state.Meta.Conceptid + "/reset", {})
            .then(function (res) {
            _this.parseInput.call(_this, res.data);
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    InvoiceEdit.prototype.finalize = function (e) {
        var _this = this;
        e.preventDefault();
        axios_1.default.post("/api/v1/invoice/" + this.state.Meta.Conceptid + "/finalize", {})
            .then(function (res) {
            _this.parseInput.call(_this, res.data);
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    InvoiceEdit.prototype.pdf = function () {
        if (this.state.Meta.Status !== 'FINAL') {
            console.log("PDF only available in finalized invoices");
            return;
        }
        var url = "/api/v1/invoice/" + this.props.params["id"] + "/pdf?bucket=" + this.props.params["bucket"];
        console.log("Open PDF " + url);
        location.assign(url);
    };
    InvoiceEdit.prototype.render = function () {
        var inv = this.state;
        var that = this;
        var lines = [];
        inv.Lines.forEach(function (line, idx) {
            console.log(inv.Meta.Status);
            lines.push(React.createElement("tr", { key: "line" + idx },
                React.createElement("td", null,
                    React.createElement("button", { disabled: inv.Meta.Status === 'FINAL', className: "btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-danger faa-parent animated-hover' : ''), onClick: that.lineRemove.bind(that, idx) },
                        React.createElement("i", { className: "fa fa-trash faa-flash" }))),
                React.createElement("td", null,
                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Lines." + idx + ".Description", onChange: that.handleChange.bind(that), value: line.Description })),
                React.createElement("td", null,
                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Lines." + idx + ".Quantity", onChange: that.handleChange.bind(that), value: line.Quantity })),
                React.createElement("td", null,
                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Lines." + idx + ".Price", onChange: that.handleChange.bind(that), value: line.Price })),
                React.createElement("td", null,
                    React.createElement("input", { className: "form-control", readOnly: true, disabled: true, type: "text", "data-key": "Lines." + idx + ".Total", value: line.Total }))));
        });
        return React.createElement("form", null,
            React.createElement("div", { className: "normalheader" },
                React.createElement("div", { className: "hpanel hblue" },
                    React.createElement("div", { className: "panel-heading hbuilt" },
                        React.createElement("div", { className: "panel-tools" },
                            React.createElement("div", { className: "btn-group nm7" },
                                React.createElement("button", { className: "btn btn-default btn-hover-success", disabled: this.revisions.length === 0 || inv.Meta.Status === "FINAL", onClick: this.save.bind(this) },
                                    React.createElement("i", { className: "fa fa-floppy-o" }),
                                    " Save"),
                                React.createElement("button", { className: "btn btn-default btn-hover-danger", disabled: inv.Meta.Status !== "CONCEPT", onClick: this.finalize.bind(this) },
                                    React.createElement("i", { className: "fa fa-lock" }),
                                    " Finalize"),
                                React.createElement("a", { className: "btn btn-default btn-hover-success", disabled: inv.Meta.Status !== "FINAL", onClick: this.pdf.bind(this) },
                                    React.createElement("i", { className: "fa fa-file-pdf-o" }),
                                    " PDF"),
                                React.createElement("button", { className: "btn btn-default btn-hover-danger", disabled: inv.Meta.Status !== "FINAL", onClick: this.reset.bind(this) },
                                    React.createElement("i", { className: "fa fa-unlock" }),
                                    " Reset"))),
                        "New Invoice"),
                    React.createElement("div", { className: "panel-body" },
                        React.createElement("div", { className: "invoice group " + (inv.Meta.Status === 'FINAL' ? 'o50' : '') },
                            React.createElement("div", { className: "row" },
                                React.createElement("div", { className: "company col-sm-4" },
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Company", onChange: that.handleChange.bind(this), value: inv.Company })),
                                React.createElement("div", { className: "col-sm-offset-3 col-sm-1" }, "From"),
                                React.createElement("div", { className: "entity col-sm-4" },
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Entity.Name", onChange: that.handleChange.bind(this), value: inv.Entity.Name }),
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Entity.Street1", onChange: that.handleChange.bind(this), value: inv.Entity.Street1 }),
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Entity.Street2", onChange: that.handleChange.bind(this), value: inv.Entity.Street2 }))),
                            React.createElement("div", { className: "row" },
                                React.createElement("div", { className: "col-sm-1" }, "Invoice For"),
                                React.createElement("div", { className: "col-sm-3" },
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Customer.Name", onChange: that.handleChange.bind(this), value: inv.Customer.Name }),
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Customer.Street1", onChange: that.handleChange.bind(this), value: inv.Customer.Street1 }),
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Customer.Street2", onChange: that.handleChange.bind(this), value: inv.Customer.Street2 }),
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Customer.Vat", onChange: that.handleChange.bind(this), value: inv.Customer.Vat, placeholder: "VAT-number" }),
                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Customer.Coc", onChange: that.handleChange.bind(this), value: inv.Customer.Coc, placeholder: "Chamber Of Commerce (CoC)" })),
                                React.createElement("div", { className: "meta col-sm-offset-3 col-sm-5" },
                                    React.createElement("table", { className: "table" },
                                        React.createElement("tbody", null,
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "Invoice ID"),
                                                React.createElement("td", null,
                                                    React.createElement("div", { className: "input-group" },
                                                        React.createElement("input", { className: "form-control", disabled: inv.Meta.InvoiceidL, type: "text", "data-key": "Meta.Invoiceid", onChange: that.handleChange.bind(that), value: inv.Meta.Invoiceid, placeholder: "AUTOGENERATED" }),
                                                        React.createElement("div", { className: "input-group-addon" },
                                                            React.createElement("a", { className: "", onClick: that.toggleChange('Meta.InvoiceidL', inv.Meta.InvoiceidL) },
                                                                React.createElement("i", { className: "fa " + (inv.Meta.InvoiceidL ? "fa-lock" : "fa-unlock") })))))),
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "Issue Date"),
                                                React.createElement("td", null,
                                                    React.createElement("div", { className: "input-group" },
                                                        React.createElement(DatePicker, { className: "form-control", disabled: inv.Meta.IssuedateL, dateFormat: "YYYY-MM-DD", selected: inv.Meta.Issuedate, placeholderText: "AUTOGENERATED", onChange: that.handleChangeDate('Meta.Issuedate').bind(that) }),
                                                        React.createElement("div", { className: "input-group-addon" },
                                                            React.createElement("a", { className: "", onClick: that.toggleChange('Meta.IssuedateL', inv.Meta.IssuedateL) },
                                                                React.createElement("i", { className: "fa " + (inv.Meta.IssuedateL ? "fa-lock" : "fa-unlock") })))))),
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "PO Number"),
                                                React.createElement("td", null,
                                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Meta.Ponumber", onChange: that.handleChange.bind(that), value: inv.Meta.Ponumber }))),
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "Due Date"),
                                                React.createElement("td", null,
                                                    React.createElement(DatePicker, { className: "form-control", dateFormat: "YYYY-MM-DD", selected: inv.Meta.Duedate, onChange: that.handleChangeDate('Meta.Duedate').bind(that) }))))))),
                            React.createElement("table", { className: "table table-striped" },
                                React.createElement("thead", null,
                                    React.createElement("tr", null,
                                        React.createElement("th", null, "\u00A0"),
                                        React.createElement("th", null, "Description"),
                                        React.createElement("th", null, "Quantity"),
                                        React.createElement("th", null, "Price"),
                                        React.createElement("th", null, "Line Total"))),
                                React.createElement("tbody", null, lines),
                                React.createElement("tfoot", null,
                                    React.createElement("tr", null,
                                        React.createElement("td", { colSpan: 3, className: "text" },
                                            React.createElement("button", { disabled: inv.Meta.Status === 'FINAL', className: "btn btn-default " + (inv.Meta.Status !== 'FINAL' ? 'btn-hover-success faa-parent animated-hover' : ''), onClick: this.lineAdd.bind(this) },
                                                React.createElement("i", { className: "fa fa-plus faa-bounce" }),
                                                " Add row")),
                                        React.createElement("td", { className: "text" }, "Total (ex tax)"),
                                        React.createElement("td", null,
                                            React.createElement("input", { className: "form-control", disabled: true, type: "text", "data-key": "Total.Ex", readOnly: true, value: inv.Total.Ex }))),
                                    React.createElement("tr", null,
                                        React.createElement("td", { colSpan: 3 }),
                                        React.createElement("td", { className: "text" }, "Tax (21%)"),
                                        React.createElement("td", null,
                                            React.createElement("input", { className: "form-control", onChange: this.handleChange.bind(this), disabled: true, type: "text", "data-key": "Total.Tax", readOnly: true, value: inv.Total.Tax }))),
                                    React.createElement("tr", null,
                                        React.createElement("td", { colSpan: 3 }, "\u00A0"),
                                        React.createElement("td", { className: "text" }, "Total"),
                                        React.createElement("td", null,
                                            React.createElement("input", { className: "form-control", onChange: this.handleChange.bind(this), disabled: true, type: "text", "data-key": "Total.Total", readOnly: true, value: inv.Total.Total }))))),
                            React.createElement("div", { className: "row notes col-sm-12" },
                                React.createElement("p", null, "Notes"),
                                React.createElement("textarea", { className: "form-control", "data-key": "Notes", onChange: this.handleChange.bind(this), value: inv.Notes })),
                            React.createElement("div", { className: "row banking" },
                                React.createElement("div", { className: "col-sm-4" },
                                    React.createElement("p", null, "Banking details"),
                                    React.createElement("table", { className: "table" },
                                        React.createElement("tbody", null,
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "VAT"),
                                                React.createElement("td", null,
                                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Bank.Vat", onChange: this.handleChange.bind(this), value: inv.Bank.Vat }))),
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "CoC"),
                                                React.createElement("td", null,
                                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Bank.Coc", onChange: this.handleChange.bind(this), value: inv.Bank.Coc }))),
                                            React.createElement("tr", null,
                                                React.createElement("td", { className: "text" }, "IBAN"),
                                                React.createElement("td", null,
                                                    React.createElement("input", { className: "form-control", type: "text", "data-key": "Bank.Iban", onChange: this.handleChange.bind(this), value: inv.Bank.Iban }))))))))))));
    };
    return InvoiceEdit;
}(React.Component));
exports.default = InvoiceEdit;
