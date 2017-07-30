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
var Hours = (function (_super) {
    __extends(Hours, _super);
    function Hours() {
        var _this = _super.call(this) || this;
        _this.state = {
            "pagination": {
                "from": "",
                "count": 50
            },
            "hours": null
        };
        return _this;
    }
    Hours.prototype.componentDidMount = function () {
        this.ajax();
    };
    Hours.prototype.ajax = function () {
        var _this = this;
        axios_1.default.get('/api/v1/hours', { params: this.state.pagination })
            .then(function (res) {
            _this.setState({ hours: res.data });
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    Hours.prototype.delete = function (e) {
        e.preventDefault();
        var id = dom_1.DOM.eventFilter(e, "A").dataset["target"];
        axios_1.default.delete("/api/v1/hour/" + id)
            .then(function (res) {
            location.reload();
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    Hours.prototype.render = function () {
        var res = [];
        var that = this;
        console.log("hours=", this.state.hours);
        if (this.state.hours && this.state.hours.length > 0) {
            this.state.hours.forEach(function (elem) {
                res.push(React.createElement("tr", { key: elem },
                    React.createElement("td", null, elem),
                    React.createElement("td", null,
                        React.createElement("a", { className: "btn btn-default btn-hover-primary", href: "#hour-add/" + elem },
                            React.createElement("i", { className: "fa fa-pencil" })),
                        React.createElement("a", { className: "btn btn-default btn-hover-danger faa-parent animated-hover", "data-target": elem, onClick: that.delete.bind(that) },
                            React.createElement("i", { className: "fa fa-trash faa-flash" })))));
            });
        }
        else {
            res.push(React.createElement("tr", { key: "empty" },
                React.createElement("td", { colSpan: 4 }, "No hours yet :)")));
        }
        return React.createElement("div", { className: "normalheader" },
            React.createElement("div", { className: "hpanel hblue" },
                React.createElement("div", { className: "panel-heading hbuilt" },
                    React.createElement("div", { className: "panel-tools" },
                        React.createElement("div", { className: "btn-group nm7" },
                            React.createElement("a", { href: "#hour-add", className: "btn btn-default btn-hover-primary showhide" },
                                React.createElement("i", { className: "fa fa-plus" }),
                                " New"))),
                    "Hour registration"),
                React.createElement("div", { className: "panel-body" },
                    React.createElement("table", { className: "table table-striped" },
                        React.createElement("thead", null,
                            React.createElement("tr", null,
                                React.createElement("th", null, "Name"),
                                React.createElement("th", null, "I/O"))),
                        React.createElement("tbody", null, res)))));
    };
    return Hours;
}(React.Component));
exports.default = Hours;
