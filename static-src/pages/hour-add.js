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
require("react-datepicker/dist/react-datepicker.css");
var HourEdit = (function (_super) {
    __extends(HourEdit, _super);
    function HourEdit(props) {
        var _this = _super.call(this, props) || this;
        _this.state = {
            start: "",
            stop: "",
            description: "",
            day: Moment(),
            Lines: [],
            Name: ""
        };
        return _this;
    }
    HourEdit.prototype.componentDidMount = function () {
        console.log("componentDidMount", this.props);
        if (this.props.params["id"]) {
            console.log("Load hour name=" + this.props.params["id"]);
            this.ajax(this.props.params["id"]);
        }
    };
    HourEdit.prototype.ajax = function (name) {
        var _this = this;
        axios_1.default.get("/api/v1/hour/" + name)
            .then(function (res) {
            _this.setState(res.data);
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    HourEdit.prototype.recalc = function (e) {
        e.preventDefault();
        if (this.state.start === "" || this.state.stop === "") {
            console.log("Empty state");
            return;
        }
        var start = Moment.duration(this.state.start);
        var stop = Moment.duration(this.state.stop);
        var sum = stop.subtract(start);
        console.log("Start=" + start + " Stop=" + stop + " to hours=" + sum.humanize());
        this.state.Lines.push({
            Start: this.state.start,
            Stop: this.state.stop,
            Hours: sum.asHours(),
            Description: this.state.description,
            Day: this.state.day.format("YYYY-MM-DD")
        });
        this.setState({
            Lines: this.state.Lines
        });
    };
    HourEdit.prototype.updateDate = function (date) {
        this.setState({ day: date });
    };
    HourEdit.prototype.update = function (e) {
        console.log(e.target.value);
        var elem = e.target;
        if (elem.id === "hour-start") {
            this.setState({ start: e.target.value });
        }
        if (elem.id === "hour-stop") {
            this.setState({ stop: e.target.value });
        }
        if (elem.id === "hour-description") {
            this.setState({ description: e.target.value });
        }
        if (elem.id === "hour-name") {
            this.setState({ Name: e.target.value });
        }
        if (elem.id === "hour-day") {
            this.setState({ day: Moment(e.target.value) });
        }
    };
    HourEdit.prototype.lineRemove = function (key) {
        console.log("Remove hour line with key=" + key);
        console.log("Deleted idx ", this.state.Lines.splice(key, 1)[0]);
        this.setState({ Lines: this.state.Lines });
    };
    HourEdit.prototype.save = function (e) {
        var _this = this;
        e.preventDefault();
        axios_1.default.post('/api/v1/hour', this.state)
            .then(function (res) {
            console.log(res.data);
            _this.props.params["id"] = res.data.Name;
            history.replaceState({}, "", "#/hour-add/" + res.data.Name);
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    HourEdit.prototype.shortHand = function (d) {
        var date = new Date(1000 * 60 * 60 * d);
        var str = '';
        if (date.getUTCDate() - 1 > 0) {
            str += date.getUTCDate() - 1 + "d";
        }
        str += date.getUTCHours() + "h";
        str += date.getUTCMinutes() + "m";
        return str;
    };
    HourEdit.prototype.render = function () {
        var lines = [];
        var that = this;
        var sum = new Big("0.00");
        this.state.Lines.forEach(function (item, idx) {
            sum = sum.plus(item.Hours);
            lines.push(React.createElement("tr", { key: idx },
                React.createElement("td", null,
                    React.createElement("button", { className: "btn btn-default btn-hover-danger faa-parent animated-hover", onClick: that.lineRemove.bind(that, idx) },
                        React.createElement("i", { className: "fa fa-trash faa-flash" }))),
                React.createElement("td", null, item.Day),
                React.createElement("td", null, item.Start),
                React.createElement("td", null, item.Stop),
                React.createElement("td", null, that.shortHand(item.Hours)),
                React.createElement("td", null, item.Description)));
        });
        return React.createElement("form", null,
            React.createElement("div", { className: "normalheader" },
                React.createElement("div", { className: "hpanel hblue" },
                    React.createElement("div", { className: "panel-heading hbuilt" }, "Project Hour Calc"),
                    React.createElement("div", { className: "panel-body" },
                        React.createElement("div", { className: "col-sm-2" },
                            React.createElement(DatePicker, { id: "hour-day", className: "form-control", dateFormat: "YYYY-MM-DD", selected: this.state.day, onChange: this.updateDate.bind(this) })),
                        React.createElement("div", { className: "col-sm-2" },
                            React.createElement("input", { type: "text", id: "hour-start", className: "form-control", placeholder: "HH:mm", value: this.state.start, onChange: this.update.bind(this) })),
                        React.createElement("div", { className: "col-sm-2" },
                            React.createElement("input", { type: "text", id: "hour-stop", className: "form-control", placeholder: "HH:mm", value: this.state.stop, onChange: this.update.bind(this) })),
                        React.createElement("div", { className: "col-sm-5" },
                            React.createElement("input", { type: "text", id: "hour-description", className: "form-control", placeholder: "Description", value: this.state.description, onChange: this.update.bind(this) })),
                        React.createElement("div", { className: "col-sm-1" },
                            React.createElement("button", { onClick: this.recalc.bind(this), className: "btn btn-default btn-hover-success" },
                                React.createElement("i", { className: "fa fa-plus" }),
                                "\u00A0Add"))))),
            React.createElement("div", { className: "normalheader" },
                React.createElement("div", { className: "hpanel hblue" },
                    React.createElement("div", { className: "panel-heading hbuilt" },
                        React.createElement("div", { className: "panel-tools" },
                            React.createElement("div", { className: "btn-group nm7" },
                                React.createElement("button", { className: "btn btn-default btn-hover-success", onClick: this.save.bind(this) },
                                    React.createElement("i", { className: "fa fa-floppy-o" }),
                                    "\u00A0Save"))),
                        "Sum (",
                        sum.toFixed(2).toString(),
                        " hours)"),
                    React.createElement("div", { className: "panel-body" },
                        React.createElement("input", { type: "text", id: "hour-name", className: "form-control", placeholder: "Hour name", value: this.state.Name, onChange: this.update.bind(this) }),
                        React.createElement("table", { className: "table table-striped" },
                            React.createElement("thead", null,
                                React.createElement("tr", null,
                                    React.createElement("th", null, "#"),
                                    React.createElement("th", null, "Day"),
                                    React.createElement("th", null, "From"),
                                    React.createElement("th", null, "To"),
                                    React.createElement("th", null, "Hours"),
                                    React.createElement("th", null, "Description"))),
                            React.createElement("tbody", null, lines))))));
    };
    return HourEdit;
}(React.Component));
exports.default = HourEdit;
